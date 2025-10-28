package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Module: helpers.go
// Purpose: Helper functions for the model
// Responsibilities:
// - Getting currently selected file across different display modes
// - Utility functions for cursor management
// - Path formatting utilities

// getCurrentFile returns the currently selected file based on cursor position
// This handles the complexity of tree view with expanded folders
func (m model) getCurrentFile() *fileItem {
	if len(m.files) == 0 || m.cursor < 0 {
		return nil
	}

	// In tree view, we need to map cursor to the flattened tree
	if m.displayMode == modeTree {
		files := m.getFilteredFiles()
		treeItems := m.buildTreeItems(files, 0, []bool{})
		if m.cursor < len(treeItems) {
			return &treeItems[m.cursor].file
		}
		return nil
	}

	// In other views, use filtered files
	files := m.getFilteredFiles()
	if m.cursor < len(files) {
		return &files[m.cursor]
	}

	return nil
}

// getMaxCursor returns the maximum valid cursor position for the current display mode
func (m model) getMaxCursor() int {
	if m.displayMode == modeTree {
		files := m.getFilteredFiles()
		treeItems := m.buildTreeItems(files, 0, []bool{})
		return len(treeItems) - 1
	}

	files := m.getFilteredFiles()
	return len(files) - 1
}

// getDisplayPath returns a user-friendly path with home directory replaced by ~
func getDisplayPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	// Replace home directory with ~
	if strings.HasPrefix(path, homeDir) {
		if path == homeDir {
			return "~"
		}
		return "~" + strings.TrimPrefix(path, homeDir)
	}

	return path
}

// isDualPaneCompatible checks if the current display mode supports dual-pane view
// All display modes now support dual-pane with accordion layout
func (m model) isDualPaneCompatible() bool {
	return m.displayMode == modeList || m.displayMode == modeTree || m.displayMode == modeDetail
}

// isNarrowTerminal checks if the terminal width is too narrow for wide views
// Returns true if width < 100 (typical phone/Termux scenario)
func (m model) isNarrowTerminal() bool {
	return m.width < 100
}

// isPromptFile checks if a file is a prompt file (.prompty, .yaml, .md, .txt)
// Only files in special directories (.claude/, ~/.prompts/) are considered prompts
// Exception: .prompty files are always prompts (Microsoft Prompty format)
func isPromptFile(item fileItem) bool {
	if item.isDir {
		return false
	}

	ext := strings.ToLower(filepath.Ext(item.name))

	// .prompty is always a prompt file (Microsoft Prompty format)
	if ext == ".prompty" {
		return true
	}

	// For other extensions, only consider them prompts if in special directories
	if ext == ".md" || ext == ".yaml" || ext == ".yml" || ext == ".txt" {
		// Exclude .claude/agents/ - those are documentation files, not prompt templates
		if strings.Contains(item.path, "/.claude/agents/") {
			return false
		}

		// Check if in .claude/ or any subfolder
		if strings.Contains(item.path, "/.claude/") || strings.HasSuffix(item.path, "/.claude") {
			return true
		}
		// Check if in ~/.prompts/ or any subfolder
		homeDir, _ := os.UserHomeDir()
		promptsDir := filepath.Join(homeDir, ".prompts")
		if strings.HasPrefix(item.path, promptsDir) {
			return true
		}
	}

	return false
}

// performPreviewSearch searches the preview content for the current query
// and populates searchMatches with line numbers
func (m *model) performPreviewSearch() {
	m.preview.searchMatches = nil
	m.preview.currentMatch = -1

	if m.preview.searchQuery == "" {
		m.setStatusMessage("🔍 Search: (type to search, Enter/n: next, Esc: exit)", false)
		return
	}

	queryLower := strings.ToLower(m.preview.searchQuery)

	// Search through preview content
	for i, line := range m.preview.content {
		if strings.Contains(strings.ToLower(line), queryLower) {
			m.preview.searchMatches = append(m.preview.searchMatches, i)
		}
	}

	if len(m.preview.searchMatches) > 0 {
		m.preview.currentMatch = 0
		// Scroll to first match
		m.preview.scrollPos = m.preview.searchMatches[0]
		m.setStatusMessage(fmt.Sprintf("🔍 Found %d matches (1/%d) - n: next, Shift+n: prev, Esc: exit", len(m.preview.searchMatches), len(m.preview.searchMatches)), false)
	} else {
		m.setStatusMessage(fmt.Sprintf("🔍 No matches for '%s' - Esc: exit", m.preview.searchQuery), false)
	}
}

// findNextSearchMatch navigates to the next search match
func (m *model) findNextSearchMatch() {
	if len(m.preview.searchMatches) == 0 {
		m.setStatusMessage("🔍 No matches found", false)
		return
	}

	m.preview.currentMatch++
	if m.preview.currentMatch >= len(m.preview.searchMatches) {
		m.preview.currentMatch = 0 // Wrap around
	}

	// Scroll to the match
	m.preview.scrollPos = m.preview.searchMatches[m.preview.currentMatch]
	m.setStatusMessage(fmt.Sprintf("🔍 Match %d/%d - n: next, Shift+n: prev, Esc: exit", m.preview.currentMatch+1, len(m.preview.searchMatches)), false)
}

// findPreviousSearchMatch navigates to the previous search match
func (m *model) findPreviousSearchMatch() {
	if len(m.preview.searchMatches) == 0 {
		m.setStatusMessage("🔍 No matches found", false)
		return
	}

	m.preview.currentMatch--
	if m.preview.currentMatch < 0 {
		m.preview.currentMatch = len(m.preview.searchMatches) - 1 // Wrap around
	}

	// Scroll to the match
	m.preview.scrollPos = m.preview.searchMatches[m.preview.currentMatch]
	m.setStatusMessage(fmt.Sprintf("🔍 Match %d/%d - n: next, Shift+n: prev, Esc: exit", m.preview.currentMatch+1, len(m.preview.searchMatches)), false)
}

// getHelpSectionName returns the appropriate help section name based on current context
// This is used for context-aware F1 help navigation
func (m model) getHelpSectionName() string {
	// Check context in priority order (most specific first)
	if m.promptEditMode {
		return "## Prompt Templates & Fillable Fields"
	}
	if m.contextMenuOpen {
		return "## Context Menu"
	}
	if m.commandFocused {
		return "## Command Prompt (Vim-Style)"
	}
	if m.viewMode == viewFullPreview {
		return "## Preview & Full-Screen Mode"
	}
	if m.viewMode == viewDualPane {
		return "## Dual-Pane Mode"
	}
	// Default to Navigation section for single-pane mode
	return "## Navigation"
}

// findSectionLine searches for a section heading in content and returns its line number
// Returns -1 if not found, otherwise returns the 0-based line index
func findSectionLine(content []string, sectionName string) int {
	for i, line := range content {
		if strings.Contains(line, sectionName) {
			return i
		}
	}
	return -1
}

// autofillDefaults populates DATE and TIME variables with current values
// Called when entering prompt edit mode for the first time
func (m *model) autofillDefaults() {
	if m.preview.promptTemplate == nil {
		return
	}

	contextVars := getContextVariables(m)

	for _, varName := range m.preview.promptTemplate.variables {
		varNameLower := strings.ToLower(varName)

		// Auto-fill DATE and TIME from context
		if varNameLower == "date" || varNameLower == "time" {
			if value, exists := contextVars[varName]; exists {
				m.filledVariables[varName] = value
			}
		}
	}

	// Invalidate cache to force header re-render with auto-filled variables
	m.preview.cacheValid = false
	m.populatePreviewCache()
}

// scrollToFocusedVariable scrolls the preview to show the currently focused variable
// This is called when navigating between variables with Tab/Shift+Tab in edit mode
func (m *model) scrollToFocusedVariable() {
	if m.focusedVariableIndex < 0 || m.preview.promptTemplate == nil {
		return
	}

	if m.focusedVariableIndex >= len(m.preview.promptTemplate.variables) {
		return
	}

	// Get focused variable name
	varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]

	// We need to search in the RENDERED content (after processing and wrapping)
	// because scrollPos is applied to the wrapped lines, not raw content

	// Calculate box content width (same logic as renderPromptPreview)
	var boxContentWidth int
	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6
	} else {
		boxContentWidth = m.rightWidth - 2
	}

	// Calculate available width for content (prompts don't show line numbers)
	availableWidth := boxContentWidth - 2 // Just padding
	if availableWidth < 20 {
		availableWidth = 20
	}

	// Process content the same way as renderPromptPreview
	var contentLines []string
	if m.promptEditMode {
		// In edit mode, use rendered template with inline variables
		renderedTemplate := m.renderInlineVariables(m.preview.promptTemplate.template)
		contentLines = strings.Split(renderedTemplate, "\n")
	} else {
		// Before edit mode, use highlighted template
		highlightedTemplate := m.highlightVariablesBeforeEdit(m.preview.promptTemplate.template)
		contentLines = strings.Split(highlightedTemplate, "\n")
	}

	// Wrap content lines (same as renderPromptPreview)
	var wrappedLines []string
	for _, line := range contentLines {
		wrapped := wrapLine(line, availableWidth)
		wrappedLines = append(wrappedLines, wrapped...)
	}

	// Search for the variable in wrapped lines
	// Look for the variable name (it appears without {{}} in edit mode, or with {{}} before edit)
	targetLine := -1
	var searchPatterns []string

	if m.promptEditMode {
		// In edit mode, variables appear with ANSI styling (no brackets)
		// We're looking for the FOCUSED variable, which has specific ANSI codes:
		// Background 235 + Foreground 220
		//
		// IMPORTANT: Don't search for the full line content because long lines get wrapped
		// and no single wrapped line will contain the full pattern.
		// Instead, search for just the highlight ANSI codes + a short prefix.
		filledValue, hasFilled := m.filledVariables[varName]
		if hasFilled && filledValue != "" {
			// Search for focused highlight marker + first few chars
			// This works even when long lines are wrapped
			searchValue := filledValue
			if strings.Contains(filledValue, "\n") {
				searchValue = strings.Split(filledValue, "\n")[0]
			}
			// Use only first 20 runes (not bytes) to avoid wrap issues and Unicode breaks
			runes := []rune(searchValue)
			if len(runes) > 20 {
				searchValue = string(runes[:20])
			}
			searchPatterns = []string{
				fmt.Sprintf("\033[48;5;235m\033[38;5;220m%s", searchValue),
			}
		} else {
			// Search for focused variable with variable name (usually short, no wrap issue)
			searchPatterns = []string{
				fmt.Sprintf("\033[48;5;235m\033[38;5;220m%s\033[0m", varName),
			}
		}
	} else {
		// Before edit mode, variables appear with {{}}
		searchPatterns = []string{"{{" + varName + "}}"}
	}

	// Search in wrapped lines
	for i, line := range wrappedLines {
		for _, pattern := range searchPatterns {
			if strings.Contains(line, pattern) {
				targetLine = i
				break
			}
		}
		if targetLine >= 0 {
			break
		}
	}

	if targetLine >= 0 {
		// Calculate maxVisible the same way as renderDualPane/renderFullPreview
		var maxVisible int
		if m.viewMode == viewDualPane {
			headerLines := 4  // title + toolbar + command + blank separator
			footerLines := 4  // blank after panes + 2 status lines + optional message/search
			maxVisible = m.height - headerLines - footerLines
			if maxVisible < 5 {
				maxVisible = 5
			}
			// Account for borders
			maxVisible = maxVisible - 2
		} else if m.viewMode == viewFullPreview {
			maxVisible = m.height - 4 - 0 // Reserve space for header (if shown), help, and borders
			contentHeight := maxVisible - 2 // Content area accounting for borders
			maxVisible = contentHeight
		} else {
			// Shouldn't happen, but default to dual-pane calculation
			maxVisible = m.height - 10
		}

		// Calculate actual header height (same logic as renderPromptPreview)
		tmpl := m.preview.promptTemplate
		var actualHeaderHeight int

		// Count header lines based on prompt metadata
		if tmpl.name != "" {
			actualHeaderHeight += 1 // Name line (or more if wrapped, but simplified here)
			actualHeaderHeight += 1 // Blank line
		}
		if tmpl.description != "" {
			actualHeaderHeight += 1 // Description line (or more if wrapped)
			actualHeaderHeight += 1 // Blank line
		}
		actualHeaderHeight += 1 // Source indicator line
		if len(tmpl.variables) > 0 {
			actualHeaderHeight += 1 // Variables line (or more if wrapped)
		}
		actualHeaderHeight += 1 // Separator line

		// Content height is maxVisible minus actual header height
		contentHeight := maxVisible - actualHeaderHeight
		if contentHeight < 5 {
			contentHeight = 5
		}

		// Try to center the variable in the visible content area
		centerOffset := contentHeight / 2
		newScrollPos := targetLine - centerOffset

		if newScrollPos < 0 {
			newScrollPos = 0
		}

		// Don't scroll past the end of wrapped content
		maxScroll := len(wrappedLines) - contentHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		if newScrollPos > maxScroll {
			newScrollPos = maxScroll
		}

		m.preview.scrollPos = newScrollPos
	}
}

// navigateToPath changes the current path and automatically exits special modes (trash, favorites, etc)
// This ensures users don't get stuck in filter modes when navigating
func (m *model) navigateToPath(newPath string) {
	// If we're in trash mode and navigating away, auto-exit and restore
	if m.showTrashOnly {
		m.showTrashOnly = false
		// Don't change path if user is still in trash view
		// Only exit trash mode, stay in current directory
		if m.trashRestorePath != "" {
			m.currentPath = m.trashRestorePath
			m.trashRestorePath = ""
		}
		m.cursor = 0
		m.loadFiles()
		return
	}

	// Normal navigation
	m.currentPath = newPath
	m.cursor = 0
	m.loadFiles()
}

// getFileListVisibleLines returns the number of file items visible in the file list
// This accounts for header, footer, and borders
func (m model) getFileListVisibleLines() int {
	var visibleLines int

	if m.viewMode == viewDualPane {
		// Dual-pane mode: account for header, borders, footer
		visibleLines = m.height - 8  // Conservative estimate
	} else {
		// Single-pane mode: header (4) + footer (2-3)
		visibleLines = m.height - 6
	}

	if visibleLines < 1 {
		visibleLines = 1
	}

	return visibleLines
}

// getPreviewVisibleLines returns the number of content lines visible in the preview pane
// This accounts for headers, borders, and the scroll indicator line reservation in dual-pane mode
func (m model) getPreviewVisibleLines() int {
	totalLines := m.getWrappedLineCount()
	if totalLines == 0 {
		return 0
	}

	var visibleLines int

	if m.viewMode == viewFullPreview {
		// Full preview mode: m.height - 4 (borders/help) - headerLines (title/info when mouse enabled)
		headerLines := 0
		if m.previewMouseEnabled {
			headerLines = 2
		}
		maxVisible := m.height - 4 - headerLines
		visibleLines = maxVisible - 2 // Account for border
	} else if m.viewMode == viewDualPane {
		// Dual-pane mode calculation must match renderPreview() targetLines exactly!
		// Layout: header(4) + panes(maxVisible) + footer(4) = m.height
		// maxVisible = m.height - 8
		// contentHeight = maxVisible - 2 (borders) = m.height - 10
		// targetLines = contentHeight - 1 (scroll indicator) = m.height - 11
		// This fixes the "missing 3 lines" bug where getPreviewVisibleLines() was
		// returning m.height - 8, but renderPreview() was using m.height - 11.
		visibleLines = m.height - 11
		// Reserve one more line if content fits (no scroll indicator needed)
		if totalLines <= visibleLines+1 {
			visibleLines++ // No scroll indicator needed, can show one more line
		}
	} else {
		// Single-pane preview mode (shouldn't happen, but default to safe value)
		visibleLines = m.height - 6
	}

	if visibleLines < 1 {
		visibleLines = 1
	}

	return visibleLines
}

// renderToolbarRow renders the emoji button toolbar row
// Shows: [🏠] [⭐/✨] [📄/📊/🌲] [⬜/⬌] [>_] [🔍] [📝] [🔀] [🗑/♻]
// This function is shared between single-pane (view.go) and dual-pane (render_preview.go) views
func (m model) renderToolbarRow() string {
	var s strings.Builder

	homeDir, _ := os.UserHomeDir()

	// Home button (highlight if in home directory)
	homeIcon := "🏠"
	if m.currentPath == homeDir {
		// Active: gray background (in home directory)
		activeHomeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeHomeStyle.Render("[" + homeIcon + "]"))
	} else {
		// Inactive: normal styling
		homeButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(homeButtonStyle.Render("[" + homeIcon + "]"))
	}
	s.WriteString(" ")

	// Favorites filter toggle button
	starIcon := "⭐"
	if m.showFavoritesOnly {
		starIcon = "✨" // Different icon when filter is active
	}
	favButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(favButtonStyle.Render("[" + starIcon + "]"))
	s.WriteString(" ")

	// View mode toggle button (cycles List → Detail → Tree)
	// Show different emoji based on current display mode
	viewIcon := "📊" // Detail view (default)
	switch m.displayMode {
	case modeList:
		viewIcon = "📄" // Document icon for simple list view
	case modeDetail:
		viewIcon = "📊" // Bar chart icon for detailed columns
	case modeTree:
		viewIcon = "🌲" // Tree icon for hierarchical view
	}
	viewButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(viewButtonStyle.Render("[" + viewIcon + "]"))
	s.WriteString(" ")

	// Pane toggle button (toggles single ↔ dual-pane)
	paneIcon := "⬜"
	if m.viewMode == viewDualPane {
		paneIcon = "⬌"
	}
	paneButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(paneButtonStyle.Render("[" + paneIcon + "]"))
	s.WriteString(" ")

	// Command mode toggle button with green >_ and blue brackets
	if m.commandFocused {
		// Active: gray background
		bracketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Background(lipgloss.Color("237"))
		termStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true).Background(lipgloss.Color("237"))
		s.WriteString(bracketStyle.Render("["))
		s.WriteString(termStyle.Render(">_"))
		s.WriteString(bracketStyle.Render("]"))
	} else {
		// Inactive: normal styling
		bracketStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		termStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
		s.WriteString(bracketStyle.Render("["))
		s.WriteString(termStyle.Render(">_"))
		s.WriteString(bracketStyle.Render("]"))
	}
	s.WriteString(" ")

	// Context-aware search button (in-file search when viewing, directory filter when browsing)
	// Highlight when search is active (either in-file or directory filter)
	searchIcon := "🔍"
	if m.preview.searchActive || m.searchMode {
		// Active: gray background
		activeSearchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeSearchStyle.Render("[" + searchIcon + "]"))
	} else {
		// Inactive: normal styling
		searchButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(searchButtonStyle.Render("[" + searchIcon + "]"))
	}
	s.WriteString(" ")

	// Prompts filter toggle button
	promptIcon := "📝"
	if m.showPromptsOnly {
		// Active: gray background (like command mode)
		activeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeStyle.Render("[" + promptIcon + "]"))
	} else {
		// Inactive: normal styling
		promptButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(promptButtonStyle.Render("[" + promptIcon + "]"))
	}
	s.WriteString(" ")

	// Git repositories toggle button
	gitIcon := "🔀"
	if m.showGitReposOnly {
		// Active: gray background (like other active toggles)
		activeStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeStyle.Render("[" + gitIcon + "]"))
	} else {
		// Inactive: normal styling
		gitButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(gitButtonStyle.Render("[" + gitIcon + "]"))
	}
	s.WriteString(" ")

	// Trash/Recycle bin button
	trashIcon := "🗑"
	if m.showTrashOnly {
		trashIcon = "♻" // Recycle icon when viewing trash
	}
	trashButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(trashButtonStyle.Render("[" + trashIcon + "]"))

	return s.String()
}
