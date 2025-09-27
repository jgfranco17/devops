package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetVersion(t *testing.T) {
	tcases := []struct {
		name            string
		partDefinitions []byte
		expectedError   string
		expectedVersion string
	}{{
		name:            "empty partDefinitions",
		expectedError:   "unexpected end of JSON input",
		expectedVersion: "undefined",
	}, {
		name:            "invalid json partDefinitions",
		partDefinitions: []byte("invalid"),
		expectedError:   "invalid character",
		expectedVersion: "undefined",
	}, {
		name:            "version missing in partDefinitions",
		partDefinitions: []byte("{}"),
		expectedError:   ".arene/part_definitions.json does not have a version",
		expectedVersion: "undefined",
	}, {
		name:            "all good",
		partDefinitions: []byte(`{"version":"1.2.3"}`),
		expectedVersion: "1.2.3",
	}}

	for _, tc := range tcases {
		t.Run(tc.name, func(t *testing.T) {
			// Don't check the returned version as buildinfo.Version is checked already
			v, err := GetVersion(tc.partDefinitions)
			if tc.expectedError != "" {
				assert.ErrorContains(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, v, tc.expectedVersion)
		})
	}
}
