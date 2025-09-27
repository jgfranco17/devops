package buildinfo

import (
	"encoding/json"
)

// GetVersion returns the version found in a JSON definition.
func GetVersion(definition []byte) (version string, err error) {
	partInfo := struct {
		Version string
	}{}
	if err := json.Unmarshal(definition, &partInfo); err != nil {
		return "undefined", err
	}
	if partInfo.Version == "" {
		return "undefined", err
	}
	return partInfo.Version, nil
}
