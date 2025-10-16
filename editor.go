package main

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// editorAvailable checks if an editor command is available
func editorAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// getAvailableEditor returns the first available editor
func getAvailableEditor() string {
	editors := []string{"micro", "nano", "vim", "vi"}
	for _, editor := range editors {
		if editorAvailable(editor) {
			return editor
		}
	}
	return ""
}

// openEditor opens a file in an external editor
func openEditor(editor, path string) tea.Cmd {
	c := exec.Command(editor, path)
	return tea.Sequence(
		tea.ClearScreen,
		tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err}
		}),
	)
}

// openTUITool opens a TUI application in the specified directory
func openTUITool(tool, dir string) tea.Cmd {
	c := exec.Command(tool)
	c.Dir = dir // Set working directory
	return tea.Sequence(
		tea.ClearScreen,
		tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err}
		}),
	)
}

// copyToClipboard copies text to the system clipboard
func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	// Try different clipboard commands based on platform
	if editorAvailable("termux-clipboard-set") {
		// Termux (Android)
		cmd = exec.Command("termux-clipboard-set")
	} else if editorAvailable("xclip") {
		cmd = exec.Command("xclip", "-selection", "clipboard")
	} else if editorAvailable("xsel") {
		cmd = exec.Command("xsel", "--clipboard", "--input")
	} else if editorAvailable("pbcopy") {
		// macOS
		cmd = exec.Command("pbcopy")
	} else if editorAvailable("clip.exe") {
		// Windows/WSL
		cmd = exec.Command("clip.exe")
	} else {
		return fmt.Errorf("no clipboard utility found (install termux-api, xclip, xsel, or use WSL)")
	}

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := pipe.Write([]byte(text)); err != nil {
		return err
	}

	if err := pipe.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}
