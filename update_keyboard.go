package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Module: update_keyboard.go
// Purpose: Keyboard event handling for TFE
// Responsibilities:
// - Processing all keyboard input events
// - Preview mode key handling
// - Dialog input processing
// - Context menu keyboard navigation
// - Command prompt input
// - File browser keyboard shortcuts

// handleKeyEvent processes all keyboard input
func (m model) handleKeyEvent(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle preview mode keys first
	if m.viewMode == viewFullPreview {
		switch msg.String() {
		case "f10", "ctrl+c", "esc":
			// Exit preview mode (F10 replaces q)
			m.viewMode = viewSinglePane
			m.calculateLayout()
			return m, tea.ClearScreen

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

		case "pageup", "pgup":
			m.preview.scrollPos -= m.height - 6
			if m.preview.scrollPos < 0 {
				m.preview.scrollPos = 0
			}

		case "pagedown", "pgdn":
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

	// Handle dialog input if dialog is open
	if m.showDialog {
		switch m.dialog.dialogType {
		case dialogInput:
			// Handle text input dialog
			switch msg.String() {
			case "esc":
				// Cancel dialog
				m.showDialog = false
				m.dialog = dialogModel{}
				return m, tea.ClearScreen

			case "enter":
				// Confirm input
				if m.dialog.title == "Create Directory" {
					// Handle F7 directory creation
					if err := m.createDirectory(m.dialog.input); err != nil {
						m.setStatusMessage(fmt.Sprintf("Error: %s", err), true)
					} else {
						m.setStatusMessage(fmt.Sprintf("Created directory: %s", m.dialog.input), false)
						m.loadFiles()
						// Move cursor to newly created directory
						for i, f := range m.files {
							if f.name == m.dialog.input {
								m.cursor = i
								break
							}
						}
					}
				}
				m.showDialog = false
				m.dialog = dialogModel{}
				return m, tea.ClearScreen

			case "backspace":
				// Delete last character
				if len(m.dialog.input) > 0 {
					m.dialog.input = m.dialog.input[:len(m.dialog.input)-1]
				}
				return m, nil

			default:
				// Add printable characters to input
				// Use msg.Runes to avoid brackets from msg.String() on paste events
				text := string(msg.Runes)
				if len(text) > 0 {
					// Check if all characters are printable
					isPrintable := true
					for _, r := range msg.Runes {
						if r < 32 || r > 126 {
							isPrintable = false
							break
						}
					}
					if isPrintable {
						m.dialog.input += text
					}
				}
				return m, nil
			}

		case dialogConfirm:
			// Handle confirmation dialog
			switch msg.String() {
			case "esc", "n", "N":
				// Cancel dialog
				m.showDialog = false
				m.dialog = dialogModel{}
				return m, tea.ClearScreen

			case "y", "Y":
				// Confirm action
				if m.dialog.title == "Delete file" || m.dialog.title == "Delete directory" {
					// Handle F8 deletion
					if m.contextMenuFile != nil {
						// Delete from context menu
						if err := m.deleteFileOrDir(m.contextMenuFile.path, m.contextMenuFile.isDir); err != nil {
							m.setStatusMessage(fmt.Sprintf("Error: %s", err), true)
						} else {
							itemType := "file"
							if m.contextMenuFile.isDir {
								itemType = "directory"
							}
							m.setStatusMessage(fmt.Sprintf("Deleted %s: %s", itemType, m.contextMenuFile.name), false)
							m.loadFiles()
							// Adjust cursor if needed
							if m.cursor >= len(m.files) {
								m.cursor = len(m.files) - 1
								if m.cursor < 0 {
									m.cursor = 0
								}
							}
						}
						m.contextMenuFile = nil
						m.contextMenuOpen = false
					} else if currentFile := m.getCurrentFile(); currentFile != nil {
						// Delete from F8 key
						if err := m.deleteFileOrDir(currentFile.path, currentFile.isDir); err != nil {
							m.setStatusMessage(fmt.Sprintf("Error: %s", err), true)
						} else {
							itemType := "file"
							if currentFile.isDir {
								itemType = "directory"
							}
							m.setStatusMessage(fmt.Sprintf("Deleted %s: %s", itemType, currentFile.name), false)
							m.loadFiles()
							// Adjust cursor if needed
							if m.cursor >= len(m.files) {
								m.cursor = len(m.files) - 1
								if m.cursor < 0 {
									m.cursor = 0
								}
							}
						}
					}
				}
				m.showDialog = false
				m.dialog = dialogModel{}
				return m, tea.ClearScreen
			}
			return m, nil
		}
	}

	// Handle context menu input if menu is open
	if m.contextMenuOpen {
		switch msg.String() {
		case "esc", "q":
			// Close context menu and clear screen to remove visual artifacts
			m.contextMenuOpen = false
			return m, tea.ClearScreen

		case "up", "k":
			// Navigate up in menu, skipping separators
			menuItems := m.getContextMenuItems()
			for {
				if m.contextMenuCursor > 0 {
					m.contextMenuCursor--
				} else {
					break
				}
				// Stop if we're not on a separator
				if menuItems[m.contextMenuCursor].action != "separator" {
					break
				}
			}
			// Clear screen to prevent ANSI overlay artifacts
			return m, tea.ClearScreen

		case "down", "j":
			// Navigate down in menu, skipping separators
			menuItems := m.getContextMenuItems()
			for {
				if m.contextMenuCursor < len(menuItems)-1 {
					m.contextMenuCursor++
				} else {
					break
				}
				// Stop if we're not on a separator
				if menuItems[m.contextMenuCursor].action != "separator" {
					break
				}
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

	// Check if user is typing/pasting in command prompt
	// If commandInput has text, prioritize adding to it over hotkeys
	// This allows typing 'e', 'v', 'f', etc. in commands

	// Handle typing/pasting while command is active
	if len(m.commandInput) > 0 {
		// Use msg.Runes to get raw text (Bubble Tea handles escape sequences for us)
		// This avoids the brackets that msg.String() adds around paste events
		text := string(msg.Runes)

		// Only process if not a special key
		if len(text) > 0 && !isSpecialKey(msg.String()) {
			// Check if it's printable text
			isPrintable := true
			for _, r := range msg.Runes {
				if r < 32 || r == 127 { // Control characters
					isPrintable = false
					break
				}
			}
			if isPrintable {
				m.commandInput += text
				m.historyPos = len(m.commandHistory)
				return m, nil
			}
		}
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
					// Clear screen for clean rendering
					return m, tea.ClearScreen
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
				// Clear screen for clean rendering
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
		// F3: Open in browser (images/HTML) or full-screen preview
		if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
			// Check if this is an image or HTML file
			if isBrowserFile(currentFile.path) {
				// Open in browser
				return m, openInBrowser(currentFile.path)
			} else {
				// Open in full-screen preview
				m.loadPreview(currentFile.path)
				m.viewMode = viewFullPreview
				// Clear screen for clean rendering
				return m, tea.ClearScreen
			}
		}

	case "pageup", "pgup":
		// Page up in dual-pane mode (only works when right pane focused)
		if m.viewMode == viewDualPane && m.focusedPane == rightPane {
			visibleLines := m.height - 7
			m.preview.scrollPos -= visibleLines
			if m.preview.scrollPos < 0 {
				m.preview.scrollPos = 0
			}
		}

	case "pagedown", "pgdn":
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

	// Note: 's' key removed to allow typing 's' in command prompt
	// To toggle favorites, use F2 (context menu) or right-click → "☆ Add Favorite"

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
			// Clear screen for clean rendering
			return m, tea.ClearScreen
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
		// F7: Create directory
		m.dialog = dialogModel{
			dialogType: dialogInput,
			title:      "Create Directory",
			message:    "Enter directory name:",
			input:      "",
		}
		m.showDialog = true
		return m, tea.ClearScreen

	case "f8":
		// F8: Delete file/folder
		if len(m.files) == 0 || m.cursor >= len(m.files) {
			return m, nil
		}

		currentFile := m.getCurrentFile()
		if currentFile == nil || currentFile.name == ".." {
			return m, nil // Can't delete parent
		}

		// Show confirmation dialog
		fileType := "file"
		if currentFile.isDir {
			fileType = "directory"
		}
		m.dialog = dialogModel{
			dialogType: dialogConfirm,
			title:      "Delete " + fileType,
			message:    fmt.Sprintf("Delete '%s'?\nThis cannot be undone.", currentFile.name),
		}
		m.showDialog = true
		return m, tea.ClearScreen

	default:
		// MC-style: any printable character(s) go to command prompt
		// This handles typing (starting a new command)

		// Use msg.Runes to get raw text (Bubble Tea handles escape sequences for us)
		// This avoids the brackets that msg.String() adds around paste events
		text := string(msg.Runes)

		// Only process if not a special key
		if len(text) > 0 && !isSpecialKey(msg.String()) {
			// Check if it's printable text
			isPrintable := true
			for _, r := range msg.Runes {
				if r < 32 || r == 127 { // Control characters
					isPrintable = false
					break
				}
			}
			if isPrintable {
				m.commandInput += text
				m.historyPos = len(m.commandHistory)
				return m, nil
			}
		}
	}

	return m, nil
}
