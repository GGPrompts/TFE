package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle preview mode keys first
		if m.viewMode == viewFullPreview {
			switch msg.String() {
			case "q", "ctrl+c", "esc":
				// Exit preview mode
				m.viewMode = viewSinglePane
				m.calculateLayout()
				return m, nil

			case "e", "E":
				// Edit file in external editor from preview
				if m.preview.loaded && m.preview.filePath != "" {
					editor := getAvailableEditor()
					if editor == "" {
						return m, nil
					}
					if editorAvailable("micro") {
						editor = "micro"
					}
					return m, openEditor(editor, m.preview.filePath)
				}

			case "n", "N":
				// Edit file in nano from preview
				if m.preview.loaded && m.preview.filePath != "" && editorAvailable("nano") {
					return m, openEditor("nano", m.preview.filePath)
				}

			case "y", "c":
				// Copy file path from preview
				if m.preview.loaded && m.preview.filePath != "" {
					_ = copyToClipboard(m.preview.filePath)
				}

			case "up", "k":
				// Scroll preview up
				if m.preview.scrollPos > 0 {
					m.preview.scrollPos--
				}

			case "down", "j":
				// Scroll preview down
				maxScroll := len(m.preview.content) - (m.height - 6)
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.preview.scrollPos < maxScroll {
					m.preview.scrollPos++
				}

			case "pageup":
				m.preview.scrollPos -= m.height - 6
				if m.preview.scrollPos < 0 {
					m.preview.scrollPos = 0
				}

			case "pagedown":
				maxScroll := len(m.preview.content) - (m.height - 6)
				if maxScroll < 0 {
					maxScroll = 0
				}
				m.preview.scrollPos += m.height - 6
				if m.preview.scrollPos > maxScroll {
					m.preview.scrollPos = maxScroll
				}
			}
			return m, nil
		}

		// Regular file browser keys
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc":
			// Exit dual-pane mode if active
			if m.viewMode == viewDualPane {
				m.viewMode = viewSinglePane
				m.calculateLayout()
			}

		case "up", "k":
			if m.viewMode == viewDualPane {
				// In dual-pane mode, check which pane is focused
				if m.focusedPane == leftPane {
					// Scroll file list
					if m.cursor > 0 {
						m.cursor--
						// Update preview if file selected
						if len(m.files) > 0 && !m.files[m.cursor].isDir {
							m.loadPreview(m.files[m.cursor].path)
						}
					}
				} else {
					// Scroll preview up
					if m.preview.scrollPos > 0 {
						m.preview.scrollPos--
					}
				}
			} else {
				// Single-pane mode: just move cursor
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "j":
			if m.viewMode == viewDualPane {
				// In dual-pane mode, check which pane is focused
				if m.focusedPane == leftPane {
					// Scroll file list
					if m.cursor < len(m.files)-1 {
						m.cursor++
						// Update preview if file selected
						if len(m.files) > 0 && !m.files[m.cursor].isDir {
							m.loadPreview(m.files[m.cursor].path)
						}
					}
				} else {
					// Scroll preview down
					// Calculate visible lines: m.height - 5 (header) - 2 (preview title) = m.height - 7
			visibleLines := m.height - 7
			maxScroll := len(m.preview.content) - visibleLines
					if maxScroll < 0 {
						maxScroll = 0
					}
					if m.preview.scrollPos < maxScroll {
						m.preview.scrollPos++
					}
				}
			} else {
				// Single-pane mode: just move cursor
				if m.cursor < len(m.files)-1 {
					m.cursor++
				}
			}

		case "enter":
			if len(m.files) > 0 {
				if m.files[m.cursor].isDir {
					// Navigate into directory
					m.currentPath = m.files[m.cursor].path
					m.cursor = 0
					m.loadFiles()
				} else {
					// Preview file
					if m.viewMode == viewDualPane {
						// Already in dual-pane, just load preview
						m.loadPreview(m.files[m.cursor].path)
					} else {
						// Enter full-screen preview
						m.loadPreview(m.files[m.cursor].path)
						m.viewMode = viewFullPreview
					}
				}
			}

		case "tab":
			// In dual-pane mode: switch focus between panes
			// In single-pane mode: enter dual-pane mode
			if m.viewMode == viewDualPane {
				// Toggle focus between left and right panes
				if m.focusedPane == leftPane {
					m.focusedPane = rightPane
				} else {
					m.focusedPane = leftPane
				}
			} else if m.viewMode == viewSinglePane {
				// Enter dual-pane mode
				m.viewMode = viewDualPane
				m.focusedPane = leftPane
				m.calculateLayout()
				// Load preview of current file
				if len(m.files) > 0 && !m.files[m.cursor].isDir {
					m.loadPreview(m.files[m.cursor].path)
				}
			}

		case " ":
			// Space: toggle dual-pane mode on/off
			if m.viewMode == viewSinglePane {
				m.viewMode = viewDualPane
				m.focusedPane = leftPane
				m.calculateLayout()
				// Load preview of current file
				if len(m.files) > 0 && !m.files[m.cursor].isDir {
					m.loadPreview(m.files[m.cursor].path)
				}
			} else if m.viewMode == viewDualPane {
				m.viewMode = viewSinglePane
				m.calculateLayout()
			}

		case "f":
			// Force full-screen preview
			if len(m.files) > 0 && !m.files[m.cursor].isDir {
				m.loadPreview(m.files[m.cursor].path)
				m.viewMode = viewFullPreview
			}

		case "pageup":
			// Page up in dual-pane mode (only works when right pane focused)
			if m.viewMode == viewDualPane && m.focusedPane == rightPane {
				visibleLines := m.height - 7
				m.preview.scrollPos -= visibleLines
				if m.preview.scrollPos < 0 {
					m.preview.scrollPos = 0
				}
			}

		case "pagedown":
			// Page down in dual-pane mode (only works when right pane focused)
			if m.viewMode == viewDualPane && m.focusedPane == rightPane {
				// Calculate visible lines: m.height - 5 (header) - 2 (preview title) = m.height - 7
			visibleLines := m.height - 7
			maxScroll := len(m.preview.content) - visibleLines
				if maxScroll < 0 {
					maxScroll = 0
				}
				m.preview.scrollPos += visibleLines
				if m.preview.scrollPos > maxScroll {
					m.preview.scrollPos = maxScroll
				}
			}

		case "h", "left":
			// Go to parent directory
			if m.currentPath != "/" {
				m.currentPath = filepath.Dir(m.currentPath)
				m.cursor = 0
				m.loadFiles()
			}

		case ".", "ctrl+h":
			// Toggle hidden files
			m.showHidden = !m.showHidden
			m.loadFiles()

		case "v":
			// Cycle through display modes
			m.displayMode = (m.displayMode + 1) % 4

		case "1":
			// Switch to list view
			m.displayMode = modeList

		case "2":
			// Switch to grid view
			m.displayMode = modeGrid

		case "3":
			// Switch to detail view
			m.displayMode = modeDetail

		case "4":
			// Switch to tree view
			m.displayMode = modeTree

		case "e", "E":
			// Edit file in external editor
			if len(m.files) > 0 && !m.files[m.cursor].isDir {
				editor := getAvailableEditor()
				if editor == "" {
					// Could show error message - for now, do nothing
					return m, nil
				}
				// Prefer micro if available, otherwise use whatever was found
				if editorAvailable("micro") {
					editor = "micro"
				}
				return m, openEditor(editor, m.files[m.cursor].path)
			}

		case "n", "N":
			// Edit file in nano specifically
			if len(m.files) > 0 && !m.files[m.cursor].isDir {
				if editorAvailable("nano") {
					return m, openEditor("nano", m.files[m.cursor].path)
				}
			}

		case "y":
			// Copy file path to clipboard (vim-style "yank")
			if len(m.files) > 0 {
				path := m.files[m.cursor].path
				err := copyToClipboard(path)
				if err != nil {
					// Could show error - for now, silently continue
					// In the future, we could add a status message system
				}
				// Success - path is copied to clipboard
			}

		case "c":
			// Copy file path (alternative to y)
			if len(m.files) > 0 {
				path := m.files[m.cursor].path
				_ = copyToClipboard(path)
			}
		}

	case editorFinishedMsg:
		// Editor has closed, we're back in TFE
		// Refresh file list in case file was modified
		m.loadFiles()
		return m, nil

	case tea.MouseMsg:
		// Handle mouse wheel scrolling in full-screen preview mode
		if m.viewMode == viewFullPreview {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				if m.preview.scrollPos > 0 {
					m.preview.scrollPos--
				}
			case tea.MouseButtonWheelDown:
				maxScroll := len(m.preview.content) - (m.height - 6)
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.preview.scrollPos < maxScroll {
					m.preview.scrollPos++
				}
			}
			return m, nil
		}

		// In dual-pane mode, detect which pane was clicked to switch focus
		if m.viewMode == viewDualPane && msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			// Check if click is in left or right pane
			if msg.X < m.leftWidth {
				m.focusedPane = leftPane
			} else if msg.X > m.leftWidth { // Account for separator
				m.focusedPane = rightPane
			}
		}

		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Action == tea.MouseActionRelease {
				// In dual-pane mode, only process file clicks if within left pane
				if m.viewMode == viewDualPane && msg.X >= m.leftWidth {
					// Click is in right pane or beyond - don't select files
					break
				}

				// Calculate which item was clicked (accounting for header lines)
				// Base offset: title + path + spacing = 3 lines (both single and dual-pane)
				// Detail mode adds 2 extra lines (column header + separator)
				headerOffset := 3
				if m.displayMode == modeDetail {
					headerOffset += 2 // Add 2 for detail view's header and separator
				}
				clickedIndex := msg.Y - headerOffset
				if clickedIndex >= 0 && clickedIndex < len(m.files) {
					now := time.Now()

					// Check for double-click: same item clicked within 500ms
					const doubleClickThreshold = 500 * time.Millisecond
					isDoubleClick := clickedIndex == m.lastClickIndex &&
						now.Sub(m.lastClickTime) < doubleClickThreshold

					if isDoubleClick {
						// Double-click: navigate or preview
						if m.files[clickedIndex].isDir {
							m.currentPath = m.files[clickedIndex].path
							m.cursor = 0
							m.loadFiles()
						} else {
							// Preview file (same as Enter key)
							if m.viewMode == viewDualPane {
								m.loadPreview(m.files[clickedIndex].path)
							} else {
								m.loadPreview(m.files[clickedIndex].path)
								m.viewMode = viewFullPreview
							}
						}
						// Reset click tracking after double-click
						m.lastClickIndex = -1
						m.lastClickTime = time.Time{}
					} else {
						// Single-click: just select and update preview in dual-pane
						m.cursor = clickedIndex
						m.lastClickIndex = clickedIndex
						m.lastClickTime = now

						// Update preview in dual-pane mode
						if m.viewMode == viewDualPane && !m.files[m.cursor].isDir {
							m.loadPreview(m.files[m.cursor].path)
						}
					}
				}
			}

		case tea.MouseButtonWheelUp:
			if m.viewMode == viewDualPane && m.focusedPane == rightPane {
				// Scroll preview up when right pane focused
				if m.preview.scrollPos > 0 {
					m.preview.scrollPos--
				}
			} else {
				// Scroll file list
				if m.cursor > 0 {
					m.cursor--
					// Don't auto-update preview on wheel scroll - only on explicit selection
				}
			}

		case tea.MouseButtonWheelDown:
			if m.viewMode == viewDualPane && m.focusedPane == rightPane {
				// Scroll preview down when right pane focused
				visibleLines := m.height - 7
				maxScroll := len(m.preview.content) - visibleLines
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.preview.scrollPos < maxScroll {
					m.preview.scrollPos++
				}
			} else {
				// Scroll file list
				if m.cursor < len(m.files)-1 {
					m.cursor++
					// Don't auto-update preview on wheel scroll - only on explicit selection
				}
			}
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.calculateGridLayout() // Recalculate grid columns on resize
		m.calculateLayout()     // Recalculate pane layout on resize
	}

	return m, nil
}

// renderPreview renders the preview pane content
func (m model) renderPreview(maxVisible int) string {
	var s strings.Builder

	if !m.preview.loaded {
		s.WriteString("No file loaded")
		return s.String()
	}

	// Calculate visible range based on scroll position
	start := m.preview.scrollPos
	end := start + maxVisible
	if end > len(m.preview.content) {
		end = len(m.preview.content)
	}
	if start > len(m.preview.content) {
		start = 0
		end = maxVisible
		if end > len(m.preview.content) {
			end = len(m.preview.content)
		}
	}

	// Calculate available width for content (pane width - line number width - border - padding)
	// Line number is 8 chars: "9999 │ " (5 for number, 1 for space, 1 for │, 1 for space)
	// Borders take up 2-4 additional characters depending on lipgloss rendering
	availableWidth := m.rightWidth - 15 // More conservative: line nums (8) + borders (4) + padding (3)
	if m.viewMode == viewFullPreview {
		availableWidth = m.width - 15
	}
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// Render lines with line numbers
	for i := start; i < end; i++ {
		// Use consistent 5-character width for line numbers (up to 9999 lines)
		lineNum := fmt.Sprintf("%5d │ ", i+1)
		lineNumStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
		s.WriteString(lineNumStyle.Render(lineNum))

		// Truncate line to prevent wrapping
		line := m.preview.content[i]
		if len(line) > availableWidth {
			line = line[:availableWidth-3] + "..."
		}
		s.WriteString(line)
		s.WriteString("\n")
	}

	return s.String()
}

// renderFullPreview renders the full-screen preview mode
func (m model) renderFullPreview() string {
	var s strings.Builder

	// Title bar with file name
	previewTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("39")).
		Width(m.width).
		Padding(0, 1)

	titleText := fmt.Sprintf("Preview: %s", m.preview.fileName)
	if m.preview.tooLarge || m.preview.isBinary {
		titleText += " [Cannot Preview]"
	}
	s.WriteString(previewTitleStyle.Render(titleText))
	s.WriteString("\n")

	// File info line
	infoText := fmt.Sprintf("Size: %s | Lines: %d | Scroll: %d-%d",
		formatFileSize(m.preview.fileSize),
		len(m.preview.content),
		m.preview.scrollPos+1,
		m.preview.scrollPos+m.height-4)
	s.WriteString(pathStyle.Render(infoText))
	s.WriteString("\n")

	// Content
	maxVisible := m.height - 4 // Reserve space for title, info, and help
	s.WriteString(m.renderPreview(maxVisible))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	s.WriteString(helpStyle.Render("↑/↓: scroll • PgUp/PgDown: page • E: edit • y/c: copy path • Esc: close • q: quit"))

	return s.String()
}

// renderDualPane renders the split-pane layout using Lipgloss layout utilities
func (m model) renderDualPane() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("TFE - Terminal File Explorer [Dual-Pane]"))
	s.WriteString("\n")

	// Current path
	s.WriteString(pathStyle.Render(m.currentPath))
	s.WriteString("\n")

	// Calculate max visible for both panes
	maxVisible := m.height - 5

	// Get left pane content
	var leftContent string
	switch m.displayMode {
	case modeList:
		leftContent = m.renderListView(maxVisible)
	case modeGrid:
		leftContent = m.renderGridView(maxVisible)
	case modeDetail:
		leftContent = m.renderDetailView(maxVisible)
	case modeTree:
		leftContent = m.renderTreeView(maxVisible)
	default:
		leftContent = m.renderListView(maxVisible)
	}

	// Get right pane content
	rightContent := ""
	if m.preview.loaded {
		previewTitleText := fmt.Sprintf("Preview: %s", m.preview.fileName)
		previewTitle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Render(previewTitleText)
		separatorLine := strings.Repeat("─", len(previewTitleText))
		rightContent = previewTitle + "\n" + separatorLine + "\n"
		rightContent += m.renderPreview(maxVisible - 2)
	} else {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)
		rightContent = emptyStyle.Render("No preview available\n\nSelect a file to preview")
	}

	// Create styled boxes for left and right panes using Lipgloss
	// Highlight the focused pane with a brighter border color
	leftBorderColor := "241"  // dim gray
	rightBorderColor := "241" // dim gray
	if m.focusedPane == leftPane {
		leftBorderColor = "39" // bright blue for focused pane
	} else {
		rightBorderColor = "39" // bright blue for focused pane
	}

	leftPaneStyle := lipgloss.NewStyle().
		Width(m.leftWidth).
		MaxWidth(m.leftWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderRight(true).
		BorderForeground(lipgloss.Color(leftBorderColor))

	rightPaneStyle := lipgloss.NewStyle().
		Width(m.rightWidth).
		MaxWidth(m.rightWidth).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderLeft(true).
		BorderForeground(lipgloss.Color(rightBorderColor))

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
	statusText := fmt.Sprintf("%s%s • %s", itemsInfo, hiddenIndicator, m.displayMode.String())
	s.WriteString(statusStyle.Render(statusText))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	focusInfo := ""
	if m.focusedPane == leftPane {
		focusInfo = "[LEFT focused]"
	} else {
		focusInfo = "[RIGHT focused]"
	}
	s.WriteString(helpStyle.Render(fmt.Sprintf("%s • ↑/↓: scroll %s • Tab: switch pane • Space: exit • E: edit • y/c: copy", focusInfo,
		map[bool]string{true: "list", false: "preview"}[m.focusedPane == leftPane])))

	return s.String()
}

// ansiRegex matches ANSI escape codes
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripAnsi removes ANSI escape codes from a string
func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// visibleWidth returns the visible width of a string (without ANSI codes)
func visibleWidth(s string) int {
	return len(stripAnsi(s))
}

// truncateOrPad truncates or pads a string to the specified width
// This function strips ANSI codes to ensure consistent alignment
func truncateOrPad(s string, width int) string {
	// Strip ANSI codes first for consistent width calculation and rendering
	stripped := stripAnsi(s)
	visible := len(stripped)

	if visible > width {
		// Truncate to width
		return stripped[:width]
	}

	// Pad to exact width
	padding := width - visible
	if padding > 0 {
		return stripped + strings.Repeat(" ", padding)
	}
	return stripped
}

func (m model) View() string {
	// Dispatch to appropriate view based on viewMode
	switch m.viewMode {
	case viewFullPreview:
		return m.renderFullPreview()
	case viewDualPane:
		return m.renderDualPane()
	default:
		// Single-pane mode (original view)
		return m.renderSinglePane()
	}
}

// renderSinglePane renders the original single-pane file browser
func (m model) renderSinglePane() string {
	var s strings.Builder

	// Debug: Write file count to understand if something is off
	// Title - ALWAYS render this
	title := titleStyle.Render("TFE - Terminal File Explorer")
	s.WriteString(title)
	s.WriteString("\n")

	// Current path
	s.WriteString(pathStyle.Render(m.currentPath))
	s.WriteString("\n")

	// File list - render based on current display mode
	// Calculate maxVisible: m.height - (title=1 + path=1 + path_padding=1 + status=1 + help=1) = m.height - 5
	// Note: pathStyle has PaddingBottom(1) which adds an extra rendered line
	maxVisible := m.height - 5 // Reserve space for title, path (with padding), status, and help

	switch m.displayMode {
	case modeList:
		s.WriteString(m.renderListView(maxVisible))
	case modeGrid:
		s.WriteString(m.renderGridView(maxVisible))
	case modeDetail:
		s.WriteString(m.renderDetailView(maxVisible))
	case modeTree:
		s.WriteString(m.renderTreeView(maxVisible))
	default:
		s.WriteString(m.renderListView(maxVisible))
	}

	// Status bar
	s.WriteString("\n")

	// Count directories and files
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

	// Selected file info
	var selectedInfo string
	if len(m.files) > 0 && m.cursor < len(m.files) {
		selected := m.files[m.cursor]
		if selected.isDir {
			selectedInfo = fmt.Sprintf("Selected: %s (folder)", selected.name)
		} else {
			selectedInfo = fmt.Sprintf("Selected: %s (%s, %s)",
				selected.name,
				formatFileSize(selected.size),
				formatModTime(selected.modTime))
		}
	}

	itemsInfo := fmt.Sprintf("%d items", len(m.files))
	if dirCount > 0 || fileCount > 0 {
		itemsInfo = fmt.Sprintf("%d folders, %d files", dirCount, fileCount)
	}

	hiddenIndicator := ""
	if m.showHidden {
		hiddenIndicator = " • showing hidden"
	}

	// View mode indicator
	viewModeText := fmt.Sprintf(" • view: %s", m.displayMode.String())

	statusText := fmt.Sprintf("%s%s%s | %s", itemsInfo, hiddenIndicator, viewModeText, selectedInfo)
	s.WriteString(statusStyle.Render(statusText))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	s.WriteString(helpStyle.Render("↑/↓: nav • enter: preview • E: edit • y/c: copy path • Tab: dual-pane • f: full • v: views • q: quit"))

	return s.String()
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
