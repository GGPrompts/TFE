package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// renderDialog renders a dialog overlay centered on screen
func (m model) renderDialog() string {
	switch m.dialog.dialogType {
	case dialogInput:
		return m.renderInputDialog()
	case dialogConfirm:
		return m.renderConfirmDialog()
	default:
		return ""
	}
}

// renderInputDialog renders a text input dialog
func (m model) renderInputDialog() string {
	// Calculate dialog dimensions
	width := 50
	if width > m.width-4 {
		width = m.width - 4
	}

	// Dialog styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(width)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(width)

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Width(width - 4)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(width)

	// Build dialog content
	var content strings.Builder
	content.WriteString(titleStyle.Render(m.dialog.title))
	content.WriteString("\n\n")
	if m.dialog.message != "" {
		content.WriteString(m.dialog.message)
		content.WriteString("\n\n")
	}
	content.WriteString(inputStyle.Render(m.dialog.input + "█"))
	content.WriteString("\n\n")
	content.WriteString(hintStyle.Render("Enter: confirm | Esc: cancel"))

	return borderStyle.Render(content.String())
}

// renderConfirmDialog renders a yes/no confirmation dialog
func (m model) renderConfirmDialog() string {
	// Calculate dialog dimensions
	width := 50
	if width > m.width-4 {
		width = m.width - 4
	}

	// Dialog styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")). // Red for warnings
		Padding(1, 2).
		Width(width)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196")).
		Align(lipgloss.Center).
		Width(width)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Width(width)

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(width)

	// Build dialog content
	var content strings.Builder
	content.WriteString(titleStyle.Render(m.dialog.title))
	content.WriteString("\n\n")
	content.WriteString(messageStyle.Render(m.dialog.message))
	content.WriteString("\n\n")
	content.WriteString(hintStyle.Render("[Y]es / [N]o / [Esc]"))

	return borderStyle.Render(content.String())
}

// getDialogPosition calculates centered position for dialog
func (m model) getDialogPosition() (int, int) {
	// Rough estimate: dialog is about 50 chars wide, 10 lines tall
	dialogWidth := 54  // 50 + border
	dialogHeight := 10 // Approximate

	x := (m.width - dialogWidth) / 2
	y := (m.height - dialogHeight) / 2

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	return x, y
}

// positionDialog wraps dialog content with ANSI positioning codes
func (m model) positionDialog(dialogContent string) string {
	x, y := m.getDialogPosition()

	// Replace newlines with newline + cursor positioning to maintain X coordinate
	// This keeps the dialog together while preserving lipgloss styling
	cursorToX := fmt.Sprintf("\n\033[%dG", x+1) // Move to column x+1 after each newline
	dialogPositioned := strings.ReplaceAll(dialogContent, "\n", cursorToX)

	// Position the start of the dialog
	// ANSI escape codes use 1-based indexing, so add 1 to coordinates
	return fmt.Sprintf("\033[%d;%dH%s", y+1, x+1, dialogPositioned)
}
