package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// getVisibleRange calculates the start and end indices for visible items in the file list
func (m model) getVisibleRange(maxVisible int) (start, end int) {
	start = 0
	end = len(m.files)

	if len(m.files) > maxVisible {
		start = m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(m.files) {
			end = len(m.files)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}
	return start, end
}

// renderListView renders files in a vertical list (current default view)
func (m model) renderListView(maxVisible int) string {
	var s strings.Builder

	// Get filtered files (respects favorites filter)
	files := m.getFilteredFiles()

	// Calculate visible range (simple scrolling)
	start := 0
	end := len(files)
	if len(files) > maxVisible {
		start = m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(files) {
			end = len(files)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		file := files[i]

		// Get icon based on file type
		icon := getFileIcon(file)
		style := fileStyle

		if file.isDir {
			style = folderStyle
		}

		// Override with orange color if it's a Claude context file
		if isClaudeContextFile(file.name) {
			style = claudeContextStyle
		}

		// Add star indicator for favorites
		favIndicator := ""
		if m.isFavorite(file.path) {
			favIndicator = "⭐"
		}

		// Truncate long filenames to prevent wrapping
		// In dual-pane mode, use narrower width to fit in left pane
		displayName := file.name
		maxNameLen := 40 // Default for single-pane
		if m.viewMode == viewDualPane {
			// Account for left pane width, icon (2), spaces (2), and padding
			maxNameLen = m.leftWidth - 10
			if maxNameLen < 20 {
				maxNameLen = 20 // Minimum reasonable length
			}
		}
		if len(displayName) > maxNameLen {
			displayName = displayName[:maxNameLen-2] + ".."
		}

		// Build the line
		line := fmt.Sprintf("  %s%s %s", icon, favIndicator, displayName)

		// Apply selection style
		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else {
			line = style.Render(line)
		}

		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

// renderGridView renders files in a multi-column grid layout
func (m model) renderGridView(maxVisible int) string {
	var s strings.Builder

	// Get filtered files (respects favorites filter)
	files := m.getFilteredFiles()

	// Calculate how many rows we need
	totalItems := len(files)
	rows := (totalItems + m.gridColumns - 1) / m.gridColumns

	// Calculate visible range
	start := 0
	end := rows

	if rows > maxVisible {
		cursorRow := m.cursor / m.gridColumns
		start = cursorRow - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > rows {
			end = rows
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	// Render rows
	for row := start; row < end; row++ {
		for col := 0; col < m.gridColumns; col++ {
			idx := row*m.gridColumns + col
			if idx >= totalItems {
				break
			}

			file := files[idx]
			icon := getFileIcon(file)

			// Add star indicator for favorites
			favIndicator := ""
			if m.isFavorite(file.path) {
				favIndicator = "⭐"
			}

			// Truncate long names
			displayName := file.name
			maxNameLen := 12
			if len(displayName) > maxNameLen {
				displayName = displayName[:maxNameLen-2] + ".."
			}

			style := fileStyle
			if file.isDir {
				style = folderStyle
			}
			if isClaudeContextFile(file.name) {
				style = claudeContextStyle
			}

			// Build cell content
			cell := fmt.Sprintf("%s%s %-12s", icon, favIndicator, displayName)

			// Apply selection style
			if idx == m.cursor {
				cell = selectedStyle.Render(cell)
			} else {
				cell = style.Render(cell)
			}

			s.WriteString(cell)
			s.WriteString("  ")
		}
		s.WriteString("\n")
	}

	return s.String()
}

// renderDetailView renders files in a detailed table with columns
func (m model) renderDetailView(maxVisible int) string {
	var s strings.Builder

	// Get filtered files (respects favorites filter)
	files := m.getFilteredFiles()

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		PaddingLeft(2)

	header := fmt.Sprintf("%-30s  %-10s  %-12s  %-15s", "Name", "Size", "Modified", "Type")
	s.WriteString(headerStyle.Render(header))
	s.WriteString("\n")

	// Separator - use left pane width in dual-pane mode to prevent wrapping
	separatorWidth := m.width - 4
	if m.viewMode == viewDualPane {
		separatorWidth = m.leftWidth - 4
	}
	separator := strings.Repeat("─", separatorWidth)
	s.WriteString(pathStyle.Render(separator))
	s.WriteString("\n")

	// Calculate visible range
	start := 0
	end := len(files)

	if len(files) > maxVisible-2 { // -2 for header and separator
		start = m.cursor - (maxVisible-2)/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible - 2
		if end > len(files) {
			end = len(files)
			start = end - (maxVisible - 2)
			if start < 0 {
				start = 0
			}
		}
	}

	// Render rows
	for i := start; i < end; i++ {
		file := files[i]
		icon := getFileIcon(file)

		// Add star indicator for favorites
		favIndicator := ""
		if m.isFavorite(file.path) {
			favIndicator = "⭐"
		}

		// Truncate long names
		displayName := file.name
		maxNameLen := 25
		if len(displayName) > maxNameLen {
			displayName = displayName[:maxNameLen-2] + ".."
		}

		name := fmt.Sprintf("%s%s %s", icon, favIndicator, displayName)
		size := "-"
		if file.isDir {
			// Show item count for directories
			if file.name == ".." {
				size = "-"
			} else {
				count := getDirItemCount(file.path)
				if count == 0 {
					size = "empty"
				} else if count == 1 {
					size = "1 item"
				} else {
					size = fmt.Sprintf("%d items", count)
				}
			}
		} else {
			size = formatFileSize(file.size)
		}
		modified := formatModTime(file.modTime)
		fileType := "File"
		if file.isDir {
			fileType = "Folder"
		}

		line := fmt.Sprintf("%-30s  %-10s  %-12s  %-15s", name, size, modified, fileType)

		style := fileStyle
		if file.isDir {
			style = folderStyle
		}
		if isClaudeContextFile(file.name) {
			style = claudeContextStyle
		}

		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else {
			line = style.Render(line)
		}

		s.WriteString("  ")
		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

// buildTreeItems builds a flattened list of tree items including expanded directories
func (m model) buildTreeItems(files []fileItem, depth int, parentLasts []bool) []treeItem {
	items := make([]treeItem, 0)

	for i, file := range files {
		isLast := i == len(files)-1

		// Add current item
		item := treeItem{
			file:        file,
			depth:       depth,
			isLast:      isLast,
			parentLasts: append([]bool{}, parentLasts...), // Copy parent lasts
		}
		items = append(items, item)

		// If this is an expanded directory, recursively add its contents
		if file.isDir && file.name != ".." && m.expandedDirs[file.path] {
			// Load subdirectory contents
			subFiles := m.loadSubdirFiles(file.path)
			if len(subFiles) > 0 {
				// Update parentLasts for children
				newParentLasts := append(parentLasts, isLast)
				subItems := m.buildTreeItems(subFiles, depth+1, newParentLasts)
				items = append(items, subItems...)
			}
		}
	}

	return items
}

// renderTreeView renders files in a hierarchical tree structure with expandable folders
func (m model) renderTreeView(maxVisible int) string {
	var s strings.Builder

	// Get filtered files (respects favorites filter)
	files := m.getFilteredFiles()

	// Build tree structure with expanded directories and cache it in model
	// Note: We're modifying the model here which is unusual in a render function,
	// but necessary for cursor-to-file mapping to work correctly
	treeItems := m.buildTreeItems(files, 0, []bool{})

	// Calculate visible range
	start := 0
	end := len(treeItems)
	if len(treeItems) > maxVisible {
		start = m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(treeItems) {
			end = len(treeItems)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		item := treeItems[i]
		file := item.file

		// Build indentation with tree characters
		var indent strings.Builder
		indent.WriteString("  ") // Base padding

		// Draw vertical lines for parent levels
		for j := 0; j < item.depth; j++ {
			if j < len(item.parentLasts) && !item.parentLasts[j] {
				indent.WriteString("│  ")
			} else {
				indent.WriteString("   ")
			}
		}

		// Draw tree connector
		var prefix string
		if file.name == ".." {
			prefix = "↑  "
		} else if item.isLast {
			prefix = "└─ "
		} else {
			prefix = "├─ "
		}

		// Add expansion indicator for directories
		expansionIndicator := ""
		if file.isDir && file.name != ".." {
			if m.expandedDirs[file.path] {
				expansionIndicator = "▼ " // Expanded
			} else {
				expansionIndicator = "▶ " // Collapsed
			}
		}

		icon := getFileIcon(file)
		style := fileStyle

		if file.isDir {
			style = folderStyle
		}

		if isClaudeContextFile(file.name) {
			style = claudeContextStyle
		}

		// Add star indicator for favorites
		favIndicator := ""
		if m.isFavorite(file.path) {
			favIndicator = "⭐"
		}

		// Truncate long filenames to prevent wrapping
		displayName := file.name
		maxNameLen := 25 - (item.depth * 3) // Reduce for deeper nesting
		if m.viewMode == viewDualPane {
			maxNameLen = m.leftWidth - 20 - (item.depth * 3)
			if maxNameLen < 15 {
				maxNameLen = 15
			}
		}
		if len(displayName) > maxNameLen {
			displayName = displayName[:maxNameLen-2] + ".."
		}

		line := fmt.Sprintf("%s%s%s%s%s %s", indent.String(), prefix, expansionIndicator, icon, favIndicator, displayName)

		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else {
			line = style.Render(line)
		}

		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}
