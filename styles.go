package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			PaddingLeft(2)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(2)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"}).
			Background(lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}).
			Bold(true)

	folderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#005faf", Dark: "#5fd7ff"})

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#444444", Dark: "#d0d0d0"})

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(2)

	claudeContextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")) // Orange
)
