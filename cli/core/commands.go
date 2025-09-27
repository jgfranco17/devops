package core

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/devops/cli/config"
	"github.com/jgfranco17/devops/cli/executor"
)

type BashExecutor interface {
	Exec(ctx context.Context, command string) (executor.Result, error)
	AddEnv(env []string)
}

func GetBuildCommand(shellExecutor BashExecutor) *cobra.Command {
	var noInstall bool

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Run the build operations",
		Long:  "Read the config file and run the build operations defined in it.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logging.FromContext(cmd.Context())
			ctx := cmd.Context()
			cfg := config.FromContext(ctx)
			if err := cfg.Build(ctx, shellExecutor); err != nil {
				return fmt.Errorf("build failed: %w", err)
			}
			logger.Info("Build completed successfully")
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Flags().BoolVar(&noInstall, "no-install", false, "Install codebase dependencies before building")
	return cmd
}
