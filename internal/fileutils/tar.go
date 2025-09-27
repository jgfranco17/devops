package fileutils

import (
	"bytes"
	"os/exec"
)

func UntarFile(src, dest string) error {
	cmd := exec.Command("tar", "-xzf", src, "-C", dest)
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func CreateTarGz(src, dest string) (err error) {
	var buf bytes.Buffer
	cmd := exec.Command("tar", "-czf", dest, "-C", src, ".")
	cmd.Stderr = &buf
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
