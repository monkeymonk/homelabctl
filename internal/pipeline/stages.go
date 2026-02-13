package pipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"homelabctl/internal/fs"
	"homelabctl/internal/stacks"
	"homelabctl/internal/inventory"
	"homelabctl/internal/secrets"
	"homelabctl/internal/render"
	"homelabctl/internal/compose"
	"homelabctl/internal/paths"
)

// LoadStacksStage loads enabled stacks and validates dependencies
func LoadStacksStage() Stage {
	return func(ctx *Context) error {
		fmt.Println("Loading stacks...")

		// Load enabled stacks from filesystem
		enabled, err := fs.GetEnabledStacks()
		if err != nil {
			return fmt.Errorf("failed to get enabled stacks: %w", err)
		}

		if len(enabled) == 0 {
			return fmt.Errorf("no stacks enabled")
		}

		// Sort by category order for proper deployment sequence
		sorted, err := stacks.SortByCategory(enabled)
		if err != nil {
			return fmt.Errorf("failed to sort stacks: %w", err)
		}

		fmt.Printf("Found %d enabled stack(s) (sorted by category)\n", len(sorted))

		// Validate dependencies
		if err := stacks.ValidateDependencies(sorted); err != nil {
			return fmt.Errorf("dependency validation failed: %w", err)
		}

		ctx.EnabledStacks = sorted
		return nil
	}
}

// LoadInventoryStage loads global inventory variables and state
func LoadInventoryStage() Stage {
	return func(ctx *Context) error {
		fmt.Println("Loading inventory...")

		// Load inventory vars
		inventoryVars, err := inventory.LoadVars()
		if err != nil {
			return fmt.Errorf("failed to load inventory vars: %w", err)
		}
		ctx.InventoryVars = inventoryVars

		// Load disabled services
		disabledServices, err := inventory.GetDisabledServices()
		if err != nil {
			return fmt.Errorf("failed to load disabled services: %w", err)
		}

		// Build map for quick lookup
		ctx.DisabledServices = make(map[string]bool)
		for _, svc := range disabledServices {
			ctx.DisabledServices[svc] = true
		}

		if len(disabledServices) > 0 {
			fmt.Printf("Loaded %d disabled service(s)\n", len(disabledServices))
		}

		return nil
	}
}

// MergeVariablesStage merges variables for all stacks
func MergeVariablesStage() Stage {
	return func(ctx *Context) error {
		fmt.Println("Merging variables...")

		for _, stackName := range ctx.EnabledStacks {
			fmt.Printf("Processing stack: %s\n", stackName)

			// Load stack
			stack, err := stacks.LoadStack(stackName)
			if err != nil {
				return fmt.Errorf("failed to load stack %s: %w", stackName, err)
			}

			// Validate service definitions
			if err := stacks.ValidateServiceDefinitions(stackName); err != nil {
				return fmt.Errorf("invalid services in %s: %w", stackName, err)
			}

			// Load stack vars
			stackVars, err := stacks.GetStackVars(stackName)
			if err != nil {
				return fmt.Errorf("failed to get vars for %s: %w", stackName, err)
			}

			// Load secrets (optional)
			stackSecrets, err := secrets.LoadSecrets(stackName)
			if err != nil {
				return fmt.Errorf("failed to load secrets for %s: %w", stackName, err)
			}

			// Merge according to precedence (including category defaults)
			mergedVars, err := stacks.MergeWithCategoryDefaults(stackName, stackVars, ctx.InventoryVars, stackSecrets)
			if err != nil {
				return fmt.Errorf("failed to merge vars for %s: %w", stackName, err)
			}

			// Store in context
			ctx.StackConfigs[stackName] = &StackConfig{
				Name:       stackName,
				MergedVars: mergedVars,
				Services:   stack.Services,
			}
		}

		return nil
	}
}

// FilterServicesStage reports disabled services but doesn't filter variables
// Variables are kept so templates can render successfully
// Actual service removal happens in FilterDisabledComposeStage after rendering
func FilterServicesStage() Stage {
	return func(ctx *Context) error {
		if len(ctx.DisabledServices) == 0 {
			// No disabled services, just copy MergedVars to FilteredVars
			for _, config := range ctx.StackConfigs {
				config.FilteredVars = config.MergedVars
			}
			return nil
		}

		fmt.Println("Disabled services will be filtered from final compose:")

		for stackName, config := range ctx.StackConfigs {
			// Keep all variables for template rendering
			config.FilteredVars = config.MergedVars

			// Just report which services are disabled in this stack
			for _, svc := range config.Services {
				if ctx.DisabledServices[svc] {
					fmt.Printf("  - %s (from %s)\n", svc, stackName)
				}
			}
		}

		return nil
	}
}

// RenderTemplatesStage renders all templates for all stacks
func RenderTemplatesStage() Stage {
	return func(ctx *Context) error {
		fmt.Println("Rendering templates...")

		// Ensure runtime directory exists
		if err := fs.EnsureDir(paths.Runtime); err != nil {
			return fmt.Errorf("failed to create runtime dir: %w", err)
		}

		for stackName, config := range ctx.StackConfigs {
			// Build template context
			templateCtx := &render.Context{
				Vars: config.FilteredVars,
				Stack: map[string]interface{}{
					"name":     stackName,
					"category": "", // Load from stack if needed
				},
				Stacks: map[string]interface{}{
					"enabled": ctx.EnabledStacks,
				},
			}

			// Render main compose template
			composeTemplate := paths.StackComposeTemplate(stackName)
			composeOutput := paths.RuntimeComposeFile(stackName)

			if err := render.RenderToFile(composeTemplate, composeOutput, templateCtx); err != nil {
				return fmt.Errorf("failed to render compose for %s: %w", stackName, err)
			}

			ctx.RenderedFiles = append(ctx.RenderedFiles, composeOutput)
			ctx.RenderedCompose[stackName] = composeOutput

			// Render Traefik contributions
			if err := renderContributions(stackName, "traefik", templateCtx, ctx); err != nil {
				return err
			}

			// Render config files
			if err := renderConfigs(stackName, templateCtx, ctx); err != nil {
				return err
			}
		}

		return nil
	}
}

// Helper function for rendering contributions
func renderContributions(stackName, provider string, templateCtx *render.Context, ctx *Context) error {
	contributeDir := paths.StackContributeDir(stackName, provider)

	info, err := os.Stat(contributeDir)
	if err != nil || !info.IsDir() {
		return nil // No contributions, skip
	}

	entries, err := os.ReadDir(contributeDir)
	if err != nil {
		return fmt.Errorf("failed to read contribute/%s for %s: %w", provider, stackName, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != paths.TemplateExt {
			continue
		}

		tmplPath := filepath.Join(contributeDir, entry.Name())
		outputName := strings.TrimSuffix(entry.Name(), paths.TemplateExt)
		outputPath := paths.TraefikContributionFile(stackName, outputName)

		if err := render.RenderToFile(tmplPath, outputPath, templateCtx); err != nil {
			return fmt.Errorf("failed to render %s contribution for %s: %w", provider, stackName, err)
		}

		fmt.Printf("  ✓ Rendered %s contribution: %s\n", provider, outputName)
	}

	return nil
}

// Helper function for rendering config files
func renderConfigs(stackName string, templateCtx *render.Context, ctx *Context) error {
	configDir := paths.StackConfigDir(stackName)

	info, err := os.Stat(configDir)
	if err != nil || !info.IsDir() {
		return nil // No configs, skip
	}

	return filepath.Walk(configDir, func(tmplPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || filepath.Ext(tmplPath) != paths.TemplateExt {
			return nil
		}

		relPath, err := filepath.Rel(configDir, tmplPath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		outputRelPath := strings.TrimSuffix(relPath, paths.TemplateExt)
		outputPath := paths.RuntimeConfigFile(stackName, outputRelPath)

		outputDir := filepath.Dir(outputPath)
		if err := fs.EnsureDir(outputDir); err != nil {
			return fmt.Errorf("failed to create config output dir: %w", err)
		}

		if err := render.RenderToFile(tmplPath, outputPath, templateCtx); err != nil {
			return fmt.Errorf("failed to render config %s: %w", relPath, err)
		}

		fmt.Printf("  ✓ Rendered config: %s\n", outputRelPath)
		return nil
	})
}

// MergeComposeStage merges all rendered compose files
func MergeComposeStage() Stage {
	return func(ctx *Context) error {
		fmt.Println("Merging compose files...")

		// Collect rendered compose file paths
		var composeFiles []string
		for _, path := range ctx.RenderedCompose {
			composeFiles = append(composeFiles, path)
		}

		// Merge all compose files
		merged, err := compose.MergeComposeFiles(composeFiles)
		if err != nil {
			return fmt.Errorf("failed to merge compose files: %w", err)
		}

		ctx.MergedCompose = merged
		return nil
	}
}

// FilterDisabledComposeStage removes disabled services from the merged compose file
func FilterDisabledComposeStage() Stage {
	return func(ctx *Context) error {
		if len(ctx.DisabledServices) == 0 {
			return nil
		}

		// Convert disabled services map to slice
		var disabled []string
		for svc := range ctx.DisabledServices {
			disabled = append(disabled, svc)
		}

		// Filter disabled services from the merged compose
		removed := compose.FilterDisabledServices(ctx.MergedCompose, disabled)
		if len(removed) > 0 {
			fmt.Printf("Removed %d disabled service(s) from final compose: %v\n", len(removed), removed)
		}

		return nil
	}
}

// WriteOutputStage writes the final docker-compose.yml
func WriteOutputStage() Stage {
	return func(ctx *Context) error {
		fmt.Println("Writing output...")

		if err := compose.WriteComposeFile(paths.DockerCompose, ctx.MergedCompose); err != nil {
			return fmt.Errorf("failed to write compose file: %w", err)
		}

		fmt.Printf("\n✓ Generation complete\n")
		fmt.Printf("✓ Written: %s\n", paths.DockerCompose)

		return nil
	}
}

// CleanupStage removes temporary files
// Set skip=true to preserve files for debugging
func CleanupStage(skip bool) Stage {
	return func(ctx *Context) error {
		if skip {
			fmt.Println("Skipping cleanup (temporary files preserved)")
			return nil
		}

		if len(ctx.RenderedFiles) == 0 {
			return nil
		}

		fmt.Println("Cleaning up temporary files...")

		for _, file := range ctx.RenderedFiles {
			if err := os.Remove(file); err != nil {
				// Log but don't fail on cleanup errors
				fmt.Printf("Warning: failed to remove %s: %v\n", file, err)
			}
		}

		return nil
	}
}
