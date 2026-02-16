package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// validateAgentName returns an error if the agent name is invalid (empty, contains path/space).
func validateAgentName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}
	if strings.ContainsAny(name, "/\\ \t\n") {
		return fmt.Errorf("agent name cannot contain path separators or spaces")
	}
	return nil
}

func sha256Hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
