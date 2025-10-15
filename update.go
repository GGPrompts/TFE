package main

import (
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
				return m, tea.ClearScreen

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

		// Handle command prompt input (MC-style: always active, no focus needed)
		// Special keys that interact with command prompt
		switch msg.String() {
		case "enter":
			// Execute command if there's input
			if m.commandInput != "" {
				cmd := m.commandInput
				m.addToHistory(cmd)
				m.commandInput = ""
				return m, runCommand(cmd, m.currentPath)
			}
			// If no command input, handle Enter for file navigation (below)

		case "backspace":
			// Delete last character from command if there's input
			if len(m.commandInput) > 0 {
				m.commandInput = m.commandInput[:len(m.commandInput)-1]
				return m, nil
			}
			// If no command input, backspace does nothing (could navigate up later)

		case "esc":
			// Clear command input if there's any
			if m.commandInput != "" {
				m.commandInput = ""
				return m, nil
			}
			// If no command input, handle Esc for dual-pane exit (below)
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

		case "up":
			// If command input exists, navigate command history
			if m.commandInput != "" || len(m.commandHistory) > 0 {
				m.commandInput = m.getPreviousCommand()
				return m, nil
			}
			// Otherwise fall through to file navigation
			fallthrough
		case "k":
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

		case "down":
			// If command input exists or history available, navigate command history
			if m.commandInput != "" || len(m.commandHistory) > 0 {
				m.commandInput = m.getNextCommand()
				return m, nil
			}
			// Otherwise fall through to file navigation
			fallthrough
		case "j":
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
					// Enter full-screen preview (regardless of current mode)
					m.loadPreview(m.files[m.cursor].path)
					m.viewMode = viewFullPreview
					return m, tea.ClearScreen
				}
			}

		case "tab":
			// In dual-pane mode: cycle focus between left and right pane
			// In single-pane mode: enter dual-pane mode
			if m.viewMode == viewDualPane {
				// Cycle through: left → right → left
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
				return m, tea.ClearScreen
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

		default:
			// MC-style: any single character goes to command prompt
			if len(msg.String()) == 1 {
				m.commandInput += msg.String()
				m.historyPos = len(m.commandHistory)
			}
		}

	case editorFinishedMsg:
		// Editor has closed, we're back in TFE
		// Refresh file list in case file was modified
		m.loadFiles()
		// Force a refresh and re-enable mouse support (external editors disable it)
		return m, tea.Batch(
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)

	case commandFinishedMsg:
		// Command has finished, we're back in TFE
		// Refresh file list in case command modified files
		m.loadFiles()
		// Force a refresh and re-enable mouse support (shell commands may disable it)
		return m, tea.Batch(
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)

	case tea.MouseMsg:
		// Handle mouse wheel scrolling in full-screen preview mode
		if m.viewMode == viewFullPreview {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				// Scroll 3 lines per wheel tick for smoother scrolling
				m.preview.scrollPos -= 3
				if m.preview.scrollPos < 0 {
					m.preview.scrollPos = 0
				}
			case tea.MouseButtonWheelDown:
				maxScroll := len(m.preview.content) - (m.height - 6)
				if maxScroll < 0 {
					maxScroll = 0
				}
				// Scroll 3 lines per wheel tick for smoother scrolling
				m.preview.scrollPos += 3
				if m.preview.scrollPos > maxScroll {
					m.preview.scrollPos = maxScroll
				}
			}
			return m, nil
		}

		// In dual-pane mode, detect which pane was clicked to switch focus
		if m.viewMode == viewDualPane && msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			// Check if click is in left or right pane (not in header or status bar)
			if msg.Y >= 4 && msg.Y < m.height-1 { // Skip header (4 lines) and status bar (1 line)
				if msg.X < m.leftWidth {
					m.focusedPane = leftPane
				} else if msg.X > m.leftWidth { // Account for separator
					m.focusedPane = rightPane
				}
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
				// Base offset: title(0) + path(1) + command(2) + separator(3) + file_list_starts(4) = 4 lines
				// Detail mode adds 2 extra lines (column header + separator)
				headerOffset := 4
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
						// Double-click: navigate or full-screen preview
						if m.files[clickedIndex].isDir {
							m.currentPath = m.files[clickedIndex].path
							m.cursor = 0
							m.loadFiles()
						} else {
							// Enter full-screen preview (same as Enter key)
							m.loadPreview(m.files[clickedIndex].path)
							m.viewMode = viewFullPreview
							// Reset click tracking after double-click
							m.lastClickIndex = -1
							m.lastClickTime = time.Time{}
							return m, tea.ClearScreen
						}
						// Reset click tracking after double-click (for directory navigation)
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
				// Scroll preview up when right pane focused (3 lines per tick)
				m.preview.scrollPos -= 3
				if m.preview.scrollPos < 0 {
					m.preview.scrollPos = 0
				}
			} else {
				// Scroll file list
				if m.cursor > 0 {
					m.cursor--
					// Update preview in dual-pane mode
					if m.viewMode == viewDualPane && len(m.files) > 0 && !m.files[m.cursor].isDir {
						m.loadPreview(m.files[m.cursor].path)
					}
				}
			}

		case tea.MouseButtonWheelDown:
			if m.viewMode == viewDualPane && m.focusedPane == rightPane {
				// Scroll preview down when right pane focused (3 lines per tick)
				visibleLines := m.height - 7
				maxScroll := len(m.preview.content) - visibleLines
				if maxScroll < 0 {
					maxScroll = 0
				}
				m.preview.scrollPos += 3
				if m.preview.scrollPos > maxScroll {
					m.preview.scrollPos = maxScroll
				}
			} else {
				// Scroll file list
				if m.cursor < len(m.files)-1 {
					m.cursor++
					// Update preview in dual-pane mode
					if m.viewMode == viewDualPane && len(m.files) > 0 && !m.files[m.cursor].isDir {
						m.loadPreview(m.files[m.cursor].path)
					}
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
