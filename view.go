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
		// Position menu using ANSI escape codes
		// Move cursor to menu position and render
		x, y := m.contextMenuX, m.contextMenuY

		// Ensure menu stays on screen with proper margins
		if x < 1 {
			x = 1 // Minimum left margin to show border
		}
		if x > m.width-25 {
			x = m.width - 25
		}
		if y < 1 {
			y = 1 // Minimum top margin
		}
		if y > m.height-10 {
			y = m.height - 10
		}

		// Replace newlines with newline + cursor positioning to maintain X coordinate
		// This keeps the menu together while fixing the alignment issue
		cursorToX := fmt.Sprintf("\n\033[%dG", x+1) // Move to column x+1 after each newline
		menuPositioned := strings.ReplaceAll(menu, "\n", cursorToX)

		// Position the start of the menu
		// ANSI escape codes use 1-based indexing, so add 1 to coordinates
		menuOverlay := fmt.Sprintf("\033[%d;%dH%s", y+1, x+1, menuPositioned)
		baseView += menuOverlay
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
	s.WriteString("\n")

	// Current path
	s.WriteString(pathStyle.Render(m.currentPath))
	s.WriteString("\n")

	// Command prompt (left-aligned on its own line)
	promptPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("$ ")
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)

	s.WriteString(promptPrefix)
	s.WriteString(inputStyle.Render(m.commandInput))
	// Always show cursor (MC-style: command prompt is always active)
	s.WriteString(cursorStyle.Render("█"))
	s.WriteString("\n")

	// Separator line between command prompt and file tree
	s.WriteString("\n")

	// File list - render based on current display mode
	// Calculate maxVisible to fit within terminal height:
	// title=1 + path=1 + command=1 + separator=1 + filelist=maxVisible + spacer=1 + status=1 = m.height
	// Therefore: maxVisible = m.height - 6
	maxVisible := m.height - 6 // Reserve space for all UI elements

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
				selectedInfo = fmt.Sprintf("Selected: %s (%s, %s)",
					currentFile.name,
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

		statusText := fmt.Sprintf("%s%s%s%s%s | %s", itemsInfo, hiddenIndicator, favoritesIndicator, viewModeText, helpHint, selectedInfo)
		s.WriteString(statusStyle.Render(statusText))
	}

	return s.String()
}
