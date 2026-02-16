package cli

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/jinashrouniyar-268/agent-composer/internal/api"
	"github.com/jinashrouniyar-268/agent-composer/internal/config"
)

func runDelete(agentName string, _ []string) error {
	if err := validateAgentName(agentName); err != nil {
		return err
	}
	entry, err := config.GetAgent(agentName)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("agent %q not found", agentName)
	}
	creds, err := config.LoadCredentials()
	if err != nil || creds == nil || creds.APIKey == "" {
		return fmt.Errorf("no API key; run 'agent init <name>' and paste your key")
	}
	var confirmed bool
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Are you sure you want to delete agent %q? This cannot be undone.", agentName),
		Default: false,
	}
	if err := survey.AskOne(prompt, &confirmed); err != nil {
		return err
	}
	if !confirmed {
		return nil
	}
	client := api.NewClient(creds.APIKey)
	if err := client.DeleteAgent(entry.AgentID); err != nil {
		return fmt.Errorf("delete agent: %w", err)
	}
	if err := config.RemoveAgent(agentName); err != nil {
		return fmt.Errorf("remove from config: %w", err)
	}
	yamlPath, err := config.AgentYAMLPath(agentName)
	if err == nil {
		_ = os.Remove(yamlPath)
	}
	fmt.Printf("Agent %q deleted successfully.\n", agentName)
	return nil
}
