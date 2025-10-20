package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
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

	// Title with mode indicator and GitHub link
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
	s.WriteString(homeButtonStyle.Render("[👁️]"))
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

	// Fuzzy search button
	s.WriteString(homeButtonStyle.Render("[🔍]"))
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

	// Show helper text when not focused and empty, otherwise show input
	if !m.commandFocused && m.commandInput == "" {
		helperStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		s.WriteString(helperStyle.Render(": to focus"))
	} else {
		s.WriteString(inputStyle.Render(m.commandInput))
	}

	// Show cursor only when command mode is active
	if m.commandFocused {
		cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		s.WriteString(cursorStyle.Render("█"))
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

	// Check if we should show status message (auto-dismiss after 3s)
	if m.statusMessage != "" && time.Since(m.statusTime) < 3*time.Second {
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

		// Help hint
		helpHint := " • F1: help"

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

	// Ensure menu stays on screen with proper margins
	if x < 1 {
		x = 1
	}
	if x > m.width-25 {
		x = m.width - 25
	}
	if y < 1 {
		y = 1
	}
	if y > m.height-10 {
		y = m.height - 10
	}

	// Split both views into lines
	baseLines := strings.Split(baseView, "\n")
	menuLines := strings.Split(strings.TrimSpace(menuContent), "\n")

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
		for bytePos < len(baseRunes) && visualPos < x {
			if baseRunes[bytePos] == '\033' {
				inAnsi = true
			}

			if inAnsi {
				if baseRunes[bytePos] >= 'A' && baseRunes[bytePos] <= 'Z' ||
					baseRunes[bytePos] >= 'a' && baseRunes[bytePos] <= 'z' {
					inAnsi = false
				}
			} else {
				visualPos++
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

		// Add the menu line
		newLine.WriteString(menuLine)

		baseLines[targetLine] = newLine.String()
	}

	return strings.Join(baseLines, "\n")
}
