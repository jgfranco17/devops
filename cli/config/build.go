package config

import (
	"context"
	"fmt"
	"time"

	"github.com/jgfranco17/dev-tooling-go/logging"
)

type BuildOptions struct {
	NoInstall bool
}

func Build(ctx context.Context, shellExecutor ShellExecutor, config *ProjectDefinition, opts *BuildOptions) error {
	logger := logging.FromContext(ctx)
	startTime := time.Now()

	if opts == nil {
		opts = &BuildOptions{}
	}
	if !opts.NoInstall {
		logger.Debug("Installing codebase dependencies")
		if err := config.Codebase.Install.Run(ctx, shellExecutor); err != nil {
			return fmt.Errorf("failed to install codebase dependencies: %w", err)
		}
	} else {
		logger.Info("Skipping codebase dependency installation")
	}
	if len(config.Codebase.Build.Steps) == 0 {
		logger.Warn("No build steps defined in the configuration.")
	}
	if err := config.Codebase.Build.Run(ctx, shellExecutor); err != nil {
		return fmt.Errorf("failed to run build steps: %w", err)
	}
	duration := time.Since(startTime)
	logger.Infof("Build completed successfully in %dms", duration.Milliseconds())
	return nil
}
