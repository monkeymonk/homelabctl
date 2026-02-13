package paths

import "path/filepath"

// Directory paths
const (
	Stacks    = "stacks"
	Enabled   = "enabled"
	Inventory = "inventory"
	Secrets   = "secrets"
	Runtime   = "runtime"
)

// File paths
const (
	InventoryVars     = "inventory/vars.yaml"
	InventoryState    = "inventory/state.yaml"
	DockerCompose     = "runtime/docker-compose.yml"
	TraefikDynamicDir = "runtime/traefik/dynamic"
)

// File names
const (
	StackYAML       = "stack.yaml"
	ComposeTemplate = "compose.yml.tmpl"
	SecretsEncExt   = ".enc.yaml"
	SecretsExt      = ".yaml"
)

// Template extensions
const (
	TemplateExt = ".tmpl"
)

// File permissions
const (
	DirPermissions        = 0755
	FilePermissions       = 0644
	SecureFilePermissions = 0600 // For sensitive files (state, secrets, temp files)
)

// Stack-related path helpers

// StackDir returns the path to a stack directory
func StackDir(name string) string {
	return filepath.Join(Stacks, name)
}

// StackYAMLPath returns the path to a stack's stack.yaml file
func StackYAMLPath(name string) string {
	return filepath.Join(Stacks, name, StackYAML)
}

// StackComposeTemplate returns the path to a stack's compose.yml.tmpl
func StackComposeTemplate(name string) string {
	return filepath.Join(Stacks, name, ComposeTemplate)
}

// StackContributeDir returns the path to a stack's contribute directory for a provider
func StackContributeDir(stackName, provider string) string {
	return filepath.Join(Stacks, stackName, "contribute", provider)
}

// EnabledStackLink returns the path to a stack's symlink in enabled/
func EnabledStackLink(name string) string {
	return filepath.Join(Enabled, name)
}

// SecretsFilePath returns the path to a stack's secrets file (with extension)
func SecretsFilePath(stackName, ext string) string {
	return filepath.Join(Secrets, stackName+ext)
}

// RuntimeComposeFile returns the path to a stack's temporary compose file in runtime/
func RuntimeComposeFile(stackName string) string {
	return filepath.Join(Runtime, stackName+"-compose.yml")
}

// TraefikContributionFile returns the path to a Traefik contribution file in runtime/
func TraefikContributionFile(stackName, filename string) string {
	return filepath.Join(TraefikDynamicDir, stackName+"-"+filename)
}

// StackConfigDir returns the path to a stack's config/ directory
func StackConfigDir(stackName string) string {
	return filepath.Join(Stacks, stackName, "config")
}

// RuntimeStackDir returns the path to a stack's directory in runtime/
func RuntimeStackDir(stackName string) string {
	return filepath.Join(Runtime, stackName)
}

// RuntimeConfigFile returns the path to a generated config file in runtime/<stack>/
func RuntimeConfigFile(stackName, filename string) string {
	return filepath.Join(Runtime, stackName, filename)
}
