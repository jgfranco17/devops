package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/sirupsen/logrus"
)

const (
	DefinitionFile = "devops-definition.yaml"
)

// GetFilePath returns the path to the project definition file.
func GetFilePath() (string, error) {
	currentWorkingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working directory: %w", err)
	}
	projectConfigPath := filepath.Join(currentWorkingDir, DefinitionFile)
	_, err = os.Stat(projectConfigPath)
	return projectConfigPath, err
}

// WithTempEnv sets environment variables from the provided map,
// saves any existing values, and restores them after the callback.
func WithTempEnv(ctx context.Context, vars map[string]string) (func(), error) {
	logger := logging.FromContext(ctx)

	originals := make(map[string]*string)
	for key, val := range vars {
		if existing, ok := os.LookupEnv(key); ok {
			originals[key] = &existing
		} else {
			originals[key] = nil
		}
		err := os.Setenv(key, val)
		logger.Infof("Using: %s=%s", key, val)
		if err != nil {
			return nil, err
		}
	}

	// Restore original environment
	restoreFunc := func() {
		for key, val := range originals {
			if val == nil {
				_ = os.Unsetenv(key)
			} else {
				_ = os.Setenv(key, *val)
			}
		}
		logger.WithFields(logrus.Fields{
			"count": len(originals),
		}).Debug("Restored environment")
	}

	return restoreFunc, nil
}
