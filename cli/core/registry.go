package core

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type CommandRegistry struct {
	rootCmd   *cobra.Command
	verbosity int
}

// NewCommandRegistry creates a new instance of CommandRegistry
func NewCommandRegistry(name string, description string, version string) *CommandRegistry {
	var verbosity int
	root := &cobra.Command{
		Use:     name,
		Version: version,
		Short:   description,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			verbosity, _ := cmd.Flags().GetCount("verbose")
			var level logrus.Level
			switch verbosity {
			case 1:
				level = logrus.InfoLevel
			case 2:
				level = logrus.DebugLevel
			case 3:
				level = logrus.TraceLevel
			default:
				level = logrus.WarnLevel
			}

			logger := logging.New(cmd.ErrOrStderr(), level)
			ctx := logging.WithContext(cmd.Context(), logger)
			ctx, cancel := context.WithCancel(ctx)
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
			go func() {
				select {
				case <-c:
					cancel()
				case <-ctx.Done():
				}
			}()
			cmd.SetContext(ctx)
		},
	}

	root.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity (-v or -vv)")

	return &CommandRegistry{
		rootCmd:   root,
		verbosity: verbosity,
	}
}

func (cr *CommandRegistry) GetMain() *cobra.Command {
	return cr.rootCmd
}

// RegisterCommand registers a new command with the CommandRegistry
func (cr *CommandRegistry) RegisterCommands(commands []*cobra.Command) {
	for _, cmd := range commands {
		cr.rootCmd.AddCommand(cmd)
	}
}

// Execute executes the root command
func (cr *CommandRegistry) Execute() error {
	return cr.rootCmd.Execute()
}
