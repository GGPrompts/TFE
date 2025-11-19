package main

// Module: render_layout.go
// Purpose: Full-screen layout rendering for dual-pane and full-preview modes
// Responsibilities:
// - Rendering full-screen preview mode
// - Rendering dual-pane split layout (horizontal and vertical)
// - Managing accordion-style focus for panes
// - Status bar and search UI rendering

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// renderFullPreview renders the full-screen preview mode
func (m model) renderFullPreview() string {
	var s strings.Builder

	// Only show title bar and info line when mouse is enabled (not in text selection mode)
	headerLines := 0
	if m.previewMouseEnabled {
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

		// File info line with scroll position percentage
		var infoText string

		// Calculate scroll percentage
		totalLines := m.getWrappedLineCount()
		// Calculate how many lines will be visible (need to calculate early for percentage)
		maxVisible := m.height - 4 - 2 // headerLines = 2 when mouse enabled
		contentHeight := maxVisible - 2

		var scrollPercent int
		var lastVisibleLine int
		if totalLines > 0 {
			// Calculate percentage based on how far through scrollable content we are
			maxScrollPos := totalLines - contentHeight
			if maxScrollPos <= 0 {
				// Content fits in one screen
				scrollPercent = 100
				lastVisibleLine = totalLines
			} else {
				scrollPercent = (m.preview.scrollPos * 100) / maxScrollPos
				if scrollPercent > 100 {
					scrollPercent = 100
				}
				// Show the last visible line number (not the top line)
				lastVisibleLine = min(m.preview.scrollPos+contentHeight, totalLines)
			}
		}

		if m.preview.isMarkdown {
			// Show scroll position for markdown too
			if totalLines > 0 {
				infoText = fmt.Sprintf("Size: %s | Markdown Rendered | Line %d/%d (%d%%)",
					formatFileSize(m.preview.fileSize),
					lastVisibleLine,
					totalLines,
					scrollPercent)
			} else {
				infoText = fmt.Sprintf("Size: %s | Markdown Rendered",
					formatFileSize(m.preview.fileSize))
			}
		} else {
			// Show scroll position for regular text
			if totalLines > 0 {
				infoText = fmt.Sprintf("Size: %s | Lines: %d (wrapped) | Line %d/%d (%d%%)",
					formatFileSize(m.preview.fileSize),
					len(m.preview.content),
					lastVisibleLine,
					totalLines,
					scrollPercent)
			} else {
				infoText = fmt.Sprintf("Size: %s | Lines: %d (wrapped)",
					formatFileSize(m.preview.fileSize),
					len(m.preview.content))
			}
		}
		s.WriteString(pathStyle.Render(infoText))
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")

		headerLines = 2 // title + info line
	}

	// Content with border
	// Reserve space based on whether header is shown
	maxVisible := m.height - 4 - headerLines // Reserve space for header (if shown), help, and borders
	contentHeight := maxVisible - 2          // Content area accounting for borders
	previewContent := m.renderPreview(contentHeight)

	// Wrap preview in bordered box with fixed dimensions
	// Content is constrained to contentHeight lines to fit within the box
	// When mouse is disabled (for text selection), remove border for cleaner copying
	previewBoxStyle := lipgloss.NewStyle().
		Width(m.width - 6).   // Leave margin for borders
		Height(contentHeight) // Content area height (borders added by Lipgloss)

	if m.previewMouseEnabled {
		// Mouse enabled: show decorative border
		previewBoxStyle = previewBoxStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.AdaptiveColor{
				Light: "#00af87", // Teal for light
				Dark:  "#5faf87", // Light teal for dark
			})
	} else {
		// Mouse disabled (text selection mode): no border for cleaner copying
		previewBoxStyle = previewBoxStyle.Padding(0, 1) // Just add side padding
	}

	s.WriteString(previewBoxStyle.Render(previewContent))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)

	// Show different F5 text based on file type
	f5Text := "copy path"
	if m.preview.isPrompt {
		f5Text = "copy rendered prompt"
	} else if !m.preview.isBinary && len(m.preview.content) > 0 {
		f5Text = "copy content"
	}

	// Mouse toggle indicator - show what 'm' key does
	var modeText, helpText string
	if m.previewMouseEnabled {
		modeText = "🖱 text select" // Press m to enable text selection
	} else {
		modeText = "⌨ mouse scroll" // Press m to enable mouse scrolling
	}

	// Build help text
	if m.preview.isBinary && isImageFile(m.preview.filePath) {
		helpText = fmt.Sprintf("F1: help • V: view image • m: %s • F4: edit • Esc: close", modeText)
	} else {
		helpText = fmt.Sprintf("F1: help • ↑/↓: scroll • m: %s • F4: edit • F5: %s • Esc: close", modeText, f5Text)
	}
	// Truncate help text to terminal width to prevent wrapping/corruption
	if m.visualWidthCompensated(helpText) > m.width-4 {
		helpText = m.truncateToWidthCompensated(helpText, m.width-4)
	}
	s.WriteString(helpStyle.Render(helpText))
	s.WriteString("\033[0m") // Reset ANSI codes

	// Show search input if search is active
	if m.preview.searchActive {
		s.WriteString("\n")
		searchStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("33")). // Blue background
			Foreground(lipgloss.Color("0")).  // Black text
			Bold(true).
			Padding(0, 1)

		matchCount := len(m.preview.searchMatches)
		var searchText string
		if matchCount > 0 {
			currentPos := m.preview.currentMatch + 1
			searchText = fmt.Sprintf("🔍 Search: %s█ (%d/%d matches)", m.preview.searchQuery, currentPos, matchCount)
		} else if m.preview.searchQuery == "" {
			searchText = "🔍 Search: █ (type to search, n/Shift+N: navigate, Esc: exit)"
		} else {
			searchText = fmt.Sprintf("🔍 Search: %s█ (no matches)", m.preview.searchQuery)
		}

		// Truncate search text to terminal width to prevent wrapping/corruption
		if m.visualWidthCompensated(searchText) > m.width-4 {
			searchText = m.truncateToWidthCompensated(searchText, m.width-4)
		}
		s.WriteString(searchStyle.Render(searchText))
		s.WriteString("\033[0m") // Reset ANSI codes
	} else if m.statusMessage != "" && (m.promptEditMode || m.filePickerMode || time.Since(m.statusTime) < 3*time.Second) {
		// Show status message if present (auto-dismiss after 3s, except in edit mode or file picker mode) and search not active
		s.WriteString("\n")
		msgStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("28")). // Green
			Foreground(lipgloss.Color("15")). // White for better contrast
			Bold(true).
			Padding(0, 1)

		if m.statusIsError {
			msgStyle = msgStyle.Background(lipgloss.Color("196")) // Red
		}

		// Truncate status message to terminal width to prevent wrapping/corruption
		statusMsg := m.statusMessage
		if m.visualWidthCompensated(statusMsg) > m.width-4 {
			statusMsg = m.truncateToWidthCompensated(statusMsg, m.width-4)
		}
		s.WriteString(msgStyle.Render(statusMsg))
		s.WriteString("\033[0m") // Reset ANSI codes
	}

	return s.String()
}

// renderDualPane renders the split-pane layout using Lipgloss layout utilities
func (m model) renderDualPane() string {
	var s strings.Builder

	// Check if we should show GitHub link (first 5 seconds) or menu bar
	showGitHub := time.Since(m.startupTime) < 5*time.Second

	if showGitHub {
		// Title with mode indicator (first 5 seconds) + terminal type for debugging
		titleText := fmt.Sprintf("(T)erminal (F)ile (E)xplorer [Dual-Pane] (%s)", m.terminalType.String())
		if m.commandFocused {
			titleText += " [Command Mode]"
		}
		if m.filePickerMode {
			if m.filePickerCopySource != "" {
				titleText += " [📋 Copy Mode - Select Destination]"
			} else {
				titleText += " [📁 File Picker]"
			}
		}

		// Right side: Update notification or GitHub link
		var rightLink string
		var displayText string

		if m.updateAvailable {
			// Show update available with clickable link
			displayText = fmt.Sprintf("🎉 Update Available: %s (click for details)", m.updateVersion)
			// Use special marker URL so we can detect clicks in mouse handler
			rightLink = fmt.Sprintf("\033]8;;update-available\033\\%s\033]8;;\033\\", displayText)
		} else {
			// Show GitHub link
			githubURL := "https://github.com/GGPrompts/TFE"
			displayText = githubURL
			rightLink = fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\", githubURL, githubURL)
		}

		// Calculate spacing to right-align
		availableWidth := m.width - len(titleText) - len(displayText) - 2
		if availableWidth < 1 {
			availableWidth = 1
		}
		spacing := strings.Repeat(" ", availableWidth)

		// Render title on left, link/update on right
		title := titleStyle.Render(titleText) + spacing + titleStyle.Render(rightLink)
		s.WriteString(title)
		s.WriteString("\033[0m") // Reset ANSI codes
		s.WriteString("\n")
	} else {
		// Show menu bar after 5 seconds
		menuBar := m.renderMenuBar()
		s.WriteString(menuBar)
		s.WriteString("\n")
	}

	// Toolbar buttons row
	s.WriteString(m.renderToolbarRow())

	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Command prompt with path (terminal-style)
	promptPrefix := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true).Render("$ ")
	pathPromptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))

	s.WriteString(promptPrefix)
	s.WriteString(pathPromptStyle.Render(getDisplayPath(m.currentPath)))
	s.WriteString(" ")

	// Show helper text based on focus state
	helperStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	if !m.commandFocused && m.commandInput == "" {
		// Not focused - show how to enter command mode
		s.WriteString(helperStyle.Render(": to focus"))
	} else if m.commandFocused && m.commandInput == "" {
		// Focused but no input - show ! prefix hint and cursor
		s.WriteString(helperStyle.Render("! prefix to run & exit"))
		cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
		s.WriteString(cursorStyle.Render("█"))
	} else {
		// Has input - show the command with cursor at correct position
		if m.commandFocused {
			// Render text before cursor, cursor, text after cursor
			beforeCursor := m.commandInput[:m.commandCursorPos]
			afterCursor := m.commandInput[m.commandCursorPos:]

			// Handle ! prefix coloring
			if strings.HasPrefix(beforeCursor, "!") {
				prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				s.WriteString(prefixStyle.Render("!"))
				s.WriteString(inputStyle.Render(beforeCursor[1:]))
			} else {
				s.WriteString(inputStyle.Render(beforeCursor))
			}

			// Render cursor
			cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true)
			s.WriteString(cursorStyle.Render("█"))

			// Render text after cursor
			s.WriteString(inputStyle.Render(afterCursor))
		} else {
			// Not focused - just show the text
			if strings.HasPrefix(m.commandInput, "!") {
				prefixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
				s.WriteString(prefixStyle.Render("!"))
				s.WriteString(inputStyle.Render(m.commandInput[1:]))
			} else {
				s.WriteString(inputStyle.Render(m.commandInput))
			}
		}
	}
	// Explicitly reset styling after cursor to prevent ANSI code leakage
	s.WriteString("\033[0m")
	s.WriteString("\n")

	// Blank line separator between command prompt and panes
	s.WriteString("\n")

	// Calculate max visible for both panes
	// Layout: title(1) + toolbar(1) + command(1) + blank(1) + panes(maxVisible) + blank_after(1) + status(2) + optional(1)
	// Total: 4 + maxVisible + 1 + 2 + (0-1) = maxVisible + 7-8
	// Use worst case (8) to ensure panes never overflow
	headerLines := 4 // title + toolbar + command + blank separator
	footerLines := 4 // blank after panes + 2 status lines + optional message/search
	maxVisible := m.height - headerLines - footerLines
	if maxVisible < 5 {
		maxVisible = 5 // Minimum pane height
	}

	// Content area is maxVisible - 2 (accounting for top/bottom borders)
	contentHeight := maxVisible - 2

	// Render panes based on display mode
	var panes string

	if m.displayMode == modeDetail {
		// VERTICAL SPLIT for detail view - gives full width to detail columns
		// ACCORDION: Focused pane gets 2/3 height, unfocused gets 1/3
		var topHeight, bottomHeight int
		if m.focusedPane == leftPane {
			// Top pane (detail view) is focused
			topHeight = (maxVisible * 2) / 3
			bottomHeight = maxVisible - topHeight
		} else {
			// Bottom pane (preview) is focused
			bottomHeight = (maxVisible * 2) / 3
			topHeight = maxVisible - bottomHeight
		}

		topContentHeight := topHeight - 2       // Account for borders
		bottomContentHeight := bottomHeight - 2

		// Render top pane (detail view with full width)
		topContent := m.renderDetailView(topContentHeight)

		// Render bottom pane (preview with full width)
		var bottomContent string
		if m.preview.loaded {
			bottomContent = m.renderPreview(bottomContentHeight)
		} else {
			emptyStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Italic(true)
			bottomContent = emptyStyle.Render("No preview available\n\nSelect a file to preview") + "\033[0m"
		}

		// Border colors based on focus
		topBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
		bottomBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
		if m.focusedPane == leftPane {
			topBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
		} else {
			bottomBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
		}

		// Create boxes with full width
		topPaneStyle := lipgloss.NewStyle().
			Width(m.width - 6). // Full width minus margins
			Height(topContentHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(topBorderColor)

		bottomPaneStyle := lipgloss.NewStyle().
			Width(m.width - 6). // Full width minus margins
			Height(bottomContentHeight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(bottomBorderColor)

		topPaneRendered := topPaneStyle.Render(topContent)
		bottomPaneRendered := bottomPaneStyle.Render(bottomContent)

		// Stack vertically
		panes = lipgloss.JoinVertical(lipgloss.Left, topPaneRendered, bottomPaneRendered)

	} else {
		// List/Tree view - use VERTICAL split on narrow terminals, HORIZONTAL on wide terminals

		if m.isNarrowTerminal() {
			// VERTICAL SPLIT for narrow terminals (phones) - same as detail view
			// ACCORDION: Focused pane gets 2/3 height, unfocused gets 1/3
			var topHeight, bottomHeight int
			if m.focusedPane == leftPane {
				// Top pane (file list) is focused
				topHeight = (maxVisible * 2) / 3
				bottomHeight = maxVisible - topHeight
			} else {
				// Bottom pane (preview) is focused
				bottomHeight = (maxVisible * 2) / 3
				topHeight = maxVisible - bottomHeight
			}

			topContentHeight := topHeight - 2       // Account for borders
			bottomContentHeight := bottomHeight - 2

			// Render top pane (file list with full width)
			var topContent string
			switch m.displayMode {
			case modeList:
				topContent = m.renderListView(topContentHeight)
			case modeTree:
				topContent = m.renderTreeView(topContentHeight)
			default:
				topContent = m.renderListView(topContentHeight)
			}

			// Render bottom pane (preview with full width)
			var bottomContent string
			if m.preview.loaded {
				bottomContent = m.renderPreview(bottomContentHeight)
			} else {
				emptyStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("241")).
					Italic(true)
				bottomContent = emptyStyle.Render("No preview available\n\nSelect a file to preview") + "\033[0m"
			}

			// Border colors based on focus
			topBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
			bottomBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
			if m.focusedPane == leftPane {
				topBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
			} else {
				bottomBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
			}

			// Create boxes with full width
			topPaneStyle := lipgloss.NewStyle().
				Width(m.width - 6). // Full width minus margins
				Height(topContentHeight).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(topBorderColor)

			bottomPaneStyle := lipgloss.NewStyle().
				Width(m.width - 6). // Full width minus margins
				Height(bottomContentHeight).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(bottomBorderColor)

			topPaneRendered := topPaneStyle.Render(topContent)
			bottomPaneRendered := bottomPaneStyle.Render(bottomContent)

			// Stack vertically
			panes = lipgloss.JoinVertical(lipgloss.Left, topPaneRendered, bottomPaneRendered)

		} else {
			// HORIZONTAL SPLIT for wide terminals - accordion style
			// Get left pane content - use contentHeight so content fits within the box
			var leftContent string
			switch m.displayMode {
			case modeList:
				leftContent = m.renderListView(contentHeight)
			case modeTree:
				leftContent = m.renderTreeView(contentHeight)
			default:
				leftContent = m.renderListView(contentHeight)
			}

			// Get right pane content (preview)
			var rightContent string
			if m.preview.loaded {
				rightContent = m.renderPreview(contentHeight)
			} else {
				emptyStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("241")).
					Italic(true)
				rightContent = emptyStyle.Render("No preview available\n\nSelect a file to preview") + "\033[0m"
			}

			// Border colors based on focus (accordion style)
			leftBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
			rightBorderColor := lipgloss.AdaptiveColor{Light: "#999999", Dark: "#585858"}
			if m.focusedPane == leftPane {
				leftBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
			} else {
				rightBorderColor = lipgloss.AdaptiveColor{Light: "#0087d7", Dark: "#00d7ff"}
			}

			// Use exact Width and Height to ensure panes stay perfectly aligned
			leftPaneStyle := lipgloss.NewStyle().
				Width(m.leftWidth - 2).   // Content width (borders added by Lipgloss)
				Height(contentHeight).    // Exact content height (borders added by Lipgloss)
				Border(lipgloss.RoundedBorder()).
				BorderForeground(leftBorderColor)

			rightPaneStyle := lipgloss.NewStyle().
				Width(m.rightWidth - 2).  // Content width (borders added by Lipgloss)
				Height(contentHeight).    // Exact content height (borders added by Lipgloss)
				Border(lipgloss.RoundedBorder()).
				BorderForeground(rightBorderColor)

			// Apply styles to content
			leftPaneRendered := leftPaneStyle.Render(leftContent)
			rightPaneRendered := rightPaneStyle.Render(rightContent)

			// Join panes horizontally
			panes = lipgloss.JoinHorizontal(lipgloss.Top, leftPaneRendered, rightPaneRendered)
		}
	}

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

	gitReposIndicator := ""
	if m.showGitReposOnly {
		gitReposIndicator = " • 🔀 git repos only"
	}

	// Show focused pane info in status bar
	focusInfo := ""
	if m.focusedPane == leftPane {
		focusInfo = " • [LEFT focused]"
	} else {
		focusInfo = " • [RIGHT focused]"
	}

	// Help hint - show "/" search hint only when not already searching
	helpHint := " • F1: help"
	if !m.searchMode && m.searchQuery == "" {
		helpHint += " • /: search"
	}

	// Selected file info
	var selectedInfo string
	if currentFile := m.getCurrentFile(); currentFile != nil {
		if currentFile.isDir {
			// Special handling for ".." to show parent directory name
			if currentFile.name == ".." {
				parentPath := filepath.Dir(m.currentPath)
				parentName := filepath.Base(parentPath)
				if parentName == "/" || parentName == "." {
					parentName = "root"
				}
				selectedInfo = fmt.Sprintf("Selected: .. (go up to %s)", parentName)
			} else {
				selectedInfo = fmt.Sprintf("Selected: %s (folder)", currentFile.name)
			}
		} else {
			fileType := getFileType(*currentFile)

			// For symlinks, truncate long paths to show the important trailing part
			if currentFile.isSymlink && currentFile.symlinkTarget != "" {
				// Calculate available space: terminal width minus other info
				// "Selected: filename (, size, date)"
				baseInfoLen := len("Selected: ") + len(currentFile.name) + len(", ") +
					len(formatFileSize(currentFile.size)) + len(", ") +
					len(formatModTime(currentFile.modTime)) + len(" ()") + 10 // padding

				availableForTarget := m.width - baseInfoLen
				if availableForTarget < 30 {
					availableForTarget = 30 // Minimum to show something useful
				}

				fullTarget := "Link → " + currentFile.symlinkTarget
				if len(fullTarget) > availableForTarget {
					// Show trailing end: "...filename" instead of "Link → /very/long/pa..."
					fileType = "..." + fullTarget[len(fullTarget)-(availableForTarget-3):]
				}
			}

			selectedInfo = fmt.Sprintf("Selected: %s (%s, %s, %s)",
				currentFile.name,
				fileType,
				formatFileSize(currentFile.size),
				formatModTime(currentFile.modTime))
		}
	}

	// Split status into two lines to prevent truncation
	// Line 1: Counts, indicators, view mode, focus, help
	statusLine1 := fmt.Sprintf("%s%s%s%s%s • %s%s%s", itemsInfo, hiddenIndicator, favoritesIndicator, promptsIndicator, gitReposIndicator, m.displayMode.String(), focusInfo, helpHint)
	// Truncate to terminal width to prevent wrapping/corruption on narrow terminals
	if m.visualWidthCompensated(statusLine1) > m.width-4 {
		statusLine1 = m.truncateToWidthCompensated(statusLine1, m.width-4)
	}
	s.WriteString(statusStyle.Render(statusLine1))
	s.WriteString("\033[0m") // Reset ANSI codes
	s.WriteString("\n")

	// Line 2: Selected file info
	statusLine2 := selectedInfo
	// Truncate to terminal width to prevent wrapping/corruption on narrow terminals
	if m.visualWidthCompensated(statusLine2) > m.width-4 {
		statusLine2 = m.truncateToWidthCompensated(statusLine2, m.width-4)
	}
	s.WriteString(statusStyle.Render(statusLine2))
	s.WriteString("\033[0m") // Reset ANSI codes

	// Show status message if present (auto-dismiss after 3s, except in edit mode or file picker mode)
	if m.statusMessage != "" && (m.promptEditMode || m.filePickerMode || time.Since(m.statusTime) < 3*time.Second) {
		s.WriteString("\n")
		msgStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("28")). // Green
			Foreground(lipgloss.Color("15")). // White for better contrast
			Bold(true).
			Padding(0, 1)

		if m.statusIsError {
			msgStyle = msgStyle.Background(lipgloss.Color("196")) // Red
		}

		// Truncate status message to terminal width to prevent wrapping/corruption
		statusMsg := m.statusMessage
		if m.visualWidthCompensated(statusMsg) > m.width-4 {
			statusMsg = m.truncateToWidthCompensated(statusMsg, m.width-4)
		}
		s.WriteString(msgStyle.Render(statusMsg))
		s.WriteString("\033[0m") // Reset ANSI codes
	} else if m.searchMode || m.searchQuery != "" {
		// Show search status
		s.WriteString("\n")
		searchStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("33")).  // Blue background
			Foreground(lipgloss.Color("255")). // Bright white for high contrast
			Bold(true).
			Padding(0, 1)

		// Calculate match count (exclude parent directory "..")
		matchCount := len(m.filteredIndices)
		if matchCount > 0 {
			matchCount-- // Exclude ".." which is always included
		}

		var searchStatus string
		if m.searchMode {
			// Active search mode with cursor
			searchStatus = fmt.Sprintf("Search: %s█ (%d matches)", m.searchQuery, matchCount)
		} else {
			// Search accepted (filter active but not in input mode)
			searchStatus = fmt.Sprintf("Filtered: %s (%d matches)", m.searchQuery, matchCount)
		}

		// Truncate search status to terminal width to prevent wrapping/corruption
		if m.visualWidthCompensated(searchStatus) > m.width-4 {
			searchStatus = m.truncateToWidthCompensated(searchStatus, m.width-4)
		}
		s.WriteString(searchStyle.Render(searchStatus))
		s.WriteString("\033[0m") // Reset ANSI codes
	}

	return s.String()
}
