package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderListView renders files in a vertical list (current default view)
func (m model) renderListView(maxVisible int) string {
	var s strings.Builder

	// Calculate visible range (simple scrolling)
	start := 0
	end := len(m.files)

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

	for i := start; i < end; i++ {
		file := m.files[i]

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

		// Build the line
		line := fmt.Sprintf("  %s %s", icon, file.name)

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

	// Calculate how many rows we need
	totalItems := len(m.files)
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

			file := m.files[idx]
			icon := getFileIcon(file)

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
			cell := fmt.Sprintf("%s %-12s", icon, displayName)

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

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		PaddingLeft(2)

	header := fmt.Sprintf("%-30s  %-10s  %-12s  %-15s", "Name", "Size", "Modified", "Type")
	s.WriteString(headerStyle.Render(header))
	s.WriteString("\n")

	// Separator
	separator := strings.Repeat("─", m.width-4)
	s.WriteString(pathStyle.Render(separator))
	s.WriteString("\n")

	// Calculate visible range
	start := 0
	end := len(m.files)

	if len(m.files) > maxVisible-2 { // -2 for header and separator
		start = m.cursor - (maxVisible-2)/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible - 2
		if end > len(m.files) {
			end = len(m.files)
			start = end - (maxVisible - 2)
			if start < 0 {
				start = 0
			}
		}
	}

	// Render rows
	for i := start; i < end; i++ {
		file := m.files[i]
		icon := getFileIcon(file)

		// Truncate long names
		displayName := file.name
		maxNameLen := 25
		if len(displayName) > maxNameLen {
			displayName = displayName[:maxNameLen-2] + ".."
		}

		name := fmt.Sprintf("%s %s", icon, displayName)
		size := "-"
		if !file.isDir {
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

// renderTreeView renders files in a hierarchical tree structure
func (m model) renderTreeView(maxVisible int) string {
	var s strings.Builder

	// For now, render a simplified tree view similar to list view
	// In the future, this could show expanded subdirectories
	start := 0
	end := len(m.files)

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

	for i := start; i < end; i++ {
		file := m.files[i]

		// Use tree-style prefix
		prefix := "├─ "
		if i == len(m.files)-1 {
			prefix = "└─ "
		}
		if file.name == ".." {
			prefix = "↑  "
		}

		icon := getFileIcon(file)
		style := fileStyle

		if file.isDir {
			style = folderStyle
		}

		if isClaudeContextFile(file.name) {
			style = claudeContextStyle
		}

		line := fmt.Sprintf("  %s%s %s", prefix, icon, file.name)

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
