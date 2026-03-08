package main

// Module: tmux.go
// Purpose: Tmux integration for spawning panes from TFE
// Responsibilities:
// - Detect if running inside tmux
// - Smart split decision (horizontal vs vertical vs new window)
// - Spawn tmux panes with commands at specific directories

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// tmuxSplitMsg is sent when a tmux split operation completes
type tmuxSplitMsg struct {
	paneID string
	err    error
}

// isInsideTmux checks whether TFE is running inside a tmux session
func isInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// getCurrentPaneDimensions returns the width and height of the current tmux pane
func getCurrentPaneDimensions() (width int, height int, err error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{pane_width}|#{pane_height}").Output()
	if err != nil {
		return 0, 0, fmt.Errorf("tmux display-message failed: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(out)), "|")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected tmux output: %s", string(out))
	}

	width, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid pane width: %w", err)
	}

	height, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid pane height: %w", err)
	}

	return width, height, nil
}

// tmuxSmartSplit spawns a new tmux pane using the best available split strategy.
// It checks current pane dimensions and chooses:
//   - Horizontal split if width/2 >= 40
//   - Vertical split if height/2 >= 12
//   - New window as fallback
//
// The cmd argument is optional; if empty, the new pane gets a default shell.
// The cwd argument sets the working directory for the new pane.
// This does NOT use tea.ExecProcess -- TFE stays running while the pane is created.
func tmuxSmartSplit(cmd, cwd string) tea.Cmd {
	return func() tea.Msg {
		w, h, err := getCurrentPaneDimensions()
		if err != nil {
			return tmuxSplitMsg{err: err}
		}

		if w/2 >= 40 {
			return runTmuxSplit("-h", cmd, cwd)
		}
		if h/2 >= 12 {
			return runTmuxSplit("-v", cmd, cwd)
		}
		return runTmuxNewWindow(cmd, cwd)
	}
}

// tmuxSplitRight forces a horizontal (right) split in tmux
func tmuxSplitRight(cmd, cwd string) tea.Cmd {
	return func() tea.Msg {
		return runTmuxSplit("-h", cmd, cwd)
	}
}

// tmuxSplitBelow forces a vertical (below) split in tmux
func tmuxSplitBelow(cmd, cwd string) tea.Cmd {
	return func() tea.Msg {
		return runTmuxSplit("-v", cmd, cwd)
	}
}

// tmuxNewWindow forces a new tmux window (adjacent to current)
func tmuxNewWindow(cmd, cwd string) tea.Cmd {
	return func() tea.Msg {
		return runTmuxNewWindow(cmd, cwd)
	}
}

// runTmuxSplit executes a tmux split-window command with the given orientation flag (-h or -v).
// Returns a tmuxSplitMsg with the new pane ID on success, or an error on failure.
func runTmuxSplit(orientation, cmd, cwd string) tmuxSplitMsg {
	args := []string{"split-window", orientation, "-P", "-F", "#{pane_id}"}
	if cwd != "" {
		args = append(args, "-c", cwd)
	}
	if cmd != "" {
		args = append(args, cmd)
	}

	out, err := exec.Command("tmux", args...).Output()
	if err != nil {
		return tmuxSplitMsg{err: fmt.Errorf("tmux split-window %s failed: %w", orientation, err)}
	}

	paneID := strings.TrimSpace(string(out))
	return tmuxSplitMsg{paneID: paneID}
}

// runTmuxNewWindow executes a tmux new-window command (adjacent to current window).
// Returns a tmuxSplitMsg with the new pane ID on success, or an error on failure.
func runTmuxNewWindow(cmd, cwd string) tmuxSplitMsg {
	args := []string{"new-window", "-a", "-P", "-F", "#{pane_id}"}
	if cwd != "" {
		args = append(args, "-c", cwd)
	}
	if cmd != "" {
		args = append(args, cmd)
	}

	out, err := exec.Command("tmux", args...).Output()
	if err != nil {
		return tmuxSplitMsg{err: fmt.Errorf("tmux new-window failed: %w", err)}
	}

	paneID := strings.TrimSpace(string(out))
	return tmuxSplitMsg{paneID: paneID}
}
