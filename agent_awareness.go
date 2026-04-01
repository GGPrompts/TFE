package main

// Module: agent_awareness.go
// Purpose: Detect active AI agent sessions and correlate them with changed files
// Responsibilities:
// - Parse agent session state files from /tmp/claude-code-state/
// - Match file paths to agent working directories
// - Provide short labels for display in changes mode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// agentStateDir is the directory where Claude Code writes session state files
const agentStateDir = "/tmp/claude-code-state"

// AgentSession represents a parsed agent session state file
type AgentSession struct {
	SessionID       string    `json:"session_id"`
	Status          string    `json:"status"`
	CurrentTool     string    `json:"current_tool"`
	SubagentCount   int       `json:"subagent_count"`
	WorkingDir      string    `json:"working_dir"`
	LastUpdated     time.Time `json:"last_updated"`
	PID             int       `json:"pid"`
	HookType        string    `json:"hook_type"`
	ContextPercent  int       `json:"context_percent"`
	Workspace       int       `json:"workspace"`
	ParentSessionID string    `json:"parent_session_id"`
	AgentID         string    `json:"agent_id"`
	AgentType       string    `json:"agent_type"`
}

// getAgentSessions reads all session state files from /tmp/claude-code-state/.
// Returns an empty slice (never an error) if the directory doesn't exist or is empty.
func getAgentSessions() []AgentSession {
	entries, err := os.ReadDir(agentStateDir)
	if err != nil {
		// Directory doesn't exist or can't be read — graceful fallback
		return nil
	}

	sessions := make([]AgentSession, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		fullPath := filepath.Join(agentStateDir, entry.Name())
		data, err := os.ReadFile(fullPath)
		if err != nil {
			continue // skip unreadable files
		}

		var session AgentSession
		if err := json.Unmarshal(data, &session); err != nil {
			continue // skip malformed JSON
		}

		// Only include sessions with a valid working directory
		if session.WorkingDir != "" {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// matchFileToAgent correlates a file's absolute path to an active agent session.
// It checks whether the file resides under any session's working directory.
// Returns a short label (e.g. "CC" for Claude Code) or empty string if no match.
func matchFileToAgent(filePath string, sessions []AgentSession) string {
	if len(sessions) == 0 || filePath == "" {
		return ""
	}

	for _, s := range sessions {
		if s.WorkingDir == "" {
			continue
		}

		// Skip subagent entries — the parent session already covers the working dir
		if s.ParentSessionID != "" {
			continue
		}

		// Check if the file path is under this session's working directory
		if isUnderDir(filePath, s.WorkingDir) {
			return agentLabel(s)
		}
	}

	return ""
}

// isUnderDir returns true if filePath is equal to or a descendant of dir.
func isUnderDir(filePath, dir string) bool {
	// Clean both paths for reliable comparison
	filePath = filepath.Clean(filePath)
	dir = filepath.Clean(dir)

	if filePath == dir {
		return true
	}

	// Ensure dir ends with separator for prefix check
	prefix := dir + string(filepath.Separator)
	return strings.HasPrefix(filePath, prefix)
}

// agentLabel returns a short display label for an agent session.
// Uses the agent_type field if present, otherwise defaults to "CC" (Claude Code).
func agentLabel(s AgentSession) string {
	switch strings.ToLower(s.AgentType) {
	case "explore":
		return "CC:Explore"
	case "general-purpose":
		return "CC:Agent"
	case "":
		// Top-level Claude Code session (no agent_type)
		return "CC"
	default:
		// Unknown agent type — show it abbreviated
		label := s.AgentType
		runes := []rune(label)
		if len(runes) > 10 {
			label = string(runes[:10])
		}
		return "CC:" + label
	}
}

// buildAgentFileMap builds a lookup map from file path to agent label
// for all changed files. This avoids repeated O(n*m) scanning during rendering.
func buildAgentFileMap(changedFiles []fileItem, sessions []AgentSession) map[string]string {
	if len(sessions) == 0 || len(changedFiles) == 0 {
		return nil
	}

	result := make(map[string]string, len(changedFiles))
	for _, f := range changedFiles {
		if label := matchFileToAgent(f.path, sessions); label != "" {
			result[f.path] = label
		}
	}

	// Return nil instead of empty map to make nil checks easy
	if len(result) == 0 {
		return nil
	}
	return result
}

// isActiveAgentStatus returns true if the status indicates an agent is actively working.
func isActiveAgentStatus(status string) bool {
	switch status {
	case "processing", "tool_use":
		return true
	default:
		return false
	}
}

// checkAgentCompletions compares current agent sessions against lastKnownAgentSessions
// to detect transitions from active (processing/tool_use) to done (gone or idle).
// If a completion is detected and TFE is not in an active dialog/prompt, it auto-switches
// to changes mode. Returns a tea.Cmd if a screen refresh is needed, nil otherwise.
func (m *model) checkAgentCompletions() tea.Cmd {
	currentSessions := getAgentSessions()

	// Build a map of current top-level sessions: session_id -> status
	current := make(map[string]string, len(currentSessions))
	for _, s := range currentSessions {
		if s.ParentSessionID == "" { // Only track top-level sessions
			current[s.SessionID] = s.Status
		}
	}

	// Detect completions: any session that was active and is now gone or no longer active
	completionDetected := false
	for id, oldStatus := range m.lastKnownAgentSessions {
		if !isActiveAgentStatus(oldStatus) {
			continue // Was not active before, no transition to detect
		}

		newStatus, exists := current[id]
		if !exists {
			// Session disappeared (agent process exited) -- completion
			completionDetected = true
			break
		}
		if !isActiveAgentStatus(newStatus) {
			// Session transitioned from active to idle/done
			completionDetected = true
			break
		}
	}

	// Update the last-known state regardless of whether we detected a completion
	m.lastKnownAgentSessions = current

	if !completionDetected {
		return nil
	}

	// Don't interrupt the user if they're in an active UI state
	if m.showDialog || m.commandFocused || m.contextMenuOpen ||
		m.fuzzySearchActive || m.searchMode || m.promptEditMode || m.filePickerMode {
		return nil
	}

	// Already in changes mode -- just refresh the data
	if m.showChangesOnly {
		if changed, err := m.getChangedFiles(); err == nil {
			m.changedFiles = changed
			m.agentSessions = currentSessions
			m.agentFileMap = buildAgentFileMap(changed, currentSessions)
			m.setStatusMessage(fmt.Sprintf("Agent finished -- refreshed changes (%d files)", len(changed)), false)
		}
		return statusTimeoutCmd()
	}

	// Auto-switch to changes mode
	changed, err := m.getChangedFiles()
	if err != nil || len(changed) == 0 {
		// No changes to show, or git error -- don't switch
		return nil
	}

	// Enter changes mode (mirrors the Ctrl+G entry path)
	m.showChangesOnly = true
	m.changedFiles = changed
	m.agentSessions = currentSessions
	m.agentFileMap = buildAgentFileMap(changed, currentSessions)
	m.changesRestoreDisplay = m.displayMode
	m.displayMode = modeDetail
	m.detailScrollX = 0
	m.showDiffPreview = true
	m.calculateLayout()
	m.cursor = 0
	m.loadFiles()

	// Clear any other filter modes
	m.showFavoritesOnly = false
	m.showGitReposOnly = false
	m.showTrashOnly = false
	m.showPromptsOnly = false

	m.setStatusMessage(fmt.Sprintf("Agent finished -- showing %d changed files", len(changed)), false)
	return statusTimeoutCmd()
}

// toggleAgentView enters or exits the agent conversation viewer.
// When entering, it navigates to the active session's directory (or project dir)
// and sorts by modified time. When exiting, it restores the previous path.
func (m *model) toggleAgentView() {
	if m.showAgentView {
		// Exit agent view — restore previous state
		m.showAgentView = false
		if m.agentViewRestore != "" {
			m.currentPath = m.agentViewRestore
			m.agentViewRestore = ""
		}
		m.sortBy = m.agentViewRestoreSort
		m.sortAsc = m.agentViewRestoreAsc
		m.displayMode = m.agentViewRestoreMode
		m.calculateLayout()
		m.cursor = 0
		m.loadFiles()
		m.setStatusMessage("Agent view closed", false)
		return
	}

	// Enter agent view — find the project's .claude directory
	// Try current path first, then git root
	agentDir := getClaudeSessionDir(m.currentPath)
	if agentDir == "" {
		gitRoot, err := m.gitRevParseRoot(m.currentPath)
		if err == nil {
			agentDir = getClaudeSessionDir(gitRoot)
		}
	}
	if agentDir == "" {
		m.setStatusMessage("No Claude session found for this directory", true)
		return
	}

	// Clear other filter modes
	if m.showChangesOnly {
		m.exitChangesMode()
	}
	m.showFavoritesOnly = false
	m.showGitReposOnly = false
	m.showTrashOnly = false
	m.showPromptsOnly = false

	m.showAgentView = true
	m.agentViewRestore = m.currentPath
	m.agentViewRestoreSort = m.sortBy
	m.agentViewRestoreAsc = m.sortAsc
	m.agentViewRestoreMode = m.displayMode
	m.currentPath = agentDir
	m.sortBy = "modified"
	m.sortAsc = false // Newest first
	m.cursor = 0
	m.loadFiles()
	m.sortFiles()
	m.setStatusMessage(fmt.Sprintf("Agent sessions: %s", agentDir), false)
}

// populateAgentMetadata enriches fileItems with agent type and description
// by reading .meta.json and the first JSONL line for each agent/session file.
func (m *model) populateAgentMetadata() {
	for i := range m.files {
		file := &m.files[i]
		if file.isDir || file.name == ".." {
			continue
		}

		ext := filepath.Ext(file.name)

		if ext == ".jsonl" {
			// Read first line to get slug or initial user message
			f, err := os.Open(file.path)
			if err != nil {
				continue
			}

			// Read up to 8KB for the first line
			buf := make([]byte, 8192)
			n, _ := f.Read(buf)
			f.Close()
			if n == 0 {
				continue
			}

			// Find first newline to get just the first line
			line := string(buf[:n])
			if idx := strings.IndexByte(line, '\n'); idx >= 0 {
				line = line[:idx]
			}

			var firstEntry struct {
				Type    string `json:"type"`
				Slug    string `json:"slug"`
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}
			if json.Unmarshal([]byte(line), &firstEntry) == nil {
				if firstEntry.Slug != "" {
					file.agentDescription = firstEntry.Slug
				} else if firstEntry.Message.Content != "" {
					// Truncate long prompts to first 80 chars
					desc := firstEntry.Message.Content
					// Take first line only
					if idx := strings.IndexByte(desc, '\n'); idx >= 0 {
						desc = desc[:idx]
					}
					if len(desc) > 80 {
						desc = desc[:77] + "..."
					}
					file.agentDescription = desc
				}
			}

			// Check for corresponding .meta.json (for subagent files)
			metaPath := strings.TrimSuffix(file.path, ".jsonl") + ".meta.json"
			if metaData, err := os.ReadFile(metaPath); err == nil {
				var meta struct {
					AgentType string `json:"agentType"`
				}
				if json.Unmarshal(metaData, &meta) == nil && meta.AgentType != "" {
					file.agentType = meta.AgentType
				}
			}
		}
	}
}

// formatAgentDisplayName returns a human-readable name for agent view files.
// Replaces raw UUID filenames with type + slug or short description.
func (m *model) formatAgentDisplayName(file fileItem) string {
	ext := filepath.Ext(file.name)
	base := strings.TrimSuffix(file.name, ext)

	if file.isDir {
		// Session directories: show "session" + short UUID
		if len(base) > 8 {
			return "session " + base[:8]
		}
		return file.name
	}

	switch ext {
	case ".jsonl":
		// Subagent files: show agent type or slug
		if file.agentType != "" {
			// Clean up agent type (e.g. "opus-agents:general-purpose" → "general-purpose")
			agentType := file.agentType
			if idx := strings.LastIndex(agentType, ":"); idx >= 0 {
				agentType = agentType[idx+1:]
			}
			return agentType
		}
		if file.agentDescription != "" {
			// Use slug as display name (e.g. "peaceful-chasing-bubble")
			desc := file.agentDescription
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			return desc
		}
		// Fallback: short UUID
		if len(base) > 12 {
			return base[:12] + ".."
		}
		return base

	case ".json":
		// Meta files: hide or show as "meta"
		if strings.HasSuffix(base, ".meta") {
			parent := strings.TrimSuffix(base, ".meta")
			if len(parent) > 8 {
				return "meta " + parent[:8] + ".."
			}
			return "meta " + parent
		}
		return file.name

	default:
		return file.name
	}
}

// getClaudeProjectDir returns the path to the .claude/projects/<encoded-cwd>/ directory
// for the given working directory. Claude Code encodes paths by replacing "/" with "-".
func getClaudeProjectDir(cwd string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Claude Code encodes: /home/builder/projects/TFE → -home-builder-projects-TFE
	encoded := strings.ReplaceAll(cwd, "/", "-")
	return filepath.Join(homeDir, ".claude", "projects", encoded)
}

// getClaudeSessionDir returns the project-level .claude/projects/<encoded-cwd>/ directory.
// This directory contains session JSONL files at the top level and session subdirectories
// with subagent conversations inside them.
func getClaudeSessionDir(cwd string) string {
	projectDir := getClaudeProjectDir(cwd)
	if projectDir == "" {
		return ""
	}
	if info, err := os.Stat(projectDir); err == nil && info.IsDir() {
		return projectDir
	}
	return ""
}
