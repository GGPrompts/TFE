package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Module: helpers.go
// Purpose: Helper functions for the model
// Responsibilities:
// - Getting currently selected file across different display modes
// - Utility functions for cursor management
// - Path formatting utilities

// getCurrentFile returns the currently selected file based on cursor position
// This handles the complexity of tree view with expanded folders
func (m model) getCurrentFile() *fileItem {
	if len(m.files) == 0 || m.cursor < 0 {
		return nil
	}

	// In tree view, we need to map cursor to the flattened tree
	if m.displayMode == modeTree {
		files := m.getFilteredFiles()
		treeItems := m.buildTreeItems(files, 0, []bool{})
		if m.cursor < len(treeItems) {
			return &treeItems[m.cursor].file
		}
		return nil
	}

	// In other views, use filtered files
	files := m.getFilteredFiles()
	if m.cursor < len(files) {
		return &files[m.cursor]
	}

	return nil
}

// getMaxCursor returns the maximum valid cursor position for the current display mode
func (m model) getMaxCursor() int {
	if m.displayMode == modeTree {
		files := m.getFilteredFiles()
		treeItems := m.buildTreeItems(files, 0, []bool{})
		return len(treeItems) - 1
	}

	files := m.getFilteredFiles()
	return len(files) - 1
}

// getDisplayPath returns a user-friendly path with home directory replaced by ~
func getDisplayPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	// Replace home directory with ~
	if strings.HasPrefix(path, homeDir) {
		if path == homeDir {
			return "~"
		}
		return "~" + strings.TrimPrefix(path, homeDir)
	}

	return path
}

// isDualPaneCompatible checks if the current display mode supports dual-pane view
// All display modes now support dual-pane with accordion layout
func (m model) isDualPaneCompatible() bool {
	return m.displayMode == modeList || m.displayMode == modeTree || m.displayMode == modeDetail
}

// isNarrowTerminal checks if the terminal width is too narrow for wide views
// Returns true if width < 100 (typical phone/Termux scenario)
func (m model) isNarrowTerminal() bool {
	return m.width < 100
}

// isPromptFile checks if a file is a prompt file (.prompty, .yaml, .md, .txt)
// Only files in special directories (.claude/, ~/.prompts/) are considered prompts
// Exception: .prompty files are always prompts (Microsoft Prompty format)
func isPromptFile(item fileItem) bool {
	if item.isDir {
		return false
	}

	ext := strings.ToLower(filepath.Ext(item.name))

	// .prompty is always a prompt file (Microsoft Prompty format)
	if ext == ".prompty" {
		return true
	}

	// For other extensions, only consider them prompts if in special directories
	if ext == ".md" || ext == ".yaml" || ext == ".yml" || ext == ".txt" {
		// Exclude .claude/agents/ - those are documentation files, not prompt templates
		if strings.Contains(item.path, "/.claude/agents/") {
			return false
		}

		// Check if in .claude/ or any subfolder
		if strings.Contains(item.path, "/.claude/") || strings.HasSuffix(item.path, "/.claude") {
			return true
		}
		// Check if in ~/.prompts/ or any subfolder
		homeDir, _ := os.UserHomeDir()
		promptsDir := filepath.Join(homeDir, ".prompts")
		if strings.HasPrefix(item.path, promptsDir) {
			return true
		}
	}

	return false
}

// performPreviewSearch searches the preview content for the current query
// and populates searchMatches with line numbers
func (m *model) performPreviewSearch() {
	m.preview.searchMatches = nil
	m.preview.currentMatch = -1

	if m.preview.searchQuery == "" {
		m.setStatusMessage("🔍 Search: (type to search, Enter/n: next, Esc: exit)", false)
		return
	}

	queryLower := strings.ToLower(m.preview.searchQuery)

	// Search through preview content
	for i, line := range m.preview.content {
		if strings.Contains(strings.ToLower(line), queryLower) {
			m.preview.searchMatches = append(m.preview.searchMatches, i)
		}
	}

	if len(m.preview.searchMatches) > 0 {
		m.preview.currentMatch = 0
		// Scroll to first match
		m.preview.scrollPos = m.preview.searchMatches[0]
		m.setStatusMessage(fmt.Sprintf("🔍 Found %d matches (1/%d) - n: next, Shift+n: prev, Esc: exit", len(m.preview.searchMatches), len(m.preview.searchMatches)), false)
	} else {
		m.setStatusMessage(fmt.Sprintf("🔍 No matches for '%s' - Esc: exit", m.preview.searchQuery), false)
	}
}

// findNextSearchMatch navigates to the next search match
func (m *model) findNextSearchMatch() {
	if len(m.preview.searchMatches) == 0 {
		m.setStatusMessage("🔍 No matches found", false)
		return
	}

	m.preview.currentMatch++
	if m.preview.currentMatch >= len(m.preview.searchMatches) {
		m.preview.currentMatch = 0 // Wrap around
	}

	// Scroll to the match
	m.preview.scrollPos = m.preview.searchMatches[m.preview.currentMatch]
	m.setStatusMessage(fmt.Sprintf("🔍 Match %d/%d - n: next, Shift+n: prev, Esc: exit", m.preview.currentMatch+1, len(m.preview.searchMatches)), false)
}

// findPreviousSearchMatch navigates to the previous search match
func (m *model) findPreviousSearchMatch() {
	if len(m.preview.searchMatches) == 0 {
		m.setStatusMessage("🔍 No matches found", false)
		return
	}

	m.preview.currentMatch--
	if m.preview.currentMatch < 0 {
		m.preview.currentMatch = len(m.preview.searchMatches) - 1 // Wrap around
	}

	// Scroll to the match
	m.preview.scrollPos = m.preview.searchMatches[m.preview.currentMatch]
	m.setStatusMessage(fmt.Sprintf("🔍 Match %d/%d - n: next, Shift+n: prev, Esc: exit", m.preview.currentMatch+1, len(m.preview.searchMatches)), false)
}

// getHelpSectionName returns the appropriate help section name based on current context
// This is used for context-aware F1 help navigation
func (m model) getHelpSectionName() string {
	// Check context in priority order (most specific first)
	if m.promptEditMode {
		return "## Prompt Templates & Fillable Fields"
	}
	if m.contextMenuOpen {
		return "## Context Menu"
	}
	if m.commandFocused {
		return "## Command Prompt (Vim-Style)"
	}
	if m.viewMode == viewFullPreview {
		return "## Preview & Full-Screen Mode"
	}
	if m.viewMode == viewDualPane {
		return "## Dual-Pane Mode"
	}
	// Default to Navigation section for single-pane mode
	return "## Navigation"
}

// findSectionLine searches for a section heading in content and returns its line number
// Returns -1 if not found, otherwise returns the 0-based line index
func findSectionLine(content []string, sectionName string) int {
	for i, line := range content {
		if strings.Contains(line, sectionName) {
			return i
		}
	}
	return -1
}

// autofillDefaults populates DATE and TIME variables with current values
// Called when entering prompt edit mode for the first time
func (m *model) autofillDefaults() {
	if m.preview.promptTemplate == nil {
		return
	}

	contextVars := getContextVariables(m)

	for _, varName := range m.preview.promptTemplate.variables {
		varNameLower := strings.ToLower(varName)

		// Auto-fill DATE and TIME from context
		if varNameLower == "date" || varNameLower == "time" {
			if value, exists := contextVars[varName]; exists {
				m.filledVariables[varName] = value
			}
		}
	}
}

// scrollToFocusedVariable scrolls the preview to show the currently focused variable
// This is called when navigating between variables with Tab/Shift+Tab in edit mode
func (m *model) scrollToFocusedVariable() {
	if m.focusedVariableIndex < 0 || m.preview.promptTemplate == nil {
		return
	}

	if m.focusedVariableIndex >= len(m.preview.promptTemplate.variables) {
		return
	}

	// Get focused variable name
	varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]

	// We need to search in the RENDERED content (after processing and wrapping)
	// because scrollPos is applied to the wrapped lines, not raw content

	// Calculate box content width (same logic as renderPromptPreview)
	var boxContentWidth int
	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6
	} else {
		boxContentWidth = m.rightWidth - 2
	}

	// Calculate available width for content (prompts don't show line numbers)
	availableWidth := boxContentWidth - 2 // Just padding
	if availableWidth < 20 {
		availableWidth = 20
	}

	// Process content the same way as renderPromptPreview
	var contentLines []string
	if m.promptEditMode {
		// In edit mode, use rendered template with inline variables
		renderedTemplate := m.renderInlineVariables(m.preview.promptTemplate.template)
		contentLines = strings.Split(renderedTemplate, "\n")
	} else {
		// Before edit mode, use highlighted template
		highlightedTemplate := m.highlightVariablesBeforeEdit(m.preview.promptTemplate.template)
		contentLines = strings.Split(highlightedTemplate, "\n")
	}

	// Wrap content lines (same as renderPromptPreview)
	var wrappedLines []string
	for _, line := range contentLines {
		wrapped := wrapLine(line, availableWidth)
		wrappedLines = append(wrappedLines, wrapped...)
	}

	// Search for the variable in wrapped lines
	// Look for the variable name (it appears without {{}} in edit mode, or with {{}} before edit)
	targetLine := -1
	var searchPatterns []string

	if m.promptEditMode {
		// In edit mode, variables appear without brackets (but with ANSI codes)
		// Search for the variable name or filled value
		filledValue, hasFilled := m.filledVariables[varName]
		if hasFilled && filledValue != "" {
			searchPatterns = []string{filledValue, varName}
		} else {
			searchPatterns = []string{varName}
		}
	} else {
		// Before edit mode, variables appear with {{}}
		searchPatterns = []string{"{{" + varName + "}}"}
	}

	// Search in wrapped lines
	for i, line := range wrappedLines {
		for _, pattern := range searchPatterns {
			if strings.Contains(line, pattern) {
				targetLine = i
				break
			}
		}
		if targetLine >= 0 {
			break
		}
	}

	if targetLine >= 0 {
		// Calculate visible lines (accounting for header)
		visibleLines := m.height - 6 // Default offset for header/footer
		if m.viewMode == viewDualPane {
			visibleLines = m.height - 8
		}

		// Account for header height in prompt preview
		// Header typically takes 4-8 lines depending on metadata
		// Estimate conservatively
		headerEstimate := 6
		contentHeight := visibleLines - headerEstimate
		if contentHeight < 5 {
			contentHeight = 5
		}

		// Try to center the variable in the visible content area
		centerOffset := contentHeight / 2
		newScrollPos := targetLine - centerOffset

		if newScrollPos < 0 {
			newScrollPos = 0
		}

		// Don't scroll past the end of wrapped content
		maxScroll := len(wrappedLines) - contentHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		if newScrollPos > maxScroll {
			newScrollPos = maxScroll
		}

		m.preview.scrollPos = newScrollPos
	}
}
