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

// isClickInFileListArea checks if a mouse click is in the file list area (vs preview area)
// Handles both horizontal split and vertical split layouts
func (m model) isClickInFileListArea(mouseX, mouseY int) bool {
	if m.viewMode != viewDualPane {
		return true // Single-pane mode - all clicks are in file list
	}

	headerLines := 4
	footerLines := 4

	// Check if click is within pane area vertically
	if mouseY < headerLines || mouseY >= m.height-footerLines {
		return false // Click in header or footer
	}

	// Check if using VERTICAL split (Detail mode always uses vertical, List/Tree on narrow terminals)
	useVerticalSplit := m.displayMode == modeDetail || m.isNarrowTerminal()

	if useVerticalSplit {
		// VERTICAL split: top pane is file list, bottom pane is preview
		// ACCORDION: Calculate based on current focus (matches render_preview.go)
		maxVisible := m.height - headerLines - footerLines
		var topHeight int
		if m.focusedPane == leftPane {
			topHeight = (maxVisible * 2) / 3  // Top pane focused = 2/3
		} else {
			topHeight = maxVisible - ((maxVisible * 2) / 3)  // Top pane unfocused = 1/3
		}
		paneY := mouseY - headerLines

		return paneY < topHeight // Top pane is file list
	} else {
		// HORIZONTAL split: left pane is file list, right pane is preview
		return mouseX < m.leftWidth // Left pane is file list
	}
}

// handleMouseEvent processes all mouse input
func (m model) handleMouseEvent(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// If fuzzy search is active, don't process any mouse events
	// (go-fzf handles its own input)
	if m.fuzzySearchActive {
		return m, nil
	}

	// Handle mouse wheel scrolling for command history when command prompt is focused
	// Block file tree navigation even if no history exists
	if m.commandFocused {
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			// Scroll up in history (previous command) if history exists
			if len(m.commandHistory) > 0 {
				m.commandInput = m.getPreviousCommand()
				m.commandCursorPos = len(m.commandInput) // Move cursor to end
			}
			return m, nil
		case tea.MouseButtonWheelDown:
			// Scroll down in history (next command) if history exists
			if len(m.commandHistory) > 0 {
				m.commandInput = m.getNextCommand()
				m.commandCursorPos = len(m.commandInput) // Move cursor to end
			}
			return m, nil
		}
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
		// Check if click is in panes (not in header or status bar)
		// Header is 4 lines total (title, toolbar, command, separator)
		headerLines := 4
		footerLines := 4  // blank + 2 status lines + optional message/search

		if msg.Y >= headerLines && msg.Y < m.height-footerLines {
			oldFocus := m.focusedPane

			// Check if using VERTICAL split (Detail mode always uses vertical, List/Tree on narrow terminals)
			useVerticalSplit := m.displayMode == modeDetail || m.isNarrowTerminal()

			if useVerticalSplit {
				// Calculate pane heights for vertical split based on CURRENT focus
				// (matches render_preview.go accordion behavior)
				maxVisible := m.height - headerLines - footerLines
				var topHeight int
				if m.focusedPane == leftPane {
					topHeight = (maxVisible * 2) / 3  // Top pane currently focused = 2/3
				} else {
					topHeight = maxVisible - ((maxVisible * 2) / 3)  // Top pane unfocused = 1/3
				}

				// Calculate Y position relative to pane area start
				paneY := msg.Y - headerLines

				if paneY < topHeight {
					m.focusedPane = leftPane  // Top pane (file list)
				} else {
					m.focusedPane = rightPane // Bottom pane (preview)
				}
			} else {
				// HORIZONTAL split - List/Tree view on wide terminals (accordion layout)
				if msg.X < m.leftWidth {
					m.focusedPane = leftPane
				} else if msg.X > m.leftWidth { // Account for separator
					m.focusedPane = rightPane
				}
			}

			// If focus changed, recalculate layout and refresh preview cache
			if oldFocus != m.focusedPane {
				m.calculateLayout()
				m.populatePreviewCache()
			}
		}
	}

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionRelease {
			// Check for menu bar clicks (Y=0) - only after 5 seconds when menu bar is visible
			if time.Since(m.startupTime) >= 5*time.Second {
				if m.isInMenuBar(msg.X, msg.Y) {
					menuKey := m.getMenuAtPosition(msg.X)
					if menuKey != "" {
						if m.menuOpen && m.activeMenu == menuKey {
							// Clicking same menu closes it
							m.menuOpen = false
							m.activeMenu = ""
							m.selectedMenuItem = -1
						} else {
							// Open menu and select first non-separator item
							m.menuOpen = true
							m.activeMenu = menuKey
							m.selectedMenuItem = m.getFirstSelectableMenuItem(menuKey)
						}
						return m, nil
					}
				}

				// Dropdown menu clicks (if menu is open)
				if m.menuOpen && m.isInDropdown(msg.X, msg.Y) {
					itemIndex := m.getMenuItemAtPosition(msg.Y)
					if itemIndex >= 0 {
						menus := m.getMenus()
						menu := menus[m.activeMenu]
						if itemIndex < len(menu.Items) {
							item := menu.Items[itemIndex]
							// Execute action if not separator or disabled
							if !item.IsSeparator && !item.Disabled {
								return m.executeMenuAction(item.Action)
							}
						}
					}
					return m, nil
				}

				// Click outside menu closes it
				if m.menuOpen {
					m.menuOpen = false
					m.activeMenu = ""
					m.selectedMenuItem = -1
					// Don't return - continue processing click
				}
			}

			// Check for toolbar button clicks (Y=1)
			// Toolbar: [🏠] [⭐/✨] [V] [⬜/⬌] [>_] [🔍] [📝] [🎮] [🗑️]
			// Layout:  0-4   5-9    10-12 13-17 18-22 23-27 28-32 33-37 38-42
			// Note: Most buttons are 5 chars ([ + emoji(2) + ] + space), [V] is 3 chars
			if msg.Y == 1 {
				// Home button [🏠] (X=0-4: [ + emoji(2) + ] + space)
				if msg.X >= 0 && msg.X <= 4 {
					// Navigate to home directory
					homeDir, err := os.UserHomeDir()
					if err == nil {
						m.currentPath = homeDir
						m.cursor = 0

						// If git repos filter is active, rescan from home
						if m.showGitReposOnly {
							m.setStatusMessage("🔍 Scanning git repositories from home...", false)
							m.gitReposList = m.scanGitReposRecursive(m.currentPath, m.gitReposScanDepth, 50)
							m.gitReposLastScan = time.Now()
							m.gitReposScanRoot = m.currentPath
							m.setStatusMessage(fmt.Sprintf("Found %d git repositories", len(m.gitReposList)), false)
						}

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
				// View mode toggle button [V] (X=10-12: [ + V + ] + space)
				if msg.X >= 10 && msg.X <= 12 {
					// Cycle through display modes: List → Detail → Tree → List
					if m.displayMode == modeList {
						m.displayMode = modeDetail
						m.detailScrollX = 0 // Reset scroll when entering detail view
					} else if m.displayMode == modeDetail {
						m.displayMode = modeTree
					} else {
						m.displayMode = modeList
						// Reset tree expansion when leaving tree view
						m.expandedDirs = make(map[string]bool)
					}
					m.calculateLayout() // Recalculate widths for new display mode
					return m, nil
				}
				// Pane toggle button [⬜/⬌] (X=13-17: [ + emoji(2) + ] + space)
				if msg.X >= 13 && msg.X <= 17 {
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
				// Terminal button [>_] (X=18-22: [ + >_(2) + ] + space)
				if msg.X >= 18 && msg.X <= 22 {
					// Toggle command mode focus
					m.commandFocused = !m.commandFocused
					if !m.commandFocused {
						// Clear command input when exiting command mode via click
						m.commandInput = ""
					}
					return m, nil
				}
				// Context-aware search button [🔍] (X=23-27: [ + emoji(2) + ] + space)
				if msg.X >= 23 && msg.X <= 27 {
					// Context-aware search toggle:
					// - When viewing file (full preview or dual-pane with right pane focused): Toggle in-file search (Ctrl+F)
					// - When browsing files (left pane or single-pane): Toggle directory filter search (/)
					if m.viewMode == viewFullPreview || (m.viewMode == viewDualPane && m.focusedPane == rightPane) {
						// Toggle in-file search (Ctrl+F behavior)
						if m.preview.searchActive {
							// Deactivate search (like Esc)
							m.preview.searchActive = false
							m.preview.searchQuery = ""
							m.preview.searchMatches = nil
							m.preview.currentMatch = -1
						} else {
							// Activate search
							m.preview.searchActive = true
							m.preview.searchQuery = ""
							m.preview.searchMatches = nil
							m.preview.currentMatch = -1
						}
					} else {
						// Toggle directory filter search (/ behavior)
						if m.viewMode != viewFullPreview {
							if m.searchMode {
								// Deactivate search (like Esc)
								m.searchMode = false
								m.searchQuery = ""
								m.filteredIndices = nil
								m.cursor = 0
							} else {
								// Activate search
								m.searchMode = true
								m.searchQuery = ""
								m.filteredIndices = m.filterFilesBySearch("")
							}
						}
					}
					return m, nil
				}
				// Prompts filter button [📝] (X=28-32: [ + emoji(2) + ] + space)
				if msg.X >= 28 && msg.X <= 32 {
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
				// Game controller button [🎮] (X=33-37: [ + emoji(2) + ] + space)
				if msg.X >= 33 && msg.X <= 37 {
					// Launch TUIClassics game launcher
					homeDir, err := os.UserHomeDir()
					if err != nil {
						m.setStatusMessage("Error: Could not find home directory", true)
						return m, nil
					}

					classicsPath := filepath.Join(homeDir, "projects", "TUIClassics", "bin", "classics")

					// Check if classics launcher exists
					if _, err := os.Stat(classicsPath); err == nil {
						// Launch the classics menu
						return m, openTUITool(classicsPath, filepath.Dir(classicsPath))
					}

					// If classics doesn't exist, check for individual games in root directory
					tuiClassicsDir := filepath.Join(homeDir, "projects", "TUIClassics")
					if entries, err := os.ReadDir(tuiClassicsDir); err == nil && len(entries) > 0 {
						// Find first executable game
						for _, entry := range entries {
							if !entry.IsDir() {
								gamePath := filepath.Join(tuiClassicsDir, entry.Name())
								if info, err := os.Stat(gamePath); err == nil && info.Mode()&0111 != 0 {
									// Found an executable - launch it
									return m, openTUITool(gamePath, tuiClassicsDir)
								}
							}
						}
					}

					// No games found
					m.setStatusMessage("Games: Install from github.com/GGPrompts/TUIClassics (run: git clone ... && make build)", false)
					return m, nil
				}
				// Trash button [🗑️] or [♻️] (X=38-42: [ + emoji(2) + ] + space)
				if msg.X >= 38 && msg.X <= 42 {
					// Toggle trash view
					m.showTrashOnly = !m.showTrashOnly
					m.showFavoritesOnly = false // Disable favorites filter
					m.showPromptsOnly = false   // Disable prompts filter
					m.cursor = 0

					if m.showTrashOnly {
						// Load trash items and convert to fileItems for display
						// Default to detail view for trash (supports horizontal scrolling on narrow terminals)
						m.displayMode = modeDetail
						m.detailScrollX = 0 // Reset scroll
						m.calculateLayout() // Recalculate widths for detail view
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

			// In dual-pane mode, only process file clicks if within file list area
			if !m.isClickInFileListArea(msg.X, msg.Y) {
				// Click is in preview pane - don't select files
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
			// Dual-pane: calculate based on current accordion state
			headerLines := 4
			footerLines := 4
			totalAvailable := m.height - headerLines - footerLines

			// Check if using VERTICAL split (Detail mode always uses vertical, List/Tree on narrow terminals)
			useVerticalSplit := m.displayMode == modeDetail || m.isNarrowTerminal()

			if useVerticalSplit {
				// VERTICAL split with accordion - top pane height varies by focus
				var topHeight int
				if m.focusedPane == leftPane {
					topHeight = (totalAvailable * 2) / 3  // Top focused = 2/3
				} else {
					topHeight = totalAvailable - ((totalAvailable * 2) / 3)  // Top unfocused = 1/3
				}
				maxVisible = topHeight - 2  // Content height inside borders
			} else {
				// HORIZONTAL split (List/Tree on wide terminals) - height is fixed
				maxVisible = totalAvailable - 2  // Content area inside borders
			}
			contentHeight = maxVisible
		} else {
			// Single-pane: maxVisible = m.height - 9 (total box height INCLUDING borders)
			maxVisible = m.height - 9
			contentHeight = maxVisible - 2 // Content area inside borders
			maxVisible = contentHeight
		}

		if m.displayMode == modeDetail {
			maxVisible -= 1 // Account for detail header line
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
						m.loadFiles()                                   // Reload files with restored filter
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
								selectedPath := clickedFile.path
								m.filledVariables[varName] = selectedPath
								m.setStatusMessage(fmt.Sprintf("✓ Set %s = %s", varName, clickedFile.name), false)
							}
						}

						m.lastClickIndex = -1
						m.lastClickTime = time.Time{}
						return m, tea.ClearScreen
					}

					// Double-click: navigate or full-screen preview
					if clickedFile.isDir {
						m.currentPath = clickedFile.path
						m.cursor = 0

						// Exit favorites mode when navigating into a folder
						if m.showFavoritesOnly {
							m.showFavoritesOnly = false
						}

						// Git repos mode: rescan if "..", exit if navigating to a repo
						if m.showGitReposOnly {
							if clickedFile.name == ".." {
								// Navigating up - rescan from parent
								m.setStatusMessage("🔍 Re-scanning from parent directory...", false)
								m.gitReposList = m.scanGitReposRecursive(m.currentPath, m.gitReposScanDepth, 50)
								m.gitReposLastScan = time.Now()
								m.gitReposScanRoot = m.currentPath
								m.setStatusMessage(fmt.Sprintf("Found %d git repositories", len(m.gitReposList)), false)
							} else {
								// Navigating to a repo - exit filter mode
								m.showGitReposOnly = false
							}
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

			// Don't open menu in preview mode or if in preview pane
			if m.viewMode == viewFullPreview {
				break
			}
			if !m.isClickInFileListArea(msg.X, msg.Y) {
				break
			}

			// Calculate which item was right-clicked
			// Both modes have top borders now, so content starts at line 5
			headerOffset := 5 // +1 for top border (both modes have borders now)
			if m.displayMode == modeDetail {
				headerOffset += 1 // Add 1 for detail view's header only (separator removed)
			}

			// Must match view.go and render_preview.go calculations (same as left-click handler)
		// Calculate visible range to account for scrolling
		// Must match view.go and render_preview.go calculations
		var maxVisible int
		var contentHeight int

		if m.viewMode == viewDualPane {
			// Dual-pane: calculate based on current accordion state
			headerLines := 4
			footerLines := 4
			totalAvailable := m.height - headerLines - footerLines

			// Check if using VERTICAL split (Detail mode always uses vertical, List/Tree on narrow terminals)
			useVerticalSplit := m.displayMode == modeDetail || m.isNarrowTerminal()

			if useVerticalSplit {
				// VERTICAL split with accordion - top pane height varies by focus
				var topHeight int
				if m.focusedPane == leftPane {
					topHeight = (totalAvailable * 2) / 3  // Top focused = 2/3
				} else {
					topHeight = totalAvailable - ((totalAvailable * 2) / 3)  // Top unfocused = 1/3
				}
				maxVisible = topHeight - 2  // Content height inside borders
			} else {
				// HORIZONTAL split (List/Tree on wide terminals) - height is fixed
				maxVisible = totalAvailable - 2  // Content area inside borders
			}
			contentHeight = maxVisible
		} else {
			// Single-pane: maxVisible = m.height - 9 (total box height INCLUDING borders)
			maxVisible = m.height - 9
			contentHeight = maxVisible - 2 // Content area inside borders
			maxVisible = contentHeight
		}

		if m.displayMode == modeDetail {
			maxVisible -= 1 // Account for detail header line
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
		// If dropdown menu is open, scroll the menu
		if m.menuOpen && m.activeMenu != "" {
			if m.selectedMenuItem > 0 {
				// Skip separators when scrolling up
				m.selectedMenuItem--
				menus := m.getMenus()
				menu := menus[m.activeMenu]
				for m.selectedMenuItem > 0 && menu.Items[m.selectedMenuItem].IsSeparator {
					m.selectedMenuItem--
				}
			}
			return m, nil
		}

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
		// If dropdown menu is open, scroll the menu
		if m.menuOpen && m.activeMenu != "" {
			menus := m.getMenus()
			menu := menus[m.activeMenu]
			if m.selectedMenuItem < len(menu.Items)-1 {
				// Skip separators when scrolling down
				m.selectedMenuItem++
				for m.selectedMenuItem < len(menu.Items)-1 && menu.Items[m.selectedMenuItem].IsSeparator {
					m.selectedMenuItem++
				}
			}
			return m, nil
		}

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
