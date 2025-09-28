// Package environment provides tooling for interacting with the environment.
package environment

import "os"

var ciVariables = []string{
	"CI",
	"GITHUB_ACTIONS",
	"GITLAB_CI",
	"NODE_NAME",
}

// IsRunningInCI checks if the current environment is running in a CI
// environment. It checks for the presence of the CI environment variables
// from known providers.
func IsRunningInCI() bool {
	for _, variable := range ciVariables {
		if os.Getenv(variable) != "" {
			return true
		}
	}
	return false
}
