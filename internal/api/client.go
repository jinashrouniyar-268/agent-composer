package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const BaseURL = "https://api.contextual.ai/v1"

// Client is the Contextual AI API client.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewClient returns a client with the given API key.
func NewClient(apiKey string) *Client {
	return &Client{
		baseURL: BaseURL,
		apiKey:  apiKey,
		http:    &http.Client{},
	}
}

// CreateDatastore creates a datastore and returns its ID.
func (c *Client) CreateDatastore(name string) (string, error) {
	body := map[string]string{"name": name}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/datastores", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	c.setHeaders(req, "application/json")
	req.ContentLength = int64(len(reqBody))

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create datastore: %s: %s", resp.Status, string(msg))
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// CreateAgentRequest is the payload for POST /agents.
type CreateAgentRequest struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	DatastoreIDs []string      `json:"datastore_ids"`
	AgentConfigs *AgentConfigs `json:"agent_configs,omitempty"`
}

// AgentConfigs holds acl_config for create/update.
type AgentConfigs struct {
	ACLConfig *ACLConfig `json:"acl_config,omitempty"`
}

// ACLConfig holds acl_active and acl_yaml.
type ACLConfig struct {
	ACLActive bool   `json:"acl_active"`
	ACLYAML   string `json:"acl_yaml"`
}

// CreateAgentOutput is the response from POST /agents.
type CreateAgentOutput struct {
	ID           string   `json:"id"`
	DatastoreIDs []string `json:"datastore_ids"`
}

// CreateAgent creates an agent and returns its ID and datastore IDs.
func (c *Client) CreateAgent(name, description string, datastoreIDs []string, aclYAML string) (*CreateAgentOutput, error) {
	body := CreateAgentRequest{
		Name:         name,
		Description:  description,
		DatastoreIDs: datastoreIDs,
		AgentConfigs: &AgentConfigs{
			ACLConfig: &ACLConfig{ACLActive: true, ACLYAML: aclYAML},
		},
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/agents", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req, "application/json")
	req.ContentLength = int64(len(reqBody))

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("create agent: %s: %s", resp.Status, string(msg))
	}
	var out CreateAgentOutput
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ModifyAgentRequest is the payload for PUT /agents/{id} (partial).
type ModifyAgentRequest struct {
	AgentConfigs *AgentConfigs `json:"agent_configs,omitempty"`
}

// ModifyAgent updates an agent (e.g. acl_config only).
func (c *Client) ModifyAgent(agentID string, aclActive bool, aclYAML string) error {
	body := ModifyAgentRequest{
		AgentConfigs: &AgentConfigs{
			ACLConfig: &ACLConfig{ACLActive: aclActive, ACLYAML: aclYAML},
		},
	}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, c.baseURL+"/agents/"+agentID, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	c.setHeaders(req, "application/json")
	req.ContentLength = int64(len(reqBody))

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("modify agent: %s: %s", resp.Status, string(msg))
	}
	return nil
}

// GetAgentMetadataResponse is the response from GET /agents/{id}/metadata.
type GetAgentMetadataResponse struct {
	Name         string        `json:"name"`
	DatastoreIDs []string      `json:"datastore_ids"`
	AgentConfigs *AgentConfigs `json:"agent_configs,omitempty"`
}

// DeleteAgent deletes an agent. Expects 2xx (e.g. 200 or 204).
func (c *Client) DeleteAgent(agentID string) error {
	req, err := http.NewRequest(http.MethodDelete, c.baseURL+"/agents/"+agentID, nil)
	if err != nil {
		return err
	}
	c.setHeaders(req, "")
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete agent: %s: %s", resp.Status, string(msg))
	}
	return nil
}

// GetAgentMetadata returns agent metadata including acl_yaml if present.
func (c *Client) GetAgentMetadata(agentID string) (*GetAgentMetadataResponse, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/agents/"+agentID+"/metadata", nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req, "")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get agent metadata: %s: %s", resp.Status, string(msg))
	}
	var out GetAgentMetadataResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

// IngestDocument uploads a file to a datastore (unstructured). Returns document ID.
// Caller may poll GET /datastores/{id}/documents/{doc_id}/metadata for status.
func (c *Client) IngestDocument(datastoreID, filePath string) (documentID string, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/datastores/"+datastoreID+"/documents", &buf)
	if err != nil {
		return "", err
	}
	c.setHeaders(req, w.FormDataContentType())
	req.ContentLength = int64(buf.Len())

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		msg, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ingest document: %s: %s", resp.Status, string(msg))
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}

// QueryACLRequest is the body for POST /agents/{id}/query/acl.
type QueryACLRequest struct {
	Messages []QueryMessage `json:"messages"`
	Stream   bool           `json:"stream"`
}

// QueryMessage is one message in the query.
type QueryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// QueryACLStream POSTs to query/acl with stream=true. Caller must close resp.Body.
func (c *Client) QueryACLStream(agentID string, messages []QueryMessage) (*http.Response, error) {
	body := QueryACLRequest{Messages: messages, Stream: true}
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/agents/"+agentID+"/query/acl", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	c.setHeaders(req, "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.ContentLength = int64(len(reqBody))

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("query acl: %s: %s", resp.Status, string(msg))
	}
	return resp, nil
}

func (c *Client) setHeaders(req *http.Request, contentType string) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
}
