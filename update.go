package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle preview mode keys first
		if m.viewMode == viewFullPreview {
			switch msg.String() {
			case "f10", "ctrl+c", "esc":
				// Exit preview mode (F10 replaces q)
				m.viewMode = viewSinglePane
				m.calculateLayout()
				// Re-enable mouse for navigation
				return m, tea.Batch(tea.ClearScreen, tea.EnableMouseCellMotion)

			case "f4":
				// Edit file in external editor from preview (F4 replaces e/E)
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

			case "f5":
				// Copy file path from preview (F5 replaces y)
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
				totalLines := m.getWrappedLineCount()
				maxScroll := totalLines - (m.height - 6)
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
				totalLines := m.getWrappedLineCount()
				maxScroll := totalLines - (m.height - 6)
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

		// Handle context menu input if menu is open
		if m.contextMenuOpen {
			switch msg.String() {
			case "esc", "q":
				// Close context menu and clear screen to remove visual artifacts
				m.contextMenuOpen = false
				return m, tea.ClearScreen

			case "up", "k":
				// Navigate up in menu
				if m.contextMenuCursor > 0 {
					m.contextMenuCursor--
				}
				// Clear screen to prevent ANSI overlay artifacts
				return m, tea.ClearScreen

			case "down", "j":
				// Navigate down in menu
				menuItems := m.getContextMenuItems()
				if m.contextMenuCursor < len(menuItems)-1 {
					m.contextMenuCursor++
				}
				// Clear screen to prevent ANSI overlay artifacts
				return m, tea.ClearScreen

			case "enter":
				// Execute selected menu action
				return m.executeContextMenuAction()
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
				// Check for exit/quit commands
				cmdLower := strings.ToLower(strings.TrimSpace(cmd))
				if cmdLower == "exit" || cmdLower == "quit" {
					return m, tea.Quit
				}
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

		// Check if user is typing in command prompt
		// If commandInput has text, prioritize adding to it over hotkeys
		// This allows typing 'e', 'v', 'f', etc. in commands
		key := msg.String()
		if len(m.commandInput) > 0 && len(key) == 1 {
			// User is actively typing a command - add this character
			m.commandInput += key
			return m, nil
		}

		// Regular file browser keys
		switch msg.String() {
		case "f10", "ctrl+c":
			// F10: Quit (replaces q)
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
						if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
							m.loadPreview(currentFile.path)
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
					maxCursor := m.getMaxCursor()
					if m.cursor < maxCursor {
						m.cursor++
						// Update preview if file selected
						if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
							m.loadPreview(currentFile.path)
						}
					}
				} else {
					// Scroll preview down
					// Calculate visible lines: m.height - 5 (header) - 2 (preview title) = m.height - 7
					visibleLines := m.height - 7
					totalLines := m.getWrappedLineCount()
					maxScroll := totalLines - visibleLines
					if maxScroll < 0 {
						maxScroll = 0
					}
					if m.preview.scrollPos < maxScroll {
						m.preview.scrollPos++
					}
				}
			} else {
				// Single-pane mode: just move cursor
				maxCursor := m.getMaxCursor()
				if m.cursor < maxCursor {
					m.cursor++
				}
			}

		case "enter":
			if currentFile := m.getCurrentFile(); currentFile != nil {
				// If in favorites mode, check if we need to navigate to a different directory
				if m.showFavoritesOnly && currentFile.name != ".." {
					// Check if favorite is in a different location than current path
					fileDir := filepath.Dir(currentFile.path)
					if currentFile.isDir {
						// Navigate to the favorited directory
						m.currentPath = currentFile.path
						m.cursor = 0
						m.showFavoritesOnly = false // Exit favorites mode
						m.loadFiles()
					} else if fileDir != m.currentPath {
						// Navigate to the file's parent directory and select it
						m.currentPath = fileDir
						m.showFavoritesOnly = false // Exit favorites mode
						m.loadFiles()
						// Find and select the file
						for i, f := range m.files {
							if f.path == currentFile.path {
								m.cursor = i
								break
							}
						}
					} else {
						// File is in current directory, just preview it
						m.loadPreview(currentFile.path)
						m.viewMode = viewFullPreview
						// Disable mouse to allow text selection
						return m, tea.Batch(tea.ClearScreen, func() tea.Msg { return tea.DisableMouse() })
					}
				} else if currentFile.isDir {
					// In tree view: toggle expansion instead of navigating
					if m.displayMode == modeTree && currentFile.name != ".." {
						m.expandedDirs[currentFile.path] = !m.expandedDirs[currentFile.path]
					} else {
						// Other modes: navigate into directory
						m.currentPath = currentFile.path
						m.cursor = 0
						m.loadFiles()
					}
				} else {
					// Enter full-screen preview (regardless of current mode)
					m.loadPreview(currentFile.path)
					m.viewMode = viewFullPreview
					// Disable mouse to allow text selection
					return m, tea.Batch(tea.ClearScreen, func() tea.Msg { return tea.DisableMouse() })
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
				if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
					m.loadPreview(currentFile.path)
				}
			}

		case " ":
			// Space: toggle dual-pane mode on/off
			if m.viewMode == viewSinglePane {
				m.viewMode = viewDualPane
				m.focusedPane = leftPane
				m.calculateLayout()
				// Load preview of current file
				if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
					m.loadPreview(currentFile.path)
				}
			} else if m.viewMode == viewDualPane {
				m.viewMode = viewSinglePane
				m.calculateLayout()
			}

		case "f3":
			// F3: Force full-screen preview (replaces f)
			if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
				m.loadPreview(currentFile.path)
				m.viewMode = viewFullPreview
				// Disable mouse to allow text selection
				return m, tea.Batch(tea.ClearScreen, func() tea.Msg { return tea.DisableMouse() })
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
				totalLines := m.getWrappedLineCount()
				maxScroll := totalLines - visibleLines
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

		case "f9":
			// F9: Cycle through display modes (replaces v)
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

		case "f4":
			// F4: Edit file in external editor (replaces e/E)
			if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
				editor := getAvailableEditor()
				if editor == "" {
					// Could show error message - for now, do nothing
					return m, nil
				}
				// Prefer micro if available, otherwise use whatever was found
				if editorAvailable("micro") {
					editor = "micro"
				}
				return m, openEditor(editor, currentFile.path)
			}

		case "n", "N":
			// Edit file in nano specifically
			if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
				if editorAvailable("nano") {
					return m, openEditor("nano", currentFile.path)
				}
			}

		case "f5":
			// F5: Copy file path to clipboard (replaces y)
			if currentFile := m.getCurrentFile(); currentFile != nil {
				err := copyToClipboard(currentFile.path)
				if err != nil {
					// Could show error - for now, silently continue
					// In the future, we could add a status message system
				}
				// Success - path is copied to clipboard
			}

		case "s", "S":
			// Toggle favorite for current file/folder
			if currentFile := m.getCurrentFile(); currentFile != nil {
				m.toggleFavorite(currentFile.path)
			}

		case "f6":
			// F6: Toggle favorites filter (replaces b/B)
			m.showFavoritesOnly = !m.showFavoritesOnly

		case "f1":
			// F1: Show hotkeys reference (replaces ?)
			hotkeysPath := filepath.Join(filepath.Dir(m.currentPath), "HOTKEYS.md")
			// Try to find HOTKEYS.md in the TFE directory
			// First check if it exists in current directory
			if _, err := os.Stat(hotkeysPath); os.IsNotExist(err) {
				// Try executable directory
				if exePath, err := os.Executable(); err == nil {
					hotkeysPath = filepath.Join(filepath.Dir(exePath), "HOTKEYS.md")
				}
			}
			// Load and show the hotkeys file if it exists
			if _, err := os.Stat(hotkeysPath); err == nil {
				m.loadPreview(hotkeysPath)
				m.viewMode = viewFullPreview
				// Disable mouse to allow text selection
				return m, tea.Batch(tea.ClearScreen, func() tea.Msg { return tea.DisableMouse() })
			}

		case "f2":
			// F2: Open context menu at cursor position (keyboard alternative to right-click)
			if currentFile := m.getCurrentFile(); currentFile != nil {
				// Calculate menu position based on cursor
				headerOffset := 4
				if m.displayMode == modeDetail {
					headerOffset = 6 // Account for detail view header
				}

				// Calculate visible range to account for scrolling
				maxVisible := m.height - 6
				if m.displayMode == modeDetail {
					maxVisible -= 2
				}
				start, _ := m.getVisibleRange(maxVisible)

				// Calculate Y position relative to visible cursor position
				menuY := headerOffset + (m.cursor - start)
				menuX := 10 // Left margin for menu (increased for border visibility)

				// Open menu
				m.contextMenuOpen = true
				m.contextMenuX = menuX
				m.contextMenuY = menuY
				m.contextMenuFile = currentFile
				m.contextMenuCursor = 0
			}

		case "f7":
			// F7: Create directory (placeholder for future feature)
			// TODO: Implement directory creation

		case "f8":
			// F8: Delete file/folder (placeholder for future feature)
			// TODO: Implement file deletion with confirmation

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
				totalLines := m.getWrappedLineCount()
				maxScroll := totalLines - (m.height - 6)
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
			// Header is 4 lines total (title, path, command, separator)
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
				// Handle context menu clicks if menu is open
				if m.contextMenuOpen {
					// Calculate menu bounds
					menuItems := m.getContextMenuItems()
					menuHeight := len(menuItems) + 2 // items + top/bottom border
					// Calculate menu width from items
					maxWidth := 0
					for _, item := range menuItems {
						width := visualWidth(item.label)
						if width > maxWidth {
							maxWidth = width
						}
					}
					menuWidth := maxWidth + 4 + 2 // padding + borders

					// Check if click is within menu bounds
					if msg.X >= m.contextMenuX && msg.X <= m.contextMenuX+menuWidth &&
						msg.Y >= m.contextMenuY && msg.Y <= m.contextMenuY+menuHeight {
						// Click is inside menu - calculate which item was clicked
						clickedItemIndex := msg.Y - m.contextMenuY - 1 // -1 for top border
						if clickedItemIndex >= 0 && clickedItemIndex < len(menuItems) {
							// Update cursor and execute the clicked item
							m.contextMenuCursor = clickedItemIndex
							return m.executeContextMenuAction()
						}
					}

					// Click is outside menu - close it
					m.contextMenuOpen = false
					return m, tea.ClearScreen
				}

				// In dual-pane mode, only process file clicks if within left pane
				if m.viewMode == viewDualPane && msg.X >= m.leftWidth {
					// Click is in right pane or beyond - don't select files
					break
				}

				// Calculate which item was clicked (accounting for header lines and scrolling)
				// Both modes: title(0) + path(1) + command(2) + separator(3) = 4 lines
				// Lipgloss borders are only on sides (BorderRight/BorderLeft), not top/bottom
				// So file list starts at line 4 in both modes
				headerOffset := 4
				if m.displayMode == modeDetail {
					headerOffset += 2 // Add 2 for detail view's header and separator
				}

				// Calculate visible range to account for scrolling
				maxVisible := m.height - 6
				if m.displayMode == modeDetail {
					maxVisible -= 2 // Account for detail header
				}

				var clickedIndex int
				var clickedLine int

				// Grid view requires calculating both row and column from X,Y coordinates
				if m.displayMode == modeGrid {
					// Calculate which row was clicked
					clickedRow := msg.Y - headerOffset

					// Calculate which column was clicked
					// Each grid cell is approximately: icon(2) + space(1) + name(12) + padding(2) = 17 chars
					cellWidth := 17
					clickedCol := msg.X / cellWidth
					if clickedCol >= m.gridColumns {
						clickedCol = m.gridColumns - 1
					}

					// Calculate visible row range (grid mode uses rows, not items)
					totalItems := len(m.files)
					rows := (totalItems + m.gridColumns - 1) / m.gridColumns

					startRow := 0
					endRow := rows
					if rows > maxVisible {
						cursorRow := m.cursor / m.gridColumns
						startRow = cursorRow - maxVisible/2
						if startRow < 0 {
							startRow = 0
						}
						endRow = startRow + maxVisible
						if endRow > rows {
							endRow = rows
							startRow = endRow - maxVisible
							if startRow < 0 {
								startRow = 0
							}
						}
					}

					// Convert click to item index
					actualRow := startRow + clickedRow
					clickedIndex = actualRow*m.gridColumns + clickedCol

					// Validate the clicked index is within bounds
					if clickedRow < 0 || actualRow >= endRow || clickedIndex >= len(m.files) {
						clickedIndex = -1
					}
				} else {
					// List, Detail, and Tree modes: one item per line
					start, end := m.getVisibleRange(maxVisible)
					clickedLine = msg.Y - headerOffset
					clickedIndex = start + clickedLine

					// Validate bounds
					if clickedLine < 0 || clickedIndex >= end || clickedIndex >= len(m.files) {
						clickedIndex = -1
					}
				}

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
							// Disable mouse to allow text selection
							return m, tea.Batch(tea.ClearScreen, func() tea.Msg { return tea.DisableMouse() })
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

		case tea.MouseButtonRight:
			// Right-click: open context menu
			if msg.Action == tea.MouseActionRelease {
				// Close any existing menu first to prevent phantoms
				if m.contextMenuOpen {
					m.contextMenuOpen = false
				}

				// Don't open menu in preview mode or if in right pane
				if m.viewMode == viewFullPreview {
					break
				}
				if m.viewMode == viewDualPane && msg.X >= m.leftWidth {
					break
				}

				// Calculate which item was right-clicked
				headerOffset := 4
				if m.displayMode == modeDetail {
					headerOffset += 2
				}

				maxVisible := m.height - 6
				if m.displayMode == modeDetail {
					maxVisible -= 2
				}

				var clickedIndex int

				// Grid view: calculate row and column
				if m.displayMode == modeGrid {
					clickedRow := msg.Y - headerOffset
					cellWidth := 17
					clickedCol := msg.X / cellWidth
					if clickedCol >= m.gridColumns {
						clickedCol = m.gridColumns - 1
					}

					totalItems := len(m.files)
					rows := (totalItems + m.gridColumns - 1) / m.gridColumns

					startRow := 0
					endRow := rows
					if rows > maxVisible {
						cursorRow := m.cursor / m.gridColumns
						startRow = cursorRow - maxVisible/2
						if startRow < 0 {
							startRow = 0
						}
						endRow = startRow + maxVisible
						if endRow > rows {
							endRow = rows
							startRow = endRow - maxVisible
							if startRow < 0 {
								startRow = 0
							}
						}
					}

					actualRow := startRow + clickedRow
					clickedIndex = actualRow*m.gridColumns + clickedCol

					if clickedRow < 0 || actualRow >= endRow || clickedIndex >= len(m.files) {
						clickedIndex = -1
					}
				} else {
					// List, Detail, Tree modes: one item per line
					start, end := m.getVisibleRange(maxVisible)
					clickedLine := msg.Y - headerOffset
					clickedIndex = start + clickedLine

					if clickedLine < 0 || clickedIndex >= end || clickedIndex >= len(m.files) {
						clickedIndex = -1
					}
				}

				// Open context menu if a valid file was clicked
				if clickedIndex >= 0 && clickedIndex < len(m.files) {
					m.contextMenuOpen = true
					// Ensure menu has enough left margin for border to show
					m.contextMenuX = msg.X
					if m.contextMenuX < 2 {
						m.contextMenuX = 2
					}
					m.contextMenuY = msg.Y
					m.contextMenuFile = &m.files[clickedIndex]
					m.contextMenuCursor = 0
				}
			}

		case tea.MouseButtonWheelUp:
			// If context menu is open, scroll the menu
			if m.contextMenuOpen {
				if m.contextMenuCursor > 0 {
					m.contextMenuCursor--
				}
				return m, tea.ClearScreen
			}

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
					if m.viewMode == viewDualPane {
						if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
							m.loadPreview(currentFile.path)
						}
					}
				}
			}

		case tea.MouseButtonWheelDown:
			// If context menu is open, scroll the menu
			if m.contextMenuOpen {
				menuItems := m.getContextMenuItems()
				if m.contextMenuCursor < len(menuItems)-1 {
					m.contextMenuCursor++
				}
				return m, tea.ClearScreen
			}

			if m.viewMode == viewDualPane && m.focusedPane == rightPane {
				// Scroll preview down when right pane focused (3 lines per tick)
				visibleLines := m.height - 7
				totalLines := m.getWrappedLineCount()
				maxScroll := totalLines - visibleLines
				if maxScroll < 0 {
					maxScroll = 0
				}
				m.preview.scrollPos += 3
				if m.preview.scrollPos > maxScroll {
					m.preview.scrollPos = maxScroll
				}
			} else {
				// Scroll file list
				maxCursor := m.getMaxCursor()
				if m.cursor < maxCursor {
					m.cursor++
					// Update preview in dual-pane mode
					if m.viewMode == viewDualPane {
						if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
							m.loadPreview(currentFile.path)
						}
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.calculateGridLayout() // Recalculate grid columns on resize
		m.calculateLayout()     // Recalculate pane layout on resize

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}
