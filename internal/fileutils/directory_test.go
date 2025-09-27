package fileutils

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDir(t *testing.T) {
	tmpDir := t.TempDir()
	file, err := os.Create(path.Join(tmpDir, "file"))
	require.NoError(t, err)
	defer file.Close()

	testCases := []struct {
		description string
		testPath    string
		isDir       bool
	}{
		{
			description: "Valid Directory",
			testPath:    tmpDir,
			isDir:       true,
		},
		{
			description: "Unknown Directory",
			testPath:    path.Join(tmpDir, "unknown"),
			isDir:       false,
		},
		{
			description: "Valid File",
			testPath:    path.Join(tmpDir, "file"),
			isDir:       false,
		},
		{
			description: "Unknown File",
			testPath:    path.Join(tmpDir, "unknownFile"),
			isDir:       false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.isDir, IsDir(tc.testPath))
		})
	}
}
