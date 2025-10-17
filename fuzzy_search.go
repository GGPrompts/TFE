package main

// Module: fuzzy_search.go
// Purpose: Fuzzy file search functionality using go-fzf
// Responsibilities:
// - Building file lists for fuzzy search
// - Launching fuzzy finder interface
// - Processing search results

import (
	"path/filepath"
	"strings"

	fzf "github.com/koki-develop/go-fzf"
	tea "github.com/charmbracelet/bubbletea"
)

// launchFuzzySearch creates a fuzzy search interface for all files in current directory
func (m *model) launchFuzzySearch() tea.Cmd {
	return func() tea.Msg {
		// Build list of all files (recursively walk directory)
		items := m.buildFuzzySearchItems()

		if len(items) == 0 {
			return fuzzySearchResultMsg{
				selected: "",
				err:      nil,
			}
		}

		// Create display names (relative paths from current directory)
		displayNames := make([]string, len(items))
		for i, item := range items {
			// Remove the current path prefix to show relative paths
			relPath := strings.TrimPrefix(item.path, m.currentPath)
			relPath = strings.TrimPrefix(relPath, "/")
			if relPath == "" {
				relPath = item.name
			}
			// Add folder indicator
			if item.isDir {
				relPath += "/"
			}
			displayNames[i] = relPath
		}

		// Create fuzzy finder with optimized settings for speed
		f, err := fzf.New(
			fzf.WithLimit(8), // Reduced to 8 results for faster rendering
		)
		if err != nil {
			return fuzzySearchResultMsg{
				selected: "",
				err:      err,
			}
		}

		// Find selected item
		idxs, err := f.Find(displayNames, func(i int) string { return displayNames[i] })
		if err != nil {
			return fuzzySearchResultMsg{
				selected: "",
				err:      err,
			}
		}

		// Get selected file path
		var selected string
		if len(idxs) > 0 {
			selected = items[idxs[0]].path
		}

		return fuzzySearchResultMsg{
			selected: selected,
			err:      nil,
		}
	}
}

// buildFuzzySearchItems recursively builds a list of all files for fuzzy search
func (m *model) buildFuzzySearchItems() []fileItem {
	var items []fileItem
	const maxItems = 200 // Reduced limit for faster searching

	// Start with current directory files (skip "..")
	for _, f := range m.files {
		if f.name != ".." {
			items = append(items, f)
			if len(items) >= maxItems {
				return items
			}
		}
	}

	// Only recurse into subdirectories if we haven't hit the limit
	// and only go 1 level deep
	if len(items) < maxItems {
		m.addSubdirFiles(m.currentPath, 0, 1, &items, maxItems)
	}

	return items
}

// addSubdirFiles recursively adds files from subdirectories
func (m *model) addSubdirFiles(dirPath string, currentDepth, maxDepth int, items *[]fileItem, maxItems int) {
	if currentDepth >= maxDepth || len(*items) >= maxItems {
		return
	}

	// Get files in this directory using loadSubdirFiles pattern
	files := m.loadSubdirFiles(dirPath)

	// Process each file
	for _, file := range files {
		// Check if we've hit the limit
		if len(*items) >= maxItems {
			return
		}

		// Add file to items
		*items = append(*items, file)

		// If it's a directory, recurse into it
		if file.isDir {
			m.addSubdirFiles(file.path, currentDepth+1, maxDepth, items, maxItems)
		}
	}
}

// navigateToFuzzyResult navigates to the selected file from fuzzy search
func (m *model) navigateToFuzzyResult(selectedPath string) {
	if selectedPath == "" {
		return
	}

	// Get directory and filename
	dir := filepath.Dir(selectedPath)
	filename := filepath.Base(selectedPath)

	// If the file is in a different directory, navigate there
	if dir != m.currentPath {
		m.currentPath = dir
		m.loadFiles()
	}

	// Find the file in the current file list and move cursor to it
	for i, file := range m.files {
		if strings.EqualFold(file.name, filename) {
			m.cursor = i

			// Load preview if not a directory
			if !file.isDir && m.viewMode == viewDualPane {
				m.loadPreview(file.path)
				m.populatePreviewCache()
			}
			break
		}
	}
}
