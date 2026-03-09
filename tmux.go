package main

// Module: tmux.go
// Purpose: Tmux integration for spawning panes from TFE
// Responsibilities:
// - Detect if running inside tmux
// - Smart split decision (horizontal vs vertical vs new window)
// - Spawn tmux panes with commands at specific directories
// - Sidebar-aware layout: TFE stays narrow, spawns target largest non-TFE pane

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	tmuxSidebarWidth = 30 // Target width for TFE sidebar after spawning
	tmuxMinPaneWidth = 40 // Minimum width for each half when splitting horizontally
	tmuxMinPaneHeight = 12 // Minimum height for each half when splitting vertically
)

// tmuxSplitMsg is sent when a tmux split operation completes
type tmuxSplitMsg struct {
	paneID string
	err    error
}

// tmuxPane represents a tmux pane with its dimensions
type tmuxPane struct {
	id     string
	width  int
	height int
}

// isInsideTmux checks whether TFE is running inside a tmux session
func isInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// getTmuxPaneID returns the pane ID of the current tmux pane
func getTmuxPaneID() string {
	out, err := exec.Command("tmux", "display-message", "-p", "#{pane_id}").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// listTmuxPanes returns all panes in the current window with their dimensions
func listTmuxPanes() ([]tmuxPane, error) {
	out, err := exec.Command("tmux", "list-panes", "-F", "#{pane_id}|#{pane_width}|#{pane_height}").Output()
	if err != nil {
		return nil, fmt.Errorf("tmux list-panes failed: %w", err)
	}

	var panes []tmuxPane
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			continue
		}
		w, _ := strconv.Atoi(parts[1])
		h, _ := strconv.Atoi(parts[2])
		panes = append(panes, tmuxPane{id: parts[0], width: w, height: h})
	}
	return panes, nil
}

// findLargestNonTFEPane finds the largest pane that isn't TFE's sidebar pane.
// Returns nil if TFE is the only pane.
func findLargestNonTFEPane(panes []tmuxPane, tfePaneID string) *tmuxPane {
	var best *tmuxPane
	bestArea := 0
	for i := range panes {
		if panes[i].id == tfePaneID {
			continue
		}
		area := panes[i].width * panes[i].height
		if area > bestArea {
			bestArea = area
			best = &panes[i]
		}
	}
	return best
}

// resizeTFESidebar resizes TFE's pane to the sidebar width
func resizeTFESidebar(tfePaneID string) {
	if tfePaneID == "" {
		return
	}
	exec.Command("tmux", "resize-pane", "-t", tfePaneID, "-x", strconv.Itoa(tmuxSidebarWidth)).Run()
}

// tmuxSmartSplit spawns a new tmux pane using sidebar-aware split strategy.
//
// Behavior:
//   - If TFE is the only pane: split horizontally from TFE, then resize TFE to sidebar width
//   - If other panes exist: find the largest non-TFE pane and split it
//     (horizontal if wide enough, vertical if tall enough, new window as fallback)
//   - After every split, resize TFE back to sidebar width
//
// The cmd argument is optional; if empty, the new pane gets a default shell.
// The cwd argument sets the working directory for the new pane.
func tmuxSmartSplit(cmd, cwd string) tea.Cmd {
	return func() tea.Msg {
		tfePaneID := getTmuxPaneID()

		panes, err := listTmuxPanes()
		if err != nil {
			return tmuxSplitMsg{err: err}
		}

		target := findLargestNonTFEPane(panes, tfePaneID)

		var result tmuxSplitMsg
		if target == nil {
			// TFE is the only pane - split from TFE horizontally
			result = runTmuxSplit("-h", "", cmd, cwd)
		} else {
			// Split the largest non-TFE pane
			if target.width/2 >= tmuxMinPaneWidth {
				result = runTmuxSplit("-h", target.id, cmd, cwd)
			} else if target.height/2 >= tmuxMinPaneHeight {
				result = runTmuxSplit("-v", target.id, cmd, cwd)
			} else {
				result = runTmuxNewWindow(cmd, cwd)
			}
		}

		// Resize TFE back to sidebar width after splitting
		if result.err == nil {
			resizeTFESidebar(tfePaneID)
		}

		return result
	}
}

// tmuxSplitRight forces a horizontal (right) split in tmux
func tmuxSplitRight(cmd, cwd string) tea.Cmd {
	return func() tea.Msg {
		tfePaneID := getTmuxPaneID()
		result := runTmuxSplit("-h", "", cmd, cwd)
		if result.err == nil {
			resizeTFESidebar(tfePaneID)
		}
		return result
	}
}

// tmuxSplitBelow forces a vertical (below) split in tmux
func tmuxSplitBelow(cmd, cwd string) tea.Cmd {
	return func() tea.Msg {
		tfePaneID := getTmuxPaneID()
		result := runTmuxSplit("-v", "", cmd, cwd)
		if result.err == nil {
			resizeTFESidebar(tfePaneID)
		}
		return result
	}
}

// tmuxNewWindow forces a new tmux window (adjacent to current)
func tmuxNewWindow(cmd, cwd string) tea.Cmd {
	return func() tea.Msg {
		return runTmuxNewWindow(cmd, cwd)
	}
}

// runTmuxSplit executes a tmux split-window command with the given orientation flag (-h or -v).
// If targetPane is non-empty, the split targets that pane instead of the current one.
// Returns a tmuxSplitMsg with the new pane ID on success, or an error on failure.
func runTmuxSplit(orientation, targetPane, cmd, cwd string) tmuxSplitMsg {
	args := []string{"split-window", orientation, "-P", "-F", "#{pane_id}"}
	if targetPane != "" {
		args = append(args, "-t", targetPane)
	}
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
