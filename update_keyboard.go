package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	// IMPORTANT: Only filter short sequences - don't block pastes that happen to contain ":" and "/"
	key := msg.String()
	if len(msg.Runes) < 20 && // Only apply filter to short inputs (terminal responses are brief)
		(strings.Contains(key, "rgb:") ||
		 (strings.Contains(key, ":") && strings.Contains(key, "/"))) {
		// Ignore terminal response sequences
		return m, nil
	}

	// If fuzzy search is active, don't process any keyboard events
	// (external fzf handles its own input)
	if m.fuzzySearchActive {
		return m, nil
	}


	// Handle menu bar navigation (when menu bar is focused but dropdown not open)
	if m.menuBarFocused && !m.menuOpen {
		switch msg.String() {
		case "left", "shift+tab":
			// Navigate to previous menu in menu bar
			m.highlightedMenu = getPreviousMenu(m.highlightedMenu)
			return m, nil

		case "right", "tab":
			// Navigate to next menu in menu bar
			m.highlightedMenu = getNextMenu(m.highlightedMenu)
			return m, nil

		case "down", "enter":
			// Open the highlighted menu dropdown
			if m.highlightedMenu != "" {
				m.menuOpen = true
				m.activeMenu = m.highlightedMenu
				m.selectedMenuItem = m.getFirstSelectableMenuItem(m.activeMenu)
			}
			return m, nil

		case "esc":
			// Exit menu bar focus mode
			m.menuBarFocused = false
			m.highlightedMenu = ""
			return m, nil
		}
	}

	// Handle menu keyboard navigation (when menu is open)
	if m.menuOpen {
		switch msg.String() {
		case "esc":
			// Close dropdown and return to menu bar focus
			m.menuOpen = false
			m.selectedMenuItem = -1
			m.menuBarFocused = true
			m.highlightedMenu = m.activeMenu
			return m, nil

		case "left", "shift+tab":
			// Close current menu and open previous menu
			m.activeMenu = getPreviousMenu(m.activeMenu)
			m.highlightedMenu = m.activeMenu
			m.selectedMenuItem = m.getFirstSelectableMenuItem(m.activeMenu)
			return m, nil

		case "right", "tab":
			// Close current menu and open next menu
			m.activeMenu = getNextMenu(m.activeMenu)
			m.highlightedMenu = m.activeMenu
			m.selectedMenuItem = m.getFirstSelectableMenuItem(m.activeMenu)
			return m, nil

		case "up":
			// Navigate up in menu items (skip separators)
			menus := m.getMenus()
			menu := menus[m.activeMenu]
			if m.selectedMenuItem <= 0 {
				// Find last non-separator item
				for i := len(menu.Items) - 1; i >= 0; i-- {
					if !menu.Items[i].IsSeparator {
						m.selectedMenuItem = i
						break
					}
				}
			} else {
				// Move up to previous non-separator
				for i := m.selectedMenuItem - 1; i >= 0; i-- {
					if !menu.Items[i].IsSeparator {
						m.selectedMenuItem = i
						break
					}
				}
			}
			return m, nil

		case "down":
			// Navigate down in menu items (skip separators)
			menus := m.getMenus()
			menu := menus[m.activeMenu]
			// Find next non-separator item
			found := false
			for i := m.selectedMenuItem + 1; i < len(menu.Items); i++ {
				if !menu.Items[i].IsSeparator {
					m.selectedMenuItem = i
					found = true
					break
				}
			}
			if !found {
				// Wrap to first non-separator
				for i := 0; i < len(menu.Items); i++ {
					if !menu.Items[i].IsSeparator {
						m.selectedMenuItem = i
						break
					}
				}
			}
			return m, nil

		case "enter":
			// Execute selected menu item
			if m.selectedMenuItem >= 0 {
				menus := m.getMenus()
				menu := menus[m.activeMenu]
				if m.selectedMenuItem < len(menu.Items) {
					item := menu.Items[m.selectedMenuItem]
					if !item.IsSeparator && !item.Disabled {
						return m.executeMenuAction(item.Action)
					}
				}
			}
			return m, nil
		}
	}

	// Handle preview search mode input
	if m.viewMode == viewFullPreview && m.preview.searchActive {
		switch msg.String() {
		case "esc":
			// Exit search mode
			m.preview.searchActive = false
			m.preview.searchQuery = ""
			m.preview.searchMatches = nil
			m.preview.currentMatch = -1
			return m, nil

		case "enter", "n":
			// Find next match
			m.findNextSearchMatch()
			return m, nil

		case "shift+n":
			// Find previous match
			m.findPreviousSearchMatch()
			return m, nil

		case "backspace":
			// Delete last character from search query
			if len(m.preview.searchQuery) > 0 {
				m.preview.searchQuery = m.preview.searchQuery[:len(m.preview.searchQuery)-1]
				m.performPreviewSearch()
			}
			return m, nil

		default:
			// Add printable characters to search query
			keyStr := msg.String()
			if !isSpecialKey(keyStr) && len(keyStr) > 0 {
				m.preview.searchQuery += keyStr
				m.performPreviewSearch()
			}
			return m, nil
		}
	}

	// PRIORITY 1: Handle file picker mode (F3 from edit mode)
	// File picker has higher priority than edit mode because Esc should close picker first
	if m.filePickerMode {
		switch msg.String() {
		case "esc":
			// Cancel file picker and return to preview mode or normal view
			wasCopyMode := m.filePickerCopySource != ""

			m.filePickerMode = false
			m.filePickerCopySource = "" // Reset copy mode

			// Only restore preview mode if we came from edit mode (prompts)
			// If we came from context menu copy, just return to normal view
			if m.filePickerRestorePath != "" {
				m.showPromptsOnly = m.filePickerRestorePrompts // Restore prompts filter
				m.loadFiles()                                  // Reload files with restored filter
				m.viewMode = viewFullPreview
				m.loadPreview(m.filePickerRestorePath)
				m.populatePreviewCache()
				m.setStatusMessage("File picker cancelled", false)
			} else {
				m.loadFiles() // Just reload current directory
				if wasCopyMode {
					m.setStatusMessage("Copy cancelled", false)
				} else {
					m.setStatusMessage("File picker cancelled", false)
				}
			}
			return m, nil

		case "enter":
			// Get current file (handles tree mode correctly)
			selectedFile := m.getCurrentFile()
			if selectedFile != nil {
				// Check if we're in copy mode (context menu copy operation)
				if m.filePickerCopySource != "" {
					// Copy mode: selecting destination
					destDir := selectedFile.path

					// If selected a file, use its parent directory as destination
					if !selectedFile.isDir {
						destDir = filepath.Dir(destDir)
					}

					// Execute copy operation
					sourcePath := m.filePickerCopySource
					sourceName := filepath.Base(sourcePath)

					// Build full destination path
					var destPath string
					sourceInfo, err := os.Stat(sourcePath)
					if err != nil {
						m.setStatusMessage(fmt.Sprintf("Error: %s", err), true)
						m.filePickerMode = false
						m.filePickerCopySource = ""
						return m, nil
					}

					if sourceInfo.IsDir() {
						// Copying a directory: append directory name to destination
						destPath = filepath.Join(destDir, sourceName)
					} else {
						// Copying a file: append filename to destination
						destPath = filepath.Join(destDir, sourceName)
					}

					if err := m.copyFile(sourcePath, destPath); err != nil {
						m.setStatusMessage(fmt.Sprintf("Error copying: %s", err), true)
					} else {
						// Show success message with destination
						m.setStatusMessage(fmt.Sprintf("✓ Copied '%s' to: %s", sourceName, destDir), false)
						// Reload files to show the new copy
						m.loadFiles()
					}

					// Reset copy mode
					m.filePickerMode = false
					m.filePickerCopySource = ""
					return m, nil
				}

				if selectedFile.isDir {
					// It's a directory - navigate into it (consistent across all views)
					m.navigateToPath(selectedFile.path)
					return m, nil
				} else {
					// It's a file - insert path into focused variable
					selectedPath := selectedFile.path

					// Save edit state before reloading preview (loadPreview resets these)
					savedEditMode := m.promptEditMode
					savedFocusedIndex := m.focusedVariableIndex
					savedFilledVars := make(map[string]string)
					for k, v := range m.filledVariables {
						savedFilledVars[k] = v
					}

					// Return to preview mode
					m.filePickerMode = false
					m.showPromptsOnly = m.filePickerRestorePrompts // Restore prompts filter
					m.loadFiles()                                  // Reload files with restored filter
					m.viewMode = viewFullPreview

					// Reload the original preview
					if m.filePickerRestorePath != "" {
						m.loadPreview(m.filePickerRestorePath)
						m.populatePreviewCache()
					}

					// Restore edit state (loadPreview resets it)
					m.promptEditMode = savedEditMode
					m.focusedVariableIndex = savedFocusedIndex
					m.filledVariables = savedFilledVars

					// Set the selected file path in the focused variable
					if m.promptEditMode && m.focusedVariableIndex >= 0 && m.preview.promptTemplate != nil {
						if m.focusedVariableIndex < len(m.preview.promptTemplate.variables) {
							varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]
							m.filledVariables[varName] = selectedPath
							m.setStatusMessage(fmt.Sprintf("✓ Set %s = %s", varName, selectedFile.name), false)

							// Invalidate cache to force header re-render with updated variable colors
							m.preview.cacheValid = false
							m.populatePreviewCache()
						}
					}

					return m, nil
				}
			}
		}
		// For all other keys in file picker mode, fall through to normal navigation
	}

	// PRIORITY 2: Handle prompt edit mode input (works in ALL view modes)
	// This prevents hotkeys like F, M, V, D, E from interfering with text input
	// Must be checked AFTER file picker mode (so Esc closes picker first)
	if m.promptEditMode && m.preview.isPrompt && m.preview.promptTemplate != nil {
		switch msg.String() {
		case "esc":
			// Exit prompt edit mode
			m.promptEditMode = false
			m.setStatusMessage("Exited edit mode", false)
			return m, nil

		case "tab":
			// Navigate to next variable
			if len(m.preview.promptTemplate.variables) > 0 {
				m.focusedVariableIndex++
				if m.focusedVariableIndex >= len(m.preview.promptTemplate.variables) {
					m.focusedVariableIndex = 0 // Wrap around
				}
				// Auto-scroll to show the focused variable
				m.scrollToFocusedVariable()
			}
			return m, nil

		case "shift+tab":
			// Navigate to previous variable
			if len(m.preview.promptTemplate.variables) > 0 {
				m.focusedVariableIndex--
				if m.focusedVariableIndex < 0 {
					m.focusedVariableIndex = len(m.preview.promptTemplate.variables) - 1 // Wrap around
				}
				// Auto-scroll to show the focused variable
				m.scrollToFocusedVariable()
			}
			return m, nil

		case "backspace":
			// Delete last character from focused variable
			if m.focusedVariableIndex >= 0 && m.focusedVariableIndex < len(m.preview.promptTemplate.variables) {
				varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]
				currentValue := m.filledVariables[varName]
				if len(currentValue) > 0 {
					m.filledVariables[varName] = currentValue[:len(currentValue)-1]

					// Invalidate cache to force header re-render with updated variable colors
					m.preview.cacheValid = false
					m.populatePreviewCache()
				}
			}
			return m, nil

		case "ctrl+u":
			// Clear focused variable
			if m.focusedVariableIndex >= 0 && m.focusedVariableIndex < len(m.preview.promptTemplate.variables) {
				varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]
				m.filledVariables[varName] = ""

				// Invalidate cache to force header re-render with updated variable colors
				m.preview.cacheValid = false
				m.populatePreviewCache()
			}
			return m, nil

		case "f3":
			// File picker for focused variable
			if m.focusedVariableIndex >= 0 && m.focusedVariableIndex < len(m.preview.promptTemplate.variables) {
				m.filePickerMode = true
				m.filePickerRestorePath = m.preview.filePath
				m.filePickerRestorePrompts = m.showPromptsOnly
				m.showPromptsOnly = false // Show all files
				m.viewMode = viewSinglePane // Exit preview mode
				m.loadFiles()
				m.setStatusMessage("📁 File Picker: Arrows/double-click to navigate, Enter to select file, Esc to cancel", false)
			}
			return m, nil

		case "f5":
			// Copy rendered prompt
			if m.preview.loaded && m.preview.filePath != "" {
				// Get variables - start with context defaults
				vars := getContextVariables(&m)

				// Override with user-filled values from inline editing
				for varName, value := range m.filledVariables {
					if value != "" {
						vars[varName] = value
					}
				}

				// Render the template with variables substituted
				rendered := renderPromptTemplate(m.preview.promptTemplate, vars)

				// Copy to clipboard
				if err := copyToClipboard(rendered); err != nil {
					m.setStatusMessage(fmt.Sprintf("Failed to copy prompt: %s", err), true)
				} else {
					m.setStatusMessage("✓ Prompt copied to clipboard", false)
				}
			}
			return m, nil

		case "up", "k":
			// Scroll preview up (allow scrolling while editing)
			if m.preview.scrollPos > 0 {
				m.preview.scrollPos--
			}
			return m, nil

		case "down", "j":
			// Scroll preview down (allow scrolling while editing)
			totalLines := m.getWrappedLineCount()
			visibleLines := m.getPreviewVisibleLines()
			maxScroll := totalLines - visibleLines
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.preview.scrollPos < maxScroll {
				m.preview.scrollPos++
			}
			return m, nil

		case "pageup", "pgup":
			// Page up (allow scrolling while editing)
			visibleLines := m.getPreviewVisibleLines()
			m.preview.scrollPos -= visibleLines
			if m.preview.scrollPos < 0 {
				m.preview.scrollPos = 0
			}
			return m, nil

		case "pagedown", "pgdn", "pgdown":
			// Page down (allow scrolling while editing)
			totalLines := m.getWrappedLineCount()
			visibleLines := m.getPreviewVisibleLines()
			maxScroll := totalLines - visibleLines
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.preview.scrollPos += visibleLines
			if m.preview.scrollPos > maxScroll {
				m.preview.scrollPos = maxScroll
			}
			return m, nil

		default:
			// Handle regular character input for focused variable
			if m.focusedVariableIndex >= 0 && m.focusedVariableIndex < len(m.preview.promptTemplate.variables) {
				varName := m.preview.promptTemplate.variables[m.focusedVariableIndex]
				currentValue := m.filledVariables[varName]

				// Append typed character (use msg.Runes for proper Unicode handling)
				text := string(msg.Runes)
				if len(text) > 0 && !isSpecialKey(msg.String()) {
					// First, strip ANSI codes (they contain ESC chars which would fail printable check)
					cleanText := stripANSI(text)

					// Normalize line endings: convert \r\n and \r to \n
					cleanText = strings.ReplaceAll(cleanText, "\r\n", "\n")
					cleanText = strings.ReplaceAll(cleanText, "\r", "\n")

					// Then check if remaining characters are printable
					// Allow: newline (\n=10), tab (\t=9), and all Unicode > 31 except DEL (127)
					// Reject: other control chars (0-8, 11-31, 127)
					isPrintable := true
					for _, r := range []rune(cleanText) {
						if (r < 32 && r != '\n' && r != '\t') || r == 127 {
							isPrintable = false
							break
						}
					}

					if isPrintable {
						m.filledVariables[varName] = currentValue + cleanText

						// Invalidate cache to force header re-render with updated variable colors
						m.preview.cacheValid = false
						m.populatePreviewCache()

						// Show paste feedback for large pastes
						if len(cleanText) > 50 {
							lineCount := strings.Count(cleanText, "\n") + 1
							m.setStatusMessage(fmt.Sprintf("✓ Pasted %d chars (%d lines)", len(cleanText), lineCount), false)
						}
						return m, nil
					}
				}
			}
		}
		// If we got here, the key wasn't handled - return early to prevent hotkey processing
		return m, nil
	}

	// Handle preview mode keys
	if m.viewMode == viewFullPreview {
		// Normal preview mode keyboard handling
		switch msg.String() {
		case "f10", "ctrl+c":
			// Exit preview mode (F10 replaces q)
			m.viewMode = viewSinglePane
			m.calculateLayout()
			m.populatePreviewCache() // Refresh cache with new width
			// Clear any stray command input that might have captured terminal responses
			m.commandInput = ""
			m.commandFocused = false
			// Reset mouse mode when exiting preview
			m.previewMouseEnabled = true
			return m, tea.EnableMouseCellMotion

		case "esc":
			// Exit preview mode (edit mode ESC is handled in universal section above)
			m.viewMode = viewSinglePane
			m.calculateLayout()
			m.populatePreviewCache() // Refresh cache with new width
			m.commandInput = ""
			m.commandFocused = false
			m.previewMouseEnabled = true
			return m, tea.EnableMouseCellMotion

		case "tab":
			// Inline editing: Enter edit mode on first Tab press
			// (Navigation within edit mode is handled by the priority section above)
			if m.preview.isPrompt && m.preview.promptTemplate != nil && m.showPromptsOnly && !m.promptEditMode {
				// First Tab press - enter edit mode
				m.promptEditMode = true
				m.focusedVariableIndex = 0
				// Auto-fill defaults for DATE/TIME
				m.autofillDefaults()
				m.setStatusMessage("Edit mode: Tab/Shift+Tab to navigate, Esc to exit, F5 to copy", false)
				return m, nil
			}

		case "f4":
			// Open file with appropriate viewer/editor from preview (F4)
			if m.preview.loaded && m.preview.filePath != "" {
				path := m.preview.filePath

				// Context-aware file opening based on file type
				if isCSVFile(path) {
					return m, openCSVViewer(path)
				} else if isVideoFile(path) {
					return m, openVideoPlayer(path)
				} else if isAudioFile(path) {
					return m, openAudioPlayer(path)
				} else if isPDFFile(path) {
					return m, openPDFViewer(path)
				} else if isDatabaseFile(path) {
					return m, openDatabaseViewer(path)
				} else if isBinaryFile(path) && !isImageFile(path) {
					return m, openHexViewer(path)
				} else {
					// Text files - use editor
					editor := getAvailableEditor()
					if editor == "" {
						m.setStatusMessage("No editor available (tried micro, nano, vim, vi)", true)
						return m, nil
					}
					if editorAvailable("micro") {
						editor = "micro"
					}
					return m, openEditor(editor, path)
				}
			}

		case "n", "N":
			// Edit file in nano from preview
			if m.preview.loaded && m.preview.filePath != "" && editorAvailable("nano") {
				return m, openEditor("nano", m.preview.filePath)
			}

		case "f5":
			// Copy rendered prompt (if prompt), full content (text files), or file path (binary files)
			if m.preview.loaded && m.preview.filePath != "" {
				// If this is a prompt, copy the rendered template
				if m.preview.isPrompt && m.preview.promptTemplate != nil {
					// Get variables - start with context defaults
					vars := getContextVariables(&m)

					// Override with user-filled values from inline editing
					for varName, value := range m.filledVariables {
						if value != "" {
							vars[varName] = value
						}
					}

					// Render the template with variables substituted
					rendered := renderPromptTemplate(m.preview.promptTemplate, vars)

					// Copy to clipboard
					if err := copyToClipboard(rendered); err != nil {
						m.setStatusMessage(fmt.Sprintf("Failed to copy prompt: %s", err), true)
					} else {
						m.setStatusMessage("✓ Prompt copied to clipboard", false)
					}
				} else if !m.preview.isBinary && len(m.preview.content) > 0 {
					// Text file: copy full content
					fullContent := strings.Join(m.preview.content, "\n")
					if err := copyToClipboard(fullContent); err != nil {
						m.setStatusMessage(fmt.Sprintf("Failed to copy content: %s", err), true)
					} else {
						m.setStatusMessage("✓ File content copied to clipboard", false)
					}
				} else {
					// Binary file or empty: copy path
					if err := copyToClipboard(m.preview.filePath); err != nil {
						m.setStatusMessage(fmt.Sprintf("Failed to copy to clipboard: %s", err), true)
					} else {
						m.setStatusMessage("Path copied to clipboard", false)
					}
				}
			}

		case "m", "M":
			// Toggle mouse mode in preview
			m.previewMouseEnabled = !m.previewMouseEnabled

			if m.previewMouseEnabled {
				m.setStatusMessage("🖱️  Mouse scrolling enabled (press 'm' to enable text selection)", false)
				return m, tea.Batch(tea.EnableMouseCellMotion, statusTimeoutCmd())
			} else {
				m.setStatusMessage("📄 Text selection enabled - Use mouse to select & copy text (Ctrl+Shift+C)", false)
				return m, tea.Batch(tea.DisableMouse, statusTimeoutCmd())
			}

		case "v", "V":
			// View image in terminal viewer (for binary image files)
			if m.preview.loaded && m.preview.isBinary && isImageFile(m.preview.filePath) {
				return m, openImageViewer(m.preview.filePath)
			}

		case "f", "F":
			// Follow symlink - load target's actual content
			if m.preview.loaded && m.preview.filePath != "" {
				// Check if the current preview is a symlink
				linfo, err := os.Lstat(m.preview.filePath)
				if err == nil && linfo.Mode()&os.ModeSymlink != 0 {
					// It's a symlink - check if target is valid
					targetInfo, err := os.Stat(m.preview.filePath) // Stat follows the link
					if err != nil {
						m.setStatusMessage("Cannot follow symlink: target does not exist", true)
						return m, nil
					}

					if targetInfo.IsDir() {
						// Target is a directory - navigate to it
						if m.showTrashOnly {
							m.showTrashOnly = false
							m.trashRestorePath = ""
						}
						m.currentPath = m.preview.filePath
						m.cursor = 0
						m.viewMode = viewSinglePane
						m.loadFiles()
						m.setStatusMessage("Navigated to symlink target directory", false)
						return m, nil
					} else {
						// Target is a file - load its content by reading through the symlink
						// Temporarily load the target by reading the symlink path (os.ReadFile follows symlinks)
						content, err := os.ReadFile(m.preview.filePath)
						if err != nil {
							m.setStatusMessage(fmt.Sprintf("Cannot read target: %s", err), true)
							return m, nil
						}

						// Get target path for display
						target, _ := os.Readlink(m.preview.filePath)
						m.setStatusMessage(fmt.Sprintf("Viewing symlink target: %s", target), false)

						// Store original symlink path
						symlinkPath := m.preview.filePath

						// Load the content as if it's the target file
						// Clear symlink-specific state and reload
						m.loadPreview(symlinkPath) // This will detect it's a symlink first

						// Manually override to show target content instead
						// Detect file type and apply appropriate rendering
						if isBinaryFile(symlinkPath) {
							m.preview.isBinary = true
							m.preview.content = []string{
								"Binary file (symlink target)",
								fmt.Sprintf("Size: %s", formatFileSize(int64(len(content)))),
							}
						} else {
							// Text file - show with syntax highlighting if available
							highlighted, ok := highlightCode(string(content), symlinkPath)
							var lines []string
							if ok {
								lines = strings.Split(highlighted, "\n")
								m.preview.isSyntaxHighlighted = true
							} else {
								lines = strings.Split(string(content), "\n")
							}
							m.preview.content = lines
						}

						m.preview.loaded = true
						m.preview.fileSize = int64(len(content))
						m.populatePreviewCache()
						return m, statusTimeoutCmd()
					}
				} else {
					m.setStatusMessage("Not a symlink (press 'f' only when viewing symlinks)", true)
					return m, statusTimeoutCmd()
				}
			}

		case "ctrl+f":
			// Activate search mode in preview
			if !m.preview.searchActive {
				m.preview.searchActive = true
				m.preview.searchQuery = ""
				m.preview.searchMatches = nil
				m.preview.currentMatch = -1
			}
			return m, nil

		case "f1":
			// F1: Show hotkeys reference from preview mode
			// First check if it exists in current directory
			hotkeysPath := filepath.Join(m.currentPath, "HOTKEYS.md")
			if _, err := os.Stat(hotkeysPath); os.IsNotExist(err) {
				// Try executable directory
				if exePath, err := os.Executable(); err == nil {
					hotkeysPath = filepath.Join(filepath.Dir(exePath), "HOTKEYS.md")
				}
			}
			// Load and show the hotkeys file if it exists
			if _, err := os.Stat(hotkeysPath); err == nil {
				// Store current preview state to restore later
				previousPath := m.preview.filePath
				previousScrollPos := m.preview.scrollPos

				// Load hotkeys file
				m.loadPreview(hotkeysPath)

				// Context-aware help: Jump to relevant section based on current mode
				sectionName := m.getHelpSectionName()
				if sectionLine := findSectionLine(m.preview.content, sectionName); sectionLine >= 0 {
					m.preview.scrollPos = sectionLine
				}

				// Store the previous preview so user can return to it
				// (Note: User would press Esc to go back to file list, then re-enter preview)
				_ = previousPath
				_ = previousScrollPos

				m.calculateLayout()
				m.populatePreviewCache()
				return m, tea.ClearScreen
			}

		case "up", "k":
			// Scroll preview up
			if m.preview.scrollPos > 0 {
				m.preview.scrollPos--
			}

		case "down", "j":
			// Scroll preview down
			totalLines := m.getWrappedLineCount()
			visibleLines := m.getPreviewVisibleLines()
			maxScroll := totalLines - visibleLines
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.preview.scrollPos < maxScroll {
				m.preview.scrollPos++
			}

		case "pageup", "pgup":
			visibleLines := m.getPreviewVisibleLines()
			m.preview.scrollPos -= visibleLines
			if m.preview.scrollPos < 0 {
				m.preview.scrollPos = 0
			}

		case "pagedown", "pgdn", "pgdown":
			totalLines := m.getWrappedLineCount()
			visibleLines := m.getPreviewVisibleLines()
			maxScroll := totalLines - visibleLines
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.preview.scrollPos += visibleLines
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
				} else if m.dialog.title == "Create File" {
					// Handle file creation
					filepath := filepath.Join(m.currentPath, m.dialog.input)
					file, err := os.Create(filepath)
					if err != nil {
						m.setStatusMessage(fmt.Sprintf("Error: %s", err), true)
					} else {
						// Always close the file handle
						defer file.Close()

						m.setStatusMessage(fmt.Sprintf("Created file: %s", m.dialog.input), false)
						m.loadFiles()

						// Check if it's an image file - open in image editor
						if isImageFile(filepath) {
							editor := getAvailableImageEditor()
							if editor == "" {
								m.setStatusMessage("File created, but no image editor available (install textual-paint)", true)
							} else {
								m.showDialog = false
								m.dialog = dialogModel{}
								return m, openImageEditor(filepath)
							}
						} else {
							// Open text file in text editor
							editor := getAvailableEditor()
							if editor == "" {
								m.setStatusMessage("File created, but no editor available", true)
							} else {
								m.showDialog = false
								m.dialog = dialogModel{}
								return m, openEditor(editor, filepath)
							}
						}
					}
				} else if m.dialog.title == "Rename" {
					// Handle rename
					newName := m.dialog.input

					// Validate name
					if newName == "" || newName == m.contextMenuFile.name {
						m.setStatusMessage("Rename cancelled", false)
					} else if strings.Contains(newName, "/") {
						m.setStatusMessage("Error: Filename cannot contain '/'", true)
					} else {
						// Rename the file
						oldPath := m.contextMenuFile.path
						newPath := filepath.Join(filepath.Dir(oldPath), newName)

						if err := os.Rename(oldPath, newPath); err != nil {
							m.setStatusMessage(fmt.Sprintf("Error: %s", err), true)
						} else {
							m.setStatusMessage(fmt.Sprintf("Renamed to: %s", newName), false)
							m.loadFiles()

							// Move cursor to renamed file
							for i, f := range m.files {
								if f.name == newName {
									m.cursor = i
									break
								}
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
					// Check if all characters are printable (allow Unicode, reject control characters)
					isPrintable := true
					for _, r := range msg.Runes {
						if r < 32 || r == 127 { // Control characters only
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
				if m.dialog.title == "Permanently Delete" {
					// Permanently delete item from trash
					if m.contextMenuFile != nil {
						if err := permanentlyDeleteFromTrash(m.contextMenuFile.path); err != nil {
							m.setStatusMessage(fmt.Sprintf("Error: %s", err), true)
						} else {
							m.setStatusMessage("Item permanently deleted", false)
							m.loadFiles() // Refresh trash view
						}
						m.contextMenuFile = nil
						m.contextMenuOpen = false
					}
				} else if m.dialog.title == "Empty Trash" {
					// Empty entire trash
					if err := emptyTrash(); err != nil {
						m.setStatusMessage(fmt.Sprintf("Error emptying trash: %s", err), true)
					} else {
						m.setStatusMessage("Trash emptied successfully", false)
						m.loadFiles() // Refresh trash view
					}
				} else if m.dialog.title == "Move to Trash" {
					// Move item to trash (from context menu)
					if m.contextMenuFile != nil {
						if err := m.deleteFileOrDir(m.contextMenuFile.path, m.contextMenuFile.isDir); err != nil {
							m.setStatusMessage(fmt.Sprintf("Error: %s", err), true)
						} else {
							m.setStatusMessage(fmt.Sprintf("Moved to trash: %s", m.contextMenuFile.name), false)
							m.loadFiles()
						}
						m.contextMenuFile = nil
						m.contextMenuOpen = false
					}
				} else if m.dialog.title == "Delete file" || m.dialog.title == "Delete directory" {
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
				} else if m.dialog.title == "Pull & Rebuild TFE" {
					// Find TFE repository
					tfeRepoPath := findTFERepository()

					if tfeRepoPath == "" {
						m.setStatusMessage("❌ TFE repository not found", true)
					} else {
						// Use tea.ExecProcess like the ! command does - it handles terminal properly
						updateScript := fmt.Sprintf(`
clear
echo "🔄 Updating TFE..."
echo "   Repository: %s"
echo ""

# Run git pull and build
cd '%s' && git pull && echo "" && ./build.sh

# Check if update succeeded
if [ $? -eq 0 ]; then
    echo ""
    echo "✓ TFE updated successfully!"
    echo ""
    read -p "Restart TFE now? [Y/n]: " response
    response=$(echo "$response" | tr '[:upper:]' '[:lower:]')

    if [ -z "$response" ] || [ "$response" = "y" ] || [ "$response" = "yes" ]; then
        echo ""
        echo "🔄 Restarting TFE..."
        exec tfe
    fi
else
    echo ""
    echo "❌ Update failed"
    echo ""
    read -p "Press Enter to exit..."
fi
`, tfeRepoPath, tfeRepoPath)

						cmd := exec.Command("bash", "-c", updateScript)
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr

						// Use tea.ExecProcess to properly handle terminal state
						return m, tea.Sequence(
							tea.ClearScreen,
							tea.ExecProcess(cmd, func(err error) tea.Msg {
								// After script exits, quit TFE
								return tea.Quit()
							}),
						)
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

	// Handle command prompt input (focus-based: only active when commandFocused)
	// NOTE: File picker mode is now handled at top of function (PRIORITY 1)
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
				m.saveCommandHistory() // Save before quitting
				return m, tea.Quit
			}

			// Check for ! prefix - run command and exit TFE
			if strings.HasPrefix(cmd, "!") {
				// Remove the ! prefix
				actualCmd := strings.TrimPrefix(cmd, "!")
				actualCmd = strings.TrimSpace(actualCmd)
				if actualCmd != "" {
					return m, runCommandAndExit(actualCmd, m.currentPath)
				}
			}

			// Handle cd command specially (change TFE's directory instead of subprocess)
			if strings.HasPrefix(cmdLower, "cd ") || cmdLower == "cd" {
				// Extract the path argument
				pathArg := strings.TrimSpace(strings.TrimPrefix(cmd, "cd"))
				pathArg = strings.TrimSpace(strings.TrimPrefix(pathArg, "cd")) // Handle uppercase too

				var newPath string
				if pathArg == "" || pathArg == "~" {
					// cd with no args or cd ~ goes to home directory
					homeDir, err := os.UserHomeDir()
					if err != nil {
						m.setStatusMessage("Error: Could not find home directory", true)
						return m, nil
					}
					newPath = homeDir
				} else if pathArg == "-" {
					// cd - goes to previous directory (if we had one saved)
					m.setStatusMessage("cd -: Previous directory not implemented yet", true)
					return m, nil
				} else if strings.HasPrefix(pathArg, "~/") {
					// Expand ~ in path
					homeDir, err := os.UserHomeDir()
					if err != nil {
						m.setStatusMessage("Error: Could not find home directory", true)
						return m, nil
					}
					newPath = filepath.Join(homeDir, pathArg[2:])
				} else if filepath.IsAbs(pathArg) {
					// Absolute path
					newPath = pathArg
				} else {
					// Relative path
					newPath = filepath.Join(m.currentPath, pathArg)
				}

				// Clean the path and verify it exists
				newPath = filepath.Clean(newPath)
				if info, err := os.Stat(newPath); err == nil && info.IsDir() {
					// Auto-exit trash mode when using cd command (goes to new path, not restore)
					if m.showTrashOnly {
						m.showTrashOnly = false
						m.trashRestorePath = ""
					}
					m.currentPath = newPath
					m.cursor = 0
					m.loadFiles()
					m.setStatusMessage(fmt.Sprintf("Changed to: %s", newPath), false)
				} else {
					m.setStatusMessage(fmt.Sprintf("cd: %s: No such directory", pathArg), true)
				}
				return m, nil
			}

			// Normal command - suspend TFE and return
			return m, runCommand(cmd, m.currentPath)
		}
		// If not in command mode or no input, handle Enter for file navigation (below)

	case "backspace":
		// Delete character at cursor position from command if focused and has input
		if m.commandFocused && len(m.commandInput) > 0 && m.commandCursorPos > 0 {
			// Delete character before cursor
			m.commandInput = m.commandInput[:m.commandCursorPos-1] + m.commandInput[m.commandCursorPos:]
			m.commandCursorPos--
			return m, nil
		}
		// If no command input, backspace does nothing

	case "delete":
		// Delete character at cursor position (forward delete)
		if m.commandFocused && m.commandCursorPos < len(m.commandInput) {
			m.commandInput = m.commandInput[:m.commandCursorPos] + m.commandInput[m.commandCursorPos+1:]
			return m, nil
		}

	case "home", "ctrl+a":
		// Move cursor to beginning of command input
		if m.commandFocused {
			m.commandCursorPos = 0
			return m, nil
		}

	case "end", "ctrl+e":
		// Move cursor to end of command input
		if m.commandFocused {
			m.commandCursorPos = len(m.commandInput)
			return m, nil
		}

	case "ctrl+k":
		// Delete from cursor to end of line
		if m.commandFocused {
			m.commandInput = m.commandInput[:m.commandCursorPos]
			return m, nil
		}

	case "ctrl+u":
		// Delete from cursor to beginning of line
		if m.commandFocused {
			m.commandInput = m.commandInput[m.commandCursorPos:]
			m.commandCursorPos = 0
			return m, nil
		}

	case "ctrl+left", "alt+left", "alt+b":
		// Move cursor one word left
		if m.commandFocused && m.commandCursorPos > 0 {
			// Find previous word boundary (skip current word, then find start of previous word)
			pos := m.commandCursorPos - 1
			// Skip whitespace
			for pos > 0 && m.commandInput[pos] == ' ' {
				pos--
			}
			// Skip word characters
			for pos > 0 && m.commandInput[pos] != ' ' {
				pos--
			}
			// If we stopped at a space, move one forward (unless at start)
			if pos > 0 {
				pos++
			}
			m.commandCursorPos = pos
			return m, nil
		}

	case "ctrl+right", "alt+right", "alt+f":
		// Move cursor one word right
		if m.commandFocused && m.commandCursorPos < len(m.commandInput) {
			// Find next word boundary
			pos := m.commandCursorPos
			// Skip current word characters
			for pos < len(m.commandInput) && m.commandInput[pos] != ' ' {
				pos++
			}
			// Skip whitespace
			for pos < len(m.commandInput) && m.commandInput[pos] == ' ' {
				pos++
			}
			m.commandCursorPos = pos
			return m, nil
		}

	case "esc":
		// Exit command mode if focused
		if m.commandFocused {
			m.commandInput = ""
			m.commandCursorPos = 0
			m.commandFocused = false
			return m, nil
		}
		// If there's leftover command input (but not focused), clear it
		if m.commandInput != "" {
			m.commandInput = ""
			m.commandCursorPos = 0
			return m, nil
		}
		// If no command input, handle Esc for dual-pane exit (below)

	case ":":
		// Enter command mode (vim-style)
		if !m.commandFocused {
			m.commandFocused = true
			m.commandInput = ""
			m.commandCursorPos = 0
			return m, nil
		}
		// If already in command mode, add the colon to input
	}

	// Handle typing/pasting while command prompt is focused
	// Only capture input when commandFocused is true
	if m.commandFocused {
		// Special case: if command input is empty and space is pressed, allow it to fall through
		// to the main switch to toggle dual-pane mode (user might have command mode focused accidentally)
		if msg.String() == " " && m.commandInput == "" {
			// Fall through to main switch - don't capture this space
		} else {
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
					cleanText := stripANSI(text)
					// Insert at cursor position
					m.commandInput = m.commandInput[:m.commandCursorPos] + cleanText + m.commandInput[m.commandCursorPos:]
					m.commandCursorPos += len(cleanText)
					m.historyPos = len(m.commandHistory)
					return m, nil
				}
			}
		}
	}

	// Regular file browser keys
	switch msg.String() {
	case "ctrl+p":
		// Ctrl+P: Fuzzy file search
		m.fuzzySearchActive = true
		// Clear screen before launching fuzzy search to ensure clean terminal state
		return m, tea.Sequence(
			tea.ClearScreen,
			m.launchFuzzySearch(),
		)

	case "/":
		// /: Enter directory search mode (filter files by name)
		// Only activate in file list view, not in full preview mode
		if m.viewMode != viewFullPreview {
			m.searchMode = true
			m.searchQuery = ""
			m.filteredIndices = m.filterFilesBySearch("")
		}
		return m, nil

	case "f10", "ctrl+c":
		// F10: Quit (replaces q)
		m.saveCommandHistory() // Save before quitting
		return m, tea.Quit

	case "ctrl+z":
		// Ctrl+Z: Suspend TFE and drop to shell
		// User can check background processes, view logs, etc.
		// Type 'fg' to resume TFE
		return m, tea.Suspend

	case "ctrl+o":
		// Ctrl+O: Open current directory in system file explorer
		// WSL: Opens in Windows Explorer
		// Linux: Opens in default file manager (via xdg-open)
		// macOS: Opens in Finder
		return m, openInFileExplorer(m.currentPath)

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
			m.navigateToPath(filepath.Dir(m.currentPath))
		}

	case "up":
		// If in command mode, navigate command history (or just block navigation if no history)
		if m.commandFocused {
			if len(m.commandHistory) > 0 {
				m.commandInput = m.getPreviousCommand()
				m.commandCursorPos = len(m.commandInput) // Move cursor to end
			}
			return m, nil
		}
		// Otherwise fall through to file navigation
		fallthrough
	case "k":
		// Block vim navigation when command is focused
		if m.commandFocused {
			return m, nil
		}
		// Clear menu focus when navigating to files
		if m.menuBarFocused || m.highlightedMenu != "" {
			m.menuBarFocused = false
			m.highlightedMenu = ""
		}
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
		// If in command mode, navigate command history (or just block navigation if no history)
		if m.commandFocused {
			if len(m.commandHistory) > 0 {
				m.commandInput = m.getNextCommand()
				m.commandCursorPos = len(m.commandInput) // Move cursor to end
			}
			return m, nil
		}
		// Otherwise fall through to file navigation
		fallthrough
	case "j":
		// Block vim navigation when command is focused
		if m.commandFocused {
			return m, nil
		}
		// Clear menu focus when navigating to files
		if m.menuBarFocused || m.highlightedMenu != "" {
			m.menuBarFocused = false
			m.highlightedMenu = ""
		}
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
				visibleLines := m.getPreviewVisibleLines()
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
			// Check if this is the prompts setup helper
			if m.showPromptsOnly && strings.HasPrefix(currentFile.name, "💡 Setup:") {
				// Create ~/.prompts/ folder
				if err := os.MkdirAll(currentFile.path, 0755); err != nil {
					m.setStatusMessage(fmt.Sprintf("Failed to create folder: %s", err), true)
				} else {
					m.setStatusMessage("✓ Created ~/.prompts/ folder! Add .prompty, .yaml, .md, or .txt files here.", false)
					// Reload files to show the new folder instead of the helper
					m.loadFiles()
				}
				return m, nil
			}

			// If in favorites mode, check if we need to navigate to a different directory
			if m.showFavoritesOnly && currentFile.name != ".." {
				// Check if favorite is in a different location than current path
				fileDir := filepath.Dir(currentFile.path)
				if currentFile.isDir {
					// Navigate to the favorited directory
					if m.showTrashOnly {
						m.showTrashOnly = false
						m.trashRestorePath = ""
					}
					m.currentPath = currentFile.path
					m.cursor = 0
					m.showFavoritesOnly = false // Exit favorites mode
					m.loadFiles()
				} else if fileDir != m.currentPath {
					// Navigate to the file's parent directory and select it
					if m.showTrashOnly {
						m.showTrashOnly = false
						m.trashRestorePath = ""
					}
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
			} else if m.showGitReposOnly {
				if currentFile.name == ".." {
					// Navigating up while git filter is active - rescan from parent
					if m.showTrashOnly {
						m.showTrashOnly = false
						m.trashRestorePath = ""
					}
					m.currentPath = currentFile.path
					m.cursor = 0
					m.setStatusMessage("🔍 Re-scanning from parent directory...", false)
					m.gitReposList = m.scanGitReposRecursive(m.currentPath, m.gitReposScanDepth, 50)
					m.gitReposLastScan = time.Now()
					m.gitReposScanRoot = m.currentPath
					m.setStatusMessage(fmt.Sprintf("Found %d git repositories", len(m.gitReposList)), false)
					m.loadFiles()
				} else if currentFile.isDir {
					// Navigate to the repo and exit filter mode
					if m.showTrashOnly {
						m.showTrashOnly = false
						m.trashRestorePath = ""
					}
					m.currentPath = currentFile.path
					m.cursor = 0
					m.showGitReposOnly = false // Exit git repos mode
					m.loadFiles()
				}
			} else if currentFile.isDir {
				// Navigate into directory (consistent across all views)
				// Arrow keys (←/→) handle tree expansion/collapse
				m.navigateToPath(currentFile.path)
			} else {
				// Enter full-screen preview (regardless of current mode)
				m.loadPreview(currentFile.path)
				m.viewMode = viewFullPreview
				// Clear any active search mode (file list search doesn't apply in preview)
				m.searchMode = false
				m.searchQuery = ""
				m.filteredIndices = nil
				m.calculateLayout() // Update widths for full-screen
				// Populate cache synchronously for full preview (user expects instant display)
				m.populatePreviewCache()
				return m, nil
			}
		}

	case "tab":
		// Priority 0: Prompt edit mode in dual-pane (when right pane focused on a prompt)
		if m.viewMode == viewDualPane && m.focusedPane == rightPane && m.preview.isPrompt && m.preview.promptTemplate != nil && m.showPromptsOnly {
			if !m.promptEditMode {
				// First Tab press - enter edit mode
				m.promptEditMode = true
				m.focusedVariableIndex = 0
				m.autofillDefaults()
				m.setStatusMessage("Edit mode: Tab/Shift+Tab to navigate, Esc to exit, F5 to copy", false)
			} else {
				// Already in edit mode - navigate to next variable
				if len(m.preview.promptTemplate.variables) > 0 {
					m.focusedVariableIndex++
					if m.focusedVariableIndex >= len(m.preview.promptTemplate.variables) {
						m.focusedVariableIndex = 0 // Wrap around
					}
				}
			}
			return m, nil
		}

		// Priority 1: In dual-pane mode: cycle focus between left and right pane
		// Priority 2: In single-pane mode: enter dual-pane mode
		if m.viewMode == viewDualPane {
			// Cycle through: left → right → left
			if m.focusedPane == leftPane {
				m.focusedPane = rightPane
			} else {
				m.focusedPane = leftPane
			}
			// Recalculate accordion layout and refresh cache when switching focus
			m.calculateLayout()
			m.populatePreviewCache()
		} else if m.viewMode == viewSinglePane {
			// Check if current display mode supports dual-pane
			if !m.isDualPaneCompatible() {
				m.setStatusMessage("Dual-pane mode requires List or Tree view (press 1 or 3)", true)
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
				m.setStatusMessage("Dual-pane mode requires List or Tree view (press 1 or 3)", true)
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
				// Clear any active search mode (file list search doesn't apply in preview)
				m.searchMode = false
				m.searchQuery = ""
				m.filteredIndices = nil
				m.calculateLayout() // Update widths for full-screen
				m.populatePreviewCache() // Repopulate cache with correct width
				// Clear screen for clean rendering
				return m, tea.ClearScreen
			}
		}

	case "pageup", "pgup":
		// If command prompt is focused, don't navigate
		if m.commandFocused {
			return m, nil
		}
		if m.viewMode == viewDualPane {
			// In dual-pane mode, check which pane is focused
			if m.focusedPane == leftPane {
				// Page up in file list
				visibleLines := m.getFileListVisibleLines()
				m.cursor -= visibleLines
				if m.cursor < 0 {
					m.cursor = 0
				}
				// Update preview if file selected
				if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
					m.loadPreview(currentFile.path)
					m.populatePreviewCache() // Populate cache with dual-pane width
				}
			} else {
				// Page up in preview pane
				visibleLines := m.getPreviewVisibleLines()
				m.preview.scrollPos -= visibleLines
				if m.preview.scrollPos < 0 {
					m.preview.scrollPos = 0
				}
			}
		} else if m.viewMode != viewFullPreview {
			// Single-pane mode: page up in file list
			visibleLines := m.getFileListVisibleLines()
			m.cursor -= visibleLines
			if m.cursor < 0 {
				m.cursor = 0
			}
		}

	case "pagedown", "pgdn", "pgdown":
		// If command prompt is focused, don't navigate
		if m.commandFocused {
			return m, nil
		}
		if m.viewMode == viewDualPane {
			// In dual-pane mode, check which pane is focused
			if m.focusedPane == leftPane {
				// Page down in file list
				visibleLines := m.getFileListVisibleLines()
				m.cursor += visibleLines
				if m.cursor >= len(m.files) {
					m.cursor = len(m.files) - 1
				}
				if m.cursor < 0 {
					m.cursor = 0
				}
				// Update preview if file selected
				if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
					m.loadPreview(currentFile.path)
					m.populatePreviewCache() // Populate cache with dual-pane width
				}
			} else {
				// Page down in preview pane
				visibleLines := m.getPreviewVisibleLines()
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
		} else if m.viewMode != viewFullPreview {
			// Single-pane mode: page down in file list
			visibleLines := m.getFileListVisibleLines()
			m.cursor += visibleLines
			if m.cursor >= len(m.files) {
				m.cursor = len(m.files) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
		}

	case "left":
		// If command prompt is focused, move cursor left in command input
		if m.commandFocused {
			if m.commandCursorPos > 0 {
				m.commandCursorPos--
			}
			return m, nil
		}
		// PRIORITY: In detail mode on narrow terminals, scroll left (most important use case)
		// On narrow terminals (phones), horizontal scrolling is more useful than tree navigation
		if m.displayMode == modeDetail && m.isNarrowTerminal() {
			if m.detailScrollX > 0 {
				// Scroll by 4 chars (even number to avoid splitting emojis which are 2 cols wide)
				m.detailScrollX -= 4
				if m.detailScrollX < 0 {
					m.detailScrollX = 0
				}
			}
			return m, nil
		}
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
							m.navigateToPath(filepath.Dir(m.currentPath))
						}
					}
				} else {
					// Not a directory or is "..", go to parent
					if m.currentPath != "/" {
						m.navigateToPath(filepath.Dir(m.currentPath))
					}
				}
			}
		} else {
			// Non-tree modes: go to parent directory
			if m.currentPath != "/" {
				m.navigateToPath(filepath.Dir(m.currentPath))
			}
		}

	case "h", "H":
		// H: Open Help menu (replaces vim-style parent navigation - use left/backspace instead)
		if !m.menuBarFocused && !m.menuOpen {
			m.menuBarFocused = true
			m.menuOpen = true
			m.activeMenu = "help"
			m.highlightedMenu = "help"
			m.selectedMenuItem = m.getFirstSelectableMenuItem("help")
			return m, nil
		}

	case "right":
		// If command prompt is focused, move cursor right in command input
		if m.commandFocused {
			if m.commandCursorPos < len(m.commandInput) {
				m.commandCursorPos++
			}
			return m, nil
		}
		// PRIORITY: In detail mode on narrow terminals, scroll right (most important use case)
		// On narrow terminals (phones), horizontal scrolling is more useful than navigation
		if m.displayMode == modeDetail && m.isNarrowTerminal() {
			// Calculate available width (viewport)
			// Must match the logic in render_file_list.go
			availableWidth := m.width
			if m.viewMode == viewDualPane {
				availableWidth = m.leftWidth - 6 // Account for borders and padding
			} else {
				// Single-pane mode: terminal-specific box width calculation
				// WezTerm/Termux: lipgloss Width() includes borders
				// Windows Terminal: lipgloss Width() excludes borders
				if m.terminalType == terminalWezTerm {
					availableWidth = m.width - 8 // Box width - borders - margin
				} else {
					availableWidth = m.width - 6 // Windows Terminal
				}
			}

			// Calculate render width (content width) - same logic as render_file_list.go
			renderWidth := availableWidth
			if m.isNarrowTerminal() && availableWidth < 120 {
				renderWidth = 120 // Fixed width for detail view on narrow terminals
			} else if availableWidth < 60 {
				renderWidth = 60 // Minimum width for wider terminals
			}

			// Calculate maximum scroll offset
			maxScroll := renderWidth - availableWidth
			if maxScroll < 0 {
				maxScroll = 0
			}

			// Scroll by 4 chars (even number to avoid splitting emojis which are 2 cols wide)
			m.detailScrollX += 4

			// Clamp to maximum scroll offset
			if m.detailScrollX > maxScroll {
				m.detailScrollX = maxScroll
			}

			return m, nil
		}
		// In tree mode: expand folder or navigate into it
		// In other modes: navigate into selected directory
		if currentFile := m.getCurrentFile(); currentFile != nil && currentFile.isDir && currentFile.name != ".." {
			if m.displayMode == modeTree {
				// If directory is collapsed, expand it
				if !m.expandedDirs[currentFile.path] {
					m.expandedDirs[currentFile.path] = true
				} else {
					// Already expanded, navigate into it
					m.navigateToPath(currentFile.path)
				}
			} else {
				// Non-tree modes: navigate into directory
				m.navigateToPath(currentFile.path)
			}
		}

	case "l":
		// 'l' always navigates into directory (vim-style)
		if currentFile := m.getCurrentFile(); currentFile != nil && currentFile.isDir {
			m.navigateToPath(currentFile.path)
		}

	case ".", "ctrl+h":
		// Toggle hidden files
		m.showHidden = !m.showHidden
		m.loadFiles()

	case "alt", "f9":
		// Alt or F9: Enter menu bar navigation mode
		if !m.menuBarFocused && !m.menuOpen {
			m.menuBarFocused = true
			m.highlightedMenu = "file" // Start with first menu
		}

	case "f", "F":
		// F: Open File menu
		if !m.menuBarFocused && !m.menuOpen {
			m.menuBarFocused = true
			m.menuOpen = true
			m.activeMenu = "file"
			m.highlightedMenu = "file"
			m.selectedMenuItem = m.getFirstSelectableMenuItem("file")
			return m, nil
		}

	case "e", "E":
		// E: Open Edit menu
		if !m.menuBarFocused && !m.menuOpen {
			m.menuBarFocused = true
			m.menuOpen = true
			m.activeMenu = "edit"
			m.highlightedMenu = "edit"
			m.selectedMenuItem = m.getFirstSelectableMenuItem("edit")
			return m, nil
		}

	case "v", "V":
		// V: Open View menu
		if !m.menuBarFocused && !m.menuOpen {
			m.menuBarFocused = true
			m.menuOpen = true
			m.activeMenu = "view"
			m.highlightedMenu = "view"
			m.selectedMenuItem = m.getFirstSelectableMenuItem("view")
			return m, nil
		}

	case "t", "T":
		// T: Open Tools menu
		if !m.menuBarFocused && !m.menuOpen {
			m.menuBarFocused = true
			m.menuOpen = true
			m.activeMenu = "tools"
			m.highlightedMenu = "tools"
			m.selectedMenuItem = m.getFirstSelectableMenuItem("tools")
			return m, nil
		}

	case "1":
		// Switch to list view
		m.displayMode = modeList
		// Reset tree expansion when leaving tree view
		m.expandedDirs = make(map[string]bool)

	case "2":
		// Switch to detail view
		m.displayMode = modeDetail
		// Reset tree expansion when leaving tree view
		m.expandedDirs = make(map[string]bool)
		// Recalculate layout to ensure correct width for detail view columns
		m.calculateLayout()
		// Refresh preview cache if in dual-pane mode
		if m.viewMode == viewDualPane {
			m.populatePreviewCache()
		}

	case "3":
		// Switch to tree view
		m.displayMode = modeTree

	case "ctrl+w":
		// Ctrl+W: Collapse all expanded folders in tree view
		if m.displayMode == modeTree {
			// Clear all expanded directories to reset tree view
			m.expandedDirs = make(map[string]bool)
			m.setStatusMessage("All folders collapsed", false)
		} else {
			// Not in tree view - show helpful message
			m.setStatusMessage("Collapse all only works in tree view (press 3)", false)
		}

	case "f4":
		// F4: Open file with appropriate viewer/editor
		if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
			path := currentFile.path

			// Context-aware file opening based on file type
			if isCSVFile(path) {
				return m, openCSVViewer(path)
			} else if isVideoFile(path) {
				return m, openVideoPlayer(path)
			} else if isAudioFile(path) {
				return m, openAudioPlayer(path)
			} else if isPDFFile(path) {
				return m, openPDFViewer(path)
			} else if isDatabaseFile(path) {
				return m, openDatabaseViewer(path)
			} else if isBinaryFile(path) && !isImageFile(path) {
				return m, openHexViewer(path)
			} else {
				// Text files - use editor
				editor := getAvailableEditor()
				if editor == "" {
					m.setStatusMessage("No editor available (tried micro, nano, vim, vi)", true)
					return m, nil
				}
				// Prefer micro if available, otherwise use whatever was found
				if editorAvailable("micro") {
					editor = "micro"
				}
				return m, openEditor(editor, path)
			}
		}

	case "n", "N":
		// Edit file in nano specifically
		if currentFile := m.getCurrentFile(); currentFile != nil && !currentFile.isDir {
			if editorAvailable("nano") {
				return m, openEditor("nano", currentFile.path)
			}
		}

	case "f5":
		// F5: Copy rendered prompt (prompts), full content (text files), or file path (binary/not previewed)
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

			// If preview is loaded and it's a text file: copy full content
			if m.preview.loaded && !m.preview.isBinary && len(m.preview.content) > 0 {
				fullContent := strings.Join(m.preview.content, "\n")
				if err := copyToClipboard(fullContent); err != nil {
					m.setStatusMessage(fmt.Sprintf("Failed to copy content: %s", err), true)
				} else {
					m.setStatusMessage("✓ File content copied to clipboard", false)
				}
			} else {
				// Binary file or not previewed: copy file path
				if err := copyToClipboard(currentFile.path); err != nil {
					m.setStatusMessage(fmt.Sprintf("Failed to copy to clipboard: %s", err), true)
				} else {
					m.setStatusMessage("Path copied to clipboard", false)
				}
			}
		}

	// Note: 's' key removed to allow typing 's' in command prompt
	// To toggle favorites, use F2 (context menu) or right-click → "☆ Add Favorite"

	case "f6":
		// F6: Toggle favorites filter (replaces b/B)
		// Auto-exit trash mode when toggling favorites
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}
		m.showFavoritesOnly = !m.showFavoritesOnly

	case "f11":
		// F11: Toggle prompts filter (show only .yaml, .md, .txt files)
		// Auto-exit trash mode when toggling prompts filter
		if m.showTrashOnly {
			m.showTrashOnly = false
			m.trashRestorePath = ""
		}
		m.showPromptsOnly = !m.showPromptsOnly

		// Auto-expand ~/.prompts when filter is turned on
		if m.showPromptsOnly {
			if homeDir, err := os.UserHomeDir(); err == nil {
				globalPromptsDir := filepath.Join(homeDir, ".prompts")
				// Check if ~/.prompts exists
				if info, err := os.Stat(globalPromptsDir); err == nil && info.IsDir() {
					// Expand the ~/.prompts directory
					m.expandedDirs[globalPromptsDir] = true
				} else {
					// ~/.prompts doesn't exist - show helpful message
					m.setStatusMessage("💡 Tip: Create ~/.prompts/ folder for global prompts (see helper below)", false)
				}
			}
		}


	case "f12":
		// F12: Navigate to trash view (or exit if already in trash)
		if m.showTrashOnly {
			// Already in trash - exit and restore previous path
			m.showTrashOnly = false
			if m.trashRestorePath != "" {
				m.currentPath = m.trashRestorePath
				m.trashRestorePath = ""
			}
			m.cursor = 0
			m.loadFiles()
		} else {
			// Enter trash view - save current path
			m.trashRestorePath = m.currentPath
			m.showTrashOnly = true
			m.showFavoritesOnly = false // Disable favorites filter
			m.showPromptsOnly = false   // Disable prompts filter
			m.cursor = 0
			// Default to detail view for trash
			m.displayMode = modeDetail
			m.calculateLayout() // Recalculate widths for detail view
			m.loadFiles()
		}

	case "f1":
		// F1: Show hotkeys reference (replaces ?)
		// First check if it exists in current directory
		hotkeysPath := filepath.Join(m.currentPath, "HOTKEYS.md")
		if _, err := os.Stat(hotkeysPath); os.IsNotExist(err) {
			// Try executable directory
			if exePath, err := os.Executable(); err == nil {
				hotkeysPath = filepath.Join(filepath.Dir(exePath), "HOTKEYS.md")
			}
		}
		// Load and show the hotkeys file if it exists
		if _, err := os.Stat(hotkeysPath); err == nil {
			m.loadPreview(hotkeysPath)

			// Context-aware help: Jump to relevant section based on current mode
			sectionName := m.getHelpSectionName()
			if sectionLine := findSectionLine(m.preview.content, sectionName); sectionLine >= 0 {
				m.preview.scrollPos = sectionLine
			}

			// Stay in dual-pane mode if already there, otherwise go full-screen
			if m.viewMode != viewDualPane {
				m.viewMode = viewFullPreview
			}
			m.searchMode = false // Disable search mode in preview
			m.calculateLayout() // Update widths for appropriate view mode
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
