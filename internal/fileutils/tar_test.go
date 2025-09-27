package fileutils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUntarSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	tarFile, err := os.Create(tmpDir + "test_sample.tar.gz")
	require.NoError(t, err)
	defer func() {
		err := tarFile.Close()
		require.NoError(t, err)
	}()

	files := map[string][]byte{
		"file_1.txt":     []byte("Hello from file 1"),
		"sub/file_2.txt": []byte("Hello from file 2"),
	}
	createTar(t, tarFile, files)

	err = UntarFile(tarFile.Name(), tmpDir)
	assert.NoError(t, err)

	// Assert on tar content
	for fileName, expectedContent := range files {
		untarFilePath := filepath.Join(tmpDir, fileName)

		assert.FileExists(t, untarFilePath)

		fileContent, err := os.ReadFile(untarFilePath)
		require.NoError(t, err)

		assert.Equal(t, expectedContent, fileContent)
	}
}

func TestCreateTarGz(t *testing.T) {
	tempDir := t.TempDir()
	srcDir := filepath.Join(tempDir, "src")
	destFile := filepath.Join(tempDir, "test.tar.gz")

	err := os.MkdirAll(srcDir, 0755)
	require.NoError(t, err)
	testFilePath := filepath.Join(srcDir, "test.txt")
	err = os.WriteFile(testFilePath, []byte("Hello, World!"), 0644)
	require.NoError(t, err)

	err = CreateTarGz(srcDir, destFile)
	require.NoError(t, err)

	_, err = os.Stat(destFile)
	assert.NoError(t, err)
}

func writeTar(t *testing.T, tw *tar.Writer, name string, contents []byte, mode int64) {
	hdr := &tar.Header{
		Name:    name,
		Size:    int64(len(contents)),
		Mode:    mode,
		ModTime: time.Now(),
	}
	err := tw.WriteHeader(hdr)
	require.NoError(t, err)
	_, err = tw.Write(contents)
	require.NoError(t, err)
}

func createTar(t *testing.T, writer io.Writer, files map[string][]byte) {
	gw := gzip.NewWriter(writer)
	defer func() {
		err := gw.Close()
		require.NoError(t, err)
	}()

	tw := tar.NewWriter(gw)
	defer func() {
		err := tw.Close()
		require.NoError(t, err)
	}()

	for name, contents := range files {
		writeTar(t, tw, name, contents, 0644)
	}
}
