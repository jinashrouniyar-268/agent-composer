package cli

import (
	"github.com/spf13/cobra"
)

// version is set at build time via ldflags (e.g. -X .../cli.version=0.1.2); default for dev builds.
var version = "0.0.1"

var (
	flagList          bool
	flagUnstructured  bool
	flagDatastoreName string
	flagVerbose       bool
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "Agent Composer CLI - Create research agents from the terminal",
	Long: `Agent Composer lets you create research agents directly from your terminal
using the Contextual AI platform.

Commands:
  init <agent-name>              Initialize a new agent
  tools --list                  List available tool types
  <agent-name> ingest <file>    Ingest a document (-U for unstructured)
  <agent-name> add <tool-type>   Add a tool (web-search, unstructured-search, structured-search)
  <agent-name> run "<query>"     Run the agent with a query (--verbose for trace)
  <agent-name> delete            Delete the agent (with confirmation)`,
	SilenceUsage: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return nil
		}
		return nil
	},
	RunE: rootRunE,
}

func init() {
	rootCmd.Version = version
	rootCmd.Flags().BoolVarP(&flagList, "list", "l", false, "List available tool types (use with: agent tools --list)")
	rootCmd.Flags().BoolVarP(&flagUnstructured, "unstructured", "U", false, "Ingest as unstructured document")
	rootCmd.Flags().StringVar(&flagDatastoreName, "datastore-name", "", "Datastore name for ingest (default: agent default)")
	rootCmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Show workflow/tool/thinking in run output")
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// rootRunE dispatches: init, tools, or <agent-name> ingest|add|run
func rootRunE(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	first := args[0]
	switch first {
	case "init":
		if len(args) < 2 {
			return cmd.Help()
		}
		return runInit(cmd, args[1:])
	case "tools":
		return runTools(cmd, args[1:])
	default:
		agentName := first
		if len(args) < 2 {
			return cmd.Help()
		}
		sub := args[1]
		switch sub {
		case "ingest":
			return runIngest(agentName, args[2:])
		case "add":
			return runAdd(agentName, args[2:])
		case "run":
			return runRun(agentName, args[2:])
		case "delete":
			return runDelete(agentName, args[2:])
		default:
			return cmd.Help()
		}
	}
}
