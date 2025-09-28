package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (r *Result) PrintStdOut() {
	if r.Stdout != "" {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", r.Stdout)
	}
}

func (r *Result) PrintStdErr() {
	if r.Stderr != "" {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", r.Stderr)
	}
}

type DefaultExecutor struct {
	Env []string
}

func (c *DefaultExecutor) Exec(ctx context.Context, command string) (Result, error) {
	var stdoutBuf, stderrBuf bytes.Buffer

	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()

	exitCode := 0
	if err != nil {
		// Get exit code if available
		if exitErr, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode = ws.ExitStatus()
			} else {
				// Unix-only, fallback
				exitCode = -1
			}
		} else {
			// Non-exit error, e.g., binary not found
			exitCode = -1
		}
	}

	return Result{
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
		ExitCode: exitCode,
	}, err
}

func (c *DefaultExecutor) AddEnv(envs []string) {
	baseEnv := os.Environ()
	if len(envs) > 0 {
		baseEnv = append(baseEnv, envs...)
	}
	c.Env = baseEnv
}
