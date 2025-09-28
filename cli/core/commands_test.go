package core

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/devops/cli/config"
	"github.com/jgfranco17/devops/cli/executor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type CliCommandFunction func() *cobra.Command

type CommandRunner func(cmd *cobra.Command, args []string)

type CliRunResult struct {
	ShellOutput string
	Error       error
}

// MockShellExecutor is a mock implementation of BashExecutor
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

// Helper function to simulate CLI execution
func ExecuteCommand(t *testing.T, cmd *cobra.Command, args ...string) CliRunResult {
	t.Helper()

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	_, err := cmd.ExecuteC()
	return CliRunResult{
		ShellOutput: buf.String(),
		Error:       err,
	}
}

func TestGetTestCommand(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockShellExecutor)
		configSetup    func() config.ProjectDefinition
		expectedError  string
		expectWarnings bool
	}{
		{
			name: "successful test execution",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Test: config.Operation{
							Steps: []string{"go test ./...", "go test -race ./..."},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "go test ./...").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
				m.On("Exec", mock.Anything, "go test -race ./...").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
			},
		},
		{
			name: "test with no steps should warn",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Test: config.Operation{
							Steps: []string{},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				// No expectations for empty steps
			},
			expectWarnings: true,
		},
		{
			name: "test failure should return error",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Test: config.Operation{
							Steps: []string{"go test ./..."},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "go test ./...").Return(executor.Result{ExitCode: 1, Stderr: "test failed"}, nil)
			},
			expectedError: "tests failed",
		},
		{
			name: "test with environment variables",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Test: config.Operation{
							Env: map[string]string{
								"TEST_ENV":    "test_value",
								"GO111MODULE": "on",
							},
							Steps: []string{"go test ./..."},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.MatchedBy(func(env []string) bool {
					// Check that our env vars are included
					envStr := ""
					for _, e := range env {
						envStr += e + " "
					}
					return contains(envStr, "TEST_ENV=test_value") &&
						contains(envStr, "GO111MODULE=on")
				})).Return()
				m.On("Exec", mock.Anything, "go test ./...").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
			},
		},
		{
			name: "test with fail_fast enabled",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Test: config.Operation{
							FailFast: true,
							Steps:    []string{"go test ./pkg1", "go test ./pkg2"},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "go test ./pkg1").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
				m.On("Exec", mock.Anything, "go test ./pkg2").Return(executor.Result{ExitCode: 1, Stderr: "test failed"}, nil)
			},
			expectedError: "tests failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockShellExecutor{}
			tt.mockSetup(mockExecutor)

			// Create test command
			cmd := GetTestCommand(mockExecutor)

			// Create context with project definition
			logger := logging.New(os.Stderr, logrus.InfoLevel)
			ctx := logging.WithContext(context.Background(), logger)
			projectDef := tt.configSetup()
			ctx = config.WithContext(ctx, projectDef)
			cmd.SetContext(ctx)

			// Execute command
			err := cmd.Execute()

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

func TestGetTestCommand_CommandProperties(t *testing.T) {
	mockExecutor := &MockShellExecutor{}
	cmd := GetTestCommand(mockExecutor)

	// Test command properties
	assert.Equal(t, "test", cmd.Use)
	assert.Equal(t, "Run the test operations", cmd.Short)
	assert.Equal(t, "Run the designated test operations.", cmd.Long)
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)

	// Test that command accepts exactly 0 arguments
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err)

	// Test that command rejects arguments
	err = cmd.Args(cmd, []string{"extra-arg"})
	assert.Error(t, err)
}

func TestGetBuildCommand(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockShellExecutor)
		configSetup    func() config.ProjectDefinition
		expectedError  string
		expectWarnings bool
	}{
		{
			name: "successful build execution",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "build-project",
					Codebase: config.Codebase{
						Build: config.Operation{
							Steps: []string{"go build ./...", "go build -o ./bin/app ."},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "go build ./...").Return(executor.Result{ExitCode: 0, Stdout: "built"}, nil)
				m.On("Exec", mock.Anything, "go build -o ./bin/app .").Return(executor.Result{ExitCode: 0, Stdout: "built"}, nil)
			},
		},
		{
			name: "build with no steps should warn",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "build-project",
					Codebase: config.Codebase{
						Build: config.Operation{
							Steps: []string{},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				// No expectations for empty steps
			},
			expectWarnings: true,
		},
		{
			name: "build failure should return error",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "build-project",
					Codebase: config.Codebase{
						Build: config.Operation{
							Steps: []string{"go build ./..."},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "go build ./...").Return(executor.Result{ExitCode: 1, Stderr: "build failed"}, nil)
			},
			expectedError: "build failed",
		},
		{
			name: "build with environment variables",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "build-project",
					Codebase: config.Codebase{
						Build: config.Operation{
							Env: map[string]string{
								"BUILD_ENV":   "production",
								"GO111MODULE": "on",
							},
							Steps: []string{"go build ./..."},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.MatchedBy(func(env []string) bool {
					// Check that our env vars are included
					envStr := ""
					for _, e := range env {
						envStr += e + " "
					}
					return contains(envStr, "BUILD_ENV=production") &&
						contains(envStr, "GO111MODULE=on")
				})).Return()
				m.On("Exec", mock.Anything, "go build ./...").Return(executor.Result{ExitCode: 0, Stdout: "built"}, nil)
			},
		},
		{
			name: "build with fail_fast enabled",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "build-project",
					Codebase: config.Codebase{
						Build: config.Operation{
							FailFast: true,
							Steps:    []string{"go build ./pkg1", "go build ./pkg2"},
						},
					},
				}
			},
			mockSetup: func(m *MockShellExecutor) {
				m.On("AddEnv", mock.AnythingOfType("[]string")).Return()
				m.On("Exec", mock.Anything, "go build ./pkg1").Return(executor.Result{ExitCode: 0, Stdout: "built"}, nil)
				m.On("Exec", mock.Anything, "go build ./pkg2").Return(executor.Result{ExitCode: 1, Stderr: "build failed"}, nil)
			},
			expectedError: "build failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockShellExecutor{}
			tt.mockSetup(mockExecutor)

			// Create build command
			cmd := GetBuildCommand(mockExecutor)

			// Create context with project definition
			logger := logging.New(os.Stderr, logrus.InfoLevel)
			ctx := logging.WithContext(context.Background(), logger)
			projectDef := tt.configSetup()
			ctx = config.WithContext(ctx, projectDef)
			cmd.SetContext(ctx)

			// Execute command
			err := cmd.Execute()

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

func TestGetBuildCommand_CommandProperties(t *testing.T) {
	mockExecutor := &MockShellExecutor{}
	cmd := GetBuildCommand(mockExecutor)

	// Test command properties
	assert.Equal(t, "build", cmd.Use)
	assert.Equal(t, "Run the build operations", cmd.Short)
	assert.Equal(t, "Build the project according to the configuration..", cmd.Long)
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)

	// Test that command accepts exactly 0 arguments (cobra.NoArgs)
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err)

	// Test that command rejects arguments
	err = cmd.Args(cmd, []string{"extra-arg"})
	assert.Error(t, err)
}

func TestGetBuildCommand_NoContext(t *testing.T) {
	mockExecutor := &MockShellExecutor{}
	cmd := GetBuildCommand(mockExecutor)

	// Execute without context should panic
	assert.Panics(t, func() {
		cmd.Execute()
	})
}

func TestGetBuildCommand_Integration(t *testing.T) {
	mockExecutor := &MockShellExecutor{}

	// Setup mock expectations
	mockExecutor.On("AddEnv", mock.AnythingOfType("[]string")).Return()
	mockExecutor.On("Exec", mock.Anything, "go clean -testcache").Return(executor.Result{ExitCode: 0, Stdout: "cleaned"}, nil)
	mockExecutor.On("Exec", mock.Anything, "go test -cover ./...").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
	mockExecutor.On("Exec", mock.Anything, "go build -ldflags=\"-s -w\" -o ./devops .").Return(executor.Result{ExitCode: 0, Stdout: "built"}, nil)
	mockExecutor.On("Exec", mock.Anything, "chmod +x ./devops").Return(executor.Result{ExitCode: 0, Stdout: "executable"}, nil)

	// Create build command
	cmd := GetBuildCommand(mockExecutor)

	// Create context with project definition
	logger := logging.New(os.Stderr, logrus.InfoLevel)
	ctx := logging.WithContext(context.Background(), logger)
	projectDef := config.ProjectDefinition{
		Name: "integration-build",
		Codebase: config.Codebase{
			Build: config.Operation{
				Steps: []string{"go clean -testcache", "go test -cover ./...", "go build -ldflags=\"-s -w\" -o ./devops .", "chmod +x ./devops"},
			},
		},
	}
	ctx = config.WithContext(ctx, projectDef)
	cmd.SetContext(ctx)

	// Execute command
	err := cmd.Execute()
	assert.NoError(t, err)

	mockExecutor.AssertExpectations(t)
}

func TestGetTestCommand_Integration(t *testing.T) {
	mockExecutor := &MockShellExecutor{}

	// Setup mock expectations
	mockExecutor.On("AddEnv", mock.AnythingOfType("[]string")).Return()
	mockExecutor.On("Exec", mock.Anything, "go test ./...").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)
	mockExecutor.On("Exec", mock.Anything, "go test -race ./...").Return(executor.Result{ExitCode: 0, Stdout: "PASS"}, nil)

	// Create test command
	cmd := GetTestCommand(mockExecutor)

	// Create context with project definition
	logger := logging.New(os.Stderr, logrus.InfoLevel)
	ctx := logging.WithContext(context.Background(), logger)
	projectDef := config.ProjectDefinition{
		Name: "integration-test",
		Codebase: config.Codebase{
			Test: config.Operation{
				Steps: []string{"go test ./...", "go test -race ./..."},
			},
		},
	}
	ctx = config.WithContext(ctx, projectDef)
	cmd.SetContext(ctx)

	// Execute command
	err := cmd.Execute()
	assert.NoError(t, err)

	mockExecutor.AssertExpectations(t)
}

func TestGetTestCommand_NoContext(t *testing.T) {
	mockExecutor := &MockShellExecutor{}
	cmd := GetTestCommand(mockExecutor)

	// Execute without context should panic
	assert.Panics(t, func() {
		cmd.Execute()
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGetDoctorCommand(t *testing.T) {
	tests := []struct {
		name           string
		configSetup    func() config.ProjectDefinition
		expectedError  string
		expectWarnings bool
	}{
		{
			name: "successful validation with complete config",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name:        "test-project",
					Description: "A test project",
					Version:     "1.0.0",
					RepoUrl:     "https://github.com/test/project",
					Codebase: config.Codebase{
						Language:     "go",
						Dependencies: []string{"github.com/stretchr/testify"},
						Install: config.Operation{
							Steps: []string{"go mod download"},
						},
						Test: config.Operation{
							Steps: []string{"go test ./..."},
						},
						Build: config.Operation{
							Steps: []string{"go build ./..."},
						},
					},
				}
			},
		},
		{
			name: "validation with missing language should fail",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Test: config.Operation{
							Steps: []string{"go test ./..."},
						},
						Build: config.Operation{
							Steps: []string{"go build ./..."},
						},
					},
				}
			},
			expectedError: "validation failed",
		},
		{
			name: "validation with missing test steps should warn but pass",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Language: "go",
						Build: config.Operation{
							Steps: []string{"go build ./..."},
						},
					},
				}
			},
			expectWarnings: true,
		},
		{
			name: "validation with missing build steps should warn but pass",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Language: "go",
						Test: config.Operation{
							Steps: []string{"go test ./..."},
						},
					},
				}
			},
			expectWarnings: true,
		},
		{
			name: "validation with missing dependencies should warn but pass",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Language: "go",
						Test: config.Operation{
							Steps: []string{"go test ./..."},
						},
						Build: config.Operation{
							Steps: []string{"go build ./..."},
						},
					},
				}
			},
			expectWarnings: true,
		},
		{
			name: "validation with all optional fields missing should warn but pass",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Language: "go",
					},
				}
			},
			expectWarnings: true,
		},
		{
			name: "validation with empty language should fail",
			configSetup: func() config.ProjectDefinition {
				return config.ProjectDefinition{
					Name: "test-project",
					Codebase: config.Codebase{
						Language: "",
						Test: config.Operation{
							Steps: []string{"go test ./..."},
						},
						Build: config.Operation{
							Steps: []string{"go build ./..."},
						},
					},
				}
			},
			expectedError: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := &MockShellExecutor{}

			// Create doctor command
			cmd := GetDoctorCommand(mockExecutor)

			// Create context with project definition
			logger := logging.New(os.Stderr, logrus.InfoLevel)
			ctx := logging.WithContext(context.Background(), logger)
			projectDef := tt.configSetup()
			ctx = config.WithContext(ctx, projectDef)
			cmd.SetContext(ctx)

			// Capture output using cmd.SetOut and cmd.SetErr
			var buf bytes.Buffer
			cmd.SetOut(&buf)
			cmd.SetErr(&buf)

			// Execute command
			err := cmd.Execute()

			output := buf.String()

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectWarnings {
				// Check for warning messages in output
				assert.Contains(t, output, "[~]")
			}

			// Verify no shell executor calls were made (doctor only validates config)
			mockExecutor.AssertExpectations(t)
		})
	}
}

func TestGetDoctorCommand_CommandProperties(t *testing.T) {
	mockExecutor := &MockShellExecutor{}
	cmd := GetDoctorCommand(mockExecutor)

	// Test command properties
	assert.Equal(t, "doctor", cmd.Use)
	assert.Equal(t, "Validate your configuration", cmd.Short)
	assert.Equal(t, "Run checks on your configuration file to ensure it is ready for use.", cmd.Long)
	assert.True(t, cmd.SilenceUsage)
	assert.True(t, cmd.SilenceErrors)

	// Test that command accepts exactly 0 arguments (cobra.NoArgs)
	err := cmd.Args(cmd, []string{})
	assert.NoError(t, err)

	// Test that command rejects arguments
	err = cmd.Args(cmd, []string{"extra-arg"})
	assert.Error(t, err)
}

func TestGetDoctorCommand_NoContext(t *testing.T) {
	mockExecutor := &MockShellExecutor{}
	cmd := GetDoctorCommand(mockExecutor)

	// Execute without context should panic
	assert.Panics(t, func() {
		cmd.Execute()
	})
}

func TestGetDoctorCommand_Integration(t *testing.T) {
	mockExecutor := &MockShellExecutor{}

	// Create doctor command
	cmd := GetDoctorCommand(mockExecutor)

	// Create context with comprehensive project definition
	logger := logging.New(os.Stderr, logrus.InfoLevel)
	ctx := logging.WithContext(context.Background(), logger)
	projectDef := config.ProjectDefinition{
		Name:        "integration-doctor",
		Description: "Integration test project",
		Version:     "2.0.0",
		RepoUrl:     "https://github.com/integration/test",
		Codebase: config.Codebase{
			Language:     "go",
			Dependencies: []string{"github.com/stretchr/testify", "github.com/spf13/cobra"},
			Install: config.Operation{
				Steps: []string{"go mod download", "go mod tidy"},
			},
			Test: config.Operation{
				Steps: []string{"go test ./...", "go test -race ./..."},
			},
			Build: config.Operation{
				Steps: []string{"go build ./...", "go build -o ./bin/app ."},
			},
		},
	}
	ctx = config.WithContext(ctx, projectDef)
	cmd.SetContext(ctx)

	// Capture output using cmd.SetOut and cmd.SetErr
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute command
	err := cmd.Execute()

	output := buf.String()

	assert.NoError(t, err)

	// Check for success indicators in output
	assert.Contains(t, output, "[✔] Language: go")
	assert.Contains(t, output, "[✔] Dependencies:")
	assert.Contains(t, output, "[✔] Install steps")
	assert.Contains(t, output, "[✔] Test steps")
	assert.Contains(t, output, "[✔] Build steps")

	// Verify no shell executor calls were made
	mockExecutor.AssertExpectations(t)
}
