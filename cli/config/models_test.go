package config

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/devops/cli/executor"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShellExecutor is a mock implementation of ShellExecutor
type MockShellExecutor struct {
	mock.Mock
}

func (m *MockShellExecutor) Exec(ctx context.Context, command string) (executor.Result, error) {
	args := m.Called(ctx, command)
	return args.Get(0).(executor.Result), args.Error(1)
}

func (m *MockShellExecutor) AddEnv(env []string) {
	m.Called(env)
}

func TestProjectDefinition_Build(t *testing.T) {
	tests := []struct {
		name           string
		project        ProjectDefinition
		mockSetup      func(*MockShellExecutor)
		expectedError  string
		expectWarnings bool
	}{
		{
			name: "successful build with steps",
			project: ProjectDefinition{
				Name: "test-project",
				Codebase: Codebase{
					Build: Operation{
						Steps: []string{"echo hello", "echo world"},
					},
				},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "echo hello").Return(executor.Result{ExitCode: 0, Stdout: "hello"}, nil)
				m.On("Exec", mock.Anything, "echo world").Return(executor.Result{ExitCode: 0, Stdout: "world"}, nil)
			},
		},
		{
			name: "build with no steps should warn",
			project: ProjectDefinition{
				Name: "test-project",
				Codebase: Codebase{
					Build: Operation{
						Steps: []string{},
					},
				},
			},
			mockSetup:      func(m *MockShellExecutor) {},
			expectWarnings: true,
		},
		{
			name: "build failure should return error",
			project: ProjectDefinition{
				Name: "test-project",
				Codebase: Codebase{
					Build: Operation{
						Steps: []string{"false"},
					},
				},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "false").Return(executor.Result{ExitCode: 1, Stderr: "command failed"}, nil)
			},
			expectedError: "failed to run steps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockShellExecutor{}
			tt.mockSetup(mockExecutor)

			buf := new(bytes.Buffer)
			ctx := logging.WithContext(context.Background(), logging.New(buf, logrus.InfoLevel))
			err := tt.project.Build(ctx, mockExecutor)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		expectError bool
		validate    func(t *testing.T, cfg *ProjectDefinition)
	}{
		{
			name: "valid YAML",
			yamlContent: `
name: test-project
description: A test project
version: 1.0.0
repo_url: https://github.com/test/project
codebase:
  language: go
  dependencies: go.mod
  install:
    steps:
      - go mod download
  build:
    steps:
      - go build ./...
`,
			expectError: false,
			validate: func(t *testing.T, cfg *ProjectDefinition) {
				assert.Equal(t, "test-project", cfg.Name)
				assert.Equal(t, "A test project", cfg.Description)
				assert.Equal(t, "1.0.0", cfg.Version)
				assert.Equal(t, "https://github.com/test/project", cfg.RepoUrl)
				assert.Equal(t, "go", cfg.Codebase.Language)
				assert.Equal(t, "go.mod", cfg.Codebase.Dependencies)
				assert.Len(t, cfg.Codebase.Install.Steps, 1)
				assert.Len(t, cfg.Codebase.Build.Steps, 1)
			},
		},
		{
			name: "invalid YAML",
			yamlContent: `
name: test-project
description: A test project
version: 1.0.0
repo_url: https://github.com/test/project
codebase:
  language: go
  dependencies: go.mod
  install:
    steps:
      - go mod download
  build:
    steps:
      - go build ./...
invalid: [unclosed array
`,
			expectError: true,
		},
		{
			name:        "empty YAML",
			yamlContent: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.yamlContent)
			cfg, err := Load(reader)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cfg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cfg)
				if tt.validate != nil {
					tt.validate(t, cfg)
				}
			}
		})
	}
}

func TestOperation_Run(t *testing.T) {
	tests := []struct {
		name          string
		operation     Operation
		mockSetup     func(*MockShellExecutor)
		expectedError string
	}{
		{
			name: "successful execution",
			operation: Operation{
				Steps: []string{"echo hello", "echo world"},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "echo hello").Return(executor.Result{ExitCode: 0, Stdout: "hello"}, nil)
				m.On("Exec", mock.Anything, "echo world").Return(executor.Result{ExitCode: 0, Stdout: "world"}, nil)
			},
		},
		{
			name: "execution with environment variables",
			operation: Operation{
				Env: map[string]string{
					"TEST_VAR": "test_value",
					"ANOTHER":  "value",
				},
				Steps: []string{"echo $TEST_VAR"},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.MatchedBy(func(env []string) bool {
					// Check that our env vars are included
					envStr := strings.Join(env, " ")
					return strings.Contains(envStr, "TEST_VAR=test_value") &&
						strings.Contains(envStr, "ANOTHER=value")
				})).Return()
				m.On("Exec", mock.Anything, "echo $TEST_VAR").Return(executor.Result{ExitCode: 0, Stdout: "test_value"}, nil)
			},
		},
		{
			name: "fail fast on error",
			operation: Operation{
				FailFast: true,
				Steps:    []string{"echo hello", "false", "echo world"},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "echo hello").Return(executor.Result{ExitCode: 0, Stdout: "hello"}, nil)
				m.On("Exec", mock.Anything, "false").Return(executor.Result{ExitCode: 1, Stderr: "command failed"}, nil)
			},
			expectedError: "error while running 'false'",
		},
		{
			name: "collect failed steps when not fail fast",
			operation: Operation{
				FailFast: false,
				Steps:    []string{"echo hello", "false", "echo world", "invalid_command"},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "echo hello").Return(executor.Result{ExitCode: 0, Stdout: "hello"}, nil)
				m.On("Exec", mock.Anything, "false").Return(executor.Result{ExitCode: 1, Stderr: "command failed"}, nil)
				m.On("Exec", mock.Anything, "echo world").Return(executor.Result{ExitCode: 0, Stdout: "world"}, nil)
				m.On("Exec", mock.Anything, "invalid_command").Return(executor.Result{ExitCode: 127, Stderr: "command not found"}, nil)
			},
			expectedError: "failed to run steps",
		},
		{
			name: "execution error",
			operation: Operation{
				Steps: []string{"echo hello"},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "echo hello").Return(executor.Result{}, errors.New("execution failed"))
			},
			expectedError: "failed to run steps",
		},
		{
			name: "empty steps",
			operation: Operation{
				Steps: []string{},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockShellExecutor{}
			tt.mockSetup(mockExecutor)

			logger := logging.New(os.Stderr, logrus.InfoLevel)
			ctx := logging.WithContext(context.Background(), logger)
			err := tt.operation.Run(ctx, mockExecutor)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestOperation_Run_OutputHandling(t *testing.T) {
	mockExecutor := &MockShellExecutor{}
	mockExecutor.On("AddEnv", mock.AnythingOfType("[]string")).Return()
	mockExecutor.On("Exec", mock.Anything, "test_command").Return(
		executor.Result{
			ExitCode: 0,
			Stdout:   "stdout output",
			Stderr:   "stderr output",
		}, nil)

	operation := Operation{
		Steps: []string{"test_command"},
	}

	logger := logging.New(os.Stderr, logrus.InfoLevel)
	ctx := logging.WithContext(context.Background(), logger)
	err := operation.Run(ctx, mockExecutor)

	assert.NoError(t, err)
	mockExecutor.AssertExpectations(t)
}
