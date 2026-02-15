package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run your agent",
	Long:  `Run your research agent. Ensure you have run 'agent init' and added tools first.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Stub implementation
		fmt.Println("Running agent...")
		fmt.Println("Connect to the Contextual AI platform to run your research agent.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
