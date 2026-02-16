package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

const (
	SetupQuick  = "Quick setup"
	SetupManual = "Manually configure later"
)

var toolOptions = []string{"web-search", "unstructured-search", "structured-search"}

// PromptSetupMode asks the user to choose Quick setup or Manually configure later.
// Returns true for Quick setup, false for Manual.
func PromptSetupMode() (quick bool, err error) {
	var choice string
	prompt := &survey.Select{
		Message: "How do you want to set up your agent?",
		Options: []string{SetupQuick, SetupManual},
	}
	if err := survey.AskOne(prompt, &choice); err != nil {
		return false, err
	}
	return choice == SetupQuick, nil
}

// PromptToolSelection asks the user to select at least one tool (space to toggle, Enter to confirm).
func PromptToolSelection() (selected []string, err error) {
	prompt := &survey.MultiSelect{
		Message: "Select tools (space to toggle, at least one required):",
		Options: toolOptions,
	}
	if err := survey.AskOne(prompt, &selected, survey.WithValidator(survey.MinItems(1))); err != nil {
		return nil, err
	}
	return selected, nil
}

// PromptReferenceDocuments asks for file paths (space or comma separated). Empty input is allowed (skip ingest).
// Validates that each path exists and has an allowed extension; returns error on first invalid path.
func PromptReferenceDocuments() (paths []string, err error) {
	var input string
	prompt := &survey.Input{
		Message: "Reference document paths (space or comma separated, or leave empty to skip):",
		Help:    "Allowed: .pdf, .html, .htm, .mhtml, .doc, .docx, .ppt, .pptx, .png, .jpg, .jpeg",
	}
	if err := survey.AskOne(prompt, &input); err != nil {
		return nil, err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}
	// Split by comma or space
	parts := strings.FieldsFunc(input, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		info, err := os.Stat(p)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file not found: %s", p)
			}
			return nil, err
		}
		if info.IsDir() {
			return nil, fmt.Errorf("not a file (directory): %s", p)
		}
		ext := strings.ToLower(filepath.Ext(p))
		if !AllowedUnstructuredExt[ext] {
			return nil, fmt.Errorf("unsupported extension %s for %s; allowed: .pdf, .html, .htm, .mhtml, .doc, .docx, .ppt, .pptx, .png, .jpg, .jpeg", ext, p)
		}
		paths = append(paths, p)
	}
	return paths, nil
}
