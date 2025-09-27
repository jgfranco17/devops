package fileutils

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
)

func TestCopyDirectory_Success(t *testing.T) {
	contents := []byte("my file contents")
	tmpDir := t.TempDir()

	fileSystem := fstest.MapFS{
		"my-folder/my-file.txt": {
			Mode: fs.ModePerm,
			Data: contents,
		},
		"my-folder/to-be-ignored/ignore.txt": {
			Mode: fs.ModePerm,
		},
	}

	destFolderPath := filepath.Join(tmpDir, "copied-folder")
	err := CopyDirectory(fileSystem, "my-folder", destFolderPath, []string{"to-be-ignored/"})

	assert.NoError(t, err)
	destFilePath := filepath.Join(destFolderPath, "my-file.txt")
	destContent, err := os.ReadFile(destFilePath)
	assert.NoError(t, err)
	assert.Equal(t, contents, destContent)

	destIgnorePath := filepath.Join(destFolderPath, "to-be-ignored/ignore.txt")
	_, err = os.ReadFile(destIgnorePath)
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestCopyDirectoryWithSymlink_Success(t *testing.T) {
	contents := []byte("my file contents")
	tmpDir := t.TempDir()

	fileSystem := fstest.MapFS{
		"my-folder/my-file.txt": {
			Mode: fs.ModePerm,
			Data: contents,
		},
		"my-folder/.direnv/nix_flake": {
			Mode: fs.ModeIrregular,
		},
	}

	destFolderPath := filepath.Join(tmpDir, "copied-folder")
	err := CopyDirectory(fileSystem, "my-folder", destFolderPath, nil)

	assert.NoError(t, err)
	destFilePath := filepath.Join(destFolderPath, "my-file.txt")
	destContent, err := os.ReadFile(destFilePath)
	assert.NoError(t, err)
	assert.Equal(t, contents, destContent)
	irregularFilePath := filepath.Join(destFolderPath, ".direnv", "nix_flake")
	_, err = os.Open(irregularFilePath)
	assert.True(t, os.IsNotExist(err))
}

func TestCopyDirectory_CannotReadSrc(t *testing.T) {
	fileSystem := fstest.MapFS{}
	err := CopyDirectory(fileSystem, "/path/does/not/exist", "/some/path", nil)
	assert.ErrorContains(t, err, "sub /path/does/not/exist: invalid argument")
}

func TestCopyLocalFile_Success(t *testing.T) {
	contents := []byte("my file contents")
	tmpDir := t.TempDir()

	fileSystem := fstest.MapFS{
		"my-file.txt": {
			Mode: fs.ModePerm,
			Data: contents,
		},
	}

	destFilePath := filepath.Join(tmpDir, "mycopiedfile.txt")
	err := CopyFile(fileSystem, "my-file.txt", destFilePath)

	assert.NoError(t, err)
	destContent, err := os.ReadFile(destFilePath)
	assert.NoError(t, err)
	assert.Equal(t, contents, destContent)
}

func TestCopyLocalFile_CannotReadSrc(t *testing.T) {
	fileSystem := fstest.MapFS{}
	err := CopyFile(fileSystem, "/path/does/not/exist.txt", "destination.txt")
	assert.ErrorContains(t, err, "file does not exist")
}
