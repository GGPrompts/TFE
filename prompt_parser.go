package main

// Module: prompt_parser.go
// Purpose: Prompt file parsing and template rendering
// Responsibilities:
// - Parse .prompty format (YAML frontmatter)
// - Parse simple YAML format
// - Parse plain text (.md, .txt)
// - Template variable substitution
// - Context variable providers

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// parsePromptFile parses a prompt file and returns a promptTemplate
// Supports three formats:
// 1. .prompty - Microsoft Prompty format (YAML frontmatter between --- markers)
// 2. .yaml/.yml - Simple YAML with 'template' field
// 3. .md/.txt - Plain text with {{variables}}
func parsePromptFile(path string) (*promptTemplate, error) {
	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	ext := strings.ToLower(filepath.Ext(path))
	filename := filepath.Base(path)

	// Determine source location
	source := determinePromptSource(path)

	var tmpl promptTemplate
	tmpl.source = source
	tmpl.raw = contentStr

	// Parse based on format
	switch ext {
	case ".prompty":
		// Microsoft Prompty format: YAML frontmatter between --- markers
		if err := parsePromptyFormat(contentStr, &tmpl); err != nil {
			return nil, err
		}

	case ".yaml", ".yml":
		// Simple YAML format
		if err := parseYAMLFormat(contentStr, &tmpl); err != nil {
			return nil, err
		}

	case ".md", ".txt":
		// Plain text format - just the template
		tmpl.template = contentStr
		// Derive name from filename (remove extension)
		tmpl.name = strings.TrimSuffix(filename, ext)
		tmpl.description = ""

	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}

	// Extract variables from template
	tmpl.variables = extractVariables(tmpl.template)

	return &tmpl, nil
}

// parsePromptyFormat parses Microsoft Prompty format (YAML frontmatter between ---)
// Format:
// ---
// name: Prompt Name
// description: Description here
// inputs:
//   var1:
//     type: string
// ---
// system:
// Template content here with {{var1}}
func parsePromptyFormat(content string, tmpl *promptTemplate) error {
	// Split by --- markers
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid .prompty format: missing --- markers")
	}

	// Parse YAML frontmatter (parts[1])
	var metadata struct {
		Name        string                 `yaml:"name"`
		Description string                 `yaml:"description"`
		Inputs      map[string]interface{} `yaml:"inputs"`
	}

	if err := yaml.Unmarshal([]byte(parts[1]), &metadata); err != nil {
		return fmt.Errorf("failed to parse YAML frontmatter: %w", err)
	}

	tmpl.name = metadata.Name
	tmpl.description = metadata.Description

	// Template is everything after second ---
	tmpl.template = strings.TrimSpace(parts[2])

	return nil
}

// parseYAMLFormat parses simple YAML format
// Format:
// name: Prompt Name
// description: Description here
// template: |
//   Template content here with {{variables}}
func parseYAMLFormat(content string, tmpl *promptTemplate) error {
	var data struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Template    string `yaml:"template"`
	}

	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	tmpl.name = data.Name
	tmpl.description = data.Description
	tmpl.template = data.Template

	return nil
}

// extractVariables finds all {{VARIABLE}} placeholders in template
func extractVariables(template string) []string {
	// Match {{variable}} pattern (case-insensitive)
	re := regexp.MustCompile(`\{\{([a-zA-Z0-9_]+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	// Extract unique variable names
	varMap := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			varMap[match[1]] = true
		}
	}

	// Convert to slice
	vars := make([]string, 0, len(varMap))
	for v := range varMap {
		vars = append(vars, v)
	}

	return vars
}

// renderPromptTemplate renders a template by substituting variables
// Variables are case-insensitive ({{file}}, {{FILE}}, {{File}} all match)
func renderPromptTemplate(tmpl *promptTemplate, vars map[string]string) string {
	result := tmpl.template

	// Replace each variable (case-insensitive)
	for varName, value := range vars {
		// Try all case variations
		patterns := []string{
			fmt.Sprintf("{{%s}}", varName),
			fmt.Sprintf("{{%s}}", strings.ToUpper(varName)),
			fmt.Sprintf("{{%s}}", strings.ToLower(varName)),
			fmt.Sprintf("{{%s}}", strings.Title(strings.ToLower(varName))),
		}

		for _, pattern := range patterns {
			result = strings.ReplaceAll(result, pattern, value)
		}
	}

	return result
}

// getContextVariables returns a map of all context variables for rendering
func getContextVariables(m *model) map[string]string {
	vars := make(map[string]string)

	// Get current file info
	currentFile := m.getCurrentFile()
	if currentFile != nil {
		vars["file"] = currentFile.path
		vars["filename"] = currentFile.name
	} else {
		vars["file"] = ""
		vars["filename"] = ""
	}

	// Project info
	vars["project"] = filepath.Base(m.currentPath)
	vars["path"] = m.currentPath

	// Date and time
	now := time.Now()
	vars["DATE"] = now.Format("2006-01-02") // YYYY-MM-DD
	vars["TIME"] = now.Format("15:04")      // HH:MM

	return vars
}

// determinePromptSource determines the source location of a prompt file
func determinePromptSource(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "local"
	}

	// Check if in home directory ~/.prompts/
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalPromptsDir := filepath.Join(homeDir, ".prompts")
		if strings.HasPrefix(absPath, globalPromptsDir) {
			return "global"
		}
	}

	// Check if in .claude/commands/
	if strings.Contains(absPath, "/.claude/commands/") {
		return "command"
	}

	// Check if in .claude/agents/
	if strings.Contains(absPath, "/.claude/agents/") {
		return "agent"
	}

	// Check if in .claude/skills/
	if strings.Contains(absPath, "/.claude/skills/") {
		return "skill"
	}

	// Default to local
	return "local"
}

// detectFieldType classifies a variable name into a field type
// Rules:
// - File type: file, filepath, filename, path → fieldTypeFile
// - Long type: code, content, text, body, snippet → fieldTypeLong
// - Short type: everything else → fieldTypeShort
func detectFieldType(varName string) inputFieldType {
	nameLower := strings.ToLower(varName)

	// File-related variables
	fileKeywords := []string{"file", "filepath", "filename", "path"}
	for _, keyword := range fileKeywords {
		if strings.Contains(nameLower, keyword) {
			return fieldTypeFile
		}
	}

	// Long text variables
	longKeywords := []string{"code", "content", "text", "body", "snippet"}
	for _, keyword := range longKeywords {
		if strings.Contains(nameLower, keyword) {
			return fieldTypeLong
		}
	}

	// Default to short
	return fieldTypeShort
}

// getFieldColor assigns a color code based on variable name and type
// Rules:
// - File variables: blue ("39")
// - Date/Time (auto-filled): green ("34")
// - Custom variables: yellow ("220")
func getFieldColor(varName string, fieldType inputFieldType) string {
	nameLower := strings.ToLower(varName)

	// Auto-filled date/time variables (green)
	if nameLower == "date" || nameLower == "time" {
		return "34" // Green
	}

	// File variables (blue)
	if fieldType == fieldTypeFile {
		return "39" // Blue
	}

	// Custom variables (yellow)
	return "220" // Yellow
}

// getFilledVariables returns a map of variables filled from input fields
// Uses user-entered values if available, otherwise defaults to context values
func getFilledVariables(fields []promptInputField, m *model) map[string]string {
	vars := getContextVariables(m)

	// Override with user-entered values
	for _, field := range fields {
		if field.hasContent() {
			// User entered a value - use it
			vars[field.name] = field.value
		} else {
			// No user value - use default (already in vars from getContextVariables)
			// But ensure it's set even if not in context
			if _, exists := vars[field.name]; !exists {
				vars[field.name] = field.defaultValue
			}
		}
	}

	return vars
}

// createInputFields creates input fields for a prompt template
// Auto-populates defaults from getContextVariables()
// DATE and TIME are auto-filled but still editable (user can override)
func createInputFields(tmpl *promptTemplate, m *model) []promptInputField {
	if tmpl == nil || len(tmpl.variables) == 0 {
		return nil
	}

	// Get context variables for defaults
	contextVars := getContextVariables(m)

	fields := make([]promptInputField, 0)

	for _, varName := range tmpl.variables {
		// Detect field type
		fieldType := detectFieldType(varName)

		// Get default value from context (case-insensitive lookup)
		defaultValue := ""
		for ctxKey, ctxVal := range contextVars {
			if strings.EqualFold(ctxKey, varName) {
				defaultValue = ctxVal
				break
			}
		}

		// Get color
		color := getFieldColor(varName, fieldType)

		// Calculate display width (will be updated when rendering)
		// For now, set a reasonable default
		displayWidth := 60

		// Create field
		field := promptInputField{
			name:         varName,
			value:        "",
			defaultValue: defaultValue,
			fieldType:    fieldType,
			displayWidth: displayWidth,
			color:        color,
		}

		fields = append(fields, field)
	}

	return fields
}
