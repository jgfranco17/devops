package doc

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GenerateMarkdown generates markdown documentation for the CLI
func GenerateMarkdown(rootCmd *cobra.Command) (string, error) {
	var docs strings.Builder

	// Write header
	docs.WriteString(fmt.Sprintf("# %s CLI Documentation\n\n", rootCmd.Name()))
	docs.WriteString(fmt.Sprintf("**Version:** %s\n\n", rootCmd.Version))
	docs.WriteString(fmt.Sprintf("**Description:** %s\n\n", rootCmd.Short))

	// Write usage
	docs.WriteString("## Usage\n\n")
	docs.WriteString(fmt.Sprintf("```bash\n%s [command] [flags] [arguments]\n```\n\n", rootCmd.Name()))

	// Write global flags
	if rootCmd.PersistentFlags().HasFlags() {
		docs.WriteString("## Global Flags\n\n")
		docs.WriteString("| Flag | Short | Type | Description |\n")
		docs.WriteString("|------|-------|------|-------------|\n")

		rootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
			short := ""
			if flag.Shorthand != "" {
				short = "-" + flag.Shorthand
			}
			docs.WriteString(fmt.Sprintf("| --%s | %s | %s | %s |\n",
				flag.Name, short, flag.Value.Type(), flag.Usage))
		})
		docs.WriteString("\n")
	}

	// Write commands
	docs.WriteString("## Commands\n\n")
	writeCommandsToDocs(&docs, rootCmd, 0)

	return docs.String(), nil
}

// writeCommandsToDocs recursively writes command documentation
func writeCommandsToDocs(docs *strings.Builder, cmd *cobra.Command, level int) {
	// Skip root command itself
	if cmd != cmd.Root() {
		indent := strings.Repeat("  ", level)

		// Write command header
		docs.WriteString(fmt.Sprintf("%s### %s\n\n", indent, cmd.Name()))

		// Write description
		if cmd.Short != "" {
			docs.WriteString(fmt.Sprintf("%s**Description:** %s\n\n", indent, cmd.Short))
		}

		// Write long description
		if cmd.Long != "" && cmd.Long != cmd.Short {
			docs.WriteString(fmt.Sprintf("%s%s\n\n", indent, cmd.Long))
		}

		// Write usage
		docs.WriteString(fmt.Sprintf("%s**Usage:**\n", indent))
		docs.WriteString(fmt.Sprintf("%s```bash\n", indent))
		docs.WriteString(fmt.Sprintf("%s%s\n", indent, cmd.UseLine()))
		docs.WriteString(fmt.Sprintf("%s```\n\n", indent))

		// Write flags if any
		if cmd.Flags().HasFlags() {
			docs.WriteString(fmt.Sprintf("%s**Flags:**\n\n", indent))
			docs.WriteString(fmt.Sprintf("%s| Flag | Short | Type | Description |\n", indent))
			docs.WriteString(fmt.Sprintf("%s|------|-------|------|-------------|\n", indent))

			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				short := ""
				if flag.Shorthand != "" {
					short = "-" + flag.Shorthand
				}
				docs.WriteString(fmt.Sprintf("%s| --%s | %s | %s | %s |\n",
					indent, flag.Name, short, flag.Value.Type(), flag.Usage))
			})
			docs.WriteString("\n")
		}

		docs.WriteString("\n")
	}

	// Write subcommands
	for _, subCmd := range cmd.Commands() {
		// Skip hidden commands unless specifically requested
		if subCmd.Hidden {
			continue
		}
		writeCommandsToDocs(docs, subCmd, level+1)
	}
}
