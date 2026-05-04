package runner

import (
	"context"
	"time"
)

// Result holds the output from a container execution.
type Result struct {
	Output    map[string]interface{}
	Stderr    string
	ExitCode  int
	Duration  time.Duration
	OOMKilled bool
	TimedOut  bool
}

// RunParams contains all parameters for a single container execution.
type RunParams struct {
	Language string
	Code     string
	Inputs   map[string]interface{}
	TenantID string
	RunID    string
	StepID   string
}

// ContainerRunner defines the interface for executing code in isolated containers.
type ContainerRunner interface {
	Run(ctx context.Context, params RunParams) (*Result, error)
	Close() error
}
