package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Style variables — initialized by initStyles() in theme.go from the active theme.
// Do NOT hardcode colors here; all colors come from currentTheme.
var (
	// Title bar styling (left-aligned, no padding)
	titleStyle lipgloss.Style

	// Path display styling
	pathStyle lipgloss.Style

	// Status bar styling
	statusStyle lipgloss.Style

	// Selected item styling
	selectedStyle lipgloss.Style

	// Narrow terminal selection styling (matrix green - no background to prevent wrapping)
	narrowSelectedStyle lipgloss.Style

	// Folder styling
	folderStyle lipgloss.Style

	// File styling
	fileStyle lipgloss.Style

	// Claude context file styling (orange)
	claudeContextStyle lipgloss.Style

	// Agents file styling (purple)
	agentsStyle lipgloss.Style

	// Prompts folder styling (bright magenta/pink)
	promptsFolderStyle lipgloss.Style

	// Obsidian vault styling (teal/cyan)
	obsidianVaultStyle lipgloss.Style

	// Diff preview styles (git diff coloring in changes mode)
	diffAddedStyle      lipgloss.Style // Green for added lines (+)
	diffRemovedStyle    lipgloss.Style // Red for removed lines (-)
	diffHunkHeaderStyle lipgloss.Style // Cyan for @@ hunk headers
	diffMetaStyle       lipgloss.Style // Dim for diff/index/---/+++ headers
)
