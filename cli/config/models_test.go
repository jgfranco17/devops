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

func TestProjectDefinition_Test(t *testing.T) {
	tests := []struct {
		name           string
		project        ProjectDefinition
		mockSetup      func(*MockShellExecutor)
		expectedError  string
		expectWarnings bool
	}{
		{
			name: "successful test with steps",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Test: Operation{
						Steps: []string{"go test ./...", "go test -race ./..."},
					},
				},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "go test ./...").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
				m.On("Exec", mock.Anything, "go test -race ./...").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
			},
		},
		{
			name: "test with no steps should warn",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Test: Operation{
						Steps: []string{},
					},
				},
			},
			mockSetup: func(m *MockShellExecutor) {
				// No expectations for empty steps
			},
			expectWarnings: true,
		},
		{
			name: "test failure should return error",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Test: Operation{
						Steps: []string{"go test ./..."},
					},
				},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "go test ./...").Return(executor.Result{ExitCode: 1, Stderr: "test failed"}, nil)
			},
			expectedError: "failed to run test steps",
		},
		{
			name: "test with environment variables",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Test: Operation{
						Env: map[string]string{
							"TEST_ENV":    "test_value",
							"GO111MODULE": "on",
						},
						Steps: []string{"go test ./..."},
					},
				},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.MatchedBy(func(env []string) bool {
					// Check that our env vars are included
					envStr := ""
					for _, e := range env {
						envStr += e + " "
					}
					return strings.Contains(envStr, "TEST_ENV=test_value") &&
						strings.Contains(envStr, "GO111MODULE=on")
				})).Return()
				m.On("Exec", mock.Anything, "go test ./...").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
			},
		},
		{
			name: "test with fail_fast enabled",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Test: Operation{
						FailFast: true,
						Steps:    []string{"go test ./pkg1", "go test ./pkg2"},
					},
				},
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "go test ./pkg1").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
				m.On("Exec", mock.Anything, "go test ./pkg2").Return(executor.Result{ExitCode: 1, Stderr: "test failed"}, nil)
			},
			expectedError: "failed to run test steps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockShellExecutor{}
			tt.mockSetup(mockExecutor)

			logger := logging.New(os.Stderr, logrus.InfoLevel)
			ctx := logging.WithContext(context.Background(), logger)
			err := tt.project.Test(ctx, mockExecutor)

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
				ID: "test-project",
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
				ID: "test-project",
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
				ID: "test-project",
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
			yamlContent: `---
id: test-project
description: A test project
version: 1.0.0
repo_url: https://github.com/test/project
codebase:
  language: go
  dependencies: [go.mod]
  install:
    steps:
      - go mod download
  build:
    steps:
      - go build ./...
`,
			expectError: false,
			validate: func(t *testing.T, cfg *ProjectDefinition) {
				assert.Equal(t, "test-project", cfg.ID)
				assert.Equal(t, "A test project", cfg.Description)
				assert.Equal(t, "1.0.0", cfg.Version)
				assert.Equal(t, "https://github.com/test/project", cfg.RepoUrl)
				assert.Equal(t, "go", cfg.Codebase.Language)
				assert.Equal(t, []string{"go.mod"}, cfg.Codebase.Dependencies)
				assert.Len(t, cfg.Codebase.Install.Steps, 1)
				assert.Len(t, cfg.Codebase.Build.Steps, 1)
			},
		},
		{
			name: "invalid YAML",
			yamlContent: `
id: test-project
description: A test project
version: 1.0.0
repo_url: https://github.com/test/project
codebase:
  language: go
  dependencies: [go.mod]
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

func TestProjectDefinition_Validate(t *testing.T) {
	tests := []struct {
		name           string
		project        ProjectDefinition
		expectedError  string
		expectWarnings bool
		outputChecks   []string // strings that should be present in output
	}{
		{
			name: "complete valid configuration",
			project: ProjectDefinition{
				ID:          "test-project",
				Description: "A test project",
				Version:     "1.0.0",
				RepoUrl:     "https://github.com/test/project",
				Codebase: Codebase{
					Language:     "go",
					Dependencies: []string{"github.com/stretchr/testify"},
					Install: Operation{
						Steps: []string{"go mod download"},
					},
					Test: Operation{
						Steps: []string{"go test ./..."},
					},
					Build: Operation{
						Steps: []string{"go build ./..."},
					},
				},
			},
			outputChecks: []string{
				"Language: go",
				"Dependencies:",
				"Install steps (1)",
				"Test steps (1)",
				"Build steps (1)",
			},
		},
		{
			name: "missing language should fail",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Test: Operation{
						Steps: []string{"go test ./..."},
					},
					Build: Operation{
						Steps: []string{"go build ./..."},
					},
				},
			},
			expectedError: "found 1 required fixes",
			outputChecks: []string{
				"Language is required",
				"Fixes:",
				"Set a language in the codebase",
			},
		},
		{
			name: "empty language should fail",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Language: "",
					Test: Operation{
						Steps: []string{"go test ./..."},
					},
					Build: Operation{
						Steps: []string{"go build ./..."},
					},
				},
			},
			expectedError: "found 1 required fixes",
			outputChecks: []string{
				"Language is required",
			},
		},
		{
			name: "missing dependencies should warn but pass",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Language: "go",
					Test: Operation{
						Steps: []string{"go test ./..."},
					},
					Build: Operation{
						Steps: []string{"go build ./..."},
					},
				},
			},
			expectWarnings: true,
			outputChecks: []string{
				"Language: go",
				"No dependencies defined",
				"Test steps (1)",
				"Build steps (1)",
			},
		},
		{
			name: "missing test steps should warn but pass",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Language:     "go",
					Dependencies: []string{"github.com/stretchr/testify"},
					Build: Operation{
						Steps: []string{"go build ./..."},
					},
				},
			},
			expectWarnings: true,
			outputChecks: []string{
				"Language: go",
				"Dependencies:",
				"No test steps defined",
				"Build steps (1)",
				"Suggestions:",
				"Set test steps in the codebase",
			},
		},
		{
			name: "missing build steps should warn but pass",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Language:     "go",
					Dependencies: []string{"github.com/stretchr/testify"},
					Test: Operation{
						Steps: []string{"go test ./..."},
					},
				},
			},
			expectWarnings: true,
			outputChecks: []string{
				"Language: go",
				"Dependencies:",
				"Test steps (1)",
				"No build steps defined",
				"Suggestions:",
				"Set build steps in the codebase",
			},
		},
		{
			name: "missing install steps should not warn (optional)",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Language:     "go",
					Dependencies: []string{"github.com/stretchr/testify"},
					Test: Operation{
						Steps: []string{"go test ./..."},
					},
					Build: Operation{
						Steps: []string{"go build ./..."},
					},
				},
			},
			outputChecks: []string{
				"Language: go",
				"Dependencies:",
				"Test steps (1)",
				"Build steps (1)",
			},
		},
		{
			name: "minimal valid configuration with only language",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Language: "go",
				},
			},
			expectWarnings: true,
			outputChecks: []string{
				"Language: go",
				"No dependencies defined",
				"No test steps defined",
				"No build steps defined",
				"Suggestions:",
			},
		},
		{
			name: "multiple warnings should be grouped",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Language: "go",
				},
			},
			expectWarnings: true,
			outputChecks: []string{
				"Language: go",
				"No dependencies defined",
				"No test steps defined",
				"No build steps defined",
				"Suggestions:",
				"Set test steps in the codebase",
				"Set build steps in the codebase",
			},
		},
		{
			name: "nil dependencies should not cause issues",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Language:     "go",
					Dependencies: nil,
					Test: Operation{
						Steps: []string{"go test ./..."},
					},
					Build: Operation{
						Steps: []string{"go build ./..."},
					},
				},
			},
			expectWarnings: true,
			outputChecks: []string{
				"Language: go",
				"No dependencies defined",
				"Test steps (1)",
				"Build steps (1)",
			},
		},
		{
			name: "nil steps should not cause issues",
			project: ProjectDefinition{
				ID: "test-project",
				Codebase: Codebase{
					Language:     "go",
					Dependencies: []string{"github.com/stretchr/testify"},
					Install: Operation{
						Steps: nil,
					},
					Test: Operation{
						Steps: nil,
					},
					Build: Operation{
						Steps: nil,
					},
				},
			},
			expectWarnings: true,
			outputChecks: []string{
				"Language: go",
				"Dependencies:",
				"No test steps defined",
				"No build steps defined",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output using a buffer
			var buf bytes.Buffer
			logger := logging.New(os.Stderr, logrus.InfoLevel)
			ctx := logging.WithContext(context.Background(), logger)

			err := tt.project.ValidateTo(ctx, &buf)

			output := buf.String()

			// Check expected error
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Check for warning indicators
			if tt.expectWarnings {
				assert.Contains(t, output, "[~]")
			}

			// Check specific output content
			for _, check := range tt.outputChecks {
				assert.Contains(t, output, check, "Expected output to contain: %s", check)
			}

			// Verify terminal line separator is present
			assert.Contains(t, output, "=")
		})
	}
}

func TestValidateProjectName(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid simple name",
			projectName: "test",
			expectError: false,
		},
		{
			name:        "valid name with underscores",
			projectName: "test_project",
			expectError: false,
		},
		{
			name:        "valid name with dashes",
			projectName: "test-project",
			expectError: false,
		},
		{
			name:        "valid name with numbers",
			projectName: "test123",
			expectError: false,
		},
		{
			name:        "valid mixed alphanumeric with underscores and dashes",
			projectName: "my_test-project2",
			expectError: false,
		},
		{
			name:        "valid name at max length (29 chars)",
			projectName: "thisIsAVeryLongProjectName29",
			expectError: false,
		},
		{
			name:        "valid single character",
			projectName: "a",
			expectError: false,
		},
		{
			name:        "valid uppercase start",
			projectName: "TestProject",
			expectError: false,
		},
		{
			name:        "empty name",
			projectName: "",
			expectError: true,
			errorMsg:    "ID cannot be empty",
		},
		{
			name:        "name too long",
			projectName: "thisIsAnExtremelyLongProjectNameThatExceedsThirtyCharacters",
			expectError: true,
			errorMsg:    "ID must be under 30 characters",
		},
		{
			name:        "starts with number",
			projectName: "1test",
			expectError: true,
			errorMsg:    "ID must start with a letter",
		},
		{
			name:        "starts with dash",
			projectName: "-test",
			expectError: true,
			errorMsg:    "ID must start with a letter",
		},
		{
			name:        "contains space",
			projectName: "test project",
			expectError: true,
			errorMsg:    "ID cannot contain whitespace",
		},
		{
			name:        "leading space",
			projectName: " test",
			expectError: true,
			errorMsg:    "ID must start with a letter", // Space character is not a letter
		},
		{
			name:        "trailing space",
			projectName: "test ",
			expectError: true,
			errorMsg:    "ID cannot contain whitespace",
		},

		// Invalid names - invalid characters
		{
			name:        "contains special characters",
			projectName: "test@project",
			expectError: true,
			errorMsg:    "ID can only contain letters, numbers, dashes, and underscores",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProjectName(tt.projectName)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProjectDefinition_ValidateNameIntegration(t *testing.T) {
	tests := []struct {
		name           string
		projectName    string
		expectError    bool
		outputContains []string
	}{
		{
			name:        "valid name shows checkmark",
			projectName: "validProject",
			expectError: false,
			outputContains: []string{
				"ID: validProject",
			},
		},
		{
			name:        "empty name shows error",
			projectName: "",
			expectError: true,
			outputContains: []string{
				"ID is required",
				"Set an ID for the project",
			},
		},
		{
			name:        "invalid name shows validation error",
			projectName: "123invalid",
			expectError: true,
			outputContains: []string{
				"Invalid ID: ID must start with a letter",
				"Use a valid project ID (alphanumeric/dashes/underscores, starts with letter, no whitespace, under 30 chars)",
			},
		},
		{
			name:        "name with spaces shows whitespace error",
			projectName: "invalid name",
			expectError: true,
			outputContains: []string{
				"Invalid ID: ID cannot contain whitespace",
				"Use a valid project ID (alphanumeric/dashes/underscores, starts with letter, no whitespace, under 30 chars)",
			},
		},
		{
			name:        "name too long shows length error",
			projectName: "thisNameIsWayTooLongAndExceedsThirtyCharacterLimit",
			expectError: true,
			outputContains: []string{
				"Invalid ID: ID must be under 30 characters",
				"Use a valid project ID (alphanumeric/dashes/underscores, starts with letter, no whitespace, under 30 chars)",
			},
		},
		{
			name:        "name with special characters shows character error",
			projectName: "invalid@name",
			expectError: true,
			outputContains: []string{
				"Invalid ID: ID can only contain letters, numbers, dashes, and underscores",
				"Use a valid project ID (alphanumeric/dashes/underscores, starts with letter, no whitespace, under 30 chars)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := logging.New(os.Stderr, logrus.InfoLevel)
			ctx := logging.WithContext(context.Background(), logger)

			project := ProjectDefinition{
				ID: tt.projectName,
				Codebase: Codebase{
					Language: "go", // Valid language to focus on name validation
				},
			}

			err := project.ValidateTo(ctx, &buf)
			output := buf.String()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			for _, expectedOutput := range tt.outputContains {
				assert.Contains(t, output, expectedOutput, "Expected output to contain: %s", expectedOutput)
			}
		})
	}
}

func TestProjectDefinition_Validate_EdgeCases(t *testing.T) {
	t.Run("validation with empty project definition", func(t *testing.T) {
		var buf bytes.Buffer
		logger := logging.New(os.Stderr, logrus.InfoLevel)
		ctx := logging.WithContext(context.Background(), logger)

		project := ProjectDefinition{}
		err := project.ValidateTo(ctx, &buf)

		output := buf.String()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "found 2 required fixes") // Now includes name validation
		assert.Contains(t, output, "ID is required")
		assert.Contains(t, output, "Language is required")
	})

	t.Run("validation with whitespace language should pass", func(t *testing.T) {
		var buf bytes.Buffer
		logger := logging.New(os.Stderr, logrus.InfoLevel)
		ctx := logging.WithContext(context.Background(), logger)

		project := ProjectDefinition{
			ID: "test-project",
			Codebase: Codebase{
				Language: "   ", // whitespace only
			},
		}
		err := project.ValidateTo(ctx, &buf)

		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "ID: test-project")
		assert.Contains(t, output, "Language:    ") // Should show the whitespace
	})

	t.Run("validation with complex dependencies", func(t *testing.T) {
		var buf bytes.Buffer
		logger := logging.New(os.Stderr, logrus.InfoLevel)
		ctx := logging.WithContext(context.Background(), logger)

		project := ProjectDefinition{
			ID: "test-project",
			Codebase: Codebase{
				Language: "go",
				Dependencies: []string{
					"github.com/stretchr/testify",
					"github.com/spf13/cobra",
					"github.com/sirupsen/logrus",
				},
			},
		}
		err := project.ValidateTo(ctx, &buf)

		output := buf.String()

		assert.NoError(t, err)
		assert.Contains(t, output, "ID: test-project")
		assert.Contains(t, output, "Language: go")
		assert.Contains(t, output, "Dependencies:")
	})
}
