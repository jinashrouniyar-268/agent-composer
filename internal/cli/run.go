package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jinashrouniyar-268/agent-composer/internal/api"
	"github.com/jinashrouniyar-268/agent-composer/internal/config"
	"github.com/jinashrouniyar-268/agent-composer/internal/stream"
)

func runRun(agentName string, args []string) error {
	if err := validateAgentName(agentName); err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("run requires a query: agent %s run \"<query>\"", agentName)
	}
	query := strings.TrimSpace(strings.Join(args, " "))
	if query == "" {
		return fmt.Errorf("query cannot be empty")
	}

	entry, err := config.GetAgent(agentName)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("agent %q not found; run 'agent init %s' first", agentName, agentName)
	}

	creds, err := config.LoadCredentials()
	if err != nil || creds == nil || creds.APIKey == "" {
		return fmt.Errorf("no API key; run 'agent init <name>' and paste your key")
	}
	client := api.NewClient(creds.APIKey)

	yamlPath, err := config.AgentYAMLPath(agentName)
	if err != nil {
		return err
	}
	localYAML, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("read local YAML: %w", err)
	}
	localStr := strings.TrimSpace(string(localYAML))
	localHash := sha256Hash(localStr)

	// Optional cache: if local file unchanged since last sync, skip GET/PUT
	if entry.LocalYAMLHash != "" && entry.LocalYAMLHash == localHash {
		// Assume in sync
	} else {
		meta, err := client.GetAgentMetadata(entry.AgentID)
		if err != nil {
			return fmt.Errorf("get agent metadata: %w", err)
		}
		cloudStr := ""
		if meta.AgentConfigs != nil && meta.AgentConfigs.ACLConfig != nil {
			cloudStr = strings.TrimSpace(meta.AgentConfigs.ACLConfig.ACLYAML)
		}
		if cloudStr != localStr {
			if err := client.ModifyAgent(entry.AgentID, true, localStr); err != nil {
				return fmt.Errorf("sync YAML to cloud: %w", err)
			}
			entry.LastSyncedAt = time.Now()
			entry.LocalYAMLHash = localHash
			_ = config.SetAgent(agentName, *entry)
		} else {
			entry.LocalYAMLHash = localHash
			_ = config.SetAgent(agentName, *entry)
		}
	}

	resp, err := client.QueryACLStream(entry.AgentID, []api.QueryMessage{{Role: "user", Content: query}})
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}
	defer resp.Body.Close()

	onDelta := func(delta string) {
		fmt.Print(delta)
	}
	onComplete := func() {
		fmt.Println()
	}
	if err := stream.StreamQueryACL(resp.Body, flagVerbose, onDelta, onComplete); err != nil {
		return err
	}
	return nil
}
