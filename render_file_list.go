package main

import (
	"fmt"
	"os"
	"path/filepath"
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

		// Override with purple color if it's an AGENTS.md file
		if isAgentsFile(file.name) {
			style = agentsStyle
		}

		// Override with bright pink if it's the .prompts folder or global prompts virtual folder
		if isPromptsFolder(file.name) || isGlobalPromptsVirtualFolder(file.name) {
			style = promptsFolderStyle
		}

		// Override with bright pink if it's a .claude prompts subfolder (commands, agents, skills)
		if file.isDir && isClaudePromptsSubfolder(file.name) {
			style = promptsFolderStyle
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
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	}

	return strings.TrimRight(s.String(), "\n")
}

// renderDetailView renders files in a detailed table with columns
func (m model) renderDetailView(maxVisible int) string {
	var s strings.Builder

	// Get filtered files (respects favorites filter)
	files := m.getFilteredFiles()

	// Header with sort indicators
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("87")). // Bright blue for header
		PaddingLeft(2)

	// Determine sort indicator (arrow)
	sortIndicator := ""
	if m.sortAsc {
		sortIndicator = " ↑" // Ascending
	} else {
		sortIndicator = " ↓" // Descending
	}

	// Build header with sort indicators
	var header string
	if m.showTrashOnly {
		// Trash mode: Name, Size, Deleted, Original Location
		nameHeader := "Name"
		sizeHeader := "Size"
		deletedHeader := "Deleted"
		locationHeader := "Original Location"

		// Add indicator to active column
		switch m.sortBy {
		case "name":
			nameHeader += sortIndicator
		case "size":
			sizeHeader += sortIndicator
		case "modified": // DeletedAt is stored in modTime
			deletedHeader += sortIndicator
		}

		header = fmt.Sprintf("%-25s  %-10s  %-12s  %-30s", nameHeader, sizeHeader, deletedHeader, locationHeader)
	} else if m.showFavoritesOnly {
		// Favorites mode: Name, Size, Modified, Location
		nameHeader := "Name"
		sizeHeader := "Size"
		modifiedHeader := "Modified"
		locationHeader := "Location"

		// Add indicator to active column
		switch m.sortBy {
		case "name":
			nameHeader += sortIndicator
		case "size":
			sizeHeader += sortIndicator
		case "modified":
			modifiedHeader += sortIndicator
		}

		header = fmt.Sprintf("%-25s  %-10s  %-12s  %-25s", nameHeader, sizeHeader, modifiedHeader, locationHeader)
	} else {
		// Regular mode: Name, Size, Modified, Type
		nameHeader := "Name"
		sizeHeader := "Size"
		modifiedHeader := "Modified"
		typeHeader := "Type"

		// Add indicator to active column
		switch m.sortBy {
		case "name":
			nameHeader += sortIndicator
		case "size":
			sizeHeader += sortIndicator
		case "modified":
			modifiedHeader += sortIndicator
		case "type":
			typeHeader += sortIndicator
		}

		header = fmt.Sprintf("%-30s  %-10s  %-12s  %-15s", nameHeader, sizeHeader, modifiedHeader, typeHeader)
	}

	// Render header with sort indicators
	s.WriteString(headerStyle.Render(header))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Calculate visible range
	start := 0
	end := len(files)

	if len(files) > maxVisible-1 { // -1 for header only (separator removed)
		start = m.cursor - (maxVisible-1)/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible - 1
		if end > len(files) {
			end = len(files)
			start = end - (maxVisible - 1)
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

		// Show different columns based on view mode
		var line string
		if m.showTrashOnly {
			// Trash mode: Name, Size, Deleted, Original Location
			deleted := formatModTime(file.modTime) // DeletedAt is stored in modTime

			// Look up original location from trash metadata
			location := "-"
			if trashItem, found := getTrashItemByPath(m.trashItems, file.path); found {
				location = filepath.Dir(trashItem.OriginalPath)
				// Shorten home directory to ~
				homeDir, _ := os.UserHomeDir()
				if homeDir != "" && strings.HasPrefix(location, homeDir) {
					location = "~" + strings.TrimPrefix(location, homeDir)
				}
				// Truncate long paths
				if len(location) > 28 {
					location = "..." + location[len(location)-25:]
				}
			}

			line = fmt.Sprintf("%-25s  %-10s  %-12s  %-30s", name, size, deleted, location)
		} else if m.showFavoritesOnly {
			// Favorites mode: Name, Size, Modified, Location
			// Get parent directory path for location
			location := filepath.Dir(file.path)
			// Shorten home directory to ~
			homeDir, _ := os.UserHomeDir()
			if homeDir != "" && strings.HasPrefix(location, homeDir) {
				location = "~" + strings.TrimPrefix(location, homeDir)
			}
			// Truncate long paths
			if len(location) > 23 {
				location = "..." + location[len(location)-20:]
			}
			line = fmt.Sprintf("%-25s  %-10s  %-12s  %-25s", name, size, modified, location)
		} else {
			// Regular mode: Name, Size, Modified, Type
			fileType := getFileType(file)
			line = fmt.Sprintf("%-30s  %-10s  %-12s  %-15s", name, size, modified, fileType)
		}

		style := fileStyle
		if file.isDir {
			style = folderStyle
		}
		if isClaudeContextFile(file.name) {
			style = claudeContextStyle
		}
		if isAgentsFile(file.name) {
			style = agentsStyle
		}
		if isPromptsFolder(file.name) || isGlobalPromptsVirtualFolder(file.name) {
			style = promptsFolderStyle
		}
		if file.isDir && isClaudePromptsSubfolder(file.name) {
			style = promptsFolderStyle
		}

		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else {
			// Add alternating row background for easier reading
			// Even rows (0, 2, 4...) get a subtle dark background
			if i%2 == 0 {
				alternateStyle := style.Copy().Background(lipgloss.Color("235")) // Very dark gray
				line = alternateStyle.Render(line)
			} else {
				line = style.Render(line)
			}
		}

		s.WriteString("  ")
		s.WriteString(line)
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	}

	return strings.TrimRight(s.String(), "\n")
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

			// Apply prompts filtering if active
			if m.showPromptsOnly {
				filteredSubFiles := make([]fileItem, 0)
				for _, subFile := range subFiles {
					if subFile.isDir {
						// Always include important dev folders
						importantFolders := []string{".claude", ".prompts", ".config"}
						isImportant := false
						for _, folder := range importantFolders {
							if subFile.name == folder {
								isImportant = true
								break
							}
						}
						if isImportant {
							filteredSubFiles = append(filteredSubFiles, subFile)
							continue
						}

						// Include directory if it contains prompt files
						if directoryContainsPrompts(subFile.path) {
							filteredSubFiles = append(filteredSubFiles, subFile)
						}
					} else if isPromptFile(subFile) {
						// Only include prompt files
						filteredSubFiles = append(filteredSubFiles, subFile)
					}
				}
				subFiles = filteredSubFiles
			}

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

// updateTreeItems rebuilds the tree items cache (called before rendering tree view)
func (m *model) updateTreeItems() {
	files := m.getFilteredFiles()
	m.treeItems = m.buildTreeItems(files, 0, []bool{})
}

// renderTreeView renders files in a hierarchical tree structure with expandable folders
func (m model) renderTreeView(maxVisible int) string {
	var s strings.Builder

	// Use cached tree items (should be updated before rendering)
	treeItems := m.treeItems

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

		if isAgentsFile(file.name) {
			style = agentsStyle
		}

		if isPromptsFolder(file.name) || isGlobalPromptsVirtualFolder(file.name) {
			style = promptsFolderStyle
		}

		if file.isDir && isClaudePromptsSubfolder(file.name) {
			style = promptsFolderStyle
		}

		// Add star indicator for favorites
		favIndicator := ""
		if m.isFavorite(file.path) {
			favIndicator = "⭐"
		}

		// Truncate long filenames to prevent wrapping
		displayName := file.name

		// Calculate available width dynamically based on view mode
		var maxNameLen int
		if m.viewMode == viewDualPane {
			// In dual-pane: use left pane width minus UI elements
			// Account for: indent, tree chars, icon, favorite, padding
			indentWidth := 2 + (item.depth * 3) + 3 + 2 + 2 + 5
			maxNameLen = m.leftWidth - indentWidth
		} else {
			// In single-pane: use full width minus UI elements
			indentWidth := 2 + (item.depth * 3) + 3 + 2 + 2 + 5
			maxNameLen = m.width - indentWidth
		}

		// Set reasonable bounds
		if maxNameLen < 20 {
			maxNameLen = 20
		}
		if maxNameLen > 100 {
			maxNameLen = 100 // Reasonable maximum
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
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	}

	return strings.TrimRight(s.String(), "\n")
}
