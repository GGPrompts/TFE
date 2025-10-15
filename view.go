package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	// Dispatch to appropriate view based on viewMode
	switch m.viewMode {
	case viewFullPreview:
		return m.renderFullPreview()
	case viewDualPane:
		return m.renderDualPane()
	default:
		// Single-pane mode (original view)
		return m.renderSinglePane()
	}
}

// renderSinglePane renders the original single-pane file browser
func (m model) renderSinglePane() string {
	var s strings.Builder

	// Debug: Write file count to understand if something is off
	// Title - ALWAYS render this
	title := titleStyle.Render("TFE - Terminal File Explorer")
	s.WriteString(title)
	s.WriteString("\n")

	// Current path
	s.WriteString(pathStyle.Render(m.currentPath))
	s.WriteString("\n")

	// File list - render based on current display mode
	// Calculate maxVisible: m.height - (title=1 + path=1 + path_padding=1 + status=1 + help=1) = m.height - 5
	// Note: pathStyle has PaddingBottom(1) which adds an extra rendered line
	maxVisible := m.height - 5 // Reserve space for title, path (with padding), status, and help

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
	if len(m.files) > 0 && m.cursor < len(m.files) {
		selected := m.files[m.cursor]
		if selected.isDir {
			selectedInfo = fmt.Sprintf("Selected: %s (folder)", selected.name)
		} else {
			selectedInfo = fmt.Sprintf("Selected: %s (%s, %s)",
				selected.name,
				formatFileSize(selected.size),
				formatModTime(selected.modTime))
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

	// View mode indicator
	viewModeText := fmt.Sprintf(" • view: %s", m.displayMode.String())

	statusText := fmt.Sprintf("%s%s%s | %s", itemsInfo, hiddenIndicator, viewModeText, selectedInfo)
	s.WriteString(statusStyle.Render(statusText))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	s.WriteString(helpStyle.Render("↑/↓: nav • enter: preview • E: edit • y/c: copy path • Tab: dual-pane • f: full • v: views • q: quit"))

	return s.String()
}
