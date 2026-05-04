package runner

import (
	"testing"

	"github.com/lavianrose/flowforge/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestDockerRunner_ResolveImage(t *testing.T) {
	r := &DockerRunner{
		config: config.DockerConfig{
			PythonImage: "flowforge/runner-python:latest",
			NodeImage:   "flowforge/runner-nodejs:latest",
		},
	}

	tests := []struct {
		language string
		expected string
	}{
		{"python", "flowforge/runner-python:latest"},
		{"javascript", "flowforge/runner-nodejs:latest"},
		{"", "flowforge/runner-python:latest"},
		{"unknown", "flowforge/runner-python:latest"},
		{"PYTHON", "flowforge/runner-python:latest"},
	}

	for _, tt := range tests {
		t.Run("language_"+tt.language, func(t *testing.T) {
			result := r.resolveImage(tt.language)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrunc(t *testing.T) {
	tests := []struct {
		input    string
		n        int
		expected string
	}{
		{"hello", 3, "hel"},
		{"hello", 5, "hello"},
		{"hello", 10, "hello"},
		{"", 5, ""},
		{"a", 1, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := trunc(tt.input, tt.n)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMockRunner_ImplementsInterface(t *testing.T) {
	// Verify MockRunner satisfies ContainerRunner interface
	var _ ContainerRunner = &MockRunner{}
}

func TestMockRunner_Run(t *testing.T) {
	expected := &Result{
		Output:   map[string]interface{}{"key": "value"},
		ExitCode: 0,
	}

	mock := &MockRunner{Result: expected, Err: nil}

	params := RunParams{
		Language: "python",
		Code:     "print('hello')",
		Inputs:   map[string]interface{}{},
	}

	result, err := mock.Run(nil, params)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
	assert.True(t, mock.Called)
	assert.Equal(t, "python", mock.Params.Language)
	assert.Equal(t, "print('hello')", mock.Params.Code)
}

func TestMockRunner_RunError(t *testing.T) {
	mock := &MockRunner{Err: assert.AnError}

	_, err := mock.Run(nil, RunParams{})
	assert.Error(t, err)
	assert.True(t, mock.Called)
}

func TestMockRunner_Close(t *testing.T) {
	mock := &MockRunner{}
	err := mock.Close()
	assert.NoError(t, err)
}
