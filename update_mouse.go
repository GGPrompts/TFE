package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Module: update_mouse.go
// Purpose: Mouse event handling for TFE
// Responsibilities:
// - Processing all mouse input events
// - Left/right click handling
// - Double-click detection
// - Context menu mouse interaction
// - Mouse wheel scrolling
// - Clickable UI elements

// handleMouseEvent processes all mouse input
func (m model) handleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// If fuzzy search is active, don't process any mouse events
	// (go-fzf handles its own input)
	if m.fuzzySearchActive {
		return m, nil
	}

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
			// Check for toolbar button clicks (Y=1)
			// Toolbar: [🏠] [⭐/✨] [👁️] [⬜/⬌] [>_] [🔍] [📝] [🗑️]
			// Layout:  0-4   5-9    10-14 15-19 20-24 25-29 30-34 35-39
			// Note: Each button is 5 chars: [ + emoji(2) + ] + space
			if msg.Y == 1 {
				// Home button [🏠] (X=0-4: [ + emoji(2) + ] + space)
				if msg.X >= 0 && msg.X <= 4 {
					// Navigate to home directory
					homeDir, err := os.UserHomeDir()
					if err == nil {
						m.currentPath = homeDir
						m.cursor = 0
						m.loadFiles()
					}
					return m, nil
				}
				// Star button [⭐/✨] (X=5-9: [ + emoji(2) + ] + space)
				if msg.X >= 5 && msg.X <= 9 {
					// Toggle favorites filter (like F6)
					m.showFavoritesOnly = !m.showFavoritesOnly
					m.cursor = 0
					if m.showFavoritesOnly {
						m.loadFiles()
					}
					return m, nil
				}
				// View mode toggle button [👁️] (X=10-14: [ + emoji(2) + ] + space)
				if msg.X >= 10 && msg.X <= 14 {
					// Cycle through display modes (List → Detail → Tree)
					if m.displayMode == modeList {
						m.displayMode = modeDetail
					} else if m.displayMode == modeDetail {
						m.displayMode = modeTree
					} else {
						m.displayMode = modeList
					}
					return m, nil
				}
				// Pane toggle button [⬜/⬌] (X=15-19: [ + emoji(2) + ] + space)
				if msg.X >= 15 && msg.X <= 19 {
					// Toggle between single and dual-pane (like Tab or Space)
					if m.viewMode == viewDualPane {
						m.viewMode = viewSinglePane
					} else {
						m.viewMode = viewDualPane
					}
					m.calculateLayout()
					m.populatePreviewCache() // Refresh cache with new layout
					return m, nil
				}
				// Terminal button [>_] (X=20-24: [ + >_(2) + ] + space)
				if msg.X >= 20 && msg.X <= 24 {
					// Toggle command mode focus
					m.commandFocused = !m.commandFocused
					if !m.commandFocused {
						// Clear command input when exiting command mode via click
						m.commandInput = ""
					}
					return m, nil
				}
				// Fuzzy search button [🔍] (X=25-29: [ + emoji(2) + ] + space)
				if msg.X >= 25 && msg.X <= 29 {
					// Launch fuzzy search
					m.fuzzySearchActive = true
					// Clear screen before launching fuzzy search to ensure clean terminal state
					return m, tea.Sequence(
						tea.ClearScreen,
						m.launchFuzzySearch(),
					)
				}
				// Prompts filter button [📝] (X=30-34: [ + emoji(2) + ] + space)
				if msg.X >= 30 && msg.X <= 34 {
					// Toggle prompts filter
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
					return m, nil
				}
				// Trash button [🗑️] or [♻️] (X=35-39: [ + emoji(2) + ] + space)
				if msg.X >= 35 && msg.X <= 39 {
					// Toggle trash view
					m.showTrashOnly = !m.showTrashOnly
					m.showFavoritesOnly = false // Disable favorites filter
					m.showPromptsOnly = false   // Disable prompts filter
					m.cursor = 0

					if m.showTrashOnly {
						// Load trash items and convert to fileItems for display
						// Default to detail view for trash
						m.displayMode = modeDetail
						m.loadFiles()
					} else {
						// Exit trash view, reload normal files
						m.loadFiles()
					}
					return m, nil
				}
			}

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
				return m, nil
			}

			// In dual-pane mode, only process file clicks if within left pane
			if m.viewMode == viewDualPane && msg.X >= m.leftWidth {
				// Click is in right pane or beyond - don't select files
				break
			}

			// Handle column header clicks in detail view (for sorting)
			// Both modes: header at Y=5 (both have top borders now)
			detailHeaderY := 5

			if m.displayMode == modeDetail && msg.Y == detailHeaderY {
				// Adjust X for left border (both modes have borders now)
				adjustedX := msg.X - 2 // Account for left border

				// Calculate which column was clicked based on X position
				// Header format (regular): "%-30s  %-10s  %-12s  %-15s" with 2-space left padding
				// Columns: Name (2-32), Size (34-44), Modified (46-58), Type (60-75)
				// Header format (favorites): "%-25s  %-10s  %-12s  %-25s" with 2-space left padding
				// Columns: Name (2-27), Size (29-39), Modified (41-53), Location (55-80)

				var newSortBy string
				if m.showFavoritesOnly {
					// Favorites mode column ranges
					if adjustedX >= 2 && adjustedX <= 27 {
						newSortBy = "name"
					} else if adjustedX >= 29 && adjustedX <= 39 {
						newSortBy = "size"
					} else if adjustedX >= 41 && adjustedX <= 53 {
						newSortBy = "modified"
					} else if adjustedX >= 55 && adjustedX <= 80 {
						// Location column - not sortable yet, ignore
						break
					}
				} else {
					// Regular mode column ranges
					if adjustedX >= 2 && adjustedX <= 32 {
						newSortBy = "name"
					} else if adjustedX >= 34 && adjustedX <= 44 {
						newSortBy = "size"
					} else if adjustedX >= 46 && adjustedX <= 58 {
						newSortBy = "modified"
					} else if adjustedX >= 60 && adjustedX <= 75 {
						newSortBy = "type"
					}
				}

				// Apply sorting if a valid column was clicked
				if newSortBy != "" {
					if m.sortBy == newSortBy {
						// Same column - toggle sort direction
						m.sortAsc = !m.sortAsc
					} else {
						// Different column - set new sort, default to ascending
						m.sortBy = newSortBy
						m.sortAsc = true
					}

					// Re-sort files and maintain cursor position if possible
					currentFile := m.getCurrentFile()
					m.sortFiles()

					// Try to keep cursor on the same file after sorting
					if currentFile != nil {
						for i, file := range m.files {
							if file.path == currentFile.path {
								m.cursor = i
								break
							}
						}
					}

					return m, nil
				}
			}

			// Calculate which item was clicked (accounting for header lines and scrolling)
			// Both modes: title(0) + toolbar(1) + command(2) + separator(3) = 4 lines
			// Both modes now have bordered boxes, so top border adds 1 more line
			// File content starts at line 5 in both single-pane and dual-pane
			headerOffset := 5 // +1 for top border of the box (both modes have borders now)
			if m.displayMode == modeDetail {
				headerOffset += 1 // Add 1 for detail view's header only (separator removed)
			}

			// Calculate visible range to account for scrolling
			// Must match view.go and render_preview.go calculations
			var maxVisible int
			var contentHeight int

			if m.viewMode == viewDualPane {
				// Dual-pane: maxVisible = m.height - 7 (total pane height INCLUDING borders)
				// See render_preview.go:845
				maxVisible = m.height - 7
				contentHeight = maxVisible - 2 // Content area inside borders
			} else {
				// Single-pane: maxVisible = m.height - 9 (total box height INCLUDING borders)
				// See view.go:154
				maxVisible = m.height - 9
				contentHeight = maxVisible - 2 // Content area inside borders
			}

			maxVisible = contentHeight // Use contentHeight for consistency with rendering
			if m.displayMode == modeDetail {
				maxVisible -= 1 // Account for detail header only
			}

			// Get filtered files for click detection (respects favorites filter)
			// This must match what's actually rendered on screen
			var displayedFiles []fileItem
			if m.displayMode == modeTree {
				// Tree mode uses treeItems, not filtered files
				displayedFiles = nil // Will use m.treeItems instead
			} else {
				displayedFiles = m.getFilteredFiles()
			}

			var clickedIndex int
			var clickedLine int

			// List, Detail, and Tree modes: one item per line
			// In tree mode, use tree items instead of files
			var totalItems int
			if m.displayMode == modeTree {
				totalItems = len(m.treeItems)
			} else {
				totalItems = len(displayedFiles)
			}

			// Calculate visible range based on cursor and totalItems
			start := 0
			end := totalItems
			if totalItems > maxVisible {
				start = m.cursor - maxVisible/2
				if start < 0 {
					start = 0
				}
				end = start + maxVisible
				if end > totalItems {
					end = totalItems
					start = end - maxVisible
					if start < 0 {
						start = 0
					}
				}
			}

			clickedLine = msg.Y - headerOffset
			clickedIndex = start + clickedLine

			// Validate bounds
			if clickedLine < 0 || clickedIndex >= end || clickedIndex >= totalItems {
				clickedIndex = -1
			}

			// Check bounds based on display mode
			var validIndex bool
			if m.displayMode == modeTree {
				validIndex = clickedIndex >= 0 && clickedIndex < len(m.treeItems)
			} else {
				validIndex = clickedIndex >= 0 && clickedIndex < len(displayedFiles)
			}

			if validIndex {
				now := time.Now()

				// Check for double-click: same item clicked within 500ms
				const doubleClickThreshold = 500 * time.Millisecond
				isDoubleClick := clickedIndex == m.lastClickIndex &&
					now.Sub(m.lastClickTime) < doubleClickThreshold

				// Get the file item based on display mode
				var clickedFile fileItem
				if m.displayMode == modeTree {
					clickedFile = m.treeItems[clickedIndex].file
				} else {
					clickedFile = displayedFiles[clickedIndex]
				}

				if isDoubleClick {
					// In file picker mode, double-click on file should select it
					if m.filePickerMode && !clickedFile.isDir {
						// IMPORTANT: Set the value AFTER reloading preview to avoid field recreation overwriting it
						selectedPath := clickedFile.path
						selectedName := clickedFile.name

						// Return to preview mode
						m.filePickerMode = false
						m.showPromptsOnly = m.filePickerRestorePrompts // Restore prompts filter
						m.loadFiles()                                   // Reload files with restored filter
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

						m.lastClickIndex = -1
						m.lastClickTime = time.Time{}
						return m, tea.ClearScreen
					}

					// Double-click: navigate or full-screen preview
					if clickedFile.isDir {
						m.currentPath = clickedFile.path
						m.cursor = 0
						// Exit favorites mode when navigating into a folder (same as Enter key behavior)
						if m.showFavoritesOnly {
							m.showFavoritesOnly = false
						}
						m.loadFiles()
					} else if !m.filePickerMode {
						// Enter full-screen preview (only if NOT in file picker mode)
						m.loadPreview(clickedFile.path)
						m.viewMode = viewFullPreview
						m.calculateLayout() // Update widths for full-screen
						m.populatePreviewCache() // Repopulate cache with correct width
						// Reset click tracking after double-click
						m.lastClickIndex = -1
						m.lastClickTime = time.Time{}
						// Clear screen for clean rendering
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
					if m.viewMode == viewDualPane && !clickedFile.isDir {
						m.loadPreview(clickedFile.path)
						m.populatePreviewCache() // Populate cache with dual-pane width
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
			// Both modes have top borders now, so content starts at line 5
			headerOffset := 5 // +1 for top border (both modes have borders now)
			if m.displayMode == modeDetail {
				headerOffset += 1 // Add 1 for detail view's header only (separator removed)
			}

			// Must match view.go and render_preview.go calculations (same as left-click handler)
			var maxVisible int
			var contentHeight int

			if m.viewMode == viewDualPane {
				// Dual-pane: maxVisible = m.height - 7 (total pane height INCLUDING borders)
				maxVisible = m.height - 7
				contentHeight = maxVisible - 2
			} else {
				// Single-pane: maxVisible = m.height - 9 (total box height INCLUDING borders)
				maxVisible = m.height - 9
				contentHeight = maxVisible - 2
			}

			maxVisible = contentHeight
			if m.displayMode == modeDetail {
				maxVisible -= 1 // Account for detail header only
			}

			// Get filtered files for right-click detection (respects favorites filter)
			var displayedFiles []fileItem
			if m.displayMode == modeTree {
				displayedFiles = nil // Tree mode uses treeItems
			} else {
				displayedFiles = m.getFilteredFiles()
			}

			var clickedIndex int

			// List, Detail, Tree modes: one item per line
			// In tree mode, use tree items instead of files
			var totalItems int
			if m.displayMode == modeTree {
				totalItems = len(m.treeItems)
			} else {
				totalItems = len(displayedFiles)
			}

			// Calculate visible range based on cursor and totalItems
			start := 0
			end := totalItems
			if totalItems > maxVisible {
				start = m.cursor - maxVisible/2
				if start < 0 {
					start = 0
				}
				end = start + maxVisible
				if end > totalItems {
					end = totalItems
					start = end - maxVisible
					if start < 0 {
						start = 0
					}
				}
			}

			clickedLine := msg.Y - headerOffset
			clickedIndex = start + clickedLine

			if clickedLine < 0 || clickedIndex >= end || clickedIndex >= totalItems {
				clickedIndex = -1
			}

			// Open context menu if a valid file was clicked
			var validIndex bool
			if m.displayMode == modeTree {
				validIndex = clickedIndex >= 0 && clickedIndex < len(m.treeItems)
			} else {
				validIndex = clickedIndex >= 0 && clickedIndex < len(displayedFiles)
			}

			if validIndex {
				m.contextMenuOpen = true
				// Ensure menu has enough left margin for border to show
				m.contextMenuX = msg.X
				if m.contextMenuX < 2 {
					m.contextMenuX = 2
				}
				m.contextMenuY = msg.Y

				// Clear command input to remove any unwanted paste from terminal's right-click paste
				// (Many terminals paste clipboard on right-click before sending the click event)
				m.commandInput = ""

				// Get the file item based on display mode
				if m.displayMode == modeTree {
					file := m.treeItems[clickedIndex].file
					m.contextMenuFile = &file
				} else {
					m.contextMenuFile = &displayedFiles[clickedIndex]
				}
				m.contextMenuCursor = 0
			}
		}

	case tea.MouseButtonWheelUp:
		// If context menu is open, scroll the menu
		if m.contextMenuOpen {
			if m.contextMenuCursor > 0 {
				m.contextMenuCursor--
			}
			return m, nil
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
						m.populatePreviewCache() // Populate cache with dual-pane width
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
			return m, nil
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
						m.populatePreviewCache() // Populate cache with dual-pane width
					}
				}
			}
		}
	}

	return m, nil
}
