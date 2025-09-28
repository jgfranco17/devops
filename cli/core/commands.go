package core

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jgfranco17/devops/cli/config"
	"github.com/jgfranco17/devops/cli/executor"
)

type BashExecutor interface {
	Exec(ctx context.Context, command string) (executor.Result, error)
	AddEnv(env []string)
}

func GetBuildCommand(shellExecutor BashExecutor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Run the build operations",
		Long:  "Build the project according to the configuration..",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := config.FromContext(ctx)
			if err := cfg.Build(ctx, shellExecutor); err != nil {
				return fmt.Errorf("build failed: %w", err)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	return cmd
}

func GetTestCommand(shellExecutor BashExecutor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run the test operations",
		Long:  "Run the designated test operations.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := config.FromContext(ctx)
			if err := cfg.Test(ctx, shellExecutor); err != nil {
				return fmt.Errorf("tests failed: %w", err)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	return cmd
}

func GetDoctorCommand(shellExecutor BashExecutor) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Validate your configuration",
		Long:  "Run checks on your configuration file to ensure it is ready for use.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cfg := config.FromContext(ctx)
			w := cmd.OutOrStdout()
			fmt.Fprintln(w, "===== DEVOPS DOCTOR =====")
			if err := cfg.ValidateTo(ctx, w); err != nil {
				return fmt.Errorf("validation failed: %w", err)
			}
			return nil
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	return cmd
}
