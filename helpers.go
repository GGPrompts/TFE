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

// openFileAsTab opens a file as a new tab (or switches to it if already open)
// Returns true if a new tab was created, false if switched to existing
func (m *model) openFileAsTab(path, name, gitStatus string) bool {
	// Check if tab already exists
	for i, tab := range m.tabs {
		if tab.path == path {
			m.activeTab = i
			m.loadPreview(tab.path)
			m.populatePreviewCache()
			return false
		}
	}

	// Add new tab
	m.tabs = append(m.tabs, openTab{
		path:      path,
		name:      name,
		gitStatus: gitStatus,
	})
	m.activeTab = len(m.tabs) - 1

	// Load preview for the new tab
	m.loadPreview(path)
	m.populatePreviewCache()
	return true
}

// closeActiveTab closes the currently active tab
func (m *model) closeActiveTab() {
	if len(m.tabs) == 0 {
		return
	}

	// Remove the active tab
	m.tabs = append(m.tabs[:m.activeTab], m.tabs[m.activeTab+1:]...)

	// Adjust active tab index
	if m.activeTab >= len(m.tabs) {
		m.activeTab = len(m.tabs) - 1
	}

	// Load the new active tab's content, or clear preview if no tabs remain
	if len(m.tabs) > 0 {
		tab := m.tabs[m.activeTab]
		m.loadPreview(tab.path)
		m.populatePreviewCache()
	} else {
		m.preview.loaded = false
		m.preview.filePath = ""
		m.preview.fileName = ""
		m.preview.content = nil
		m.preview.cacheValid = false
	}
}

// nextTab switches to the next tab (wraps around)
func (m *model) nextTab() {
	if len(m.tabs) <= 1 {
		return
	}
	m.activeTab = (m.activeTab + 1) % len(m.tabs)
	tab := m.tabs[m.activeTab]
	m.loadPreview(tab.path)
	m.populatePreviewCache()
}

// prevTab switches to the previous tab (wraps around)
func (m *model) prevTab() {
	if len(m.tabs) <= 1 {
		return
	}
	m.activeTab--
	if m.activeTab < 0 {
		m.activeTab = len(m.tabs) - 1
	}
	tab := m.tabs[m.activeTab]
	m.loadPreview(tab.path)
	m.populatePreviewCache()
}

// parseGitStatusFromName extracts the git status code from a changed file name
// Changed file names are formatted as "[XY] relative/path"
func parseGitStatusFromName(name string) string {
	if len(name) >= 4 && name[0] == '[' && name[3] == ']' {
		return name[1:3]
	}
	return ""
}

// cleanNameFromChangedFile extracts just the filename from a changed file name
// Changed file names are formatted as "[XY] relative/path"
func cleanNameFromChangedFile(name string) string {
	if len(name) >= 5 && name[0] == '[' && name[3] == ']' {
		relPath := strings.TrimSpace(name[4:])
		return filepath.Base(relPath)
	}
	return filepath.Base(name)
}

// navigateToPath changes the current path and automatically exits special modes (trash, favorites, etc)
// This ensures users don't get stuck in filter modes when navigating
func (m *model) navigateToPath(newPath string) {
	// If we're in trash mode and navigating away, check if staying within trash
	if m.showTrashOnly {
		trashDir, err := getTrashDir()
		if err == nil {
			// Check if the new path is within the trash directory
			if strings.HasPrefix(newPath, trashDir) {
				// Still within trash - allow navigation
				m.currentPath = newPath
				m.cursor = 0
				m.loadFiles()
				return
			}
		}

		// Navigating outside trash - exit trash mode
		m.showTrashOnly = false
		if m.trashRestorePath != "" {
			m.currentPath = m.trashRestorePath
			m.trashRestorePath = ""
		}
		m.cursor = 0
		m.loadFiles()
		return
	}

	// Auto-exit changes mode when navigating to a different directory
	if m.showChangesOnly {
		m.showChangesOnly = false
		m.showDiffPreview = false
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

	if m.previewOnly {
		// Standalone preview mode: title(1) + help(1) + borders(2) = 4 lines overhead
		visibleLines = m.height - 4
		if visibleLines < 3 {
			visibleLines = 3
		}
	} else if m.viewMode == viewFullPreview {
		// Full preview mode: m.height - 4 (borders/help) - headerLines (title/info when mouse enabled)
		headerLines := 0
		if m.previewMouseEnabled {
			headerLines = 2
		}
		maxVisible := m.height - 4 - headerLines
		visibleLines = maxVisible - 2 // Account for border
	} else if m.viewMode == viewDualPane {
		// Dual-pane mode has different layouts:
		// 1. Horizontal split (tree/list on wide terminals): side-by-side panes
		// 2. Vertical split (detail mode or narrow terminals): stacked panes with accordion

		// Check if using vertical stacking (detail mode or narrow terminal)
		useVerticalSplit := m.displayMode == modeDetail || m.isNarrowTerminal()

		if useVerticalSplit {
			// VERTICAL SPLIT: Preview shares height with file list
			// Layout: header(4) + panes(maxVisible) + footer(4) = m.height
			headerLines := 4
			footerLines := 4
			maxVisible := m.height - headerLines - footerLines
			if maxVisible < 5 {
				maxVisible = 5
			}

			// Accordion: Focused pane gets 2/3, unfocused gets 1/3
			var bottomHeight int
			if m.focusedPane == leftPane {
				// File list focused, preview gets 1/3
				topHeight := (maxVisible * 2) / 3
				bottomHeight = maxVisible - topHeight
			} else {
				// Preview focused, gets 2/3
				bottomHeight = (maxVisible * 2) / 3
			}

			// Account for borders and scroll indicator
			bottomContentHeight := bottomHeight - 2
			visibleLines = bottomContentHeight - 1 // Subtract 1 for scroll indicator line
			// Note: Scroll indicator is ALWAYS shown in dual-pane mode, so always reserve the line
		} else {
			// HORIZONTAL SPLIT: Preview uses full height, side-by-side
			// Layout: header(4) + panes(maxVisible) + footer(4) = m.height
			// maxVisible = m.height - 8
			// contentHeight = maxVisible - 2 (borders) = m.height - 10
			// targetLines = contentHeight - 1 (scroll indicator) = m.height - 11
			visibleLines = m.height - 11
			// Reserve one more line if content fits (no scroll indicator needed)
			if totalLines <= visibleLines+1 {
				visibleLines++ // No scroll indicator needed, can show one more line
			}
		}
	} else {
		// Single-pane preview mode (shouldn't happen, but default to safe value)
		visibleLines = m.height - 6
	}

	// For prompt files, subtract the header height since it doesn't scroll
	if m.preview.isPrompt && m.preview.promptTemplate != nil {
		var boxContentWidth int
		if m.viewMode == viewFullPreview {
			boxContentWidth = m.width - 6
		} else {
			boxContentWidth = m.rightWidth - 6 // Match full preview calculation
		}
		promptHeaderHeight := m.getPromptHeaderHeight(boxContentWidth)
		visibleLines -= promptHeaderHeight
	}

	if visibleLines < 1 {
		visibleLines = 1
	}

	return visibleLines
}

// renderToolbarRow renders the emoji button toolbar row
// Shows: [🏠] [📄/📊/🌲] [🔃] [⬜/⬌] [>_] [🔍] [🤖] [⚡]
// This function is shared between single-pane (view.go) and dual-pane (render_preview.go) views
func (m model) renderToolbarRow() string {
	var s strings.Builder

	// Home button - navigate to home directory
	homeButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(homeButtonStyle.Render("[🏠]"))
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

	// Sort toggle button (cycles Name → Size → Modified → Type)
	sortIcon := "🔃"
	sortButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(sortButtonStyle.Render("[" + sortIcon + "]"))
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

	// Agent conversations toggle button
	agentIcon := "🤖"
	if m.showAgentView {
		activeAgentStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeAgentStyle.Render("[" + agentIcon + "]"))
	} else {
		agentButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(agentButtonStyle.Render("[" + agentIcon + "]"))
	}
	s.WriteString(" ")

	// Git changes toggle button
	changesIcon := "⚡"
	if m.showChangesOnly {
		// Active: gray background
		activeChangesStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(activeChangesStyle.Render("[" + changesIcon + "]"))
	} else {
		changesButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(changesButtonStyle.Render("[" + changesIcon + "]"))
	}
	s.WriteString(" ")

	// Pad to full terminal width to prevent ANSI escape sequence leakage into adjacent tmux panes
	// This matches the behavior of renderMenuBar() and prevents visual corruption
	toolbarContent := s.String()
	padding := m.width - lipgloss.Width(toolbarContent)
	if padding < 0 {
		padding = 0
	}

	return toolbarContent + strings.Repeat(" ", padding)
}

// findTFERepository attempts to locate the TFE git repository
// It tries multiple strategies:
// 1. Walk up from the current executable path
// 2. Check common development directories
// 3. Check go workspace locations
func findTFERepository() string {
	// Strategy 1: Find from current executable location
	if exePath, err := os.Executable(); err == nil {
		// Resolve symlinks
		if realPath, err := filepath.EvalSymlinks(exePath); err == nil {
			exePath = realPath
		}

		// Walk up the directory tree looking for TFE repo
		dir := filepath.Dir(exePath)
		for i := 0; i < 5; i++ { // Check up to 5 levels up
			// Check if this directory is a TFE git repo
			if isTFERepo(dir) {
				return dir
			}
			// Move up one directory
			parent := filepath.Dir(dir)
			if parent == dir {
				break // Reached root
			}
			dir = parent
		}
	}

	// Strategy 2: Check common locations
	// Use os.UserHomeDir() for cross-platform compatibility, with fallback to HOME env var
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	if home == "" {
		// No home directory available, skip home-based path checks
		return ""
	}
	possiblePaths := []string{
		filepath.Join(home, "TFE"),
		filepath.Join(home, "tfe"),
		filepath.Join(home, "projects", "TFE"),
		filepath.Join(home, "projects", "tfe"),
		filepath.Join(home, "Projects", "TFE"),
		filepath.Join(home, "Projects", "tfe"),
		filepath.Join(home, "dev", "TFE"),
		filepath.Join(home, "dev", "tfe"),
		filepath.Join(home, "Development", "TFE"),
		filepath.Join(home, "Development", "tfe"),
		filepath.Join(home, "code", "TFE"),
		filepath.Join(home, "code", "tfe"),
		filepath.Join(home, "go", "src", "github.com", "GGPrompts", "tfe"),
		filepath.Join(home, "go", "src", "github.com", "GGPrompts", "TFE"),
	}

	for _, path := range possiblePaths {
		if isTFERepo(path) {
			return path
		}
	}

	return ""
}

// isTFERepo checks if a directory is a TFE git repository
func isTFERepo(path string) bool {
	// Must have build.sh
	if _, err := os.Stat(filepath.Join(path, "build.sh")); err != nil {
		return false
	}

	// Must be a git repository
	if _, err := os.Stat(filepath.Join(path, ".git")); err != nil {
		return false
	}

	// Optional: Check for main.go to confirm it's the TFE project
	if _, err := os.Stat(filepath.Join(path, "main.go")); err != nil {
		return false
	}

	return true
}

// getClaudeCodePath returns the best available Claude Code executable path
// Tries local development version first, then falls back to system version
func getClaudeCodePath() string {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		localClaudePath := filepath.Join(homeDir, ".claude", "local", "claude")
		if _, err := os.Stat(localClaudePath); err == nil {
			return localClaudePath
		}
	}
	// Fall back to system version (assumes 'claude' is in PATH)
	return "claude"
}

// getClaudeCodePathForTermux returns a command to run Claude Code in Termux
// The standard 'claude' script has shebang #!/usr/bin/env node which fails
// in Termux because /usr/bin/env doesn't exist. We run node directly instead.
func getClaudeCodePathForTermux() string {
	// Try the npm global install location first
	cliPath := "/data/data/com.termux/files/usr/lib/node_modules/@anthropic-ai/claude-code/cli.js"
	if _, err := os.Stat(cliPath); err == nil {
		return "node " + cliPath
	}
	// Fall back to just 'claude' and hope for the best
	return "claude"
}

// isTermux returns true if running in Termux environment
func isTermux() bool {
	// Check for Termux-specific path
	_, err := os.Stat("/data/data/com.termux/files/usr")
	return err == nil
}

// hasParentShell returns true if TFE was launched from an interactive shell
// (as opposed to being launched from a widget/shortcut with no parent shell).
// When there's a parent shell, the tfe wrapper function handles cd after exit.
func hasParentShell() bool {
	ppid := os.Getppid()
	if ppid <= 1 {
		return false
	}
	// Check if parent process is a shell
	cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", ppid))
	if err != nil {
		return false
	}
	cmd := strings.ToLower(string(cmdline))
	return strings.Contains(cmd, "bash") || strings.Contains(cmd, "zsh") || strings.Contains(cmd, "fish") || strings.Contains(cmd, "sh")
}

// termuxNewSessionCmd returns the am startservice command to launch a new Termux session
// This enables features like Quick CD and Claude launch when TFE is started from a widget
// Requires: allow-external-apps = true in ~/.termux/termux.properties
func termuxNewSessionCmd(command string, workDir string) string {
	// Build the bash command to run in the new session
	// IMPORTANT: We must set up the environment properly:
	// 1. LD_PRELOAD for libtermux-exec - fixes shebangs like #!/usr/bin/env node
	// 2. Source .bashrc for PATH and other environment variables
	// 3. Use login shell (-l) so profile scripts are sourced

	// Dynamically detect Termux installation directory from PREFIX env var
	// Fall back to standard Termux path if PREFIX is not set
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}

	ldPreload := fmt.Sprintf("export LD_PRELOAD=%s/lib/libtermux-exec.so", prefix)
	bashPath := fmt.Sprintf("%s/bin/bash", prefix)

	var bashCmd string
	if workDir != "" {
		bashCmd = fmt.Sprintf("%s; source ~/.bashrc 2>/dev/null; cd '%s' && %s", ldPreload, workDir, command)
	} else {
		bashCmd = fmt.Sprintf("%s; source ~/.bashrc 2>/dev/null; %s", ldPreload, command)
	}

	// Use am startservice with RUN_COMMAND intent
	// --ez com.termux.RUN_COMMAND_BACKGROUND false = run in foreground terminal
	return fmt.Sprintf(
		"am startservice --user 0 "+
			"-n com.termux/com.termux.app.RunCommandService "+
			"-a com.termux.RUN_COMMAND "+
			"--es com.termux.RUN_COMMAND_PATH '%s' "+
			"--esa com.termux.RUN_COMMAND_ARGUMENTS '-c,%s' "+
			"--ez com.termux.RUN_COMMAND_BACKGROUND false",
		bashPath, bashCmd,
	)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderScrollingFooter renders footer text with horizontal scrolling if enabled
// If text fits within width, returns as-is. If scrolling is enabled and text is too long,
// creates a looping marquee effect. Otherwise truncates with "..."
func (m model) renderScrollingFooter(text string, availableWidth int) string {
	textLen := m.visualWidthCompensated(text)

	// If text fits, no modification needed
	if textLen <= availableWidth {
		return text
	}

	// If scrolling is active, create looping marquee
	if m.footerScrolling {
		// Add visual indicator and separator for smooth loop
		indicator := "⏵ " // Indicates scrolling is active
		paddedText := indicator + text + "   •   " + indicator + text

		// Convert to runes to handle multi-byte unicode characters (↑, ↓, •, etc.)
		runes := []rune(paddedText)
		runeCount := len(runes)

		// Calculate scroll position with wrapping
		scrollPos := m.footerOffset % runeCount

		// Extract visible portion (by rune, not byte)
		var result strings.Builder
		for i := 0; i < availableWidth && i < runeCount; i++ {
			charPos := (scrollPos + i) % runeCount
			result.WriteRune(runes[charPos])
		}

		return result.String()
	}

	// Not scrolling - truncate with "..."
	return m.truncateToWidthCompensated(text, availableWidth)
}
