package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
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
		dialogOverlay := m.positionDialog(dialog)
		baseView += dialogOverlay
	}

	return baseView
}

// renderSinglePane renders the original single-pane file browser
func (m model) renderSinglePane() string {
	var s strings.Builder

	// Title
	title := titleStyle.Render("TFE - Terminal File Explorer")
	s.WriteString(title)
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Home button (path moved to command prompt line)
	homeButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		Render("[🏠]")
	s.WriteString(homeButtonStyle)
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Command prompt with path (terminal-style)
	promptPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("$ ")
	pathPromptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)

	s.WriteString(promptPrefix)
	s.WriteString(pathPromptStyle.Render(getDisplayPath(m.currentPath)))
	s.WriteString(" ")
	s.WriteString(inputStyle.Render(m.commandInput))
	// Always show cursor (MC-style: command prompt is always active)
	s.WriteString(cursorStyle.Render("█"))
	// Explicitly reset styling after cursor to prevent ANSI code leakage
	s.WriteString("\033[0m")
	s.WriteString("\n")

	// Separator line between command prompt and file tree
	s.WriteString("\n")

	// File list - render based on current display mode
	// Calculate maxVisible to fit within terminal height:
	// title=1 + path=1 + command=1 + separator=1 + filelist=maxVisible + spacer=1 + status=2 = m.height
	// Therefore: maxVisible = m.height - 7
	maxVisible := m.height - 7 // Reserve space for all UI elements (including 2-line status)

	switch m.displayMode {
	case modeList:
		s.WriteString(m.renderListView(maxVisible))
	case modeGrid:
		s.WriteString(m.renderGridView(maxVisible))
	case modeDetail:
		s.WriteString(m.renderDetailView(maxVisible))
	case modeTree:
		s.WriteString(m.renderTreeView(maxVisible))
	default:
		s.WriteString(m.renderListView(maxVisible))
	}

	// Status bar
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
				selectedInfo = fmt.Sprintf("Selected: %s (folder)", currentFile.name)
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

		// View mode indicator
		viewModeText := fmt.Sprintf(" • view: %s", m.displayMode.String())

		// Help hint
		helpHint := " • F1: help"

		// Split status into two lines to prevent truncation
		// Line 1: Counts, indicators, view mode, help
		statusLine1 := fmt.Sprintf("%s%s%s%s%s", itemsInfo, hiddenIndicator, favoritesIndicator, viewModeText, helpHint)
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
