package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestGetFilePath(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T) (string, func())
		expectError  bool
		expectedPath string
	}{
		{
			name: "file exists in current directory",
			setupFunc: func(t *testing.T) (string, func()) {
				// Create a temporary directory and file
				tempDir := t.TempDir()
				configFile := filepath.Join(tempDir, DefinitionFile)
				file, err := os.Create(configFile)
				assert.NoError(t, err)
				file.Close()

				// Change to temp directory
				originalDir, err := os.Getwd()
				assert.NoError(t, err)
				err = os.Chdir(tempDir)
				assert.NoError(t, err)

				return tempDir, func() {
					os.Chdir(originalDir)
				}
			},
			expectError:  false,
			expectedPath: filepath.Join(t.TempDir(), DefinitionFile),
		},
		{
			name: "file does not exist in current directory",
			setupFunc: func(t *testing.T) (string, func()) {
				// Create a temporary directory without the config file
				tempDir := t.TempDir()

				// Change to temp directory
				originalDir, err := os.Getwd()
				assert.NoError(t, err)
				err = os.Chdir(tempDir)
				assert.NoError(t, err)

				return tempDir, func() {
					os.Chdir(originalDir)
				}
			},
			expectError:  true,
			expectedPath: filepath.Join(t.TempDir(), DefinitionFile),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, cleanup := tt.setupFunc(t)
			defer cleanup()

			path, err := GetFilePath()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no such file or directory")
			} else {
				assert.NoError(t, err)
			}

			expectedPath := filepath.Join(tempDir, DefinitionFile)
			assert.Equal(t, expectedPath, path)
		})
	}
}

func TestWithTempEnv(t *testing.T) {
	logger := logging.New(os.Stderr, logrus.DebugLevel)
	ctx := logging.WithContext(context.Background(), logger)

	tests := []struct {
		name            string
		envVars         map[string]string
		originalEnv     map[string]string
		expectedError   bool
		validateEnv     func(t *testing.T, envVars map[string]string)
		validateRestore func(t *testing.T)
	}{
		{
			name: "set new environment variables",
			envVars: map[string]string{
				"TEST_VAR1": "value1",
				"TEST_VAR2": "value2",
			},
			originalEnv: map[string]string{},
			validateEnv: func(t *testing.T, envVars map[string]string) {
				assert.Equal(t, "value1", os.Getenv("TEST_VAR1"))
				assert.Equal(t, "value2", os.Getenv("TEST_VAR2"))
			},
			validateRestore: func(t *testing.T) {
				_, exists1 := os.LookupEnv("TEST_VAR1")
				_, exists2 := os.LookupEnv("TEST_VAR2")
				assert.False(t, exists1)
				assert.False(t, exists2)
			},
		},
		{
			name: "override existing environment variables",
			envVars: map[string]string{
				"PATH":     "/custom/path",
				"TEST_VAR": "test_value",
			},
			originalEnv: map[string]string{
				"PATH": "/original/path",
			},
			validateEnv: func(t *testing.T, envVars map[string]string) {
				assert.Equal(t, "/custom/path", os.Getenv("PATH"))
				assert.Equal(t, "test_value", os.Getenv("TEST_VAR"))
			},
			validateRestore: func(t *testing.T) {
				// PATH should be restored to original value
				assert.Equal(t, "/original/path", os.Getenv("PATH"))
				// TEST_VAR should be unset
				_, exists := os.LookupEnv("TEST_VAR")
				assert.False(t, exists)
			},
		},
		{
			name:        "empty environment variables map",
			envVars:     map[string]string{},
			originalEnv: map[string]string{},
			validateEnv: func(t *testing.T, envVars map[string]string) {
				// No changes expected
			},
			validateRestore: func(t *testing.T) {
				// No changes expected
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up original environment
			for key, value := range tt.originalEnv {
				os.Setenv(key, value)
			}

			// Clean up after test
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
				for key := range tt.originalEnv {
					os.Unsetenv(key)
				}
			}()

			restore, err := WithTempEnv(ctx, tt.envVars)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, restore)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, restore)

				// Validate environment is set correctly
				tt.validateEnv(t, tt.envVars)

				// Restore and validate
				restore()
				tt.validateRestore(t)
			}
		})
	}
}

func TestWithTempEnv_ErrorHandling(t *testing.T) {
	logger := logging.New(os.Stderr, logrus.DebugLevel)
	ctx := logging.WithContext(context.Background(), logger)

	// Test with invalid environment variable name
	// This is hard to test directly since os.Setenv is quite permissive,
	// but we can test the function doesn't panic
	restore, err := WithTempEnv(ctx, map[string]string{
		"": "empty_key", // This might cause issues
	})

	// The function should handle this gracefully
	if err != nil {
		assert.Nil(t, restore)
	} else {
		assert.NotNil(t, restore)
		restore() // Clean up
	}
}

func TestWithTempEnv_MultipleCalls(t *testing.T) {
	logger := logging.New(os.Stderr, logrus.DebugLevel)
	ctx := logging.WithContext(context.Background(), logger)

	// First call
	restore1, err1 := WithTempEnv(ctx, map[string]string{
		"VAR1": "value1",
	})
	assert.NoError(t, err1)
	assert.Equal(t, "value1", os.Getenv("VAR1"))

	// Second call
	restore2, err2 := WithTempEnv(ctx, map[string]string{
		"VAR2": "value2",
	})
	assert.NoError(t, err2)
	assert.Equal(t, "value1", os.Getenv("VAR1"))
	assert.Equal(t, "value2", os.Getenv("VAR2"))

	// Restore in reverse order
	restore2()
	assert.Equal(t, "value1", os.Getenv("VAR1"))
	_, exists := os.LookupEnv("VAR2")
	assert.False(t, exists)

	restore1()
	_, exists = os.LookupEnv("VAR1")
	assert.False(t, exists)
}
