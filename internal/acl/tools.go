package acl

// Tool types that can be added via `agent <name> add <tool-type>`.
const (
	ToolWebSearch          = "web-search"
	ToolUnstructuredSearch = "unstructured-search"
	ToolStructuredSearch   = "structured-search"
)

// ToolCatalog returns a short description for each tool type (for `agent tools --list`).
func ToolCatalog() map[string]string {
	return map[string]string{
		ToolWebSearch:          "Search the web for current information.",
		ToolUnstructuredSearch: "Search uploaded documents (vector + lexical).",
		ToolStructuredSearch:   "Query structured datastores (get_schema + execute_sql_query).",
	}
}

// UnstructuredSearchTool returns the search_docs tool entry (SearchUnstructuredDataStep).
func UnstructuredSearchTool() map[string]interface{} {
	return map[string]interface{}{
		"name": "search_docs",
		"description": "Search the datastore containing user-uploaded documents. Use for relevant chunks from uploaded documents.\n",
		"step_config": map[string]interface{}{
			"type": "SearchUnstructuredDataStep",
			"config": map[string]interface{}{
				"top_k": 50,
				"lexical_alpha": 0.1,
				"semantic_alpha": 0.9,
				"reranker": "ctxl-rerank-v2-instruct-multilingual-FP8",
				"rerank_top_k": 12,
				"reranker_score_filter_threshold": 0.2,
			},
		},
	}
}

// WebSearchTool returns the tool entry for web_search (step_config type WebSearchStep).
func WebSearchTool() map[string]interface{} {
	return map[string]interface{}{
		"name": "web_search",
		"description": "Search the web for current information. Use for live data, recent events, and facts not in uploaded documents.\n",
		"step_config": map[string]interface{}{
			"type": "WebSearchStep",
		},
	}
}

// getSchemaTool returns the get_schema tool (graph_config with GetStructuredDatastoreSchemaStep).
func getSchemaTool() map[string]interface{} {
	return map[string]interface{}{
		"name": "get_schema",
		"description": `Get schema information from structured datastores (tables, databases). Use this first before writing SQL.
`,
		"graph_config": map[string]interface{}{
			"version": "0.1",
			"inputs":  map[string]interface{}{},
			"outputs": map[string]interface{}{
				"schemas": "Dict[str, Dict[str, Any]]",
			},
			"nodes": map[string]interface{}{
				"get_schema": map[string]interface{}{
					"type":         "GetStructuredDatastoreSchemaStep",
					"input_mapping": map[string]interface{}{},
				},
				"__outputs__": map[string]interface{}{
					"type": "output",
					"input_mapping": map[string]interface{}{
						"schemas": "get_schema#schemas",
					},
				},
			},
		},
	}
}

// executeSQLQueryTool returns the execute_sql_query tool (graph_config with QueryStructuredDatastoreStep).
func executeSQLQueryTool() map[string]interface{} {
	return map[string]interface{}{
		"name": "execute_sql_query",
		"description": `Execute SQL queries against structured datastores. Provide complete SQL query strings.
Examples: "SELECT * FROM t LIMIT 10", "SELECT COUNT(*) FROM t"
`,
		"graph_config": map[string]interface{}{
			"version": "0.1",
			"inputs": map[string]interface{}{
				"sql_query": "str",
			},
			"outputs": map[string]interface{}{
				"retrievals": "Retrievals",
			},
			"nodes": map[string]interface{}{
				"execute_sql_query": map[string]interface{}{
					"type": "QueryStructuredDatastoreStep",
					"input_mapping": map[string]interface{}{
						"sql_query": "__inputs__#sql_query",
					},
				},
				"__outputs__": map[string]interface{}{
					"type":       "output",
					"ui_output": "retrievals",
					"input_mapping": map[string]interface{}{
						"retrievals": "execute_sql_query#retrievals",
					},
				},
			},
		},
	}
}

// StructuredResearchGuidelinesPrompt is the research_guidelines_prompt when structured-search tools are added (hybrid agent).
const StructuredResearchGuidelinesPrompt = `You have access to the following tools:
- ` + "`search_docs`" + ` — Search the document datastore. Returns SEARCH_RESULTS with CITE_ID for citation.
- ` + "`get_schema()`" + ` — Returns schema information for all structured (SQL) tables. Use FIRST before any SQL query.
- ` + "`execute_sql_query(sql_query: str)`" + ` — Executes SQL and returns results. Input: complete SQL string.

You have access to the following data sources:
1. Document Datastore (Unstructured): Use ` + "`search_docs`" + ` for documents and text.
2. Structured Datastore (SQL): Use ` + "`get_schema()`" + ` first to discover tables/columns, then ` + "`execute_sql_query`" + ` for queries.

## Research Strategy
- For structured data: call ` + "`get_schema()`" + ` first, then ` + "`execute_sql_query`" + ` with complete SQL.
- For documents: use ` + "`search_docs`" + `.
- Use both when the question needs numbers from SQL and context from documents.
`

// HasToolByName returns true if tools_config already has a tool with the given name.
func HasToolByName(path string, name string) (bool, error) {
	m, err := LoadYAML(path)
	if err != nil {
		return false, err
	}
	list, err := getResearchToolsConfig(m)
	if err != nil {
		return false, err
	}
	for _, t := range list {
		tm, ok := t.(map[string]interface{})
		if !ok {
			continue
		}
		if n, _ := tm["name"].(string); n == name {
			return true, nil
		}
	}
	return false, nil
}

// AddWebSearch adds the web_search tool to the YAML at path.
func AddWebSearch(path string) error {
	return AddToolToYAML(path, WebSearchTool())
}

// AddUnstructuredSearch ensures one search_docs (SearchUnstructuredDataStep) exists; idempotent.
func AddUnstructuredSearch(path string) error {
	has, err := HasToolByName(path, "search_docs")
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	return AddToolToYAML(path, UnstructuredSearchTool())
}

// AddStructuredSearch adds get_schema and execute_sql_query and updates research_guidelines_prompt.
func AddStructuredSearch(path string) error {
	if err := AddToolToYAML(path, getSchemaTool()); err != nil {
		return err
	}
	if err := AddToolToYAML(path, executeSQLQueryTool()); err != nil {
		return err
	}
	return UpdateResearchGuidelinesPrompt(path, StructuredResearchGuidelinesPrompt)
}
