package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Adaptive color definitions - work in both light and dark terminals
var (
	// Title bar styling (left-aligned, no padding)
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#0087d7", // Dark blue for light backgrounds
			Dark:  "#5fd7ff", // Bright cyan for dark backgrounds
		})

	// Path display styling
	pathStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#666666", // Medium gray for light
			Dark:  "#999999", // Light gray for dark
		}).
		PaddingLeft(2)

	// Status bar styling
	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#444444",
			Dark:  "#AAAAAA",
		}).
		PaddingLeft(2)

	// Selected item styling
	selectedStyle = lipgloss.NewStyle().
		Bold(true).
		Background(lipgloss.AdaptiveColor{
			Light: "#0087d7", // Dark blue background for light
			Dark:  "#00d7ff", // Bright cyan background for dark
		}).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#FFFFFF", // White text on dark blue
			Dark:  "#000000", // Black text on bright cyan
		})

	// Narrow terminal selection styling (matrix green - no background to prevent wrapping)
	narrowSelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{
			Light: "#00AF00", // Bright green for light backgrounds
			Dark:  "#00FF00", // Matrix green for dark backgrounds
		})

	// Folder styling
	folderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#005faf", // Dark blue for light
			Dark:  "#5fd7ff", // Bright cyan for dark
		})

	// File styling
	fileStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#000000", // Black for light
			Dark:  "#FFFFFF", // White for dark
		})

	// Claude context file styling (orange)
	claudeContextStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#D75F00", // Darker orange for light
			Dark:  "#FF8700", // Bright orange for dark
		})

	// Agents file styling (purple)
	agentsStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#8B00FF", // Darker purple for light
			Dark:  "#BD93F9", // Bright purple for dark
		})

	// Prompts folder styling (bright magenta/pink)
	promptsFolderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#D7005F", // Dark magenta for light
			Dark:  "#FF79C6", // Bright pink for dark
		})

	// Obsidian vault styling (teal/cyan)
	obsidianVaultStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{
			Light: "#008B8B", // Dark cyan for light
			Dark:  "#50FAE9", // Bright teal for dark
		})
)
