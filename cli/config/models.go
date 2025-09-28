package config

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/devops/cli/executor"
	"github.com/jgfranco17/devops/internal/outputs"
	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v3"
)

type ShellExecutor interface {
	Exec(ctx context.Context, command string) (executor.Result, error)
	AddEnv(env []string)
}

type ProjectDefinition struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Version     string   `yaml:"version"`
	RepoUrl     string   `yaml:"repo_url"`
	Codebase    Codebase `yaml:"codebase"`
}

func (d *ProjectDefinition) Validate(ctx context.Context) error {
	return d.ValidateTo(ctx, os.Stdout)
}

func (d *ProjectDefinition) ValidateTo(ctx context.Context, w io.Writer) error {
	logger := logging.FromContext(ctx)
	fixes := []string{}
	suggestions := []string{}

	if d.Codebase.Language == "" {
		outputs.PrintColoredMessageTo(w, "red", "[✘] Language is required")
		fixes = append(fixes, "Set a language in the codebase")
	} else {
		outputs.PrintColoredMessageTo(w, "green", "[✔] Language: %s", d.Codebase.Language)
	}

	if d.Codebase.Dependencies != nil {
		outputs.PrintColoredMessageTo(w, "green", "[✔] Dependencies: %s", d.Codebase.Dependencies)
	} else {
		outputs.PrintColoredMessageTo(w, "yellow", "[~] No dependencies defined")
		suggestions = append(suggestions, "Set dependencies in the codebase")
	}

	if d.Codebase.Install.Steps != nil {
		outputs.PrintColoredMessageTo(w, "green", "[✔] Install steps (%d)", len(d.Codebase.Install.Steps))
	}

	if d.Codebase.Test.Steps != nil {
		outputs.PrintColoredMessageTo(w, "green", "[✔] Test steps (%d)", len(d.Codebase.Test.Steps))
	} else {
		outputs.PrintColoredMessageTo(w, "yellow", "[~] No test steps defined")
		suggestions = append(suggestions, "Set test steps in the codebase")
	}

	if d.Codebase.Build.Steps != nil {
		outputs.PrintColoredMessageTo(w, "green", "[✔] Build steps (%d)", len(d.Codebase.Build.Steps))
	} else {
		outputs.PrintColoredMessageTo(w, "yellow", "[~] No build steps defined")
		suggestions = append(suggestions, "Set build steps in the codebase")
	}

	outputs.PrintTerminalWideLineTo(w, "=")
	if len(suggestions) > 0 {
		outputs.PrintColoredMessageTo(w, "yellow", "Suggestions:")
		for _, suggestion := range suggestions {
			outputs.PrintColoredMessageTo(w, "yellow", "  - %s", suggestion)
		}
	}
	if len(fixes) > 0 {
		outputs.PrintColoredMessageTo(w, "red", "Fixes:")
		for _, fix := range fixes {
			outputs.PrintColoredMessageTo(w, "red", "  - %s", fix)
		}
		return fmt.Errorf("found %d required fixes", len(fixes))
	}

	logger.Info("Project definition validated successfully")
	return nil
}

func (d *ProjectDefinition) Test(ctx context.Context, shellExecutor ShellExecutor) error {
	logger := logging.FromContext(ctx)
	if len(d.Codebase.Test.Steps) == 0 {
		logger.Warn("No test steps defined in the configuration.")
		return nil
	}
	if err := d.Codebase.Test.Run(ctx, shellExecutor); err != nil {
		return fmt.Errorf("failed to run test steps: %w", err)
	}
	logger.Info("Tests completed successfully")
	return nil
}

func (d *ProjectDefinition) Build(ctx context.Context, shellExecutor ShellExecutor) error {
	logger := logging.FromContext(ctx)
	startTime := time.Now()

	if len(d.Codebase.Build.Steps) == 0 {
		logger.Warn("No build steps defined in the configuration.")
		return nil
	}
	if err := d.Codebase.Build.Run(ctx, shellExecutor); err != nil {
		return fmt.Errorf("failed to run build steps: %w", err)
	}
	duration := time.Since(startTime)
	logger.WithFields(logrus.Fields{
		"duration": duration,
	}).Info("Build completed successfully")
	return nil
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
	Language     string    `yaml:"language"`
	Dependencies []string  `yaml:"dependencies,omitempty"`
	Install      Operation `yaml:"install,omitempty"`
	Test         Operation `yaml:"test,omitempty"`
	Build        Operation `yaml:"build,omitempty"`
}

type Operation struct {
	FailFast bool              `yaml:"fail_fast,omitempty"`
	Env      map[string]string `yaml:"env,omitempty"`
	Steps    []string          `yaml:"steps"`
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
