package pipeline

import (
	"homelabctl/internal/compose"
)

// Context holds state that flows through the pipeline
type Context struct {
	// Input
	EnabledStacks    []string
	InventoryVars    map[string]interface{}
	DisabledServices map[string]bool

	// Intermediate state
	RenderedFiles    []string                      // For cleanup
	StackConfigs     map[string]*StackConfig       // Per-stack merged config
	RenderedCompose  map[string]string             // stack name -> compose file path

	// Output
	MergedCompose    *compose.ComposeFile
}

// StackConfig holds the processed configuration for a single stack
type StackConfig struct {
	Name         string
	MergedVars   map[string]interface{}
	FilteredVars map[string]interface{}
	Services     []string
}
