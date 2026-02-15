package cli

import (
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add components to your agent",
	Long:  `Add components such as tools to your research agent.`,
}

func init() {
	rootCmd.AddCommand(addCmd)
}
