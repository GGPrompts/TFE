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
// Grid and Detail views need full width for their column layouts
func (m model) isDualPaneCompatible() bool {
	return m.displayMode == modeList || m.displayMode == modeTree
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
