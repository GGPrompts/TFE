package main

import (
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

// overlayDialog embeds the dialog into the base view at the centered position
// This approach works with Bubble Tea's diff-based rendering
func (m model) overlayDialog(baseView, dialogContent string) string {
	x, y := m.getDialogPosition()

	// Ensure dialog stays on screen with proper margins
	if x < 1 {
		x = 1
	}
	if y < 1 {
		y = 1
	}

	// Split both views into lines
	baseLines := strings.Split(baseView, "\n")
	dialogLines := strings.Split(strings.TrimSpace(dialogContent), "\n")

	// Ensure we have enough base lines
	for len(baseLines) < m.height {
		baseLines = append(baseLines, "")
	}

	// Overlay each dialog line onto the base view
	for i, dialogLine := range dialogLines {
		targetLine := y + i
		if targetLine < 0 || targetLine >= len(baseLines) {
			continue
		}

		baseLine := baseLines[targetLine]

		// We need to overlay dialogLine at visual column x
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

		// Add the dialog line
		newLine.WriteString(dialogLine)

		// Calculate where the dialog ends visually
		dialogWidth := visualWidth(dialogLine)
		endPos := x + dialogWidth

		// Preserve the rest of the base line after the dialog
		// Continue scanning from where we left off to find the visual end position
		for bytePos < len(baseRunes) && visualPos < endPos {
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

		// Append the rest of the base line (everything after the dialog)
		if bytePos < len(baseRunes) {
			newLine.WriteString(string(baseRunes[bytePos:]))
		}

		baseLines[targetLine] = newLine.String()
	}

	return strings.Join(baseLines, "\n")
}
