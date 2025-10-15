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

	return s.String()
}
