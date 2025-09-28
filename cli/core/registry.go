package core

import (
	"context"
	"errors"
	"fmt"
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
	var path string

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

			definition, err := loadConfig(ctx, path)
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
	root.PersistentFlags().StringVarP(&path, "file", "f", config.DefinitionFile, "Path to the project definition file")
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

func loadConfig(ctx context.Context, path string) (config.ProjectDefinition, error) {
	logger := logging.FromContext(ctx)
	pathToUse := path
	_, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return config.ProjectDefinition{}, err
		}
		logger.WithFields(logrus.Fields{
			"path": path,
		}).Warn("Config not found at given path, using default")
		defaultPath, err := config.GetFilePath()
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return config.ProjectDefinition{}, err
			}
		} else {
			pathToUse = defaultPath
		}
	}
	logger.WithFields(logrus.Fields{
		"path": pathToUse,
	}).Trace("Found config file")
	file, err := os.Open(pathToUse)
	if err != nil {
		return config.ProjectDefinition{}, fmt.Errorf("failed to open config (%s): %w", pathToUse, err)
	}
	defer file.Close()

	cfg, err := config.Load(file)
	if err != nil {
		return config.ProjectDefinition{}, fmt.Errorf("failed to load config (%s): %w", pathToUse, err)
	}
	return *cfg, nil
}
