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

		// Override with orange color if it's a Claude context file or global .claude virtual folder
		if isClaudeContextFile(file.name) || isGlobalClaudeVirtualFolder(file.name) {
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

		// Override with teal if it's an Obsidian vault
		if file.isDir && isObsidianVault(file.path) {
			style = obsidianVaultStyle
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

		// Build the line with special handling for global virtual folders to preserve emoji color
		var line string
		if isGlobalPromptsVirtualFolder(file.name) || isGlobalClaudeVirtualFolder(file.name) {
			// Extract the leading emoji and render it separately to preserve its color
			var leadingEmoji string
			var restOfName string
			if strings.HasPrefix(displayName, "🌐 ") {
				leadingEmoji = "🌐 "
				restOfName = strings.TrimPrefix(displayName, "🌐 ")
			} else if strings.HasPrefix(displayName, "🤖 ") {
				leadingEmoji = "🤖 "
				restOfName = strings.TrimPrefix(displayName, "🤖 ")
			} else {
				restOfName = displayName
			}

			// Render the emoji without styling, then the rest with styling
			if i == m.cursor {
				line = fmt.Sprintf("  %s%s %s%s", icon, favIndicator, leadingEmoji, selectedStyle.Render(restOfName))
			} else {
				line = fmt.Sprintf("  %s%s %s%s", icon, favIndicator, leadingEmoji, style.Render(restOfName))
			}
		} else {
			// Normal rendering for all other files
			line = fmt.Sprintf("  %s%s %s", icon, favIndicator, displayName)

			// Apply selection style
			if i == m.cursor {
				line = selectedStyle.Render(line)
			} else {
				line = style.Render(line)
			}
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

	// Calculate available width for columns based on view mode
	availableWidth := m.width
	if m.viewMode == viewDualPane {
		availableWidth = m.leftWidth - 6 // Account for borders and padding
	}
	if availableWidth < 60 {
		availableWidth = 60 // Minimum width
	}

	// Distribute column widths dynamically (total must fit in availableWidth)
	// Leave space for icons (4), star (3), spacing (6), and padding (4) = 17 chars
	usableWidth := availableWidth - 17

	var nameWidth, sizeWidth, modifiedWidth, extraWidth int
	if m.showTrashOnly || m.showFavoritesOnly {
		// 4 columns: Name, Size, Modified/Deleted, Location
		nameWidth = usableWidth * 35 / 100    // 35%
		sizeWidth = 10                         // Fixed
		modifiedWidth = 12                     // Fixed
		extraWidth = usableWidth - nameWidth - sizeWidth - modifiedWidth
		if extraWidth < 15 {
			extraWidth = 15
		}
	} else {
		// 4 columns: Name, Size, Modified, Type
		nameWidth = usableWidth * 40 / 100    // 40%
		sizeWidth = 10                         // Fixed
		modifiedWidth = 12                     // Fixed
		extraWidth = 15                        // Type is usually short
	}

	// Ensure minimum widths
	if nameWidth < 15 {
		nameWidth = 15
	}

	// Account for icon (2) + potential star (3) in display name
	// The name field includes these, so effective text width is smaller
	maxNameTextLen := nameWidth - 5

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

	// Build header with sort indicators using dynamic widths
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

		header = fmt.Sprintf("%-*s  %-*s  %-*s  %-*s", nameWidth, nameHeader, sizeWidth, sizeHeader, modifiedWidth, deletedHeader, extraWidth, locationHeader)
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

		header = fmt.Sprintf("%-*s  %-*s  %-*s  %-*s", nameWidth, nameHeader, sizeWidth, sizeHeader, modifiedWidth, modifiedHeader, extraWidth, locationHeader)
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

		header = fmt.Sprintf("%-*s  %-*s  %-*s  %-*s", nameWidth, nameHeader, sizeWidth, sizeHeader, modifiedWidth, modifiedHeader, extraWidth, typeHeader)
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

		// Truncate long names based on dynamic width
		displayName := file.name
		// Use the pre-calculated maxNameTextLen which accounts for icon + star
		if maxNameTextLen < 10 {
			maxNameTextLen = 10
		}
		if len(displayName) > maxNameTextLen {
			displayName = displayName[:maxNameTextLen-2] + ".."
		}

		// Extract leading emoji for global virtual folders to preserve color
		var nameLeadingEmoji string
		var nameWithoutEmoji string
		if isGlobalPromptsVirtualFolder(file.name) || isGlobalClaudeVirtualFolder(file.name) {
			if strings.HasPrefix(displayName, "🌐 ") {
				nameLeadingEmoji = "🌐 "
				nameWithoutEmoji = strings.TrimPrefix(displayName, "🌐 ")
			} else if strings.HasPrefix(displayName, "🤖 ") {
				nameLeadingEmoji = "🤖 "
				nameWithoutEmoji = strings.TrimPrefix(displayName, "🤖 ")
			}
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
				// Truncate long paths based on dynamic width
				if len(location) > extraWidth {
					location = "..." + location[len(location)-(extraWidth-3):]
				}
			}

			// Use visual-width padding for name column (contains emojis), regular padding for others
		paddedName := padToVisualWidth(name, nameWidth)
		line = fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedName, sizeWidth, size, modifiedWidth, deleted, extraWidth, location)
		} else if m.showFavoritesOnly {
			// Favorites mode: Name, Size, Modified, Location
			// Get parent directory path for location
			location := filepath.Dir(file.path)
			// Shorten home directory to ~
			homeDir, _ := os.UserHomeDir()
			if homeDir != "" && strings.HasPrefix(location, homeDir) {
				location = "~" + strings.TrimPrefix(location, homeDir)
			}
			// Truncate long paths based on dynamic width
			if len(location) > extraWidth {
				location = "..." + location[len(location)-(extraWidth-3):]
			}
			// Use visual-width padding for name column (contains emojis), regular padding for others
		paddedName := padToVisualWidth(name, nameWidth)
		line = fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedName, sizeWidth, size, modifiedWidth, modified, extraWidth, location)
		} else {
			// Regular mode: Name, Size, Modified, Type
			fileType := getFileType(file)
			// Truncate file type if needed
			if len(fileType) > extraWidth {
				fileType = fileType[:extraWidth-2] + ".."
			}
			// Use visual-width padding for name column (contains emojis), regular padding for others
		paddedName := padToVisualWidth(name, nameWidth)
		line = fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedName, sizeWidth, size, modifiedWidth, modified, extraWidth, fileType)
		}

		style := fileStyle
		if file.isDir {
			style = folderStyle
		}
		if isClaudeContextFile(file.name) || isGlobalClaudeVirtualFolder(file.name) {
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
		if file.isDir && isObsidianVault(file.path) {
			style = obsidianVaultStyle
		}

		// Apply styling with special handling for global virtual folders to preserve emoji color
		if nameLeadingEmoji != "" {
			// Replace the styled portion, preserving the emoji color
			plainNameWithEmoji := fmt.Sprintf("%s%s %s%s", icon, favIndicator, nameLeadingEmoji, nameWithoutEmoji)

			if i == m.cursor {
				line = strings.Replace(line, plainNameWithEmoji, fmt.Sprintf("%s%s %s%s", icon, favIndicator, nameLeadingEmoji, selectedStyle.Render(nameWithoutEmoji)), 1)
			} else {
				if i%2 == 0 {
					alternateStyle := style.Copy().Background(lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: "#333333"})
					line = strings.Replace(line, plainNameWithEmoji, fmt.Sprintf("%s%s %s%s", icon, favIndicator, nameLeadingEmoji, alternateStyle.Render(nameWithoutEmoji)), 1)
				} else {
					line = strings.Replace(line, plainNameWithEmoji, fmt.Sprintf("%s%s %s%s", icon, favIndicator, nameLeadingEmoji, style.Render(nameWithoutEmoji)), 1)
				}
			}
		} else {
			// Normal rendering
			if i == m.cursor {
				line = selectedStyle.Render(line)
			} else {
				// Add alternating row background for easier reading
				// Even rows (0, 2, 4...) get a subtle background
				if i%2 == 0 {
					alternateStyle := style.Copy().Background(lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: "#333333"})
					line = alternateStyle.Render(line)
				} else {
					line = style.Render(line)
				}
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

			// Apply favorites filtering if active
			// Only show contents of favorited folders (let users explore inside their favorites)
			if m.showFavoritesOnly && !m.isFavorite(file.path) {
				filteredSubFiles := make([]fileItem, 0)
				for _, subFile := range subFiles {
					if m.isFavorite(subFile.path) {
						filteredSubFiles = append(filteredSubFiles, subFile)
					}
				}
				subFiles = filteredSubFiles
			}

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

		if isClaudeContextFile(file.name) || isGlobalClaudeVirtualFolder(file.name) {
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

		if file.isDir && isObsidianVault(file.path) {
			style = obsidianVaultStyle
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

		// Build the line with special handling for global virtual folders to preserve emoji color
		var line string
		if isGlobalPromptsVirtualFolder(file.name) || isGlobalClaudeVirtualFolder(file.name) {
			// Extract the leading emoji and render it separately to preserve its color
			var leadingEmoji string
			var restOfName string
			if strings.HasPrefix(displayName, "🌐 ") {
				leadingEmoji = "🌐 "
				restOfName = strings.TrimPrefix(displayName, "🌐 ")
			} else if strings.HasPrefix(displayName, "🤖 ") {
				leadingEmoji = "🤖 "
				restOfName = strings.TrimPrefix(displayName, "🤖 ")
			} else {
				restOfName = displayName
			}

			// Build base string with emoji uncolored
			baseString := fmt.Sprintf("%s%s%s%s%s %s", indent.String(), prefix, expansionIndicator, icon, favIndicator, leadingEmoji)

			// Render the emoji without styling, then the rest with styling
			if i == m.cursor {
				line = baseString + selectedStyle.Render(restOfName)
			} else {
				line = baseString + style.Render(restOfName)
			}
		} else {
			// Normal rendering for all other files
			line = fmt.Sprintf("%s%s%s%s%s %s", indent.String(), prefix, expansionIndicator, icon, favIndicator, displayName)

			if i == m.cursor {
				line = selectedStyle.Render(line)
			} else {
				line = style.Render(line)
			}
		}

		s.WriteString(line)
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	}

	return strings.TrimRight(s.String(), "\n")
}
