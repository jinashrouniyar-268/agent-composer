package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jinashrouniyar-268/agent-composer/internal/acl"
	"github.com/jinashrouniyar-268/agent-composer/internal/api"
	"github.com/jinashrouniyar-268/agent-composer/internal/config"
)

func runAdd(agentName string, args []string) error {
	if err := validateAgentName(agentName); err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("add requires tool type: agent %s add <web-search|unstructured-search|structured-search>", agentName)
	}
	toolType := strings.TrimSpace(strings.ToLower(args[0]))
	catalog := acl.ToolCatalog()
	if _, ok := catalog[toolType]; !ok {
		return fmt.Errorf("unknown tool type %q; use 'agent tools --list' to see options", toolType)
	}

	entry, err := config.GetAgent(agentName)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("agent %q not found; run 'agent init %s' first", agentName, agentName)
	}

	yamlPath, err := config.AgentYAMLPath(agentName)
	if err != nil {
		return err
	}

	switch toolType {
	case acl.ToolWebSearch:
		if err := acl.AddWebSearch(yamlPath); err != nil {
			return fmt.Errorf("add web-search: %w", err)
		}
	case acl.ToolUnstructuredSearch:
		if err := acl.AddUnstructuredSearch(yamlPath); err != nil {
			return fmt.Errorf("add unstructured-search: %w", err)
		}
	case acl.ToolStructuredSearch:
		if err := acl.AddStructuredSearch(yamlPath); err != nil {
			return fmt.Errorf("add structured-search: %w", err)
		}
	default:
		return fmt.Errorf("unsupported tool type: %s", toolType)
	}

	yamlBytes, err := readYAMLBytes(yamlPath)
	if err != nil {
		return err
	}

	creds, err := config.LoadCredentials()
	if err != nil || creds == nil || creds.APIKey == "" {
		return fmt.Errorf("no API key")
	}
	client := api.NewClient(creds.APIKey)
	if err := client.ModifyAgent(entry.AgentID, true, string(yamlBytes)); err != nil {
		return fmt.Errorf("sync to cloud: %w", err)
	}
	entry.LastSyncedAt = time.Now()
	entry.LocalYAMLHash = sha256Hash(string(yamlBytes))
	if err := config.SetAgent(agentName, *entry); err != nil {
		return err
	}
	fmt.Printf("Added %s and synced to cloud.\n", toolType)
	return nil
}

func readYAMLBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}
