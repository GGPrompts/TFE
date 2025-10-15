package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func initialModel() model {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	m := model{
		currentPath:    cwd,
		cursor:         0,
		height:         24,
		width:          80,
		showHidden:     false,
		displayMode:    modeList,
		gridColumns:    4,
		sortBy:         "name",
		sortAsc:        true,
		viewMode:       viewSinglePane,
		focusedPane:    leftPane,
		lastClickIndex: -1,
		preview: previewModel{
			maxPreview: 10000, // Max 10k lines
		},
	}

	m.loadFiles()
	m.calculateGridLayout()
	m.calculateLayout()
	return m
}

// calculateGridLayout calculates how many columns fit in grid view
func (m *model) calculateGridLayout() {
	itemWidth := 15 // Estimated width per item (icon + name + padding)
	columns := m.width / itemWidth
	if columns < 1 {
		columns = 1
	}
	if columns > 8 {
		columns = 8 // Max 8 columns for readability
	}
	m.gridColumns = columns
}

// calculateLayout calculates left and right pane widths for dual-pane mode
func (m *model) calculateLayout() {
	if m.viewMode == viewSinglePane || m.viewMode == viewFullPreview {
		m.leftWidth = m.width
		m.rightWidth = 0
	} else {
		// 40/60 split for dual-pane
		m.leftWidth = m.width * 40 / 100
		m.rightWidth = m.width - m.leftWidth - 1 // -1 for separator
		if m.leftWidth < 20 {
			m.leftWidth = 20
		}
		if m.rightWidth < 30 {
			m.rightWidth = 30
		}
	}
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
