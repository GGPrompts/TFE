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
)

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
			m.agentSessions = getAgentSessions()
			m.agentFileMap = buildAgentFileMap(changed, m.agentSessions)
			m.changesRestoreDisplay = m.displayMode
			m.displayMode = modeDetail
			m.detailScrollX = 0
			m.showDiffPreview = true
			m.calculateLayout()
			m.setStatusMessage(fmt.Sprintf("Git changes: %d files (d: toggle diff)", len(changed)), false)
		}
	} else {
		m.exitChangesMode()
	}

	m.cursor = 0
	m.loadFiles()
}

// toggleGitRepos toggles the git repositories filter, scanning recursively when enabled.
// Used by: menu (toggle-git-repos, go-git-repos).
func (m *model) toggleGitRepos() {
	if m.showTrashOnly {
		m.showTrashOnly = false
		m.trashRestorePath = ""
	}

	m.showGitReposOnly = !m.showGitReposOnly

	if m.showGitReposOnly {
		m.displayMode = modeDetail
		m.detailScrollX = 0
		m.calculateLayout()

		m.setStatusMessage("🔍 Scanning for git repositories (depth 3, max 50)...", false)
		m.gitReposList = m.scanGitReposRecursive(m.currentPath, m.gitReposScanDepth, 50)
		m.gitReposLastScan = time.Now()
		m.gitReposScanRoot = m.currentPath
		m.setStatusMessage(fmt.Sprintf("Found %d git repositories", len(m.gitReposList)), false)
	}

	m.cursor = 0
	m.loadFiles()
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
