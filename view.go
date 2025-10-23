package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

func (m model) View() string {
	// Show landing page if active
	if m.showLandingPage {
		if m.landingPage != nil {
			return m.landingPage.Render()
		}
		return "Loading..."
	}

	// If fuzzy search is active, return empty string
	// (go-fzf handles its own rendering)
	if m.fuzzySearchActive {
		return ""
	}

	var baseView string

	// Dispatch to appropriate view based on viewMode
	switch m.viewMode {
	case viewFullPreview:
		baseView = m.renderFullPreview()
	case viewDualPane:
		baseView = m.renderDualPane()
	default:
		// Single-pane mode (original view)
		baseView = m.renderSinglePane()
	}

	// Overlay dropdown menu if open
	if m.menuOpen && m.activeMenu != "" {
		dropdown := m.renderActiveDropdown()
		if dropdown != "" {
			// Position dropdown below menu bar (line 1)
			menuX := m.getMenuXPosition(m.activeMenu)
			menuY := 1 // Below menu bar on line 0
			baseView = m.overlayDropdown(baseView, dropdown, menuX, menuY)
		}
	}

	// Overlay context menu if open
	if m.contextMenuOpen {
		menu := m.renderContextMenu()
		baseView = m.overlayContextMenu(baseView, menu)
	}

	// Overlay dialog if open
	if m.showDialog {
		dialog := m.renderDialog()
		baseView = m.overlayDialog(baseView, dialog)
	}

	return baseView
}

// renderSinglePane renders the original single-pane file browser
func (m model) renderSinglePane() string {
	var s strings.Builder

	// Check if we should show GitHub link (first 5 seconds) or menu bar
	showGitHub := time.Since(m.startupTime) < 5*time.Second

	if showGitHub {
		// Title with mode indicator and GitHub link (first 5 seconds)
		titleText := "(T)erminal (F)ile (E)xplorer"
		if m.commandFocused {
			titleText += " [Command Mode]"
		}
		if m.filePickerMode {
			titleText += " [📁 File Picker]"
		}

		// Create GitHub link (OSC 8 hyperlink format)
		githubURL := "https://github.com/GGPrompts/TFE"
		githubLink := fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", githubURL, githubURL)

		// Calculate spacing to right-align GitHub link
		githubText := githubURL // Display text
		availableWidth := m.width - len(titleText) - len(githubText) - 2
		if availableWidth < 1 {
			availableWidth = 1
		}
		spacing := strings.Repeat(" ", availableWidth)

		// Render title on left, GitHub link on right
		title := titleStyle.Render(titleText) + spacing + titleStyle.Render(githubLink)
		s.WriteString(title)
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	} else {
		// Show menu bar after 5 seconds
		menuBar := m.renderMenuBar()
		s.WriteString(menuBar)
		s.WriteString("\n")
	}

	// Toolbar buttons
	homeButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)

	// Home button (highlight if in home directory)
	homeDir, _ := os.UserHomeDir()
	if m.currentPath == homeDir {
		// Active: gray background (in home directory)
		activeHomeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeHomeStyle.Render("[🏠]"))
	} else {
		// Inactive: normal styling
		s.WriteString(homeButtonStyle.Render("[🏠]"))
	}
	s.WriteString(" ")

	// Favorites filter toggle button
	starIcon := "⭐"
	if m.showFavoritesOnly {
		starIcon = "✨" // Different icon when filter is active
	}
	s.WriteString(homeButtonStyle.Render("[" + starIcon + "]"))
	s.WriteString(" ")

	// View mode toggle button (cycles List → Detail → Tree)
	s.WriteString(homeButtonStyle.Render("[V]"))
	s.WriteString(" ")

	// Pane toggle button (toggles single ↔ dual-pane)
	paneIcon := "⬜"
	if m.viewMode == viewDualPane {
		paneIcon = "⬌"
	}
	s.WriteString(homeButtonStyle.Render("[" + paneIcon + "]"))
	s.WriteString(" ")

	// Command mode toggle button with green >_ and blue brackets
	if m.commandFocused {
		// Active: gray background
		bracketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Background(lipgloss.Color("237"))
		termStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Background(lipgloss.Color("237"))
		s.WriteString(bracketStyle.Render("["))
		s.WriteString(termStyle.Render(">_"))
		s.WriteString(bracketStyle.Render("]"))
	} else {
		// Inactive: normal styling
		bracketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		termStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
		s.WriteString(bracketStyle.Render("["))
		s.WriteString(termStyle.Render(">_"))
		s.WriteString(bracketStyle.Render("]"))
	}
	s.WriteString(" ")

	// Context-aware search button (in-file search when viewing, directory filter when browsing)
	// Highlight when search is active (directory filter in single-pane)
	if m.searchMode {
		// Active: gray background
		activeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeStyle.Render("[🔍]"))
	} else {
		// Inactive: normal styling
		s.WriteString(homeButtonStyle.Render("[🔍]"))
	}
	s.WriteString(" ")

	// Prompts filter toggle button
	if m.showPromptsOnly {
		// Active: gray background (like command mode)
		activeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeStyle.Render("[📝]"))
	} else {
		// Inactive: normal styling
		s.WriteString(homeButtonStyle.Render("[📝]"))
	}
	s.WriteString(" ")

	// Games launcher button
	s.WriteString(homeButtonStyle.Render("[🎮]"))
	s.WriteString(" ")

	// Trash/Recycle bin button
	trashIcon := "🗑️"
	if m.showTrashOnly {
		trashIcon = "♻️" // Recycle icon when viewing trash
	}
	s.WriteString(homeButtonStyle.Render("[" + trashIcon + "]"))

	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Command prompt with path (terminal-style)
	promptPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("$ ")
	pathPromptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	s.WriteString(promptPrefix)
	s.WriteString(pathPromptStyle.Render(getDisplayPath(m.currentPath)))
	s.WriteString(" ")

	// Show helper text based on focus state
	helperStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	if !m.commandFocused && m.commandInput == "" {
		// Not focused - show contextual hints
		if m.displayMode == modeDetail && m.isNarrowTerminal() {
			// Detail view on narrow terminal - show scroll hint
			s.WriteString(helperStyle.Render("←→ scroll | h/l nav | : focus"))
		} else {
			// Normal - show how to enter command mode
			s.WriteString(helperStyle.Render(": to focus"))
		}
	} else if m.commandFocused && m.commandInput == "" {
		// Focused but no input - show ! prefix hint and cursor
		s.WriteString(helperStyle.Render("! prefix to run & exit"))
		cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		s.WriteString(cursorStyle.Render("█"))
	} else {
		// Has input - show the command with cursor at correct position
		if m.commandFocused {
			// Render text before cursor, cursor, text after cursor
			beforeCursor := m.commandInput[:m.commandCursorPos]
			afterCursor := m.commandInput[m.commandCursorPos:]

			// Handle ! prefix coloring
			if strings.HasPrefix(beforeCursor, "!") {
				prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				s.WriteString(prefixStyle.Render("!"))
				s.WriteString(inputStyle.Render(beforeCursor[1:]))
			} else {
				s.WriteString(inputStyle.Render(beforeCursor))
			}

			// Render cursor
			cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
			s.WriteString(cursorStyle.Render("█"))

			// Render text after cursor
			s.WriteString(inputStyle.Render(afterCursor))
		} else {
			// Not focused - just show the text
			if strings.HasPrefix(m.commandInput, "!") {
				prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				s.WriteString(prefixStyle.Render("!"))
				s.WriteString(inputStyle.Render(m.commandInput[1:]))
			} else {
				s.WriteString(inputStyle.Render(m.commandInput))
			}
		}
	}
	// Explicitly reset styling after cursor to prevent ANSI code leakage
	s.WriteString("\033[0m")
	s.WriteString("\n")

	// Separator line between command prompt and file tree
	s.WriteString("\n")

	// File list - render based on current display mode
	// Calculate maxVisible to fit within terminal height:
	// title=1 + toolbar=1 + command=1 + separator=1 + filelist=maxVisible + spacer=1 + status=2 = m.height
	// Account for border (2 lines for top/bottom)
	// Therefore: maxVisible = m.height - 9 (total box height INCLUDING borders)
	maxVisible := m.height - 9

	// Content area is maxVisible - 2 (accounting for top/bottom borders)
	contentHeight := maxVisible - 2

	// Get file list content - use contentHeight so content fits within the box
	var fileListContent string
	switch m.displayMode {
	case modeList:
		fileListContent = m.renderListView(contentHeight)
	case modeDetail:
		fileListContent = m.renderDetailView(contentHeight)
	case modeTree:
		fileListContent = m.renderTreeView(contentHeight)
	default:
		fileListContent = m.renderDetailView(contentHeight) // Default to detail view
	}

	// Wrap content in a bordered box with fixed dimensions
	// Content is constrained to contentHeight lines to fit within the box
	fileListStyle := lipgloss.NewStyle().
		Width(m.width - 6).       // Leave margin for padding
		Height(contentHeight).    // Content area height (borders added by Lipgloss)
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{
			Light: "#0087d7", // Dark blue border for light
			Dark:  "#5fd7ff",  // Bright cyan border for dark
		})

	s.WriteString(fileListStyle.Render(fileListContent))
	s.WriteString("\n")

	// Check if we should show status message (auto-dismiss after 3s, except in edit mode)
	if m.statusMessage != "" && (m.promptEditMode || time.Since(m.statusTime) < 3*time.Second) {
		msgStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("28")). // Green
			Foreground(lipgloss.Color("0")).
			Bold(true).
			Padding(0, 1)

		if m.statusIsError {
			msgStyle = msgStyle.Background(lipgloss.Color("196")) // Red
		}

		s.WriteString(msgStyle.Render(m.statusMessage))
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n") // Add blank line to maintain 2-line height
		s.WriteString(" ") // Empty second line for consistent layout
	} else if m.searchMode || m.searchQuery != "" {
		// Show search status
		searchStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("33")). // Blue background
			Foreground(lipgloss.Color("255")). // Bright white for high contrast
			Bold(true).
			Padding(0, 1)

		// Calculate match count (exclude parent directory "..")
		matchCount := len(m.filteredIndices)
		if matchCount > 0 {
			matchCount-- // Exclude ".." which is always included
		}

		var searchStatus string
		if m.searchMode {
			// Active search mode with cursor
			searchStatus = fmt.Sprintf("Search: %s█ (%d matches)", m.searchQuery, matchCount)
		} else {
			// Search accepted (filter active but not in input mode)
			searchStatus = fmt.Sprintf("Filtered: %s (%d matches)", m.searchQuery, matchCount)
		}

		s.WriteString(searchStyle.Render(searchStatus))
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n") // Add blank line to maintain 2-line height
		s.WriteString(" ") // Empty second line for consistent layout
	} else {
		// Regular status bar
		// Count directories and files
		dirCount, fileCount := 0, 0
		for _, f := range m.files {
			if f.name == ".." {
				continue
			}
			if f.isDir {
				dirCount++
			} else {
				fileCount++
			}
		}

		// Selected file info
		var selectedInfo string
		if currentFile := m.getCurrentFile(); currentFile != nil {
			if currentFile.isDir {
				// Special handling for ".." to show parent directory name
				if currentFile.name == ".." {
					parentPath := filepath.Dir(m.currentPath)
					parentName := filepath.Base(parentPath)
					if parentName == "/" || parentName == "." {
						parentName = "root"
					}
					selectedInfo = fmt.Sprintf("Selected: .. (go up to %s)", parentName)
				} else {
					selectedInfo = fmt.Sprintf("Selected: %s (folder)", currentFile.name)
				}
			} else {
				fileType := getFileType(*currentFile)

				// For symlinks, truncate long paths to show the important trailing part
				if currentFile.isSymlink && currentFile.symlinkTarget != "" {
					// Calculate available space: terminal width minus other info
					// "Selected: filename (, size, date)"
					baseInfoLen := len("Selected: ") + len(currentFile.name) + len(", ") +
						len(formatFileSize(currentFile.size)) + len(", ") +
						len(formatModTime(currentFile.modTime)) + len(" ()") + 10 // padding

					availableForTarget := m.width - baseInfoLen
					if availableForTarget < 30 {
						availableForTarget = 30 // Minimum to show something useful
					}

					fullTarget := "Link → " + currentFile.symlinkTarget
					if len(fullTarget) > availableForTarget {
						// Show trailing end: "...filename" instead of "Link → /very/long/pa..."
						fileType = "..." + fullTarget[len(fullTarget)-(availableForTarget-3):]
					}
				}

				selectedInfo = fmt.Sprintf("Selected: %s (%s, %s, %s)",
					currentFile.name,
					fileType,
					formatFileSize(currentFile.size),
					formatModTime(currentFile.modTime))
			}
		}

		itemsInfo := fmt.Sprintf("%d items", len(m.files))
		if dirCount > 0 || fileCount > 0 {
			itemsInfo = fmt.Sprintf("%d folders, %d files", dirCount, fileCount)
		}

		hiddenIndicator := ""
		if m.showHidden {
			hiddenIndicator = " • showing hidden"
		}

		favoritesIndicator := ""
		if m.showFavoritesOnly {
			favoritesIndicator = " • ⭐ favorites only"
		}

		promptsIndicator := ""
		if m.showPromptsOnly {
			promptsIndicator = " • 📝 prompts only"
		}

		// View mode indicator
		viewModeText := fmt.Sprintf(" • view: %s", m.displayMode.String())

		// Help hint - show "/" search hint only when not already searching
		helpHint := " • F1: help"
		if !m.searchMode && m.searchQuery == "" {
			helpHint += " • /: search"
		}

		// Split status into two lines to prevent truncation
		// Line 1: Counts, indicators, view mode, help
		statusLine1 := fmt.Sprintf("%s%s%s%s%s%s", itemsInfo, hiddenIndicator, favoritesIndicator, promptsIndicator, viewModeText, helpHint)
		s.WriteString(statusStyle.Render(statusLine1))
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")

		// Line 2: Selected file info
		statusLine2 := selectedInfo
		s.WriteString(statusStyle.Render(statusLine2))
		s.WriteString("\033[0m") // Reset ANSI codes
	}

	return s.String()
}

// overlayContextMenu embeds the context menu into the base view at the correct position
// This approach works with Bubble Tea's diff-based rendering without needing tea.ClearScreen
func (m model) overlayContextMenu(baseView, menuContent string) string {
	x, y := m.contextMenuX, m.contextMenuY

	// Calculate actual menu height (number of lines + 2 for borders)
	menuLines := strings.Split(strings.TrimSpace(menuContent), "\n")
	menuHeight := len(menuLines)

	// Calculate menu width for horizontal bounds checking
	menuWidth := 25 // Default estimate
	if len(menuLines) > 0 {
		menuWidth = lipgloss.Width(menuLines[0])
	}

	// Ensure menu stays on screen with proper margins
	if x < 1 {
		x = 1
	}
	if x > m.width-menuWidth {
		x = m.width - menuWidth
	}
	if y < 1 {
		y = 1
	}

	// Dynamic height checking: reposition upward if menu would extend off bottom
	// Leave 3 lines buffer to avoid collision with file tree bottom border
	maxY := m.height - menuHeight - 3
	if y > maxY {
		// Try to position above the click point instead
		newY := m.contextMenuY - menuHeight
		if newY >= 1 {
			y = newY
		} else {
			// If it doesn't fit above either, clamp to maxY
			y = maxY
			if y < 1 {
				y = 1
			}
		}
	}

	// Split base view into lines (menuLines already calculated above)
	baseLines := strings.Split(baseView, "\n")

	// Ensure we have enough base lines
	for len(baseLines) < m.height {
		baseLines = append(baseLines, "")
	}

	// Overlay each menu line onto the base view
	for i, menuLine := range menuLines {
		targetLine := y + i
		if targetLine < 0 || targetLine >= len(baseLines) {
			continue
		}

		baseLine := baseLines[targetLine]

		// We need to overlay menuLine at visual column x
		// Use a string builder to construct the new line
		var newLine strings.Builder

		// Get the part of baseLine before position x
		// We need to handle ANSI codes properly
		visualPos := 0
		bytePos := 0
		inAnsi := false
		baseRunes := []rune(baseLine)

		// Scan through base line until we reach visual position x
		// Use runewidth to properly handle wide characters (emoji like ⭐)
		for bytePos < len(baseRunes) && visualPos < x {
			if baseRunes[bytePos] == '\033' {
				inAnsi = true
			} else if inAnsi {
				if (baseRunes[bytePos] >= 'A' && baseRunes[bytePos] <= 'Z') ||
					(baseRunes[bytePos] >= 'a' && baseRunes[bytePos] <= 'z') {
					inAnsi = false
				}
			} else {
				// Use RuneWidth to get actual visual width (handles wide emoji)
				visualPos += runewidth.RuneWidth(baseRunes[bytePos])
			}
			bytePos++
		}

		// Add the left part of the base line (up to position x)
		if bytePos > 0 && bytePos <= len(baseRunes) {
			newLine.WriteString(string(baseRunes[:bytePos]))
		}

		// Pad with spaces if needed to reach position x (fixes empty space alignment)
		// This ensures we always reach position x, even on empty lines
		for visualPos < x {
			newLine.WriteRune(' ')
			visualPos++
		}

		// Add the menu line
		newLine.WriteString(menuLine)

		baseLines[targetLine] = newLine.String()
	}

	return strings.Join(baseLines, "\n")
}

// overlayDropdown overlays a dropdown menu on the base view at the specified position
// Uses proper ANSI-aware overlay to preserve background content
func (m model) overlayDropdown(baseView, dropdown string, x, y int) string {
	// Split base view into lines
	baseLines := strings.Split(baseView, "\n")
	dropdownLines := strings.Split(dropdown, "\n")

	// Ensure we have enough base lines
	for len(baseLines) < m.height {
		baseLines = append(baseLines, "")
	}

	// Overlay each dropdown line onto the base view
	for i, dropdownLine := range dropdownLines {
		targetLine := y + i
		if targetLine < 0 || targetLine >= len(baseLines) {
			continue
		}

		baseLine := baseLines[targetLine]

		// We need to overlay dropdownLine at visual column x
		// Use a string builder to construct the new line
		var newLine strings.Builder

		// Get the part of baseLine before position x
		// We need to handle ANSI codes properly
		visualPos := 0
		bytePos := 0
		inAnsi := false
		baseRunes := []rune(baseLine)

		// Scan through base line until we reach visual position x
		// Use runewidth to properly handle wide characters
		for bytePos < len(baseRunes) && visualPos < x {
			if baseRunes[bytePos] == '\033' {
				inAnsi = true
			} else if inAnsi {
				if (baseRunes[bytePos] >= 'A' && baseRunes[bytePos] <= 'Z') ||
					(baseRunes[bytePos] >= 'a' && baseRunes[bytePos] <= 'z') {
					inAnsi = false
				}
			} else {
				// Use RuneWidth to get actual visual width
				visualPos += runewidth.RuneWidth(baseRunes[bytePos])
			}
			bytePos++
		}

		// Add the left part of the base line (up to position x)
		if bytePos > 0 && bytePos <= len(baseRunes) {
			newLine.WriteString(string(baseRunes[:bytePos]))
		}

		// Pad with spaces if needed to reach position x
		for visualPos < x {
			newLine.WriteRune(' ')
			visualPos++
		}

		// Add the dropdown line
		newLine.WriteString(dropdownLine)

		// Now preserve the right side of the base line (after the dropdown)
		dropdownWidth := lipgloss.Width(dropdownLine)
		endVisualPos := x + dropdownWidth

		// Continue from where we left off and skip to the end position
		for bytePos < len(baseRunes) && visualPos < endVisualPos {
			if baseRunes[bytePos] == '\033' {
				inAnsi = true
			} else if inAnsi {
				if (baseRunes[bytePos] >= 'A' && baseRunes[bytePos] <= 'Z') ||
					(baseRunes[bytePos] >= 'a' && baseRunes[bytePos] <= 'z') {
					inAnsi = false
				}
			} else {
				visualPos += runewidth.RuneWidth(baseRunes[bytePos])
			}
			bytePos++
		}

		// Add the remaining right part of the base line
		if bytePos < len(baseRunes) {
			newLine.WriteString(string(baseRunes[bytePos:]))
		}

		baseLines[targetLine] = newLine.String()
	}

	return strings.Join(baseLines, "\n")
}
