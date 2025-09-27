package core

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"github.com/jgfranco17/dev-tooling-go/logging"
	"github.com/jgfranco17/devops/cli/config"
	"github.com/jgfranco17/devops/internal/fileutils"
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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

			definition, err := loadConfig()
			if err != nil {
				return err
			}
			ctx = config.WithContext(ctx, definition)

			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx = fileutils.ApplyRootDirToContext(ctx, os.DirFS(cwd))

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
			return nil
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

func loadConfig() (config.ProjectDefinition, error) {
	path, err := config.GetFilePath()
	if err != nil {
		return config.ProjectDefinition{}, err
	}
	file, err := os.Open(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return config.ProjectDefinition{}, err
	}
	var cfg config.ProjectDefinition
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return config.ProjectDefinition{}, err
	}
	return cfg, nil
}
