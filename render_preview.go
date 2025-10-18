package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// getWrappedLineCount calculates the total number of wrapped lines for the current preview
func (m model) getWrappedLineCount() int {
	if !m.preview.loaded {
		return 0
	}

	// Calculate available width based on file type and view mode
	var availableWidth int
	var boxContentWidth int

	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6 // Box content width in full preview
	} else {
		boxContentWidth = m.rightWidth - 2 // Box content width in dual-pane
	}

	if m.preview.isMarkdown {
		// Markdown: no line numbers or scrollbar, content uses full box width
		availableWidth = boxContentWidth
	} else {
		// Regular text: subtract line nums (6) + scrollbar (1) + space (1) = 8 chars
		availableWidth = boxContentWidth - 8
	}

	if availableWidth < 20 {
		availableWidth = 20
	}

	// Use cached line count if available and width matches
	if m.preview.cacheValid && m.preview.cachedLineCount > 0 && m.preview.cachedWidth == availableWidth {
		return m.preview.cachedLineCount
	}

	// For markdown, we need to render it to count lines
	if m.preview.isMarkdown {
		markdownContent := strings.Join(m.preview.content, "\n")
		renderer, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle("auto"),
			glamour.WithWordWrap(availableWidth),
		)
		if err == nil {
			rendered, err := renderer.Render(markdownContent)
			if err == nil {
				renderedLines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
				return len(renderedLines)
			}
		}
		// Fallback if glamour fails
	}

	// For regular text, count wrapped lines
	totalLines := 0
	for _, line := range m.preview.content {
		wrapped := wrapLine(line, availableWidth)
		totalLines += len(wrapped)
	}
	return totalLines
}

// wrapLine wraps a line of text to fit within the specified width
func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}

	// Handle empty lines
	if len(line) == 0 {
		return []string{""}
	}

	var wrapped []string
	currentLine := ""
	currentWidth := 0

	words := strings.Fields(line)
	if len(words) == 0 {
		// Line is only whitespace
		return []string{line}
	}

	for i, word := range words {
		wordWidth := visualWidth(word)
		spaceWidth := 1

		// Check if this word fits on the current line
		if currentWidth == 0 {
			// First word on line
			if wordWidth <= width {
				currentLine = word
				currentWidth = wordWidth
			} else {
				// Word is too long, force break it
				wrapped = append(wrapped, word[:width])
				currentLine = ""
				currentWidth = 0
			}
		} else if currentWidth+spaceWidth+wordWidth <= width {
			// Word fits on current line
			currentLine += " " + word
			currentWidth += spaceWidth + wordWidth
		} else {
			// Word doesn't fit, start new line
			wrapped = append(wrapped, currentLine)
			if wordWidth <= width {
				currentLine = word
				currentWidth = wordWidth
			} else {
				// Word is too long, force break it
				wrapped = append(wrapped, word[:width])
				currentLine = ""
				currentWidth = 0
			}
		}

		// If this is the last word, add the current line
		if i == len(words)-1 && currentLine != "" {
			wrapped = append(wrapped, currentLine)
		}
	}

	if len(wrapped) == 0 {
		return []string{line}
	}

	return wrapped
}

// renderPreview renders the preview pane content with scrollbar
func (m model) renderPreview(maxVisible int) string {
	var s strings.Builder

	if !m.preview.loaded {
		s.WriteString("No file loaded")
		return s.String()
	}

	// If this is a prompt file, show metadata header
	if m.preview.isPrompt && m.preview.promptTemplate != nil {
		return m.renderPromptPreview(maxVisible)
	}

	// Calculate available width for content based on file type and view mode
	var availableWidth int
	var boxContentWidth int // Width of the box content area

	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6 // Box content width in full preview
	} else {
		boxContentWidth = m.rightWidth - 2 // Box content width in dual-pane (accounting for borders)
	}

	if m.preview.isMarkdown {
		// Markdown: no line numbers or scrollbar, content uses full box width
		availableWidth = boxContentWidth
	} else {
		// Regular text: subtract line nums (6) + scrollbar (1) + space (1) = 8 chars
		availableWidth = boxContentWidth - 8
	}

	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// If markdown, render with Glamour
	if m.preview.isMarkdown && m.preview.cachedRenderedContent != "" {
		// Use cached Glamour-rendered content (no line numbers)
		if m.preview.cachedRenderedContent != "" {
			renderedLines := strings.Split(strings.TrimRight(m.preview.cachedRenderedContent, "\n"), "\n")

			// Calculate visible range based on scroll position
			totalLines := len(renderedLines)
			start := m.preview.scrollPos

			if start < 0 {
				start = 0
			}
			if start >= totalLines {
				start = max(0, totalLines-maxVisible)
			}

			end := start + maxVisible
			if end > totalLines {
				end = totalLines
			}

			// Render visible lines without line numbers for markdown
			for i := start; i < end; i++ {
				s.WriteString(renderedLines[i])
				s.WriteString("\033[0m") // Reset ANSI codes to prevent bleed
				s.WriteString("\n")
			}

			return strings.TrimRight(s.String(), "\n")
		}
		// If markdown flag is set but no rendered content, fall through to plain text rendering
		// This happens when Glamour fails or file is too large
	}

	// Wrap all lines first (use cache if available and width matches)
	var wrappedLines []string
	if m.preview.cacheValid && len(m.preview.cachedWrappedLines) > 0 && m.preview.cachedWidth == availableWidth {
		// Use cached wrapped lines
		wrappedLines = m.preview.cachedWrappedLines
	} else {
		// Wrap lines (will be slow for large files without cache)
		for _, line := range m.preview.content {
			wrapped := wrapLine(line, availableWidth)
			wrappedLines = append(wrappedLines, wrapped...)
		}
	}

	// Calculate visible range based on scroll position
	totalLines := len(wrappedLines)
	start := m.preview.scrollPos

	// Ensure start is within bounds
	if start < 0 {
		start = 0
	}
	if start >= totalLines {
		start = max(0, totalLines-maxVisible)
	}

	end := start + maxVisible
	if end > totalLines {
		end = totalLines
	}

	// Render lines with line numbers and scrollbar
	for i := start; i < end; i++ {
		// Line number (5 chars)
		lineNum := fmt.Sprintf("%5d ", i+1)
		lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		s.WriteString(lineNumStyle.Render(lineNum))

		// Scrollbar right after line number (replaces the │ separator)
		scrollbar := m.renderScrollbar(i-start, maxVisible, totalLines)
		s.WriteString(scrollbar)

		// Space after scrollbar
		s.WriteString(" ")

		// Content line
		s.WriteString(wrappedLines[i])
		s.WriteString("\033[0m") // Reset ANSI codes to prevent bleed
		s.WriteString("\n")
	}

	return strings.TrimRight(s.String(), "\n")
}

// renderPromptPreview renders a prompt file with metadata header
func (m model) renderPromptPreview(maxVisible int) string {
	var s strings.Builder
	tmpl := m.preview.promptTemplate

	// Build header lines
	var headerLines []string

	// Prompt name (if available)
	if tmpl.name != "" {
		nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		headerLines = append(headerLines, nameStyle.Render("📝 "+tmpl.name))
		headerLines = append(headerLines, "") // Blank line
	}

	// Description (if available)
	if tmpl.description != "" {
		descStyle := lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("245"))
		headerLines = append(headerLines, descStyle.Render(tmpl.description))
		headerLines = append(headerLines, "") // Blank line
	}

	// Source indicator
	sourceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	sourceIcon := ""
	sourceLabel := ""
	switch tmpl.source {
	case "global":
		sourceIcon = "🌐"
		sourceLabel = "Global Prompt (~/.prompts/)"
	case "command":
		sourceIcon = "⚙️"
		sourceLabel = "Project Command (.claude/commands/)"
	case "agent":
		sourceIcon = "🤖"
		sourceLabel = "Project Agent (.claude/agents/)"
	case "local":
		sourceIcon = "📁"
		sourceLabel = "Local Prompt"
	}
	headerLines = append(headerLines, sourceStyle.Render(sourceIcon+" "+sourceLabel))

	// Variables detected (if any)
	if len(tmpl.variables) > 0 {
		varsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
		varsLine := fmt.Sprintf("Variables: %s", strings.Join(tmpl.variables, ", "))
		headerLines = append(headerLines, varsStyle.Render(varsLine))
	}

	// Separator line
	separatorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	headerLines = append(headerLines, separatorStyle.Render("─────────────────────────────────────"))
	headerLines = append(headerLines, "") // Blank line before content

	// Calculate how many lines the header takes
	headerHeight := len(headerLines)

	// Calculate available height for content
	contentHeight := maxVisible - headerHeight
	if contentHeight < 5 {
		contentHeight = 5 // Minimum content height
	}

	// Calculate available width
	var availableWidth int
	var boxContentWidth int

	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6
	} else {
		boxContentWidth = m.rightWidth - 2
	}

	// Prompts don't show line numbers, so use full width
	availableWidth = boxContentWidth - 2 // Just padding

	if availableWidth < 20 {
		availableWidth = 20
	}

	// Wrap content lines
	var wrappedLines []string
	for _, line := range m.preview.content {
		wrapped := wrapLine(line, availableWidth)
		wrappedLines = append(wrappedLines, wrapped...)
	}

	// Calculate visible range for content
	totalLines := len(wrappedLines)
	start := m.preview.scrollPos

	if start < 0 {
		start = 0
	}
	if start >= totalLines {
		start = max(0, totalLines-contentHeight)
	}

	end := start + contentHeight
	if end > totalLines {
		end = totalLines
	}

	// Render header
	for _, line := range headerLines {
		s.WriteString(line)
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	}

	// Render content (no line numbers for prompts)
	for i := start; i < end; i++ {
		s.WriteString(wrappedLines[i])
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	}

	return strings.TrimRight(s.String(), "\n")
}

// renderScrollbar renders a scrollbar indicator for the current line
// Now renders in place of the separator between line numbers and content
func (m model) renderScrollbar(lineIndex, visibleLines, totalLines int) string {
	// Calculate scrollbar position
	// The scrollbar thumb should represent the visible portion of the content
	scrollbarHeight := visibleLines
	thumbSize := max(1, (visibleLines*scrollbarHeight)/totalLines)
	thumbStart := (m.preview.scrollPos * scrollbarHeight) / totalLines

	scrollbarStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	scrollbarThumbStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

	// Determine what to render for this line
	if lineIndex >= thumbStart && lineIndex < thumbStart+thumbSize {
		// This line is part of the scrollbar thumb (bright blue)
		return scrollbarThumbStyle.Render("│")
	} else {
		// This line is part of the scrollbar track (dim gray)
		return scrollbarStyle.Render("│")
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// renderFullPreview renders the full-screen preview mode
func (m model) renderFullPreview() string {
	var s strings.Builder

	// Title bar with file name
	previewTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#ffffff"}).
		Background(lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}).
		Width(m.width).
		Padding(0, 1)

	titleText := fmt.Sprintf("Preview: %s", m.preview.fileName)
	if m.preview.tooLarge || m.preview.isBinary {
		titleText += " [Cannot Preview]"
	}
	if m.preview.isPrompt {
		titleText += " [Prompt Template]"
	} else if m.preview.isMarkdown {
		titleText += " [Markdown]"
	}
	s.WriteString(previewTitleStyle.Render(titleText))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// File info line - update based on whether we're showing markdown or wrapped text
	var infoText string
	if m.preview.isMarkdown {
		infoText = fmt.Sprintf("Size: %s | Markdown Rendered",
			formatFileSize(m.preview.fileSize))
	} else {
		infoText = fmt.Sprintf("Size: %s | Lines: %d (wrapped) | Scroll: %d",
			formatFileSize(m.preview.fileSize),
			len(m.preview.content),
			m.preview.scrollPos+1)
	}
	s.WriteString(pathStyle.Render(infoText))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Content with border
	maxVisible := m.height - 6 // Reserve space for title, info, help (total box height INCLUDING borders)
	contentHeight := maxVisible - 2 // Content area accounting for borders
	previewContent := m.renderPreview(contentHeight)

	// Wrap preview in bordered box with fixed dimensions
	// Content is constrained to contentHeight lines to fit within the box
	previewBoxStyle := lipgloss.NewStyle().
		Width(m.width - 6).       // Leave margin for borders
		Height(contentHeight).    // Content area height (borders added by Lipgloss)
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.AdaptiveColor{
			Light: "#00af87", // Teal for light
			Dark:  "#5faf87",  // Light teal for dark
		})

	s.WriteString(previewBoxStyle.Render(previewContent))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	s.WriteString(helpStyle.Render("↑/↓: scroll • PgUp/PgDown: page • F4: edit • F5: copy path • Esc: close • F10: quit"))
	s.WriteString("\033[0m") // Reset ANSI codes

	return s.String()
}

// renderDualPane renders the split-pane layout using Lipgloss layout utilities
func (m model) renderDualPane() string {
	var s strings.Builder

	// Title with mode indicator and GitHub link
	titleText := "(T)erminal (F)ile (E)xplorer [Dual-Pane]"
	if m.commandFocused {
		titleText += " [Command Mode]"
	}

	// Create GitHub link (OSC 8 hyperlink format)
	githubURL := "https://github.com/GGPrompts/TFE"
	githubLink := fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", githubURL, githubURL)

	// Calculate spacing to right-align GitHub link
	githubText := githubURL // Display text
	availableWidth := m.width - len(titleText) - len(githubText) - 2
	if availableWidth < 1 {
		availableWidth = 1
	}
	spacing := strings.Repeat(" ", availableWidth)

	// Render title on left, GitHub link on right
	title := titleStyle.Render(titleText) + spacing + titleStyle.Render(githubLink)
	s.WriteString(title)
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Toolbar buttons
	// Home button - highlight with gray background when in home directory
	homeDir, _ := os.UserHomeDir()
	if homeDir != "" && m.currentPath == homeDir {
		// Active: gray background (in home directory)
		homeButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			Background(lipgloss.Color("237"))
		s.WriteString(homeButtonStyle.Render("[🏠]"))
	} else {
		// Inactive: normal styling
		homeButtonStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true)
		s.WriteString(homeButtonStyle.Render("[🏠]"))
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

	// CellBlocks button
	cellblocksButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(cellblocksButtonStyle.Render("[📦]"))
	s.WriteString(" ")

	// Fuzzy search button
	searchButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(searchButtonStyle.Render("[🔍]"))
	s.WriteString(" ")

	// Prompts filter toggle button
	promptIcon := "📝"
	if m.showPromptsOnly {
		promptIcon = "✨📝" // Different icon when filter is active
	}
	promptButtonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true)
	s.WriteString(promptButtonStyle.Render("[" + promptIcon + "]"))

	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Command prompt with path (terminal-style)
	promptPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("$ ")
	pathPromptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	s.WriteString(promptPrefix)
	s.WriteString(pathPromptStyle.Render(getDisplayPath(m.currentPath)))
	s.WriteString(" ")
	s.WriteString(inputStyle.Render(m.commandInput))

	// Show cursor only when command mode is active
	if m.commandFocused {
		cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		s.WriteString(cursorStyle.Render("█"))
	}
	// Explicitly reset styling after cursor to prevent ANSI code leakage
	s.WriteString("\033[0m")
	s.WriteString("\n")

	// Blank line separator between command prompt and panes
	s.WriteString("\n")

	// Calculate max visible for both panes
	// title=1 + toolbar=1 + command=1 + separator=1 + panes=maxVisible + separator=1 + status=2 = m.height
	// Therefore: maxVisible = m.height - 7 (total pane height INCLUDING borders)
	maxVisible := m.height - 7

	// Content area is maxVisible - 2 (accounting for top/bottom borders)
	contentHeight := maxVisible - 2

	// Get left pane content - use contentHeight so content fits within the box
	var leftContent string
	switch m.displayMode {
	case modeList:
		leftContent = m.renderListView(contentHeight)
	case modeGrid:
		leftContent = m.renderGridView(contentHeight)
	case modeDetail:
		leftContent = m.renderDetailView(contentHeight)
	case modeTree:
		leftContent = m.renderTreeView(contentHeight)
	default:
		leftContent = m.renderListView(contentHeight)
	}

	// Get right pane content (just the preview, no title)
	rightContent := ""
	if m.preview.loaded {
		rightContent = m.renderPreview(contentHeight)
	} else {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		rightContent = emptyStyle.Render("No preview available\n\nSelect a file to preview") + "\033[0m"
	}

	// Create styled boxes for left and right panes using Lipgloss
	// Highlight the focused pane with a brighter border color
	leftBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}  // dim gray
	rightBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"} // dim gray
	if m.focusedPane == leftPane {
		leftBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"} // bright blue for focused pane
	} else {
		rightBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"} // bright blue for focused pane
	}

	// Use fixed Width/Height for consistent borders
	// Content is constrained to contentHeight lines to fit within the box
	leftPaneStyle := lipgloss.NewStyle().
		Width(m.leftWidth - 2).   // -2 for left/right borders
		Height(contentHeight).    // Content area height (borders added by Lipgloss)
		Border(lipgloss.RoundedBorder()).
		BorderForeground(leftBorderColor)

	rightPaneStyle := lipgloss.NewStyle().
		Width(m.rightWidth - 2).  // -2 for left/right borders
		Height(contentHeight).    // Content area height (borders added by Lipgloss)
		Border(lipgloss.RoundedBorder()).
		BorderForeground(rightBorderColor)

	// Apply styles to content
	leftPaneRendered := leftPaneStyle.Render(leftContent)
	rightPaneRendered := rightPaneStyle.Render(rightContent)

	// Join panes horizontally
	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPaneRendered, rightPaneRendered)
	s.WriteString(panes)
	s.WriteString("\n")

	// Status bar (full width)
	// File counts
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

	itemsInfo := fmt.Sprintf("%d folders, %d files", dirCount, fileCount)
	hiddenIndicator := ""
	if m.showHidden {
		hiddenIndicator = " • hidden"
	}

	favoritesIndicator := ""
	if m.showFavoritesOnly {
		favoritesIndicator = " • ⭐ favorites only"
	}

	promptsIndicator := ""
	if m.showPromptsOnly {
		promptsIndicator = " • 📝 prompts only"
	}

	// Show focused pane info in status bar
	focusInfo := ""
	if m.focusedPane == leftPane {
		focusInfo = " • [LEFT focused]"
	} else {
		focusInfo = " • [RIGHT focused]"
	}

	// Help hint
	helpHint := " • F1: help"

	// Selected file info
	var selectedInfo string
	if currentFile := m.getCurrentFile(); currentFile != nil {
		if currentFile.isDir {
			selectedInfo = fmt.Sprintf("Selected: %s (folder)", currentFile.name)
		} else {
			fileType := getFileType(*currentFile)
			selectedInfo = fmt.Sprintf("Selected: %s (%s, %s, %s)",
				currentFile.name,
				fileType,
				formatFileSize(currentFile.size),
				formatModTime(currentFile.modTime))
		}
	}

	// Split status into two lines to prevent truncation
	// Line 1: Counts, indicators, view mode, focus, help
	statusLine1 := fmt.Sprintf("%s%s%s%s • %s%s%s", itemsInfo, hiddenIndicator, favoritesIndicator, promptsIndicator, m.displayMode.String(), focusInfo, helpHint)
	s.WriteString(statusStyle.Render(statusLine1))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Line 2: Selected file info
	statusLine2 := selectedInfo
	s.WriteString(statusStyle.Render(statusLine2))
	s.WriteString("\033[0m") // Reset ANSI codes

	return s.String()
}
