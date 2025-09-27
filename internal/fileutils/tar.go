package fileutils

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/arene-vertex/arene-vertex-cli/internal/derrors"
)

func UntarFile(src, dest string) error {
	cmd := exec.Command("tar", "-xzf", src, "-C", dest)
	err := cmd.Run()
	if err != nil {
		return &derrors.VertexError{
			Err:      err,
			Code:     "VXVC0065",
			ExitCode: derrors.VertexRuntimeError,
		}
	}
	return nil
}

func CreateTarGz(src, dest string) (err error) {
	defer derrors.Wrap(&err, "CreateTarGz(%s, %s)", src, dest)
	var buf bytes.Buffer
	cmd := exec.Command("tar", "-czf", dest, "-C", src, ".")
	cmd.Stderr = &buf
	err = cmd.Run()
	if err != nil {
		return &derrors.VertexError{
			Err:      fmt.Errorf("unable to run tar command: %s", buf.String()),
			Code:     "VXVC0066",
			ExitCode: derrors.VertexRuntimeError,
		}
	}
	return nil
}
