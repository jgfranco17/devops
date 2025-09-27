package environment

import "os"

// IsRunningInCI checks if the current environment is running in a CI
// environment. It checks for the presence of the CI environment variables
// from known providers.
func IsRunningInCI() bool {
	ciVariables := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"NODE_NAME",
	}
	for _, variable := range ciVariables {
		if os.Getenv(variable) != "" {
			return true
		}
	}
	return false
}
