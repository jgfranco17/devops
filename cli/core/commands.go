package core

import (
	"context"
	"fmt"
	"os"

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
	var filePath string
	var noInstall bool
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Run the build operations",
		Long:  "Read the config file and run the build operations defined in it.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logging.FromContext(cmd.Context())
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()
			logger.Debugf("Starting build with config file: %s", filePath)
			contents, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", filePath, err)
			}
			cfg, err := config.Load(contents)
			if err != nil {
				return fmt.Errorf("failed to load config from file: %w", err)
			}
			opts := &config.BuildOptions{
				NoInstall: noInstall,
			}
			if err := config.Build(ctx, shellExecutor, cfg, opts); err != nil {
				return fmt.Errorf("build failed: %w", err)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", ".opsrunner.yaml", "OpsRunner definition file")
	cmd.Flags().BoolVar(&noInstall, "no-install", false, "Install codebase dependencies before building")
	return cmd
}
