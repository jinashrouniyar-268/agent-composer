package cli

import (
	"fmt"
	"path/filepath"

	"github.com/jinashrouniyar-268/agent-composer/internal/api"
	"github.com/jinashrouniyar-268/agent-composer/internal/config"
)

// AllowedUnstructuredExt is used for validating ingest file extensions (ingest and quick-setup prompts).
var AllowedUnstructuredExt = map[string]bool{
	".pdf": true, ".html": true, ".htm": true, ".mhtml": true,
	".doc": true, ".docx": true, ".ppt": true, ".pptx": true,
	".png": true, ".jpg": true, ".jpeg": true,
}

func runIngest(agentName string, args []string) error {
	if err := validateAgentName(agentName); err != nil {
		return err
	}
	if !flagUnstructured {
		return fmt.Errorf("use -U for unstructured document ingest (structured -S deferred)")
	}
	if len(args) < 1 {
		return fmt.Errorf("ingest requires file path: agent %s ingest <filepath> -U", agentName)
	}
	filePath := args[0]
	ext := filepath.Ext(filePath)
	if !AllowedUnstructuredExt[ext] {
		return fmt.Errorf("unsupported extension %s for unstructured ingest; allowed: .pdf, .html, .htm, .mhtml, .doc, .docx, .ppt, .pptx, .png, .jpg, .jpeg", ext)
	}

	entry, err := config.GetAgent(agentName)
	if err != nil {
		return err
	}
	if entry == nil {
		return fmt.Errorf("agent %q not found; run 'agent init %s' first", agentName, agentName)
	}
	datastoreID := entry.DatastoreID
	if flagDatastoreName != "" {
		// For now we only have one datastore per agent; could look up by name later
		datastoreID = entry.DatastoreID
	}

	creds, err := config.LoadCredentials()
	if err != nil || creds == nil || creds.APIKey == "" {
		return fmt.Errorf("no API key; run 'agent init <name>' and paste your key")
	}
	client := api.NewClient(creds.APIKey)
	docID, err := client.IngestDocument(datastoreID, filePath)
	if err != nil {
		return fmt.Errorf("ingest: %w", err)
	}
	fmt.Printf("Ingestion started. Document ID: %s\n", docID)
	return nil
}
