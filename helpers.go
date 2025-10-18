package main

import (
	"os"
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
