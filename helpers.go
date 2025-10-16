package main

// Module: helpers.go
// Purpose: Helper functions for the model
// Responsibilities:
// - Getting currently selected file across different display modes
// - Utility functions for cursor management

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
