package config

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"
)

// WithTempEnv sets environment variables from the provided map,
// saves any existing values, and restores them after the callback.
func WithTempEnv(ctx context.Context, vars map[string]string) (func(), error) {
	// Save original values and set new ones
	originals := make(map[string]*string)
	for key, val := range vars {
		if existing, ok := os.LookupEnv(key); ok {
			originals[key] = &existing
		} else {
			originals[key] = nil
		}
		err := os.Setenv(key, val)
		log.Infof("Using: %s=%s", key, val)
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
		log.Infof("Restored %d envs\n", len(originals))
	}

	return restoreFunc, nil
}
