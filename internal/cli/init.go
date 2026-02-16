package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/jinashrouniyar-268/agent-composer/internal/acl"
	"github.com/jinashrouniyar-268/agent-composer/internal/api"
	"github.com/jinashrouniyar-268/agent-composer/internal/config"
	"github.com/spf13/cobra"
)

const apiKeysURL = "https://app.contextual.ai"

func runInit(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("init requires agent name: agent init <agent-name>")
	}
	agentName := strings.TrimSpace(args[0])
	if err := validateAgentName(agentName); err != nil {
		return err
	}

	creds, err := config.LoadCredentials()
	if err != nil {
		return fmt.Errorf("load credentials: %w", err)
	}
	if creds == nil || creds.APIKey == "" {
		if err := promptAndSaveAPIKey(); err != nil {
			return err
		}
	}

	quick, err := PromptSetupMode()
	if err != nil {
		return err
	}
	if !quick {
		return createAgentAndPersist(agentName)
	}
	return runQuickSetup(agentName)
}

// runQuickSetup runs the quick setup flow: tool selection, optional reference docs, then create datastore → ingest → build ACL → create agent → persist.
func runQuickSetup(agentName string) error {
	creds, err := config.LoadCredentials()
	if err != nil || creds == nil || creds.APIKey == "" {
		return fmt.Errorf("no API key: run init again and paste your key")
	}
	client := api.NewClient(creds.APIKey)

	selected, err := PromptToolSelection()
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return fmt.Errorf("at least one tool is required")
	}

	var docPaths []string
	needDocs := false
	for _, t := range selected {
		if t == acl.ToolUnstructuredSearch || t == acl.ToolStructuredSearch {
			needDocs = true
			break
		}
	}
	if needDocs {
		docPaths, err = PromptReferenceDocuments()
		if err != nil {
			return err
		}
	}

	datastoreName := agentName + "-default"
	fmt.Printf("Creating datastore: %s…\n", datastoreName)
	datastoreID, err := client.CreateDatastore(datastoreName)
	if err != nil {
		return fmt.Errorf("create datastore: %w", err)
	}
	fmt.Printf("Datastore created: %s\n", datastoreID)

	for _, path := range docPaths {
		fmt.Printf("Ingesting %s…\n", path)
		_, err := client.IngestDocument(datastoreID, path)
		if err != nil {
			return fmt.Errorf("ingest %s: %w", path, err)
		}
		fmt.Printf("  Ingested.\n")
	}

	aclYAML, err := acl.BuildACLWithTools(selected)
	if err != nil {
		return fmt.Errorf("build ACL: %w", err)
	}

	fmt.Printf("Creating agent: %s…\n", agentName)
	out, err := client.CreateAgent(agentName, "Agent created by agent-composer (quick setup)", []string{datastoreID}, aclYAML)
	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}
	fmt.Printf("Agent created: %s\n", out.ID)

	yamlPath, err := config.AgentYAMLPath(agentName)
	if err != nil {
		return err
	}
	if err := os.WriteFile(yamlPath, []byte(aclYAML), 0644); err != nil {
		return fmt.Errorf("write YAML: %w", err)
	}
	fmt.Printf("Wrote %s\n", yamlPath)

	entry := config.AgentEntry{
		AgentID:              out.ID,
		DatastoreID:          datastoreID,
		DefaultDatastoreName: datastoreName,
		YAMLPath:             agentName + ".yaml",
		LastSyncedAt:         time.Now(),
		LocalYAMLHash:        sha256Hash(aclYAML),
	}
	if err := config.SetAgent(agentName, entry); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("Agent %q is ready. Use `agent %s run \"<query>\"` to run.\n", agentName, agentName)
	return nil
}

// createAgentAndPersist creates datastore, agent with minimal ACL, writes YAML and configs.
func createAgentAndPersist(agentName string) error {
	creds, err := config.LoadCredentials()
	if err != nil || creds == nil || creds.APIKey == "" {
		return fmt.Errorf("no API key: run init again and paste your key")
	}
	client := api.NewClient(creds.APIKey)

	datastoreName := agentName + "-default"
	fmt.Printf("Creating datastore: %s…\n", datastoreName)
	datastoreID, err := client.CreateDatastore(datastoreName)
	if err != nil {
		return fmt.Errorf("create datastore: %w", err)
	}
	fmt.Printf("Datastore created: %s\n", datastoreID)

	fmt.Printf("Creating agent: %s…\n", agentName)
	out, err := client.CreateAgent(agentName, "Agent created by agent-composer", []string{datastoreID}, acl.MinimalACLYAML)
	if err != nil {
		return fmt.Errorf("create agent: %w", err)
	}
	fmt.Printf("Agent created: %s\n", out.ID)

	yamlPath, err := config.AgentYAMLPath(agentName)
	if err != nil {
		return err
	}
	if err := os.WriteFile(yamlPath, []byte(acl.MinimalACLYAML), 0644); err != nil {
		return fmt.Errorf("write YAML: %w", err)
	}
	fmt.Printf("Wrote %s\n", yamlPath)

	entry := config.AgentEntry{
		AgentID:              out.ID,
		DatastoreID:          datastoreID,
		DefaultDatastoreName: datastoreName,
		YAMLPath:             agentName + ".yaml",
		LastSyncedAt:         time.Now(),
		LocalYAMLHash:        sha256Hash(acl.MinimalACLYAML),
	}
	if err := config.SetAgent(agentName, entry); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	fmt.Printf("Agent %q is ready. Use `agent %s ingest <file>` to add documents, `agent %s add <tool>` to add tools, and `agent %s run \"<query>\"` to run.\n", agentName, agentName, agentName, agentName)
	return nil
}

func promptAndSaveAPIKey() error {
	fmt.Println("Checking Contextual AI setup…")
	fmt.Println("Press Enter to open the Contextual AI API keys page in your browser.")
	reader := bufio.NewReader(os.Stdin)
	if _, err := reader.ReadString('\n'); err != nil {
		return fmt.Errorf("read input: %w", err)
	}
	if err := openBrowser(apiKeysURL); err != nil {
		// Non-fatal: user can still paste key
		fmt.Printf("Could not open browser: %v\n", err)
	}
	fmt.Print("Paste your API key and press Enter: ")
	key, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read API key: %w", err)
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	if err := config.SaveCredentials(key); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}
	fmt.Println("API key saved.")
	return nil
}

func openBrowser(url string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", url)
	case "linux":
		c = exec.Command("xdg-open", url)
	case "windows":
		c = exec.Command("cmd", "/c", "start", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return c.Start()
}
