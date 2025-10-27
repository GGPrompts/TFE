package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
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

		// Show parent folder name for ".." entry
		if file.name == ".." {
			parentPath := filepath.Dir(m.currentPath)
			parentName := filepath.Base(parentPath)

			// Handle root directory edge case
			if parentPath == m.currentPath || parentName == "/" || parentName == "." {
				parentName = "root"
			}

			displayName = fmt.Sprintf(".. (%s)", parentName)
		}

		maxNameLen := 40 // Default for single-pane
		if m.viewMode == viewDualPane {
			// Check if using vertical split (narrow terminal) - need to account for box borders
			if m.isNarrowTerminal() {
				// Vertical split: box uses (m.width - 6) with borders (-2) = m.width - 8
				// Then subtract icon (2), spaces (2), and padding (6)
				maxNameLen = (m.width - 8) - 10
			} else {
				// Horizontal split: use left pane width
				maxNameLen = m.leftWidth - 10
			}
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
			// Don't highlight if command prompt is focused
			// Pad icon to 2 cells for consistent alignment
			paddedIcon := m.padIconToWidth(icon)
			if i == m.cursor && !m.commandFocused {
				line = fmt.Sprintf("  %s%s %s%s", paddedIcon, favIndicator, leadingEmoji, selectedStyle.Render(restOfName))
			} else {
				line = fmt.Sprintf("  %s%s %s%s", paddedIcon, favIndicator, leadingEmoji, style.Render(restOfName))
			}
		} else {
			// Normal rendering for all other files
			// Pad icon to 2 cells for consistent alignment across different emoji widths
			paddedIcon := m.padIconToWidth(icon)
			line = fmt.Sprintf("  %s%s %s", paddedIcon, favIndicator, displayName)

			// Apply selection style
			// Don't highlight if command prompt is focused
			if i == m.cursor && !m.commandFocused {
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
	// Must account for box borders and padding in BOTH single and dual-pane modes
	availableWidth := m.width
	if m.viewMode == viewDualPane {
		availableWidth = m.leftWidth - 6 // Account for borders and padding
	} else {
		// Single-pane mode: box in view.go has Width(m.width - 6) + Border()
		// Different terminals interpret lipgloss Width() differently:
		// - Windows Terminal: Width() = content width (borders added on top)
		// - WezTerm/Termux: Width() = total width (borders included in Width())
		// For WezTerm/Termux, we need to subtract borders (2 chars) from content area
		if m.terminalType == terminalWezTerm {
			availableWidth = m.width - 8 // Box width - borders (2) - margin (6) = m.width - 8
		} else {
			availableWidth = m.width - 6 // Windows Terminal and others
		}
	}

	// On narrow terminals, use fixed wide width for horizontal scrolling
	// This allows the full detail view to render and be scrollable
	renderWidth := availableWidth
	if m.isNarrowTerminal() && availableWidth < 120 {
		renderWidth = 120 // Fixed width for detail view on narrow terminals
	} else if availableWidth < 60 {
		renderWidth = 60 // Minimum width for wider terminals
	}

	// Distribute column widths dynamically (total must fit in renderWidth)
	// Leave space for icons (4), star (3), spacing (6), and padding (4) = 17 chars
	usableWidth := renderWidth - 17

	var nameWidth, sizeWidth, modifiedWidth, extraWidth int
	if m.showTrashOnly || m.showFavoritesOnly || m.showGitReposOnly {
		// 4 columns: Name, Size, Modified/Deleted, Location/Branch
		nameWidth = usableWidth * 35 / 100    // 35%
		sizeWidth = 10                         // Fixed
		modifiedWidth = 12                     // Fixed
		extraWidth = usableWidth - nameWidth - sizeWidth - modifiedWidth
		if extraWidth < 15 {
			extraWidth = 15
		}
	} else {
		// 4 columns: Name, Size, Modified, Type (or symlink target)
		nameWidth = usableWidth * 40 / 100    // 40%
		sizeWidth = 10                         // Fixed
		modifiedWidth = 12                     // Fixed
		// Make Type column dynamic too - symlink targets can be long paths
		extraWidth = usableWidth - nameWidth - sizeWidth - modifiedWidth
		if extraWidth < 15 {
			extraWidth = 15 // Minimum for "Type" header
		}
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

		paddedNameHeader := m.padToVisualWidth(nameHeader, nameWidth)
		header = fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedNameHeader, sizeWidth, sizeHeader, modifiedWidth, deletedHeader, extraWidth, locationHeader)
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

		paddedNameHeader := m.padToVisualWidth(nameHeader, nameWidth)
		header = fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedNameHeader, sizeWidth, sizeHeader, modifiedWidth, modifiedHeader, extraWidth, locationHeader)
	} else if m.showGitReposOnly {
		// Git repos mode: Name, Branch, Status, Last Commit
		nameHeader := "Name"
		branchHeader := "Branch"
		statusHeader := "Status"
		commitHeader := "Last Commit"

		// Add indicator to active column
		switch m.sortBy {
		case "name":
			nameHeader += sortIndicator
		case "branch":
			branchHeader += sortIndicator
		case "status":
			statusHeader += sortIndicator
		case "modified": // Use modified for commit time sorting
			commitHeader += sortIndicator
		}

		// Calculate dynamic widths for git repos
		// Name: 35%, Branch: 15%, Status: 20%, Last Commit: 30%
		branchWidth := usableWidth * 15 / 100
		statusWidth := usableWidth * 20 / 100
		commitWidth := usableWidth * 30 / 100

		// Apply minimum width constraints
		if branchWidth < 10 {
			branchWidth = 10
		}
		if statusWidth < 15 {
			statusWidth = 15
		}
		if commitWidth < 15 {
			commitWidth = 15
		}

		// Recalculate nameWidth after applying minimums
		nameWidth = usableWidth - branchWidth - statusWidth - commitWidth

		paddedNameHeader := m.padToVisualWidth(nameHeader, nameWidth)
		header = fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedNameHeader, branchWidth, branchHeader, statusWidth, statusHeader, commitWidth, commitHeader)
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

		paddedNameHeader := m.padToVisualWidth(nameHeader, nameWidth)
		header = fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedNameHeader, sizeWidth, sizeHeader, modifiedWidth, modifiedHeader, extraWidth, typeHeader)
	}

	// Add "  " prefix to header to match data row padding
	header = "  " + header

	// Apply styling to header - DO NOT scroll here!
	// The header will be scrolled along with data rows by applyHorizontalScroll() later
	headerLine := headerStyle.Render(header)
	s.WriteString(headerLine)
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

		// Show parent folder name for ".." entry
		if file.name == ".." {
			parentPath := filepath.Dir(m.currentPath)
			parentName := filepath.Base(parentPath)

			// Handle root directory edge case
			if parentPath == m.currentPath || parentName == "/" || parentName == "." {
				parentName = "root"
			}

			displayName = fmt.Sprintf(".. (%s)", parentName)
		}

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

		// Pad icon to 2 cells for consistent alignment across different emoji widths
		paddedIcon := m.padIconToWidth(icon)
		name := fmt.Sprintf("%s%s %s", paddedIcon, favIndicator, displayName)
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
		paddedName := m.padToVisualWidth(name, nameWidth)
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
		paddedName := m.padToVisualWidth(name, nameWidth)
		line = fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedName, sizeWidth, size, modifiedWidth, modified, extraWidth, location)
		} else if m.showGitReposOnly {
			// Git repos mode: Name (with path), Branch, Status, Last Commit
			// Get parent directory path for location
			location := filepath.Dir(file.path)
			// Shorten home directory to ~
			homeDir, _ := os.UserHomeDir()
			if homeDir != "" && strings.HasPrefix(location, homeDir) {
				location = "~" + strings.TrimPrefix(location, homeDir)
			}

			// Show path in name column if repo is not in current directory
			repoDisplayName := file.name
			if location != "~" && location != "." {
				// Show relative path from scan root
				if m.gitReposScanRoot != "" {
					relPath, err := filepath.Rel(m.gitReposScanRoot, file.path)
					if err == nil && relPath != file.name {
						repoDisplayName = relPath
					}
				}
			}

			// Truncate long paths if needed
			if len(repoDisplayName) > maxNameTextLen {
				// For paths, show trailing end
				repoDisplayName = "..." + repoDisplayName[len(repoDisplayName)-(maxNameTextLen-3):]
			}

			// Pad icon to 2 cells for consistent alignment across different emoji widths
			paddedIcon := m.padIconToWidth(icon)
			name = fmt.Sprintf("%s%s %s", paddedIcon, favIndicator, repoDisplayName)

			// Get branch, status, and last commit from fileItem git fields
			branch := "-"
			status := "-"
			lastCommit := "-"

			if file.isGitRepo && file.name != ".." {
				// Use cached git status from fileItem
				if file.gitBranch != "" {
					branch = file.gitBranch
				}

				// Format status using git fields
				gitStat := gitStatus{
					branch:        file.gitBranch,
					ahead:         file.gitAhead,
					behind:        file.gitBehind,
					dirty:         file.gitDirty,
					lastCommitTime: file.gitLastCommit,
				}
				status = formatGitStatus(gitStat)
				lastCommit = formatLastCommitTime(file.gitLastCommit)
			}

			// Calculate widths (same as header calculation)
			branchWidth := usableWidth * 15 / 100
			statusWidth := usableWidth * 20 / 100
			commitWidth := usableWidth * 30 / 100
			if branchWidth < 10 {
				branchWidth = 10
			}
			if statusWidth < 15 {
				statusWidth = 15
			}
			if commitWidth < 15 {
				commitWidth = 15
			}

			// Truncate if needed
			if len(branch) > branchWidth {
				branch = branch[:branchWidth-2] + ".."
			}
			if visualWidth(status) > statusWidth {
				status = truncateToWidth(status, statusWidth-2) + ".."
			}
			if len(lastCommit) > commitWidth {
				lastCommit = lastCommit[:commitWidth-2] + ".."
			}

			// Use visual-width padding for name column (contains emojis), regular padding for others
		paddedName := m.padToVisualWidth(name, nameWidth)
		line = fmt.Sprintf("%s  %-*s  %-*s  %-*s", paddedName, branchWidth, branch, statusWidth, status, commitWidth, lastCommit)
		} else {
			// Regular mode: Name, Size, Modified, Type
			fileType := getFileType(file)
			// Truncate file type if needed (show trailing end for long paths)
			if len(fileType) > extraWidth {
				// For symlinks showing paths, show the end (filename) rather than beginning
				if strings.HasPrefix(fileType, "Link → ") {
					// Show "...filename" instead of "Link → /very/long/pa..."
					fileType = "..." + fileType[len(fileType)-(extraWidth-3):]
				} else {
					// For regular types, truncate normally
					fileType = fileType[:extraWidth-2] + ".."
				}
			}
			// Use visual-width padding for name column (contains emojis), regular padding for others
		paddedName := m.padToVisualWidth(name, nameWidth)
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
			// Use paddedIcon for consistency
			plainNameWithEmoji := fmt.Sprintf("%s%s %s%s", paddedIcon, favIndicator, nameLeadingEmoji, nameWithoutEmoji)

			// Don't highlight if command prompt is focused
			if i == m.cursor && !m.commandFocused {
				if m.isNarrowTerminal() && renderWidth > availableWidth {
					// Use matrix green for narrow terminals (no background to prevent wrapping)
					line = strings.Replace(line, plainNameWithEmoji, fmt.Sprintf("%s%s %s%s", paddedIcon, favIndicator, nameLeadingEmoji, narrowSelectedStyle.Render(nameWithoutEmoji)), 1)
				} else {
					line = strings.Replace(line, plainNameWithEmoji, fmt.Sprintf("%s%s %s%s", paddedIcon, favIndicator, nameLeadingEmoji, selectedStyle.Render(nameWithoutEmoji)), 1)
				}
			} else {
				// Add alternating row background for easier reading on wide terminals
				// Disabled on narrow terminals to prevent wrapping issues with horizontal scroll
				if !m.isNarrowTerminal() && i%2 == 0 {
					alternateStyle := style.Copy().Background(lipgloss.AdaptiveColor{Light: "#eeeeee", Dark: "#333333"})
					line = strings.Replace(line, plainNameWithEmoji, fmt.Sprintf("%s%s %s%s", paddedIcon, favIndicator, nameLeadingEmoji, alternateStyle.Render(nameWithoutEmoji)), 1)
				} else {
					line = strings.Replace(line, plainNameWithEmoji, fmt.Sprintf("%s%s %s%s", paddedIcon, favIndicator, nameLeadingEmoji, style.Render(nameWithoutEmoji)), 1)
				}
			}
		} else {
			// Normal rendering
			// Don't highlight if command prompt is focused
			if i == m.cursor && !m.commandFocused {
				if m.isNarrowTerminal() && renderWidth > availableWidth {
					// Use matrix green for narrow terminals (no background to prevent wrapping)
					line = narrowSelectedStyle.Render(line)
				} else {
					// Use blue background for wide terminals
					line = selectedStyle.Render(line)
				}
			} else {
				// Add alternating row background for easier reading on wide terminals
				// Disabled on narrow terminals to prevent wrapping issues with horizontal scroll
				if !m.isNarrowTerminal() && i%2 == 0 {
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

	result := s.String()

	// Apply horizontal scrolling on narrow terminals
	if m.isNarrowTerminal() && renderWidth > availableWidth {
		result = m.applyHorizontalScroll(result, availableWidth)
	}

	return strings.TrimRight(result, "\n")
}

// applyHorizontalScroll applies horizontal scrolling to all lines of rendered text
// This is used for detail view on narrow terminals
// IMPORTANT: Ensures output lines never exceed viewWidth to prevent terminal wrapping
func (m model) applyHorizontalScroll(content string, viewWidth int) string {
	lines := strings.Split(content, "\n")
	var result strings.Builder

	for _, line := range lines {
		// Extract visible portion - viewWidth is the TOTAL visible width
		// The line already includes "  " padding at the start
		visibleLine := m.extractVisiblePortion(line, viewWidth)

		// Pad or truncate to exact width to prevent wrapping
		// IMPORTANT: Strip the trailing \033[0m, add padding, then re-add reset
		// This prevents padding from inheriting highlight colors
		plainVisible := stripANSI(visibleLine)
		visualLen := m.visualWidthCompensated(plainVisible) // Use terminal-aware visual width for emojis

		if visualLen < viewWidth {
			// Line is too short - add padding
			// Remove trailing ANSI reset if present
			resetSuffix := "\033[0m"
			hasReset := strings.HasSuffix(visibleLine, resetSuffix)
			if hasReset {
				visibleLine = visibleLine[:len(visibleLine)-len(resetSuffix)]
			}

			// Add padding with explicit reset to prevent color bleeding
			padding := viewWidth - visualLen // Use visual width difference
			visibleLine += resetSuffix + strings.Repeat(" ", padding)
		} else if visualLen > viewWidth {
			// Line is too long - truncate to prevent wrapping
			// Use ANSI-aware truncation that preserves styling
			visibleLine = m.truncateToVisualWidth(visibleLine, viewWidth)
		}

		result.WriteString(visibleLine)
		result.WriteString("\n")
	}

	return result.String()
}

// extractVisiblePortion extracts the visible portion of a line based on scroll offset
// Properly handles ANSI escape codes and multi-column characters (emojis)
func (m model) extractVisiblePortion(line string, viewWidth int) string {
	// Calculate visible window based on scroll offset
	scrollOffset := m.detailScrollX
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	// Build result by walking through line rune by rune
	// Track visible column position (accounting for wide chars like emojis)
	var result strings.Builder
	visibleCol := 0
	inEscape := false
	escapeSeq := strings.Builder{}

	// Track active ANSI codes to prepend them to result
	activeANSI := strings.Builder{}

	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Detect ANSI escape sequence start
		if r == '\x1b' {
			inEscape = true
			escapeSeq.Reset()
			escapeSeq.WriteRune(r)
			continue
		}

		// Inside ANSI escape sequence
		if inEscape {
			escapeSeq.WriteRune(r)
			// ANSI sequences end with a letter (m for color, other commands too)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
				// Save this ANSI code as active (for color/style)
				ansiCode := escapeSeq.String()
				if strings.Contains(ansiCode, "m") { // Color/style code
					activeANSI.WriteString(ansiCode)
				}
				// If we're in visible range, write ANSI code to result
				if visibleCol >= scrollOffset && visibleCol < scrollOffset+viewWidth {
					result.WriteString(ansiCode)
				}
			}
			continue
		}

		// Calculate visual width of this character
		// Most chars are 1 column, emojis/wide chars are 2 columns
		charWidth := m.runeWidth(r)

		// Check if this character is in the visible window
		if visibleCol+charWidth > scrollOffset && visibleCol < scrollOffset+viewWidth {
			// First character? Prepend active ANSI codes
			if result.Len() == 0 && activeANSI.Len() > 0 {
				result.WriteString(activeANSI.String())
			}

			// Only add the character if it fits completely in the visible window
			// This prevents splitting wide characters (emojis)
			if visibleCol >= scrollOffset {
				result.WriteRune(r)
			}
		}

		visibleCol += charWidth

		// Stop if we've passed the visible window
		if visibleCol >= scrollOffset+viewWidth {
			break
		}
	}

	// If we got nothing, return empty (scrolled past content)
	if result.Len() == 0 {
		return ""
	}

	// Always append ANSI reset at the end to prevent formatting bleeding
	// This ensures highlights/colors don't extend beyond the visible portion
	result.WriteString("\033[0m")

	return result.String()
}

// truncateToVisualWidth truncates a string (with ANSI codes) to a specific visual width
// Preserves ANSI styling codes while ensuring visual width doesn't exceed target
// Terminal-aware: applies WezTerm emoji compensation
func (m model) truncateToVisualWidth(s string, targetWidth int) string {
	var result strings.Builder
	visualWidth := 0
	inEscape := false
	escapeSeq := strings.Builder{}

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		// Detect ANSI escape sequence start
		if r == '\x1b' {
			inEscape = true
			escapeSeq.Reset()
			escapeSeq.WriteRune(r)
			continue
		}

		// Inside ANSI escape sequence
		if inEscape {
			escapeSeq.WriteRune(r)
			// ANSI sequences end with a letter
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
				// Write ANSI code to result (doesn't count toward visual width)
				result.WriteString(escapeSeq.String())
			}
			continue
		}

		// Calculate visual width of this character using terminal-aware method
		charWidth := m.runeWidth(r)

		// Check if adding this character would exceed target width
		if visualWidth + charWidth > targetWidth {
			// Reached target width - add reset and stop
			result.WriteString("\033[0m")
			break
		}

		// Add character and increment visual width
		result.WriteRune(r)
		visualWidth += charWidth
	}

	return result.String()
}

// runeWidth returns the visual width of a rune (1 for most, 2 for emojis/wide chars)
// Terminal-aware: Treats variation selectors correctly for Windows Terminal
// Delegates to runewidth library for consistent width calculations
func (m model) runeWidth(r rune) int {
	// Variation selectors have special handling based on terminal
	if r >= 0xFE00 && r <= 0xFE0F { // Variation selectors
		// runewidth reports VS as width 1, but Windows Terminal renders emoji+VS as 2 cells total
		// We return +1 for Windows Terminal to match its 2-cell rendering
		// WezTerm/Kitty/iTerm2/xterm/Termux render emoji+VS as 1 cell (matches runewidth), return 0
		if m.terminalType == terminalWindowsTerminal {
			return 1 // Compensate for Windows Terminal's wider rendering
		}
		return 0
	}
	if r >= 0x0300 && r <= 0x036F { // Combining diacritical marks
		return 0
	}
	if r >= 0x1AB0 && r <= 0x1AFF { // Combining diacritical marks extended
		return 0
	}
	if r >= 0x20D0 && r <= 0x20FF { // Combining diacritical marks for symbols
		return 0
	}

	// Delegate to runewidth library for all other characters
	// This ensures consistency with padToVisualWidth() and visualWidthCompensated()
	// which also use runewidth for their calculations
	return runewidth.RuneWidth(r)
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
						if m.directoryContainsPrompts(subFile.path) {
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

		// Show parent folder name for ".." entry
		if file.name == ".." {
			parentPath := filepath.Dir(m.currentPath)
			parentName := filepath.Base(parentPath)

			// Handle root directory edge case
			if parentPath == m.currentPath || parentName == "/" || parentName == "." {
				parentName = "root"
			}

			displayName = fmt.Sprintf(".. (%s)", parentName)
		}

		// Calculate available width dynamically based on view mode
		var maxNameLen int
		// Icon is always padded to 2 cells for consistent alignment
		iconWidth := 2
		indentWidth := 2 + (item.depth * 3) + 3 + iconWidth + 2 + 5

		if m.viewMode == viewDualPane {
			// Check if using vertical split (narrow terminal) - need to account for box borders
			if m.isNarrowTerminal() {
				// Vertical split: box uses (m.width - 6) with borders (-2) = m.width - 8
				// Then subtract UI elements
				maxNameLen = (m.width - 8) - indentWidth
			} else {
				// Horizontal split: use left pane width minus UI elements
				maxNameLen = m.leftWidth - indentWidth
			}
		} else {
			// Single-pane: use full width minus UI elements
			maxNameLen = m.width - indentWidth
		}

		// Set reasonable bounds
		// Allow very narrow widths when pane is narrow (important for accordion mode)
		if maxNameLen < 5 {
			maxNameLen = 5  // Absolute minimum (shows a few chars + "..")
		}
		if maxNameLen > 100 {
			maxNameLen = 100 // Reasonable maximum
		}

		if len(displayName) > maxNameLen {
			if maxNameLen > 2 {
				displayName = displayName[:maxNameLen-2] + ".."
			} else {
				displayName = displayName[:maxNameLen]  // Very narrow, no room for ".."
			}
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
			// Pad icon to 2 cells for consistent alignment
			paddedIcon := m.padIconToWidth(icon)
			baseString := fmt.Sprintf("%s%s%s%s%s %s", indent.String(), prefix, expansionIndicator, paddedIcon, favIndicator, leadingEmoji)

			// Render the emoji without styling, then the rest with styling
			// Don't highlight if command prompt is focused
			if i == m.cursor && !m.commandFocused {
				line = baseString + selectedStyle.Render(restOfName)
			} else {
				line = baseString + style.Render(restOfName)
			}
		} else {
			// Normal rendering for all other files
			// Pad icon to 2 cells for consistent alignment
			paddedIcon := m.padIconToWidth(icon)
			line = fmt.Sprintf("%s%s%s%s%s %s", indent.String(), prefix, expansionIndicator, paddedIcon, favIndicator, displayName)

			// Don't highlight if command prompt is focused
			if i == m.cursor && !m.commandFocused {
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
