package testutils

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

type CliCommandFunction func() *cobra.Command

type CommandRunner func(cmd *cobra.Command, args []string)

type ExecResult struct {
	Stdout string
	Stderr string
	RunErr error
}

// Helper function to simulate CLI execution
func RunCommand(t *testing.T, cmd *cobra.Command, args ...string) ExecResult {
	t.Helper()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)

	_, err := cmd.ExecuteC()
	return ExecResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		RunErr: err,
	}
}
