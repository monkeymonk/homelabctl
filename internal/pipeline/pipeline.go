package pipeline

import (
	"fmt"
)

// Stage is a function that processes the pipeline context
type Stage func(*Context) error

// Pipeline represents a sequence of processing stages
type Pipeline struct {
	stages []Stage
	ctx    *Context
}

// New creates a new pipeline with an initial context
func New() *Pipeline {
	return &Pipeline{
		stages: []Stage{},
		ctx: &Context{
			RenderedFiles:    []string{},
			StackConfigs:     make(map[string]*StackConfig),
			RenderedCompose:  make(map[string]string),
			DisabledServices: make(map[string]bool),
		},
	}
}

// AddStage adds a stage to the pipeline
func (p *Pipeline) AddStage(stage Stage) *Pipeline {
	p.stages = append(p.stages, stage)
	return p
}

// Execute runs all stages in sequence
func (p *Pipeline) Execute() error {
	for i, stage := range p.stages {
		if err := stage(p.ctx); err != nil {
			return fmt.Errorf("stage %d failed: %w", i+1, err)
		}
	}

	return nil
}

// Context returns the pipeline context (useful for testing)
func (p *Pipeline) Context() *Context {
	return p.ctx
}
