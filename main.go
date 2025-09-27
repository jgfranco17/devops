package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jgfranco17/devops/cli/core"
	"github.com/jgfranco17/devops/cli/executor"
)

const (
	projectName        = "devops"
	projectDescription = "DevOps: Simplifying your CI/CD pipelines."
)

var (
	version string = "0.0.0-dev.1"
)

func main() {
	executor := &executor.DefaultExecutor{}
	command := core.NewCommandRegistry(projectName, projectDescription, version)
	commandsList := []*cobra.Command{
		core.GetBuildCommand(executor),
	}
	command.RegisterCommands(commandsList)

	err := command.Execute()
	if err != nil {
		log.Error(err.Error())
	}
}
