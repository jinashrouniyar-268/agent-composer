package cli

import (
	"fmt"

	"github.com/jinashrouniyar-268/agent-composer/internal/acl"
	"github.com/spf13/cobra"
)

func runTools(cmd *cobra.Command, args []string) error {
	if !flagList {
		fmt.Fprintln(cmd.OutOrStdout(), "Use --list to see available tool types: agent tools --list")
		return nil
	}
	catalog := acl.ToolCatalog()
	fmt.Fprintln(cmd.OutOrStdout(), "Tool types:")
	for name, desc := range catalog {
		fmt.Fprintf(cmd.OutOrStdout(), "  %-22s %s\n", name, desc)
	}
	return nil
}
