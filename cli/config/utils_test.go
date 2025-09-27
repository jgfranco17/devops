package config

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithTempEnv_SetsAndRestoresVars(t *testing.T) {
	ctx := context.Background()

	// Save original env
	_ = os.Setenv("ORIGINAL_ENV", "original")
	_ = os.Unsetenv("NEW_ENV")

	restore, err := WithTempEnv(ctx, map[string]string{
		"ORIGINAL_ENV": "temporary",
		"NEW_ENV":      "newvalue",
	})

	assert.NoError(t, err)

	// Check that vars are temporarily set
	assert.Equal(t, "temporary", os.Getenv("ORIGINAL_ENV"))
	assert.Equal(t, "newvalue", os.Getenv("NEW_ENV"))

	// Restore original environment
	restore()

	// ORIGINAL_ENV should be restored
	assert.Equal(t, "original", os.Getenv("ORIGINAL_ENV"))

	// NEW_ENV should be unset
	_, isSet := os.LookupEnv("NEW_ENV")
	assert.False(t, isSet)
}
