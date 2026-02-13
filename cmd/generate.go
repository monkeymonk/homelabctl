package cmd

import (
	"fmt"
	"os"

	"github.com/monkeymonk/homelabctl/internal/fs"
	"github.com/monkeymonk/homelabctl/internal/pipeline"
)

// Generate renders all templates and creates runtime files
func Generate() error {
	fmt.Println("Generating runtime files...")

	// Verify repository
	if err := fs.VerifyRepository(); err != nil {
		return err
	}

	// Check debug mode
	debug := os.Getenv("HOMELAB_DEBUG") == "1"
	if debug {
		fmt.Println("DEBUG MODE: Temporary files will be preserved")
	}

	// Build and execute pipeline
	p := pipeline.New()
	p.AddStage(pipeline.LoadStacksStage()).
		AddStage(pipeline.LoadInventoryStage()).
		AddStage(pipeline.MergeVariablesStage()).
		AddStage(pipeline.FilterServicesStage()).
		AddStage(pipeline.RenderTemplatesStage()).
		AddStage(pipeline.MergeComposeStage()).
		AddStage(pipeline.FilterDisabledComposeStage()).
		AddStage(pipeline.WriteOutputStage()).
		AddStage(pipeline.CleanupStage(debug)) // Skip cleanup in debug mode

	return p.Execute()
}
