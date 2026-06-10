package main

// Module: actions.go
// Purpose: Shared action helpers for state mutations
// Responsibilities:
// - Centralizing duplicated logic from menu, keyboard, and context menu handlers
// - Ensuring consistent behavior for toggle/navigation actions
// - Single source of truth for state transitions

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// openFileWithBestTool opens a file with the most appropriate viewer/editor
// based on its type: CSV viewer, video/audio player, PDF viewer, database
// viewer, hex viewer for binaries, or a text editor (preferring micro).
// Returns nil (after setting a status message) if no editor is available.
// Used by: keyboard (F4 in full-preview mode, F4 in file list).
func (m *model) openFileWithBestTool(path string) tea.Cmd {
	// Context-aware file opening based on file type
	if isCSVFile(path) {
		return openCSVViewer(path)
	} else if isVideoFile(path) {
		return openVideoPlayer(path)
	} else if isAudioFile(path) {
		return openAudioPlayer(path)
	} else if isPDFFile(path) {
		return openPDFViewer(path)
	} else if isDatabaseFile(path) {
		return openDatabaseViewer(path)
	} else if isBinaryFile(path) && !isImageFile(path) {
		return openHexViewer(path)
	}

	// Text files - use editor
	editor := getAvailableEditor()
	if editor == "" {
		m.setStatusMessage("No editor available (tried micro, nano, vim, vi)", true)
		return nil
	}
	// Prefer micro if available, otherwise use whatever was found
	if editorAvailable("micro") {
		editor = "micro"
	}
	return openEditor(editor, path)
}

// toggleFavorites toggles the favorites-only filter.
// Used by: menu (toggle-favorites, go-favorites), keyboard (F6).
func (m *model) toggleFavorites() {
	if m.showTrashOnly {
		m.showTrashOnly = false
		m.trashRestorePath = ""
	}
	m.showFavoritesOnly = !m.showFavoritesOnly
	m.cursor = 0
	m.loadFiles()
}

// toggleShowHidden toggles visibility of hidden (dot) files and persists the setting.
// Used by: menu (toggle-hidden, settings-show-hidden), keyboard (".", ctrl+h).
func (m *model) toggleShowHidden() {
	m.showHidden = !m.showHidden
	m.loadFiles()
	m.persistConfig()
}

// togglePanelLock toggles the panel lock (disables accordion resizing in dual-pane).
// Returns true if the toggle was applied, false if not in dual-pane mode.
// Used by: menu (toggle-panel-lock), keyboard (ctrl+l).
func (m *model) togglePanelLock() bool {
	if m.viewMode != viewDualPane {
		return false
	}
	m.panelsLocked = !m.panelsLocked
	m.applyPanelLockEffects()
	m.persistConfig()
	return true
}

// applyPanelLockEffects applies side effects after m.panelsLocked has been set.
// Used by: settings-panel-lock (where setConfigBool already toggled the field).
func (m *model) applyPanelLockEffects() {
	if m.panelsLocked {
		useVertical := m.displayMode == modeDetail || m.isNarrowTerminal()
		if useVertical {
			if m.focusedPane == leftPane {
				m.lockedTopRatio = 2.0 / 3.0
			} else {
				m.lockedTopRatio = 1.0 / 3.0
			}
		}
	} else {
		m.lockedTopRatio = 0
		m.calculateLayout()
		m.populatePreviewCache()
	}
}

// navigateHome navigates to the user's home directory, exiting any active filter modes.
// Returns an error message if the home directory cannot be determined.
// Used by: menu (go-home).
func (m *model) navigateHome() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "Error: Could not find home directory"
	}
	if m.showTrashOnly {
		m.showTrashOnly = false
		m.trashRestorePath = ""
	}
	m.currentPath = homeDir
	m.cursor = 0
	m.showFavoritesOnly = false
	m.showPromptsOnly = false
	m.showGitReposOnly = false
	if m.showChangesOnly {
		m.exitChangesMode()
	}
	m.loadFiles()
	return ""
}

// toggleTrash toggles the trash view on/off, saving/restoring the previous path.
// Used by: menu (toggle-trash, go-trash), keyboard (F12).
func (m *model) toggleTrash() {
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
		m.showFavoritesOnly = false
		m.showPromptsOnly = false
		if m.showChangesOnly {
			m.exitChangesMode()
		}
		m.cursor = 0
		m.loadFiles()
	}
}

// togglePrompts toggles the prompts-only filter and auto-expands ~/.prompts.
// Used by: menu (toggle-prompts, go-prompts), keyboard (F11).
func (m *model) togglePrompts() {
	if m.showTrashOnly {
		m.showTrashOnly = false
		m.trashRestorePath = ""
	}
	m.showPromptsOnly = !m.showPromptsOnly
	m.cursor = 0
	m.loadFiles()

	// Auto-expand ~/.prompts when filter is turned on
	if m.showPromptsOnly {
		if homeDir, err := os.UserHomeDir(); err == nil {
			globalPromptsDir := filepath.Join(homeDir, ".prompts")
			if info, err := os.Stat(globalPromptsDir); err == nil && info.IsDir() {
				m.expandedDirs[globalPromptsDir] = true
				m.markTreeItemsDirty()
			} else {
				m.setStatusMessage("💡 Tip: Create ~/.prompts/ folder for global prompts (see helper below)", false)
			}
		}
	}
}

// toggleChangesMode toggles the git changes filter, scanning for changed files when enabled.
// Used by: menu (toggle-changes, git-changes-mode), keyboard (ctrl+g).
func (m *model) toggleChangesMode() {
	if m.showTrashOnly {
		m.showTrashOnly = false
		m.trashRestorePath = ""
	}

	m.showChangesOnly = !m.showChangesOnly

	if m.showChangesOnly {
		changed, err := m.getChangedFiles()
		if err != nil {
			m.setStatusMessage(err.Error(), true)
			m.showChangesOnly = false
		} else {
			m.changedFiles = changed
			m.invalidateDiffPreviewCache() // changedFiles refreshed: cached diff may be stale
			m.agentSessions = getAgentSessions()
			m.agentFileMap = buildAgentFileMap(changed, m.agentSessions)
			m.changesRestoreDisplay = m.displayMode
			m.showDiffPreview = true
			m.setDisplayMode(modeDetail)
			m.setStatusMessage(fmt.Sprintf("Git changes: %d files (d: toggle diff)", len(changed)), false)
		}
	} else {
		m.exitChangesMode()
	}

	m.cursor = 0
	m.loadFiles()
}

// toggleGitRepos toggles the git repositories filter, scanning recursively when enabled.
// Returns a tea.Cmd that callers must propagate: when git-repos mode is newly
// activated it starts the slow background-rescan tick (gitReposTick). The tick
// is self-perpetuating only while the mode stays active, so the caller need not
// do anything when the mode is dismissed.
// Used by: menu (toggle-git-repos, go-git-repos).
func (m *model) toggleGitRepos() tea.Cmd {
	if m.showTrashOnly {
		m.showTrashOnly = false
		m.trashRestorePath = ""
	}

	m.showGitReposOnly = !m.showGitReposOnly

	var cmd tea.Cmd
	if m.showGitReposOnly {
		m.setDisplayMode(modeDetail)

		m.setStatusMessage("🔍 Scanning for git repositories (depth 3, max 50)...", false)
		m.gitReposList = m.scanGitReposRecursive(m.currentPath, m.gitReposScanDepth, 50)
		m.gitReposLastScan = time.Now()
		m.gitReposScanRoot = m.currentPath
		m.setStatusMessage(fmt.Sprintf("Found %d git repositories", len(m.gitReposList)), false)

		// Start the background rescan tick if one isn't already running. The
		// guard prevents stacking multiple ticks across rapid toggles.
		if !m.gitReposTickActive {
			m.gitReposTickActive = true
			cmd = gitReposTick()
		}
	}

	m.cursor = 0
	m.loadFiles()
	return cmd
}

// maxPreviewScroll returns the maximum valid preview scroll position, i.e. the
// largest scrollPos that still keeps content on screen. It is the single source
// of truth for preview scroll bounds (e.g. any scroll-indicator line
// reservation belongs here). Always >= 0.
// Used by: keyboard/mouse preview scrolling helpers below.
func (m *model) maxPreviewScroll() int {
	maxScroll := m.getWrappedLineCount() - m.getPreviewVisibleLines()
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// scrollPreviewBy adjusts the preview scroll position by delta (negative scrolls
// up, positive scrolls down), clamping the result to [0, maxPreviewScroll()].
// Used by: keyboard up/down/pageup/pagedown and mouse wheel handlers.
func (m *model) scrollPreviewBy(delta int) {
	pos := m.preview.scrollPos + delta
	if pos < 0 {
		pos = 0
	}
	if max := m.maxPreviewScroll(); pos > max {
		pos = max
	}
	m.preview.scrollPos = pos
}

// scrollPreviewByPage scrolls the preview by one visible page. direction < 0
// scrolls up, direction >= 0 scrolls down. Result is clamped to valid bounds.
// Used by: keyboard pageup/pagedown handlers.
func (m *model) scrollPreviewByPage(direction int) {
	page := m.getPreviewVisibleLines()
	if direction < 0 {
		page = -page
	}
	m.scrollPreviewBy(page)
}

// scrollPreviewToBottom scrolls the preview to its last valid position.
// Used by: keyboard end/G in full-preview mode.
func (m *model) scrollPreviewToBottom() {
	m.preview.scrollPos = m.maxPreviewScroll()
}

// setDisplayMode switches the file-list display mode (list/detail/tree),
// applying the side effects every entry point used to duplicate (and had
// drifted on): it resets tree expansion when LEAVING tree view, resets the
// detail horizontal scroll offset when ENTERING detail view, always
// recalculates the layout, and refreshes the preview cache in dual-pane mode.
// Used by: keyboard (1/2/3), menu (display-list/display-detail/display-tree),
// mouse (toolbar view-mode cycle button).
func (m *model) setDisplayMode(mode displayMode) {
	leavingTree := m.displayMode == modeTree && mode != modeTree

	m.displayMode = mode

	if leavingTree {
		// Reset tree expansion when leaving tree view
		m.expandedDirs = make(map[string]bool)
		m.markTreeItemsDirty()
	}

	if mode == modeDetail {
		// Reset scroll when entering detail view
		m.detailScrollX = 0
	}

	// Recalculate widths for the new display mode
	m.calculateLayout()

	// Refresh preview cache if in dual-pane mode
	if m.viewMode == viewDualPane {
		m.populatePreviewCache()
	}
}

// toggleDualPane toggles between single-pane and dual-pane view modes.
// Used by: menu (toggle-dual-pane).
func (m *model) toggleDualPane() {
	if m.viewMode == viewDualPane {
		m.viewMode = viewSinglePane
	} else {
		m.viewMode = viewDualPane
	}
	m.calculateLayout()
	m.populatePreviewCache()
}
