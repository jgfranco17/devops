package fileutils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func CopyFile(srcFS fs.FS, srcRelPath string, dstPath string) (err error) {
	// Open the source file
	srcFile, err := srcFS.Open(srcRelPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstDirPath := path.Dir(dstPath)
	err = os.MkdirAll(dstDirPath, 0755)
	if err != nil {
		return err
	}

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}

	// Copy the contents from the source file to the destination file
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}
	return dstFile.Close()
}

// CopyDirectory recursively copies a directory from an fs.FS to a destination path.
// Paths matching skip prefixes will not be copied. skip prefixes should be
// cleaned according to https://pkg.go.dev/path/filepath#Clean.
func CopyDirectory(srcFS fs.FS, srcDir, dstDir string, skip []string) (err error) {
	subFS, err := fs.Sub(srcFS, srcDir)
	if err != nil {
		return err
	}
	err = copyFS(dstDir, subFS, skip)
	if err != nil {
		return err
	}
	return nil
}

// copyFS copies the file system fsys into the directory dir,
// creating dir if necessary.
//
// Files are created with mode 0o666 plus any execute permissions
// from the source, and directories are created with mode 0o777
// (before umask).
//
// copyFS will not overwrite existing files. If a file name in fsys
// already exists in the destination, copyFS will return an error
// such that errors.Is(err, fs.ErrExist) will be true.
//
// Symbolic links in fsys are not supported and will be ignored.
//
// New files added to fsys (including if dir is a subdirectory of fsys)
// while copyFS is running are not guaranteed to be copied.
//
// Copying stops at and returns the first error encountered.
// Implementation for os.copyFS @ v1.24.5 with ignore symlinks
// https://cs.opensource.google/go/go/+/refs/tags/go1.24.5:src/os/dir.go;l=152
func copyFS(dir string, fsys fs.FS, skip []string) (err error) {
	shouldSkip := func(path string) bool {
		for _, i := range skip {
			// filepath.Match(i, path) cannot be used because it
			// does not support globstar patterns.
			if strings.HasPrefix(path, i) {
				return true
			}
		}
		return false
	}

	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if shouldSkip(path) {
			return nil
		}

		fpath, err := filepath.Localize(path)
		if err != nil {
			return err
		}
		newPath := filepath.Join(dir, fpath)
		if d.IsDir() {
			return os.MkdirAll(newPath, 0777)
		}

		// Ignore Symlinks and other irregular files
		if !d.Type().IsRegular() {
			// Original: &PathError{Op: "CopyFS", Path: path, Err: ErrInvalid}
			return nil
		}

		r, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		info, err := r.Stat()
		if err != nil {
			return err
		}
		w, err := os.OpenFile(newPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666|info.Mode()&0777)
		if err != nil {
			return err
		}

		if _, err := io.Copy(w, r); err != nil {
			_ = w.Close()
			// Original: &PathError{Op: "Copy", Path: newPath, Err: err}
			return fmt.Errorf("unable to copy content from %s to new path %s: %w", path, newPath, err)
		}
		return w.Close()
	})
}
