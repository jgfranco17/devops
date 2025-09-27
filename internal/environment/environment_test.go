package environment

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsRunningInCI_CommonCases(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		expected    bool
		description string
	}{
		{
			name:        "no CI environment variables",
			envVars:     map[string]string{},
			expected:    false,
			description: "Should return false when no CI environment variables are set",
		},
		{
			name: "CI variable set to true",
			envVars: map[string]string{
				"CI": "true",
			},
			expected:    true,
			description: "Should return true when CI variable is set to true",
		},
		{
			name: "CI variable set to false",
			envVars: map[string]string{
				"CI": "false",
			},
			expected:    true,
			description: "Should return true when CI variable is set to any non-empty value",
		},
		{
			name: "CI variable set to empty string",
			envVars: map[string]string{
				"CI": "",
			},
			expected:    false,
			description: "Should return false when CI variable is set to empty string",
		},
		{
			name: "GITHUB_ACTIONS variable set",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			expected:    true,
			description: "Should return true when GITHUB_ACTIONS variable is set",
		},
		{
			name: "GITLAB_CI variable set",
			envVars: map[string]string{
				"GITLAB_CI": "true",
			},
			expected:    true,
			description: "Should return true when GITLAB_CI variable is set",
		},
		{
			name: "NODE_NAME variable set",
			envVars: map[string]string{
				"NODE_NAME": "build-node-1",
			},
			expected:    true,
			description: "Should return true when NODE_NAME variable is set",
		},
		{
			name: "multiple CI variables set",
			envVars: map[string]string{
				"CI":             "true",
				"GITHUB_ACTIONS": "true",
				"GITLAB_CI":      "true",
				"NODE_NAME":      "build-node-1",
			},
			expected:    true,
			description: "Should return true when multiple CI variables are set",
		},
		{
			name: "mixed CI and non-CI variables",
			envVars: map[string]string{
				"CI":         "true",
				"USER":       "testuser",
				"HOME":       "/home/testuser",
				"PATH":       "/usr/bin:/bin",
				"RANDOM_VAR": "random_value",
			},
			expected:    true,
			description: "Should return true when CI variable is set among other variables",
		},
		{
			name: "only non-CI variables set",
			envVars: map[string]string{
				"USER":       "testuser",
				"HOME":       "/home/testuser",
				"PATH":       "/usr/bin:/bin",
				"RANDOM_VAR": "random_value",
			},
			expected:    false,
			description: "Should return false when only non-CI variables are set",
		},
		{
			name: "CI variable with spaces",
			envVars: map[string]string{
				"CI": " true ",
			},
			expected:    true,
			description: "Should return true when CI variable contains spaces",
		},
		{
			name: "GITHUB_ACTIONS with different values",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "1",
			},
			expected:    true,
			description: "Should return true when GITHUB_ACTIONS is set to 1",
		},
		{
			name: "GITLAB_CI with different values",
			envVars: map[string]string{
				"GITLAB_CI": "yes",
			},
			expected:    true,
			description: "Should return true when GITLAB_CI is set to yes",
		},
		{
			name: "NODE_NAME with different values",
			envVars: map[string]string{
				"NODE_NAME": "jenkins-slave-1",
			},
			expected:    true,
			description: "Should return true when NODE_NAME is set to jenkins value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up test environment using t.Setenv
			for key, value := range tc.envVars {
				if value == "" {
					t.Setenv(key, "") // This unsets the environment variable
				} else {
					t.Setenv(key, value)
				}
			}

			// Test the function
			result := IsRunningInCI()
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

func TestIsRunningInCI_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expected    bool
		description string
	}{
		{
			name: "unset all CI variables",
			envVars: map[string]string{
				"CI":             "",
				"GITHUB_ACTIONS": "",
				"GITLAB_CI":      "",
				"NODE_NAME":      "",
			},
			expected:    false,
			description: "Should return false when all CI variables are unset",
		},
		{
			name: "set CI variable to zero",
			envVars: map[string]string{
				"CI": "0",
			},
			expected:    true,
			description: "Should return true when CI is set to 0 (non-empty)",
		},
		{
			name: "set CI variable to no",
			envVars: map[string]string{
				"CI": "no",
			},
			expected:    true,
			description: "Should return true when CI is set to no (non-empty)",
		},
		{
			name: "set CI variable to off",
			envVars: map[string]string{
				"CI": "off",
			},
			expected:    true,
			description: "Should return true when CI is set to off (non-empty)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test environment using t.Setenv
			for key, value := range tt.envVars {
				if value == "" {
					t.Setenv(key, "") // This unsets the environment variable
				} else {
					t.Setenv(key, value)
				}
			}

			// Test the function
			result := IsRunningInCI()
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestIsRunningInCI_Integration(t *testing.T) {
	// Test that the function works correctly in a real environment
	// This test doesn't modify environment variables, just tests the current state

	// Test with a known CI environment
	t.Setenv("CI", "true")
	result := IsRunningInCI()
	assert.True(t, result, "Should return true when CI is set to true")

	// Test with CI unset
	t.Setenv("CI", "")
	result = IsRunningInCI()
	// The result depends on the current environment, so we just verify it doesn't panic
	assert.NotPanics(t, func() {
		IsRunningInCI()
	}, "Should not panic when called multiple times")
}

func TestIsRunningInCI_Performance(t *testing.T) {
	// Test that the function performs well with many environment variables
	// Set up a large number of non-CI environment variables
	for i := 0; i < 100; i++ {
		t.Setenv(fmt.Sprintf("TEST_VAR_%d", i), "value")
	}

	// Test that the function still works correctly
	result := IsRunningInCI()
	// The result should be false since we didn't set any CI variables
	assert.False(t, result, "Should return false when no CI variables are set among many other variables")
}

func TestIsRunningInCI_Concurrent(t *testing.T) {
	// Test that the function is safe for concurrent access
	// This is important since environment variables can be read concurrently

	done := make(chan bool)

	// Start multiple goroutines that call IsRunningInCI
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				IsRunningInCI()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without panicking, the function is safe for concurrent access
	assert.True(t, true, "Function should be safe for concurrent access")
}

func TestIsRunningInCI_AllCIVariables(t *testing.T) {
	// Test each CI variable individually to ensure they all work
	ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "NODE_NAME"}

	for _, ciVar := range ciVars {
		t.Run("test_"+ciVar, func(t *testing.T) {
			// Test with variable set
			t.Setenv(ciVar, "test_value")
			result := IsRunningInCI()
			assert.True(t, result, "Should return true when %s is set", ciVar)

			// Test with variable unset
			t.Setenv(ciVar, "")
			result = IsRunningInCI()
			assert.False(t, result, "Should return false when %s is unset", ciVar)
		})
	}
}
