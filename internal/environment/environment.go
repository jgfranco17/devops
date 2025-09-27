package environment

import "os"

func IsRunningInCI() bool {
	ciVariables := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
	}
	for _, variable := range ciVariables {
		if os.Getenv(variable) != "" {
			return true
		}
	}
	return false
}
