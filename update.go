package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"
)

// Module: update.go
// Purpose: Main update dispatcher and initialization
// Responsibilities:
// - Bubbletea initialization (Init)
// - Main message dispatcher (Update)
// - Window resize handling
// - Editor/command finished message handling
// - Spinner tick handling
// - Helper function: isSpecialKey() for detecting non-printable keys

func (m model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.spinner.Tick,
		tickCmd(),          // Start landing page animation
		checkForUpdates(),  // Check for new releases on GitHub
	}

	// Start file watcher for the initial directory
	if watchCmd := m.startWatcher(m.currentPath); watchCmd != nil {
		cmds = append(cmds, watchCmd)
	}

	// Start agent session polling if auto-watch is enabled (TFE_AUTO_CHANGES=1)
	if m.agentAutoWatch {
		// Seed initial agent session state so we don't trigger on startup
		for _, s := range getAgentSessions() {
			if s.ParentSessionID == "" { // Only track top-level sessions
				m.lastKnownAgentSessions[s.SessionID] = s.Status
			}
		}
		cmds = append(cmds, agentCheckTick())
	}

	return tea.Batch(cmds...)
}

// tickMsg for landing page animation
type tickMsg struct{}

// tickCmd creates a command that sends tick messages for animation
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*50, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// footerTick sends periodic messages to animate footer scrolling
func footerTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return footerTickMsg{}
	})
}

// agentCheckTick sends a periodic tick to poll agent session state (every 5s)
func agentCheckTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return agentCheckTickMsg{}
	})
}

// checkForUpdates queries GitHub API for the latest release
// Only checks once per day to respect rate limits
func checkForUpdates() tea.Cmd {
	return func() tea.Msg {
		// Get config directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil // Silent fail
		}

		configDir := filepath.Join(homeDir, ".config", "tfe")
		cacheFile := filepath.Join(configDir, "update_check.json")

		// Check cache to avoid checking too often
		var cache struct {
			LastCheck time.Time `json:"last_check"`
		}

		// Only check once per day
		if data, err := os.ReadFile(cacheFile); err == nil {
			json.Unmarshal(data, &cache)
			if time.Since(cache.LastCheck) < 24*time.Hour {
				return nil // Checked recently
			}
		}

		// Fetch latest release from GitHub API
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("https://api.github.com/repos/GGPrompts/TFE/releases/latest")
		if err != nil {
			return nil // Silent fail on network issues
		}
		defer resp.Body.Close()

		var release struct {
			TagName string `json:"tag_name"`
			Body    string `json:"body"`
			HTMLURL string `json:"html_url"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil
		}

		// Update cache
		cache.LastCheck = time.Now()
		os.MkdirAll(configDir, 0755)
		if data, _ := json.Marshal(cache); data != nil {
			os.WriteFile(cacheFile, data, 0644)
		}

		// Compare versions (simple string comparison)
		latest := strings.TrimPrefix(release.TagName, "v")
		current := strings.TrimPrefix(Version, "v")

		// Simple version comparison - for semantic versioning, consider using a library
		if latest > current {
			return updateAvailableMsg{
				version:   release.TagName,
				changelog: release.Body,
				url:       release.HTMLURL,
			}
		}

		return nil
	}
}

// stripANSI removes ANSI escape codes from a string
// This prevents pasted styled text from corrupting the command line
func stripANSI(s string) string {
	// Match all common ANSI escape sequences:
	// - CSI sequences: ESC [ ... (letter or @)
	// - OSC sequences: ESC ] ... (BEL or ST)
	// - Other escape sequences: ESC followed by various characters
	ansiRegex := regexp.MustCompile(`\x1b(\[[0-9;?]*[a-zA-Z@]|\][^\x07\x1b]*(\x07|\x1b\\)|[>=<>()#])`)
	cleaned := ansiRegex.ReplaceAllString(s, "")

	// Also strip terminal response sequences that may appear without ESC prefix
	// These can leak in when terminal responds to queries (e.g., color capability checks)
	// Patterns: ";rgb:xxxx/xxxx", "1;rgb:xxxx/xxxx/xxxx", "0;rgb:...", numeric response codes
	// Match anywhere in the string, not just exact matches
	responseRegex := regexp.MustCompile(`;?rgb:[0-9a-fA-F/]+|\d+;rgb:[0-9a-fA-F/]+|\d+;\d+(?:;\d+)*`)
	cleaned = responseRegex.ReplaceAllString(cleaned, "")

	return cleaned
}

// isSpecialKey checks if a key string represents a special (non-printable) key
// that should not be added to command input
func isSpecialKey(key string) bool {
	specialKeys := []string{
		"up", "down", "left", "right",
		"home", "end", "pageup", "pagedown", "pgup", "pgdn",
		"delete", "insert",
		"backspace", "enter", "return", "tab", "esc", "escape",
		"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12",
		"ctrl+c", "ctrl+h", "ctrl+d", "ctrl+z",
		"alt+", "ctrl+", // Prefixes for modifier combinations
		"shift+",
	}

	// Check exact matches
	for _, special := range specialKeys {
		if key == special {
			return true
		}
		// Check if key starts with modifier prefix (like "ctrl+a", "alt+x")
		if len(special) > 0 && special[len(special)-1] == '+' && len(key) > len(special) {
			if key[:len(special)] == special {
				return true
			}
		}
	}

	return false
}

// Update is the main message dispatcher
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Update tree items cache if in tree mode (before processing events)
	if m.displayMode == modeTree {
		m.updateTreeItems()
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Dispatch to keyboard event handler
		return m.handleKeyEvent(msg)

	case tea.MouseMsg:
		// Dispatch to mouse event handler
		return m.handleMouseEvent(msg)

	case tea.WindowSizeMsg:
		// Handle window resize
		m.height = msg.Height
		m.width = msg.Width

		m.calculateLayout()      // Recalculate pane layout on resize
		m.populatePreviewCache() // Repopulate cache with new width

		// Reset horizontal scroll on window resize
		m.detailScrollX = 0

		// Clear screen to prevent ghost content from previous render
		// (old content can persist if new view is shorter than old view)
		return m, tea.ClearScreen

	case tickMsg:
		// Background refresh for git repos (every 60 seconds)
		if m.showGitReposOnly && !m.gitReposLastScan.IsZero() {
			elapsed := time.Since(m.gitReposLastScan)
			if elapsed >= 60*time.Second {
				// Re-scan from the same root directory
				m.gitReposList = m.scanGitReposRecursive(m.gitReposScanRoot, m.gitReposScanDepth, 50)
				m.gitReposLastScan = time.Now()
			}
		}

		return m, tickCmd() // Continue animation

	case footerTickMsg:
		// Animate footer scrolling if active
		if m.footerScrolling {
			m.footerOffset++
			return m, footerTick() // Continue scrolling
		}
		// If scrolling was stopped, don't schedule next tick

	case spinner.TickMsg:
		// Update spinner animation
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case statusTimeoutMsg:
		// Status message timeout - force full screen redraw
		// Clear screen to ensure proper redraw of footer
		return m, tea.ClearScreen

	case editorFinishedMsg:
		// Editor has closed, we're back in TFE
		// Refresh file list in case file was modified
		m.loadFiles()
		// Force a refresh and restore terminal state (alt screen + mouse support)
		// Re-entering alt screen is crucial for image viewers (viu, timg, chafa)
		// to prevent "Press any key to continue..." text from bleeding through
		return m, tea.Batch(
			tea.EnterAltScreen,       // Re-enter alternate screen (prevents text bleed)
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)

	case commandFinishedMsg:
		// Command has finished, we're back in TFE
		// Refresh file list in case command modified files
		m.loadFiles()
		// Force a refresh and restore terminal state (alt screen + mouse support)
		return m, tea.Batch(
			tea.EnterAltScreen,       // Re-enter alternate screen (required!)
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)

	case gitOperationFinishedMsg:
		// Git operation has finished, we're back in TFE
		// Refresh file list to update git status
		m.loadFiles()

		// Refresh changedFiles after any git operation (push, pull, sync, etc.)
		// so the changes view stays current
		if m.showChangesOnly {
			oldCount := len(m.changedFiles)

			if changed, err := m.getChangedFiles(); err == nil {
				// Build a set of paths still changed for fast lookup
				newChangedPaths := make(map[string]bool, len(changed))
				for _, f := range changed {
					newChangedPaths[f.path] = true
				}

				// Close tabs for files no longer in the changed list
				closedTabs := 0
				for i := len(m.tabs) - 1; i >= 0; i-- {
					if !newChangedPaths[m.tabs[i].path] {
						m.tabs = append(m.tabs[:i], m.tabs[i+1:]...)
						closedTabs++
					}
				}

				// Reset activeTab if needed
				if len(m.tabs) == 0 {
					m.activeTab = 0
				} else if m.activeTab >= len(m.tabs) {
					m.activeTab = len(m.tabs) - 1
				}

				// Load the new active tab's content, or clear preview
				if len(m.tabs) > 0 {
					tab := m.tabs[m.activeTab]
					m.loadPreview(tab.path)
					m.populatePreviewCache()
				} else if closedTabs > 0 {
					m.preview.loaded = false
					m.preview.filePath = ""
					m.preview.fileName = ""
					m.preview.content = nil
					m.preview.cacheValid = false
				}

				m.changedFiles = changed

				// Show contextual status message for push/sync operations
				cleared := oldCount - len(changed)
				if msg.err == nil && (msg.operation == "push" || msg.operation == "sync") && cleared > 0 {
					if len(changed) == 0 {
						m.setStatusMessage(fmt.Sprintf("Pushed -- %d files cleared. All changes pushed", cleared), false)
					} else {
						m.setStatusMessage(fmt.Sprintf("Pushed -- %d files cleared (%d remaining)", cleared, len(changed)), false)
					}
				} else if msg.err == nil {
					m.setStatusMessage(fmt.Sprintf("✓ Git %s completed successfully", msg.operation), false)
				} else {
					m.setStatusMessage(fmt.Sprintf("✗ Git %s failed", msg.operation), true)
				}
			} else {
				// getChangedFiles failed, fall back to default status
				if msg.err == nil {
					m.setStatusMessage(fmt.Sprintf("✓ Git %s completed successfully", msg.operation), false)
				} else {
					m.setStatusMessage(fmt.Sprintf("✗ Git %s failed", msg.operation), true)
				}
			}
		} else if m.showGitReposOnly {
			m.setStatusMessage("🔍 Refreshing git repository status...", false)
			m.gitReposList = m.scanGitReposRecursive(m.gitReposScanRoot, 3, 50)
			m.setStatusMessage(fmt.Sprintf("✓ %s completed - Found %d repositories", msg.operation, len(m.gitReposList)), false)
		} else {
			// Show operation result
			if msg.err == nil {
				m.setStatusMessage(fmt.Sprintf("✓ Git %s completed successfully", msg.operation), false)
			} else {
				m.setStatusMessage(fmt.Sprintf("✗ Git %s failed", msg.operation), true)
			}
		}

		// Force a refresh and restore terminal state (alt screen + mouse support)
		return m, tea.Batch(
			tea.EnterAltScreen,       // Re-enter alternate screen (required!)
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)

	case fuzzySearchResultMsg:
		// Fuzzy search completed
		m.fuzzySearchActive = false
		if msg.err != nil {
			m.setStatusMessage(msg.err.Error(), true)
		} else if msg.selected != "" {
			m.navigateToFuzzyResult(msg.selected)
		}
		// Force a refresh and re-enable mouse support
		return m, tea.Batch(
			tea.ClearScreen,
			tea.EnableMouseCellMotion,
		)

	case updateAvailableMsg:
		// Update notification received from GitHub
		m.updateAvailable = true
		m.updateVersion = msg.version
		m.updateChangelog = msg.changelog
		m.updateURL = msg.url
		return m, nil

	case markdownRenderedMsg:
		// Markdown rendering completed in background
		// Just return to trigger a re-render with the cached content
		return m, nil

	case browserOpenedMsg:
		// Browser opened (or failed to open)
		if msg.success {
			m.setStatusMessage("✓ Opened in browser", false)
		} else {
			m.setStatusMessage("Failed to open in browser: "+msg.err.Error(), true)
		}
		return m, statusTimeoutCmd()

	case fileExplorerOpenedMsg:
		// File explorer opened (or failed to open)
		if msg.success {
			m.setStatusMessage("✓ Opened in file explorer", false)
		} else {
			m.setStatusMessage("Failed to open file explorer: "+msg.err.Error(), true)
		}
		return m, statusTimeoutCmd()

	case tmuxSplitMsg:
		// Tmux split/window operation completed
		if msg.err != nil {
			m.setStatusMessage(fmt.Sprintf("Tmux split failed: %v", msg.err), true)
		} else {
			m.setStatusMessage("Opened in tmux split", false)
		}
		return m, statusTimeoutCmd()

	case ghostTextMsg:
		// Ghost text suggestion received from Haiku API
		// Only apply if the sequence number matches (discard stale responses)
		if msg.seq == m.ghostTextSeq {
			m.ghostTextLoading = false
			if msg.err == nil {
				m.ghostText = msg.suggestion
			} else {
				m.ghostText = ""
			}
		}
		return m, nil

	case agentCheckTickMsg:
		// Periodic poll: detect agent session completions
		if m.agentAutoWatch {
			if cmd := m.checkAgentCompletions(); cmd != nil {
				return m, tea.Batch(cmd, agentCheckTick())
			}
		}
		return m, agentCheckTick() // Re-schedule next check

	case fileChangedMsg:
		// File system change detected by fsnotify watcher
		// Refresh the file list to reflect changes (new/deleted/modified files)
		m.loadFiles()

		// Auto-refresh git changes list when in changes mode
		if m.showChangesOnly {
			if changed, err := m.getChangedFiles(); err == nil {
				// Prune stale tabs (files no longer changed)
				if len(m.tabs) > 0 {
					newChangedPaths := make(map[string]bool, len(changed))
					for _, f := range changed {
						newChangedPaths[f.path] = true
					}
					for i := len(m.tabs) - 1; i >= 0; i-- {
						if !newChangedPaths[m.tabs[i].path] {
							m.tabs = append(m.tabs[:i], m.tabs[i+1:]...)
						}
					}
					if len(m.tabs) == 0 {
						m.activeTab = 0
					} else if m.activeTab >= len(m.tabs) {
						m.activeTab = len(m.tabs) - 1
					}
				}
				m.changedFiles = changed
			}
		}

		// If preview is active, refresh it too (file content may have changed)
		if m.preview.loaded && m.preview.filePath != "" {
			// Check if the changed file is the one being previewed
			if msg.path == m.preview.filePath || msg.op&fsnotify.Create != 0 || msg.op&fsnotify.Remove != 0 {
				m.loadPreview(m.preview.filePath)
				m.populatePreviewCache()
			}
		}

		// Re-subscribe to the watcher channel for the next event
		if m.watcherChan != nil {
			return m, waitForWatcherEvent(m.watcherChan)
		}
		return m, nil
	}

	return m, nil
}
