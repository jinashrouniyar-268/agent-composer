package acl

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadYAML reads and parses an ACL YAML file into a generic structure.
func LoadYAML(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// SaveYAML writes the ACL structure to a file.
func SaveYAML(path string, m map[string]interface{}) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// getResearchToolsConfig returns the tools_config list under nodes.research.config.
// It returns (list, nil) or (nil, err). The list may be nil if not found.
func getResearchToolsConfig(m map[string]interface{}) ([]interface{}, error) {
	nodes, ok := m["nodes"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("acl: missing or invalid nodes")
	}
	research, ok := nodes["research"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("acl: missing or invalid nodes.research")
	}
	config, ok := research["config"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("acl: missing or invalid nodes.research.config")
	}
	toolsConfig, _ := config["tools_config"].([]interface{})
	return toolsConfig, nil
}

// setResearchToolsConfig sets nodes.research.config.tools_config to the given list.
func setResearchToolsConfig(m map[string]interface{}, list []interface{}) error {
	nodes, ok := m["nodes"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("acl: missing nodes")
	}
	research, ok := nodes["research"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("acl: missing nodes.research")
	}
	config, ok := research["config"].(map[string]interface{})
	if !ok {
		config = make(map[string]interface{})
		research["config"] = config
	}
	config["tools_config"] = list
	return nil
}

// getResearchGuidelinesPrompt returns research_guidelines_prompt under nodes.research.config.agent_config.
func getResearchGuidelinesPrompt(m map[string]interface{}) (string, error) {
	nodes, ok := m["nodes"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("acl: missing nodes")
	}
	research, ok := nodes["research"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("acl: missing nodes.research")
	}
	config, ok := research["config"].(map[string]interface{})
	if !ok {
		return "", nil
	}
	agentConfig, ok := config["agent_config"].(map[string]interface{})
	if !ok {
		return "", nil
	}
	s, _ := agentConfig["research_guidelines_prompt"].(string)
	return s, nil
}

// setResearchGuidelinesPrompt sets research_guidelines_prompt under nodes.research.config.agent_config.
func setResearchGuidelinesPrompt(m map[string]interface{}, prompt string) error {
	nodes, ok := m["nodes"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("acl: missing nodes")
	}
	research, ok := nodes["research"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("acl: missing nodes.research")
	}
	config, ok := research["config"].(map[string]interface{})
	if !ok {
		config = make(map[string]interface{})
		research["config"] = config
	}
	agentConfig, ok := config["agent_config"].(map[string]interface{})
	if !ok {
		agentConfig = make(map[string]interface{})
		config["agent_config"] = agentConfig
	}
	agentConfig["research_guidelines_prompt"] = prompt
	return nil
}

// AddToolToYAML appends a tool entry (map or structure that serializes to YAML) to tools_config and saves.
func AddToolToYAML(path string, tool interface{}) error {
	m, err := LoadYAML(path)
	if err != nil {
		return err
	}
	list, err := getResearchToolsConfig(m)
	if err != nil {
		return err
	}
	if list == nil {
		list = []interface{}{}
	}
	list = append(list, tool)
	if err := setResearchToolsConfig(m, list); err != nil {
		return err
	}
	return SaveYAML(path, m)
}

// UpdateResearchGuidelinesPrompt sets the research_guidelines_prompt in the YAML and saves.
func UpdateResearchGuidelinesPrompt(path string, prompt string) error {
	m, err := LoadYAML(path)
	if err != nil {
		return err
	}
	if err := setResearchGuidelinesPrompt(m, prompt); err != nil {
		return err
	}
	return SaveYAML(path, m)
}

// GetResearchGuidelinesPrompt reads the current research_guidelines_prompt from the YAML file.
func GetResearchGuidelinesPrompt(path string) (string, error) {
	m, err := LoadYAML(path)
	if err != nil {
		return "", err
	}
	return getResearchGuidelinesPrompt(m)
}
