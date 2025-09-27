package config

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/devops/cli/executor"
	"github.com/jgfranco17/devops/internal/outputs"

	"gopkg.in/yaml.v3"
)

type ShellExecutor interface {
	Exec(ctx context.Context, command string) (executor.Result, error)
	AddEnv(env []string)
}

type ProjectDefinition struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version"`
	RepoUrl     string   `json:"repo_url"`
	Codebase    Codebase `json:"codebase"`
}

// Load reads a YAML configuration from the provided reader and unmarshals
// it into a struct instance.
func Load(r io.Reader) (*ProjectDefinition, error) {
	var cfg ProjectDefinition
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}
	return &cfg, nil
}

type Codebase struct {
	Language     string    `json:"language"`
	Dependencies string    `json:"dependencies,omitempty"`
	Install      Operation `json:"install,omitempty"`
	Build        Operation `json:"build,omitempty"`
}

type Operation struct {
	FailFast bool              `json:"fail_fast,omitempty"`
	Env      map[string]string `json:"env,omitempty"`
	Steps    []string          `json:"steps"`
}

// Run executes the defined steps in the Operation using the provided envs.
func (op *Operation) Run(ctx context.Context, executor ShellExecutor) error {
	logger := logging.FromContext(ctx)

	env := os.Environ()
	if len(op.Env) > 0 {
		envsAdded := []string{}
		for k, v := range op.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
			envsAdded = append(envsAdded, k)
		}
		logger.Infof("Loading additional %d additional environment variable(s): %v", len(op.Env), envsAdded)
	}
	executor.AddEnv(env)

	var failedSteps []string
	for idx, step := range op.Steps {
		fmt.Printf("[%d] %s\n", idx+1, step)
		result, err := executor.Exec(ctx, step)
		if err != nil || result.ExitCode != 0 {
			if op.FailFast {
				return fmt.Errorf("error while running '%s' (exit code %d): %w", step, result.ExitCode, err)
			}
			failedSteps = append(failedSteps, step)
		}
		if result.Stdout != "" {
			_, _ = fmt.Fprintf(os.Stdout, "%s\n", result.Stdout)
		}
		if result.Stderr != "" {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", result.Stderr)
		}
	}
	outputs.PrintTerminalWideLine("=")
	if len(failedSteps) > 0 {
		return fmt.Errorf("failed to run steps: %v", failedSteps)
	}
	return nil
}
