package runner

import "context"

// MockRunner is a test double for ContainerRunner.
type MockRunner struct {
	Result *Result
	Err    error
	Called bool
	Params RunParams
}

func (m *MockRunner) Run(_ context.Context, params RunParams) (*Result, error) {
	m.Called = true
	m.Params = params
	return m.Result, m.Err
}

func (m *MockRunner) Close() error {
	return nil
}
