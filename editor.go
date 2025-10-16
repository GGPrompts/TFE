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

// isImageFile checks if a file is an image based on extension
func isImageFile(path string) bool {
	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp", ".ico", ".tiff", ".tif"}
	for _, ext := range imageExts {
		if len(path) >= len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// isHTMLFile checks if a file is an HTML file based on extension
func isHTMLFile(path string) bool {
	return len(path) >= 5 && (path[len(path)-5:] == ".html" || path[len(path)-4:] == ".htm")
}

// isBrowserFile checks if a file should be opened in a browser
func isBrowserFile(path string) bool {
	return isImageFile(path) || isHTMLFile(path)
}

// getAvailableBrowser returns the command to open files in the default browser
func getAvailableBrowser() string {
	// WSL - try wslview first (from wslu package), then use Windows commands
	if editorAvailable("wslview") {
		return "wslview"
	}
	// Windows via WSL
	if editorAvailable("cmd.exe") {
		return "cmd.exe"
	}
	// Linux - xdg-open is the standard
	if editorAvailable("xdg-open") {
		return "xdg-open"
	}
	// macOS
	if editorAvailable("open") {
		return "open"
	}
	return ""
}

// openInBrowser opens a file in the default browser
func openInBrowser(path string) tea.Cmd {
	browser := getAvailableBrowser()
	if browser == "" {
		return nil
	}

	return func() tea.Msg {
		var c *exec.Cmd
		if browser == "cmd.exe" {
			// Windows via WSL - use cmd.exe /c start
			c = exec.Command("cmd.exe", "/c", "start", path)
		} else {
			// Linux/macOS/wslview
			c = exec.Command(browser, path)
		}

		// Start the browser without blocking (browsers run in background)
		_ = c.Start()

		// Return a clear screen message to refresh the UI
		return tea.ClearScreen()
	}
}
