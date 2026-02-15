package cli

import (
	"github.com/spf13/cobra"
)

const version = "0.0.1"

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent Composer CLI - Create research agents from the terminal",
	Long: `Agent Composer lets you create research agents directly from your terminal
using the Contextual AI platform.

Commands:
  init        Initialize a new agent project
  add         Add components (e.g. tools) to your agent
  run         Run your agent`,
	SilenceUsage: true,
}

func init() {
	rootCmd.Version = version
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
