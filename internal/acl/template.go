package acl

// MinimalACLYAML is the default ACL with no tools (empty tools_config).
// Users add tools via `agent <name> add <tool-type>` or via quick setup.
const MinimalACLYAML = `version: 0.1
inputs:
  query: str

outputs:
  response: str

nodes:
  create_message_history:
    type: CreateMessageHistoryStep
    input_mapping:
      query: __inputs__#query

  research:
    type: AgenticResearchStep
    ui_stream_types:
      retrievals: true
    config:
      tools_config: []

      agent_config:
        agent_loop:
          num_turns: 10
          parallel_tool_calls: false
          model_name_or_path: "vertex_ai/claude-opus-4-5@20251101"
          identity_guidelines_prompt: |
            You are a retrieval-augmented assistant created by Contextual AI. You provide factual, grounded answers to user's questions by retrieving information via tools and then synthesizing a response based only on what you retrieved.

          research_guidelines_prompt: |
            You have access to tools configured for this agent. Use them as needed to gather information before answering.
            Plan your research, run tools as appropriate, and synthesize a complete answer based on the results.

    input_mapping:
      message_history: create_message_history#message_history

  generate:
    type: GenerateFromResearchStep
    ui_stream_types:
      generation: true
    config:
      model_name_or_path: "vertex_ai/claude-opus-4-5@20251101"

      identity_guidelines_prompt: |
        You are a retrieval-augmented assistant created by Contextual AI. Your purpose is to provide factual, grounded answers by retrieving information via tools and then synthesizing a response based only on what you retrieved. Always start immediately with the answer, don't begin with a preamble or thoughts.

      response_guidelines_prompt: |
        ## Output
        - Write a concise, direct, well-structured answer in **Markdown**.
        - **START IMMEDIATELY WITH THE ANSWER.** Never begin with preamble.
        - If the required fact is missing from your research results, state limitations or reply that you don't have specific information available.

    input_mapping:
      message_history: create_message_history#message_history
      research: research#research

  __outputs__:
    type: output
    input_mapping:
      response: generate#response
`
