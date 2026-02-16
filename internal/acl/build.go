package acl

import (
	"gopkg.in/yaml.v3"
)

// BuildACLWithTools builds a full ACL YAML string from the minimal template (no tools)
// by adding the selected tools. selectedTools can contain ToolWebSearch, ToolUnstructuredSearch, ToolStructuredSearch.
func BuildACLWithTools(selectedTools []string) (string, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal([]byte(MinimalACLYAML), &m); err != nil {
		return "", err
	}
	list, err := getResearchToolsConfig(m)
	if err != nil {
		return "", err
	}
	if list == nil {
		list = []interface{}{}
	}
	for _, t := range selectedTools {
		switch t {
		case ToolUnstructuredSearch:
			list = append(list, UnstructuredSearchTool())
		case ToolWebSearch:
			list = append(list, WebSearchTool())
		case ToolStructuredSearch:
			list = append(list, getSchemaTool())
			list = append(list, executeSQLQueryTool())
			if err := setResearchGuidelinesPrompt(m, StructuredResearchGuidelinesPrompt); err != nil {
				return "", err
			}
		}
	}
	if err := setResearchToolsConfig(m, list); err != nil {
		return "", err
	}
	data, err := yaml.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
