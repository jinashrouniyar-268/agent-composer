package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	ConfigDirName   = "agent-composer"
	CredentialsFile = "credentials.json"
	ConfigsFile     = "configs.json"
)

// Credentials holds the API key (do not log or commit).
type Credentials struct {
	APIKey string `json:"api_key"`
}

// AgentEntry is one agent in the registry.
type AgentEntry struct {
	AgentID              string    `json:"agent_id"`
	DatastoreID          string    `json:"datastore_id"`
	DefaultDatastoreName string    `json:"default_datastore_name"`
	YAMLPath             string    `json:"yaml_path"`
	LastSyncedAt         time.Time `json:"last_synced_at"`
	LocalYAMLHash        string    `json:"local_yaml_hash,omitempty"` // optional: skip GET when local hash unchanged
}

// Configs is the agent registry (configs.json).
type Configs struct {
	Agents map[string]AgentEntry `json:"agents"`
}

// ConfigDir returns the config directory: ~/.config/agent-composer (or platform equivalent).
func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigDirName), nil
}

// EnsureConfigDir creates the config directory if it does not exist.
func EnsureConfigDir() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// CredentialsPath returns the path to credentials.json.
func CredentialsPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, CredentialsFile), nil
}

// ConfigsPath returns the path to configs.json.
func ConfigsPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigsFile), nil
}

// LoadCredentials reads credentials.json. Returns nil if file does not exist or has no api_key.
func LoadCredentials() (*Credentials, error) {
	p, err := CredentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var c Credentials
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	if c.APIKey == "" {
		return nil, nil
	}
	return &c, nil
}

// SaveCredentials writes credentials.json. Creates config dir if needed.
func SaveCredentials(apiKey string) error {
	dir, err := EnsureConfigDir()
	if err != nil {
		return err
	}
	p := filepath.Join(dir, CredentialsFile)
	c := Credentials{APIKey: apiKey}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0600)
}

// LoadConfigs reads configs.json. Returns empty Configs if file does not exist.
func LoadConfigs() (*Configs, error) {
	p, err := ConfigsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return &Configs{Agents: make(map[string]AgentEntry)}, nil
		}
		return nil, err
	}
	var cfg Configs
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentEntry)
	}
	return &cfg, nil
}

// SaveConfigs writes configs.json.
func SaveConfigs(cfg *Configs) error {
	_, err := EnsureConfigDir()
	if err != nil {
		return err
	}
	p, err := ConfigsPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0644)
}

// GetAgent returns the agent entry for the given name, or nil if not found.
func GetAgent(name string) (*AgentEntry, error) {
	cfg, err := LoadConfigs()
	if err != nil {
		return nil, err
	}
	e, ok := cfg.Agents[name]
	if !ok {
		return nil, nil
	}
	return &e, nil
}

// SetAgent updates or adds an agent entry and saves configs.json.
func SetAgent(name string, entry AgentEntry) error {
	cfg, err := LoadConfigs()
	if err != nil {
		return err
	}
	if cfg.Agents == nil {
		cfg.Agents = make(map[string]AgentEntry)
	}
	cfg.Agents[name] = entry
	return SaveConfigs(cfg)
}

// RemoveAgent removes an agent from the registry and saves configs.json.
// Caller should ensure the agent exists (e.g. via GetAgent) before calling.
func RemoveAgent(name string) error {
	cfg, err := LoadConfigs()
	if err != nil {
		return err
	}
	if cfg.Agents != nil {
		delete(cfg.Agents, name)
	}
	return SaveConfigs(cfg)
}

// AgentYAMLPath returns the absolute path to the agent's YAML file (relative to config dir).
func AgentYAMLPath(agentName string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, agentName+".yaml"), nil
}
