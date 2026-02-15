package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new agent project",
	Long:  `Initialize a new research agent project in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Stub: create local project/config
		fmt.Println("Initializing agent project...")
		fmt.Println("Run 'agent add tool' to add tools, then 'agent run' to run your agent.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
