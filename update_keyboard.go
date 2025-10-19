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
	// Filter out terminal response sequences (color queries, etc.)
	// These are not real keypresses but terminal responses that leak through
	// Examples: "1;rgb:0000/00", "11;rgb:0000/0000/0000", "b:0000/00"
	key := msg.String()
	if strings.Contains(key, "rgb:") ||
	   (strings.Contains(key, ":") && strings.Contains(key, "/")) {
		// Ignore terminal response sequences
		return m, nil
	}

	// If fuzzy search is active, don't process any keyboard events
	// (go-fzf handles its own input)
	if m.fuzzySearchActive {
		return m, nil
	}

	// Handle input field editing FIRST (works in both dual-pane and full preview)
	// This must come before preview mode and dialog handling
	if m.inputFieldsActive && len(m.promptInputFields) > 0 {
		switch msg.String() {
		case "tab":
			// Navigate to next field
			m.focusedInputField++
			if m.focusedInputField >= len(m.promptInputFields) {
				m.focusedInputField = 0 // Wrap around
			}
			return m, nil

		case "shift+tab":
			// Navigate to previous field
			m.focusedInputField--
			if m.focusedInputField < 0 {
				m.focusedInputField = len(m.promptInputFields) - 1 // Wrap around
			}
			return m, nil

		case "backspace":
			// Delete last character from focused field
			if m.focusedInputField >= 0 && m.focusedInputField < len(m.promptInputFields) {
				field := &m.promptInputFields[m.focusedInputField]
				if len(field.value) > 0 {
					field.value = field.value[:len(field.value)-1]
				}
			}
			return m, nil

		case "ctrl+u":
			// Clear entire field
			if m.focusedInputField >= 0 && m.focusedInputField < len(m.promptInputFields) {
				m.promptInputFields[m.focusedInputField].value = ""
			}
			return m, nil

		case "f3":
			// Activate file picker mode - exit to file browser to select a file
			m.filePickerMode = true
			m.filePickerRestorePath = m.preview.filePath // Store preview path to restore later
			m.filePickerRestorePrompts = m.showPromptsOnly // Store prompts filter state
			m.showPromptsOnly = false // Disable prompts filter to show all files
			m.viewMode = viewSinglePane // Exit preview mode
			m.loadFiles() // Reload files without prompts filter
			m.setStatusMessage("📁 File Picker: Navigate and press Enter to select file (Esc to cancel)", false)
			return m, nil

		case "enter":
			// Move to next field on Enter
			m.focusedInputField++
			if m.focusedInputField >= len(m.promptInputFields) {
				m.focusedInputField = 0 // Wrap around
			}
			return m, nil

		default:
			// Handle regular character input
			keyStr := msg.String()

			// Filter out function keys and special keys
			if !isSpecialKey(keyStr) && len(keyStr) > 0 {
				// Detect paste (multiple characters at once)
				isPaste := len(keyStr) > 1

				if m.focusedInputField >= 0 && m.focusedInputField < len(m.promptInputFields) {
					field := &m.promptInputFields[m.focusedInputField]

					// Add the input to the field value
					field.value += keyStr

					// If it's a paste, show status message
					if isPaste {
						charCount := len(keyStr)
						m.setStatusMessage(fmt.Sprintf("✓ Pasted %d characters", charCount), false)
					}
				}
				return m, nil
			}
		}
	}

	// Handle preview mode keys
	if m.viewMode == viewFullPreview {
		switch msg.String() {
		case "f10", "ctrl+c", "esc":
			// Exit preview mode (F10 replaces q)
			m.viewMode = viewSinglePane
			m.calculateLayout()
			m.populatePreviewCache() // Refresh cache with new width
			// Clear any stray command input that might have captured terminal responses
			m.commandInput = ""
			m.commandFocused = false
			// Return nil to force immediate re-render
			return m, nil

		case "f4":
			// Edit file in external editor from preview (F4 replaces e/E)
			if m.preview.loaded && m.preview.filePath != "" {
				editor := getAvailableEditor()
				if editor == "" {
					m.setStatusMessage("No editor available (tried micro, nano, vim, vi)", true)
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
			// Copy rendered prompt (if prompt) or file path (regular file)
			if m.preview.loaded && m.preview.filePath != "" {
				// If this is a prompt, copy the rendered template
				if m.preview.isPrompt && m.preview.promptTemplate != nil {
					// Get variables (use filled fields if active, otherwise context defaults)
					var vars map[string]string
					if m.inputFieldsActive && len(m.promptInputFields) > 0 {
						vars = getFilledVariables(m.promptInputFields, &m)
					} else {
						vars = getContextVariables(&m)
					}

					// Render the template with variables substituted
					rendered := renderPromptTemplate(m.preview.promptTemplate, vars)

					// Copy to clipboard
					if err := copyToClipboard(rendered); err != nil {
						m.setStatusMessage(fmt.Sprintf("Failed to copy prompt: %s", err), true)
					} else {
						m.setStatusMessage("✓ Prompt copied to clipboard", false)
					}
				} else {
					// Regular file: copy path
					if err := copyToClipboard(m.preview.filePath); err != nil {
						m.setStatusMessage(fmt.Sprintf("Failed to copy to clipboard: %s", err), true)
					} else {
						m.setStatusMessage("Path copied to clipboard", false)
					}
				}
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

		case "pagedown", "pgdn", "pgdown":
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
			// Close context menu
			m.contextMenuOpen = false
			return m, nil

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
			return m, nil

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
			return m, nil

		case "enter":
			// Execute selected menu action
			return m.executeContextMenuAction()
		}
		return m, nil
	}

	// Handle search mode input (/ key for directory search)
	if m.searchMode {
		switch msg.String() {
		case "esc":
			// Exit search mode
			m.searchMode = false
			m.searchQuery = ""
			m.filteredIndices = nil
			m.cursor = 0 // Reset cursor
			return m, nil

		case "backspace":
			// Delete last character from search query
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				// Update filtered results
				m.filteredIndices = m.filterFilesBySearch(m.searchQuery)
				// Reset cursor if out of bounds
				if m.cursor >= len(m.filteredIndices) {
					m.cursor = 0
				}
			}
			return m, nil

		case "enter":
			// Accept search and exit search mode (keep filter active)
			m.searchMode = false
			return m, nil

		default:
			// Add printable characters to search query
			text := string(msg.Runes)
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
					m.searchQuery += text
					// Update filtered results
					m.filteredIndices = m.filterFilesBySearch(m.searchQuery)
					// Reset cursor if out of bounds
					if m.cursor >= len(m.filteredIndices) {
						m.cursor = 0
					}
				}
			}
			return m, nil
		}
	}

	// Handle file picker mode (F3 from input fields)
	if m.filePickerMode {
		switch msg.String() {
		case "esc":
			// Cancel file picker and return to preview mode
			m.filePickerMode = false
			m.showPromptsOnly = m.filePickerRestorePrompts // Restore prompts filter
			m.loadFiles() // Reload files with restored filter
			m.viewMode = viewFullPreview
			m.inputFieldsActive = true // Re-enable input fields
			// Reload the original preview
			if m.filePickerRestorePath != "" {
				m.loadPreview(m.filePickerRestorePath)
				m.populatePreviewCache()
			}
			m.setStatusMessage("File picker cancelled", false)
			return m, nil

		case "enter":
			// Get current file (handles tree mode correctly)
			selectedFile := m.getCurrentFile()
			if selectedFile != nil {
				if selectedFile.isDir {
					// It's a directory - navigate into it (consistent across all views)
					m.currentPath = selectedFile.path
					m.cursor = 0
					m.loadFiles()
					return m, nil
				} else {
					// It's a file - select it and populate the input field
					// IMPORTANT: Set the value AFTER reloading preview to avoid field recreation overwriting it
					selectedPath := selectedFile.path
					selectedName := selectedFile.name

					// Return to preview mode
					m.filePickerMode = false
					m.showPromptsOnly = m.filePickerRestorePrompts // Restore prompts filter
					m.loadFiles() // Reload files with restored filter
					m.viewMode = viewFullPreview
					m.inputFieldsActive = true // Re-enable input fields

					// Reload the original preview (this recreates input fields)
					if m.filePickerRestorePath != "" {
						m.loadPreview(m.filePickerRestorePath)
						m.populatePreviewCache()
					}

					// NOW set the value after fields have been recreated
					if m.focusedInputField >= 0 && m.focusedInputField < len(m.promptInputFields) {
						m.promptInputFields[m.focusedInputField].value = selectedPath
						m.setStatusMessage(fmt.Sprintf("✓ Selected: %s", selectedName), false)
					}

					return m, nil
				}
			}
		}
		// For all other keys in file picker mode, fall through to normal navigation
	}

	// Handle command prompt input (focus-based: only active when commandFocused)
	// Special keys that interact with command prompt
	switch msg.String() {
	case "enter":
		// Execute command if command prompt is focused and has input
		if m.commandFocused && m.commandInput != "" {
			cmd := m.commandInput
			m.addToHistory(cmd)
			m.commandInput = ""
			m.commandFocused = false // Exit command mode after executing
			// Check for exit/quit commands
			cmdLower := strings.ToLower(strings.TrimSpace(cmd))
			if cmdLower == "exit" || cmdLower == "quit" {
				return m, tea.Quit
			}
			return m, runCommand(cmd, m.currentPath)
		}
		// If not in command mode or no input, handle Enter for file navigation (below)

	case "backspace":
		// Delete last character from command if focused and has input
		if m.commandFocused && len(m.commandInput) > 0 {
			m.commandInput = m.commandInput[:len(m.commandInput)-1]
			return m, nil
		}
		// If no command input, backspace does nothing

	case "esc":
		// Exit command mode if focused
		if m.commandFocused {
			m.commandInput = ""
			m.commandFocused = false
			return m, nil
		}
		// If there's leftover command input (but not focused), clear it
		if m.commandInput != "" {
			m.commandInput = ""
			return m, nil
		}
		// If no command input, handle Esc for dual-pane exit (below)

	case ":":
		// Enter command mode (vim-style)
		if !m.commandFocused {
			m.commandFocused = true
			m.commandInput = ""
			return m, nil
		}
		// If already in command mode, add the colon to input
	}

	// Handle typing/pasting while command prompt is focused
	// Only capture input when commandFocused is true
	if m.commandFocused {
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
				// Strip ANSI codes to prevent pasted styled text from corrupting command line
				m.commandInput += stripANSI(text)
				m.historyPos = len(m.commandHistory)
				return m, nil
			}
		}
	}

	// Regular file browser keys
	switch msg.String() {
	case "ctrl+p":
		// Ctrl+P: Fuzzy file search
		m.fuzzySearchActive = true
		return m, m.launchFuzzySearch()

	case "/":
		// /: Enter directory search mode (filter files by name)
		m.searchMode = true
		m.searchQuery = ""
		m.filteredIndices = m.filterFilesBySearch("")
		return m, nil

	case "f10", "ctrl+c":
		// F10: Quit (replaces q)
		return m, tea.Quit

	case "esc":
		// Context-aware ESC behavior:
		// 1. Exit dual-pane mode if active
		// 2. Otherwise, go to parent directory (Windows-style back navigation)
		if m.viewMode == viewDualPane {
			m.viewMode = viewSinglePane
			m.calculateLayout()
			m.populatePreviewCache() // Refresh cache with new width
		} else if m.currentPath != "/" {
			// Go up one level
			m.currentPath = filepath.Dir(m.currentPath)
			m.cursor = 0
			m.loadFiles()
		}

	case "up":
		// If in command mode, navigate command history
		if m.commandFocused && len(m.commandHistory) > 0 {
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
						m.populatePreviewCache() // Populate cache with dual-pane width
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
		// If in command mode, navigate command history
		if m.commandFocused && len(m.commandHistory) > 0 {
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
						m.populatePreviewCache() // Populate cache with dual-pane width
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
					m.searchMode = false // Disable search mode in preview
					m.calculateLayout() // Update widths for full-screen
					// Populate cache synchronously for full preview (user expects instant display)
					m.populatePreviewCache()
					return m, nil
				}
			} else if currentFile.isDir {
				// Navigate into directory (consistent across all views)
				// Arrow keys (←/→) handle tree expansion/collapse
				m.currentPath = currentFile.path
				m.cursor = 0
				m.loadFiles()
			} else {
				// Enter full-screen preview (regardless of current mode)
				m.loadPreview(currentFile.path)
				m.viewMode = viewFullPreview
				m.calculateLayout() // Update widths for full-screen
				// Populate cache synchronously for full preview (user expects instant display)
				m.populatePreviewCache()
				return m, nil
			}
		}

	case "tab":
		// Priority 1: If input fields are active, navigate between fields
		if m.inputFieldsActive && len(m.promptInputFields) > 0 {
			m.focusedInputField++
			if m.focusedInputField >= len(m.promptInputFields) {
				m.focusedInputField = 0 // Wrap around
			}
			return m, nil
		}

		// Priority 2: In dual-pane mode: cycle focus between left and right pane
		// Priority 3: In single-pane mode: enter dual-pane mode
		if m.viewMode == viewDualPane {
			// Cycle through: left → right → left
			if m.focusedPane == leftPane {
				m.focusedPane = rightPane
			} else {
				m.focusedPane = leftPane
			}
		} else if m.viewMode == viewSinglePane {
			// Check if current display mode supports dual-pane
			if !m.isDualPaneCompatible() {
				m.setStatusMessage("Dual-pane mode requires List or Tree view (press 1 or 4)", true)
				return m, nil
			}
			// Enter dual-pane mode
			m.viewMode = viewDualPane
			m.focusedPane = leftPane
			m.calculateLayout()
			// Load preview of current file
			if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
				m.loadPreview(currentFile.path)
				m.populatePreviewCache() // Populate cache with dual-pane width
			}
		}

	case " ":
		// Space: toggle dual-pane mode on/off
		if m.viewMode == viewSinglePane {
			// Check if current display mode supports dual-pane
			if !m.isDualPaneCompatible() {
				m.setStatusMessage("Dual-pane mode requires List or Tree view (press 1 or 4)", true)
				return m, nil
			}
			m.viewMode = viewDualPane
			m.focusedPane = leftPane
			m.calculateLayout()
			// Load preview of current file
			if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
				m.loadPreview(currentFile.path)
				m.populatePreviewCache() // Populate cache with dual-pane width
			}
		} else if m.viewMode == viewDualPane {
			m.viewMode = viewSinglePane
			m.calculateLayout()
			m.populatePreviewCache() // Refresh cache with new width
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
				m.calculateLayout() // Update widths for full-screen
				m.populatePreviewCache() // Repopulate cache with correct width
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
			return m, nil
		}

	case "pagedown", "pgdn", "pgdown":
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
			return m, nil
		}

	case "left":
		// In tree mode: collapse folder or go to parent
		// In other modes: go to parent directory
		if m.displayMode == modeTree {
			if currentFile := m.getCurrentFile(); currentFile != nil {
				if currentFile.isDir && currentFile.name != ".." {
					// If directory is expanded, collapse it
					if m.expandedDirs[currentFile.path] {
						m.expandedDirs[currentFile.path] = false
					} else {
						// Already collapsed, go to parent
						if m.currentPath != "/" {
							m.currentPath = filepath.Dir(m.currentPath)
							m.cursor = 0
							m.loadFiles()
						}
					}
				} else {
					// Not a directory or is "..", go to parent
					if m.currentPath != "/" {
						m.currentPath = filepath.Dir(m.currentPath)
						m.cursor = 0
						m.loadFiles()
					}
				}
			}
		} else {
			// Non-tree modes: go to parent directory
			if m.currentPath != "/" {
				m.currentPath = filepath.Dir(m.currentPath)
				m.cursor = 0
				m.loadFiles()
			}
		}

	case "h":
		// 'h' always goes to parent (vim-style)
		if m.currentPath != "/" {
			m.currentPath = filepath.Dir(m.currentPath)
			m.cursor = 0
			m.loadFiles()
		}

	case "right":
		// In tree mode: expand folder or navigate into it
		// In other modes: navigate into selected directory
		if currentFile := m.getCurrentFile(); currentFile != nil && currentFile.isDir && currentFile.name != ".." {
			if m.displayMode == modeTree {
				// If directory is collapsed, expand it
				if !m.expandedDirs[currentFile.path] {
					m.expandedDirs[currentFile.path] = true
				} else {
					// Already expanded, navigate into it
					m.currentPath = currentFile.path
					m.cursor = 0
					m.loadFiles()
				}
			} else {
				// Non-tree modes: navigate into directory
				m.currentPath = currentFile.path
				m.cursor = 0
				m.loadFiles()
			}
		}

	case "l":
		// 'l' always navigates into directory (vim-style)
		if currentFile := m.getCurrentFile(); currentFile != nil && currentFile.isDir {
			m.currentPath = currentFile.path
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
		// Auto-exit dual-pane if switching to incompatible mode
		if m.viewMode == viewDualPane && !m.isDualPaneCompatible() {
			m.viewMode = viewSinglePane
			m.calculateLayout()
			m.populatePreviewCache()
		}

	case "1":
		// Switch to list view
		m.displayMode = modeList

	case "2":
		// Switch to grid view
		m.displayMode = modeGrid
		// Auto-exit dual-pane (grid view needs full width)
		if m.viewMode == viewDualPane {
			m.viewMode = viewSinglePane
			m.calculateLayout()
			m.populatePreviewCache()
		}

	case "3":
		// Switch to detail view
		m.displayMode = modeDetail
		// Auto-exit dual-pane (detail view needs full width)
		if m.viewMode == viewDualPane {
			m.viewMode = viewSinglePane
			m.calculateLayout()
			m.populatePreviewCache()
		}

	case "4":
		// Switch to tree view
		m.displayMode = modeTree

	case "f4":
		// F4: Edit file in external editor (replaces e/E)
		if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
			editor := getAvailableEditor()
			if editor == "" {
				m.setStatusMessage("No editor available (tried micro, nano, vim, vi)", true)
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
		// F5: Copy rendered prompt (in prompts mode) or file path (regular mode)
		if currentFile := m.getCurrentFile(); currentFile != nil {
			// Special handling for prompts mode: copy rendered prompt
			if m.showPromptsOnly && !currentFile.isDir && isPromptFile(*currentFile) {
				if m.preview.isPrompt && m.preview.promptTemplate != nil {
					// Get context variables
					contextVars := getContextVariables(&m)
					// Render the template with variables substituted
					rendered := renderPromptTemplate(m.preview.promptTemplate, contextVars)

					// Copy to clipboard
					if err := copyToClipboard(rendered); err != nil {
						m.setStatusMessage(fmt.Sprintf("Failed to copy prompt: %s", err), true)
					} else {
						m.setStatusMessage("✓ Prompt copied to clipboard", false)
					}
					return m, nil
				}
			}

			// Regular mode: copy file path
			if err := copyToClipboard(currentFile.path); err != nil {
				m.setStatusMessage(fmt.Sprintf("Failed to copy to clipboard: %s", err), true)
			} else {
				m.setStatusMessage("Path copied to clipboard", false)
			}
		}

	// Note: 's' key removed to allow typing 's' in command prompt
	// To toggle favorites, use F2 (context menu) or right-click → "☆ Add Favorite"

	case "f6":
		// F6: Toggle favorites filter (replaces b/B)
		m.showFavoritesOnly = !m.showFavoritesOnly

	case "f11":
		// F11: Toggle prompts filter (show only .yaml, .md, .txt files)
		m.showPromptsOnly = !m.showPromptsOnly

		// Auto-expand ~/.prompts when filter is turned on
		if m.showPromptsOnly {
			if homeDir, err := os.UserHomeDir(); err == nil {
				globalPromptsDir := filepath.Join(homeDir, ".prompts")
				// Check if ~/.prompts exists
				if info, err := os.Stat(globalPromptsDir); err == nil && info.IsDir() {
					// Expand the ~/.prompts directory
					m.expandedDirs[globalPromptsDir] = true
				}
			}
		}

		// If currently viewing a prompt file, create/clear input fields
		if m.preview.isPrompt && m.preview.promptTemplate != nil {
			if m.showPromptsOnly {
				// Entering prompts mode - create input fields
				m.promptInputFields = createInputFields(m.preview.promptTemplate, &m)
				m.inputFieldsActive = len(m.promptInputFields) > 0
				m.focusedInputField = 0
			} else {
				// Exiting prompts mode - clear input fields
				m.promptInputFields = nil
				m.inputFieldsActive = false
				m.focusedInputField = 0
			}
		}

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
			m.searchMode = false // Disable search mode in preview
			m.calculateLayout() // Update widths for full-screen
			m.populatePreviewCache() // Repopulate cache with correct width
			// Clear screen for clean rendering
			return m, tea.ClearScreen
		}

	case "f2":
		// F2: Open context menu at cursor position (keyboard alternative to right-click)
		if currentFile := m.getCurrentFile(); currentFile != nil {
			// Calculate menu position based on cursor
			headerOffset := 5 // Account for borders (title + toolbar + command + separator + border)
			if m.displayMode == modeDetail {
				headerOffset = 6 // Detail view has header at line 5, content starts at 6 (separator removed)
			}

			// Calculate visible range to account for scrolling
			maxVisible := m.height - 7 // Match rendering calculation
			if m.displayMode == modeDetail {
				maxVisible -= 1 // Account for detail header only (separator removed)
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

	// Default case removed - command input is now focus-based (press : to enter command mode)
	// This prevents stray characters (including terminal response sequences) from leaking into command prompt
	}

	return m, nil
}
