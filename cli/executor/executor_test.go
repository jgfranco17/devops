package executor

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResult_PrintStdOut(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected string
	}{
		{
			name: "print stdout with content",
			result: Result{
				Stdout: "Hello, World!",
			},
			expected: "Hello, World!\n",
		},
		{
			name: "print stdout with empty string",
			result: Result{
				Stdout: "",
			},
			expected: "",
		},
		{
			name: "print stdout with multiline content",
			result: Result{
				Stdout: "Line 1\nLine 2\nLine 3",
			},
			expected: "Line 1\nLine 2\nLine 3\n",
		},
		{
			name: "print stdout with special characters",
			result: Result{
				Stdout: "Special chars: !@#$%^&*()",
			},
			expected: "Special chars: !@#$%^&*()\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the method
			tt.result.PrintStdOut()

			// Close the write end and restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read the captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestResult_PrintStdErr(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected string
	}{
		{
			name: "print stderr with content",
			result: Result{
				Stderr: "Error message",
			},
			expected: "Error message\n",
		},
		{
			name: "print stderr with empty string",
			result: Result{
				Stderr: "",
			},
			expected: "",
		},
		{
			name: "print stderr with multiline content",
			result: Result{
				Stderr: "Error 1\nError 2\nError 3",
			},
			expected: "Error 1\nError 2\nError 3\n",
		},
		{
			name: "print stderr with special characters",
			result: Result{
				Stderr: "Error: !@#$%^&*()",
			},
			expected: "Error: !@#$%^&*()\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Call the method
			tt.result.PrintStdErr()

			// Close the write end and restore stderr
			w.Close()
			os.Stderr = oldStderr

			// Read the captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestDefaultExecutor_Exec(t *testing.T) {
	executor := &DefaultExecutor{}

	tests := []struct {
		name         string
		command      string
		expectedCode int
		expectError  bool
		expectedOut  string
		expectedErr  string
		timeout      time.Duration
	}{
		{
			name:         "successful command",
			command:      "echo 'Hello, World!'",
			expectedCode: 0,
			expectError:  false,
			expectedOut:  "Hello, World!\n",
			expectedErr:  "",
		},
		{
			name:         "command with stderr",
			command:      "echo 'Hello' >&2",
			expectedCode: 0,
			expectError:  false,
			expectedOut:  "",
			expectedErr:  "Hello\n",
		},
		{
			name:         "command with both stdout and stderr",
			command:      "echo 'stdout' && echo 'stderr' >&2",
			expectedCode: 0,
			expectError:  false,
			expectedOut:  "stdout\n",
			expectedErr:  "stderr\n",
		},
		{
			name:         "command that fails",
			command:      "false",
			expectedCode: 1,
			expectError:  true,
			expectedOut:  "",
			expectedErr:  "",
		},
		{
			name:         "command that doesn't exist",
			command:      "nonexistentcommand12345",
			expectedCode: 127,
			expectError:  true,
			expectedOut:  "",
			expectedErr:  "bash: line 1: nonexistentcommand12345: command not found\n",
		},
		{
			name:         "command with exit code 2",
			command:      "exit 2",
			expectedCode: 2,
			expectError:  true,
			expectedOut:  "",
			expectedErr:  "",
		},
		{
			name:         "command with exit code 127",
			command:      "exit 127",
			expectedCode: 127,
			expectError:  true,
			expectedOut:  "",
			expectedErr:  "",
		},
		{
			name:         "multiline output",
			command:      "echo -e 'Line 1\nLine 2\nLine 3'",
			expectedCode: 0,
			expectError:  false,
			expectedOut:  "Line 1\nLine 2\nLine 3\n",
			expectedErr:  "",
		},
		{
			name:         "command with special characters",
			command:      "echo 'Special: !@#$%^&*()'",
			expectedCode: 0,
			expectError:  false,
			expectedOut:  "Special: !@#$%^&*()\n",
			expectedErr:  "",
		},
		{
			name:         "command with timeout",
			command:      "sleep 10",
			expectedCode: -1,
			expectError:  true,
			expectedOut:  "",
			expectedErr:  "",
			timeout:      100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.timeout)
				defer cancel()
			}

			result, err := executor.Exec(ctx, tt.command)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCode, result.ExitCode)
			assert.Equal(t, tt.expectedOut, result.Stdout)
			assert.Equal(t, tt.expectedErr, result.Stderr)
		})
	}
}

func TestDefaultExecutor_Exec_WithEnvironment(t *testing.T) {
	executor := &DefaultExecutor{}
	executor.AddEnv([]string{"TEST_VAR=test_value", "ANOTHER_VAR=another_value"})

	// Note: AddEnv only stores the environment variables but doesn't actually set them
	// for the command execution. The command will use the current process environment.
	ctx := context.Background()
	result, err := executor.Exec(ctx, "echo $TEST_VAR $ANOTHER_VAR")

	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	// The variables won't be set in the command execution, so we expect empty output
	assert.Equal(t, "\n", result.Stdout)
}

func TestDefaultExecutor_AddEnv(t *testing.T) {
	tests := []struct {
		name        string
		envVars     []string
		expectedLen int
	}{
		{
			name:        "add single environment variable",
			envVars:     []string{"TEST_VAR=value1"},
			expectedLen: len(os.Environ()) + 1,
		},
		{
			name:        "add multiple environment variables",
			envVars:     []string{"VAR1=value1", "VAR2=value2", "VAR3=value3"},
			expectedLen: len(os.Environ()) + 3,
		},
		{
			name:        "add empty environment variables",
			envVars:     []string{},
			expectedLen: len(os.Environ()),
		},
		{
			name:        "add nil environment variables",
			envVars:     nil,
			expectedLen: len(os.Environ()),
		},
		{
			name:        "add environment variables with special characters",
			envVars:     []string{"SPECIAL_VAR=!@#$%^&*()", "PATH_VAR=/usr/bin:/bin"},
			expectedLen: len(os.Environ()) + 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &DefaultExecutor{}
			executor.AddEnv(tt.envVars)

			assert.Equal(t, tt.expectedLen, len(executor.Env))

			// Verify that the original environment is preserved
			originalEnv := os.Environ()
			for _, env := range originalEnv {
				assert.Contains(t, executor.Env, env)
			}

			// Verify that new environment variables are added
			for _, env := range tt.envVars {
				assert.Contains(t, executor.Env, env)
			}
		})
	}
}

func TestDefaultExecutor_AddEnv_Overwrite(t *testing.T) {
	executor := &DefaultExecutor{}

	// Add initial environment variables
	executor.AddEnv([]string{"TEST_VAR=initial_value"})

	// Add more environment variables (this will replace the entire Env slice)
	executor.AddEnv([]string{"NEW_VAR=new_value"})

	// Should have original + new (AddEnv replaces the entire slice)
	expectedLen := len(os.Environ()) + 1
	assert.Equal(t, expectedLen, len(executor.Env))
	assert.Contains(t, executor.Env, "NEW_VAR=new_value")
	// The previous TEST_VAR should not be present as AddEnv replaces the entire slice
	assert.NotContains(t, executor.Env, "TEST_VAR=initial_value")
}

func TestDefaultExecutor_Exec_ContextCancellation(t *testing.T) {
	executor := &DefaultExecutor{}

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel the context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	result, err := executor.Exec(ctx, "sleep 1")

	assert.Error(t, err)
	assert.Equal(t, -1, result.ExitCode)
	// The error could be either "context canceled" or "signal: killed" depending on timing
	assert.True(t, err.Error() == "context canceled" || err.Error() == "signal: killed")
}

func TestDefaultExecutor_Exec_EmptyCommand(t *testing.T) {
	executor := &DefaultExecutor{}

	ctx := context.Background()
	result, err := executor.Exec(ctx, "")

	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "", result.Stdout)
	assert.Equal(t, "", result.Stderr)
}

func TestDefaultExecutor_Exec_ComplexCommand(t *testing.T) {
	executor := &DefaultExecutor{}

	ctx := context.Background()
	result, err := executor.Exec(ctx, "echo 'Hello' && echo 'World' >&2 && echo 'Done'")

	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "Hello")
	assert.Contains(t, result.Stdout, "Done")
	assert.Contains(t, result.Stderr, "World")
}

func TestDefaultExecutor_Exec_CommandWithPipes(t *testing.T) {
	executor := &DefaultExecutor{}

	ctx := context.Background()
	result, err := executor.Exec(ctx, "echo 'Hello World' | wc -w")

	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "2") // "Hello World" has 2 words
}

func TestDefaultExecutor_Exec_CommandWithRedirects(t *testing.T) {
	executor := &DefaultExecutor{}

	ctx := context.Background()
	result, err := executor.Exec(ctx, "echo 'Output' > /dev/null && echo 'Success'")

	assert.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Contains(t, result.Stdout, "Success")
}

func TestResult_Struct(t *testing.T) {
	// Test Result struct creation and field access
	result := Result{
		Stdout:   "test output",
		Stderr:   "test error",
		ExitCode: 42,
	}

	assert.Equal(t, "test output", result.Stdout)
	assert.Equal(t, "test error", result.Stderr)
	assert.Equal(t, 42, result.ExitCode)
}

func TestDefaultExecutor_Struct(t *testing.T) {
	// Test DefaultExecutor struct creation
	executor := &DefaultExecutor{
		Env: []string{"TEST=value"},
	}

	assert.Equal(t, []string{"TEST=value"}, executor.Env)
	assert.NotNil(t, executor)
}
