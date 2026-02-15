package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addToolCmd = &cobra.Command{
	Use:   "tool",
	Short: "Add a tool to your agent",
	Long:  `Add a tool to your research agent. Tools extend what your agent can do.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Stub implementation
		fmt.Println("Adding tool to agent...")
		fmt.Println("Configure your tool in the agent project and run 'agent run' to use it.")
		return nil
	},
}

func init() {
	addCmd.AddCommand(addToolCmd)
}
