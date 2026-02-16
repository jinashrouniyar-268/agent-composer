package stream

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Handler is called for each parsed SSE event (event type and data JSON).
type Handler func(event string, data []byte) error

// ParseSSE reads SSE from r and calls h for each event. Event type is taken from JSON "event" field. Stops on "end" or error.
func ParseSSE(r io.Reader, h Handler) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" || data == "" {
			continue
		}
		var evt map[string]interface{}
		if err := json.Unmarshal([]byte(data), &evt); err != nil {
			continue
		}
		event, _ := evt["event"].(string)
		if err := h(event, []byte(data)); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
	return scanner.Err()
}

// QueryACLHandler handles Contextual query/acl SSE: metadata, message_delta, retrievals, version+event, error, end.
func QueryACLHandler(verbose bool, onDelta func(delta string), onComplete func()) Handler {
	return func(event string, data []byte) error {
		if len(data) == 0 {
			return nil
		}
		var generic map[string]interface{}
		if err := json.Unmarshal(data, &generic); err != nil {
			return nil
		}
		// Top-level event field (e.g. metadata, message_delta)
		if ev, ok := generic["event"].(string); ok {
			switch ev {
			case "metadata":
				// optional: capture conversation_id, message_id
				return nil
			case "message_delta":
				if d, ok := generic["data"].(map[string]interface{}); ok {
					if delta, ok := d["delta"].(string); ok && onDelta != nil {
						onDelta(delta)
					}
				}
				return nil
			case "message_complete", "outputs":
				if onComplete != nil {
					onComplete()
				}
				return nil
			case "retrievals":
				if verbose && generic["data"] != nil {
					// optional: print retrieval count
				}
				return nil
			case "error":
				msg := "unknown error"
				if d, ok := generic["data"].(map[string]interface{}); ok {
					if m, ok := d["message"].(string); ok {
						msg = m
					}
				}
				return fmt.Errorf("stream error: %s", msg)
			case "end":
				return io.EOF
			}
		}
		// Nested ACL events: version + event (step_start, step_end, dynamic_thinking_*, dynamic_tool_call_*, etc.)
		if version, ok := generic["version"].(string); ok && version != "" && verbose {
			if ev, ok := generic["event"].(map[string]interface{}); ok {
				printVerboseEvent(ev)
			}
		}
		return nil
	}
}

func printVerboseEvent(ev map[string]interface{}) {
	typ, _ := ev["type"].(string)
	switch typ {
	case "step_start":
		stepID, _ := ev["step_id"].(string)
		fmt.Printf("\n  [step] %s\n", stepID)
	case "step_end":
		stepID, _ := ev["step_id"].(string)
		dur, _ := ev["duration"].(float64)
		fmt.Printf("  [step] %s done (%.2fs)\n", stepID, dur)
	case "dynamic_thinking_start":
		fmt.Print("\n  💭 thinking...\n")
	case "dynamic_thinking_end":
		fmt.Print("  💭 done\n")
	case "dynamic_tool_call_created":
		toolName, _ := ev["tool_name"].(string)
		fmt.Printf("\n  🔧 tool: %s\n", toolName)
	case "dynamic_tool_call_end":
		dur, _ := ev["duration"].(float64)
		fmt.Printf("  🔧 done (%.2fs)\n", dur)
	case "dynamic_response_start":
		fmt.Print("\n  📝 generating response...\n")
	}
}

// StreamQueryACL parses SSE from query/acl and calls onDelta for each message_delta; onComplete at message_complete/outputs; stops on end or error.
func StreamQueryACL(r io.Reader, verbose bool, onDelta func(string), onComplete func()) error {
	h := QueryACLHandler(verbose, onDelta, onComplete)
	return ParseSSE(r, h)
}
