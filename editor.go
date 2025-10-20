package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

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

// isImageFile checks if a file is an image based on extension (case-insensitive)
func isImageFile(path string) bool {
	imageExts := []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".webp", ".ico", ".tiff", ".tif"}
	// Convert path to lowercase for case-insensitive comparison
	lowerPath := ""
	for _, ch := range path {
		if ch >= 'A' && ch <= 'Z' {
			lowerPath += string(ch + 32)
		} else {
			lowerPath += string(ch)
		}
	}
	for _, ext := range imageExts {
		if len(lowerPath) >= len(ext) && lowerPath[len(lowerPath)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// isHTMLFile checks if a file is an HTML file based on extension (case-insensitive)
func isHTMLFile(path string) bool {
	// Convert path to lowercase for case-insensitive comparison
	lowerPath := ""
	for _, ch := range path {
		if ch >= 'A' && ch <= 'Z' {
			lowerPath += string(ch + 32)
		} else {
			lowerPath += string(ch)
		}
	}
	return len(lowerPath) >= 5 && (lowerPath[len(lowerPath)-5:] == ".html" || lowerPath[len(lowerPath)-4:] == ".htm")
}

// isBrowserFile checks if a file should be opened in a browser
func isBrowserFile(path string) bool {
	return isImageFile(path) || isHTMLFile(path)
}

// isWSL checks if we're running in Windows Subsystem for Linux
func isWSL() bool {
	// Check for WSL-specific indicators
	if _, err := os.Stat("/proc/version"); err == nil {
		data, err := os.ReadFile("/proc/version")
		if err == nil {
			version := string(data)
			return strings.Contains(strings.ToLower(version), "microsoft") ||
				strings.Contains(strings.ToLower(version), "wsl")
		}
	}
	return false
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

// browserOpenedMsg is returned when a file is opened in browser
type browserOpenedMsg struct {
	success bool
	err     error
}

// openInBrowser opens a file in the default browser
func openInBrowser(path string) tea.Cmd {
	browser := getAvailableBrowser()
	if browser == "" {
		return func() tea.Msg {
			return browserOpenedMsg{
				success: false,
				err:     fmt.Errorf("no browser command found (install xdg-open, wslview, or open)"),
			}
		}
	}

	return func() tea.Msg {
		var c *exec.Cmd
		browserPath := path

		// In WSL, convert Linux paths to Windows paths for better compatibility
		if isWSL() && browser != "wslview" {
			// Use wslpath to convert WSL path to Windows path
			cmd := exec.Command("wslpath", "-w", path)
			output, err := cmd.Output()
			if err == nil {
				browserPath = strings.TrimSpace(string(output))
			}
			// If conversion fails, fall back to original path
		}

		if browser == "cmd.exe" {
			// Windows via WSL - use cmd.exe /c start with empty title ""
			// The empty title "" prevents cmd from treating first arg as window title
			c = exec.Command("cmd.exe", "/c", "start", "", browserPath)
		} else {
			// Linux/macOS/wslview
			c = exec.Command(browser, browserPath)
		}

		// Start the browser without blocking (browsers run in background)
		err := c.Start()
		if err != nil {
			return browserOpenedMsg{
				success: false,
				err:     err,
			}
		}

		// Success
		return browserOpenedMsg{
			success: true,
			err:     nil,
		}
	}
}

// getAvailableImageViewer returns the first available image viewer
func getAvailableImageViewer() string {
	viewers := []string{"viu", "timg", "chafa"}
	for _, viewer := range viewers {
		if editorAvailable(viewer) {
			return viewer
		}
	}
	return ""
}

// openImageViewer opens an image in a TUI viewer
func openImageViewer(path string) tea.Cmd {
	viewer := getAvailableImageViewer()
	if viewer == "" {
		return func() tea.Msg {
			return editorFinishedMsg{fmt.Errorf("no image viewer found (install viu, timg, or chafa)")}
		}
	}

	var c *exec.Cmd
	if viewer == "viu" {
		// viu with transparent background and size to terminal
		// Wrap in shell to pause after displaying image (same as command prompt)
		script := fmt.Sprintf(`viu -t '%s'
echo ""
echo "Press any key to continue..."
read -n 1 -s -r`, path)
		c = exec.Command("bash", "-c", script)
	} else if viewer == "timg" {
		// timg with grid view support
		script := fmt.Sprintf(`timg '%s'
echo ""
echo "Press any key to continue..."
read -n 1 -s -r`, path)
		c = exec.Command("bash", "-c", script)
	} else {
		// chafa or other
		script := fmt.Sprintf(`%s '%s'
echo ""
echo "Press any key to continue..."
read -n 1 -s -r`, viewer, path)
		c = exec.Command("bash", "-c", script)
	}

	// Set up stdin/stdout/stderr for interaction (same as command execution)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return tea.Sequence(
		tea.ClearScreen,
		tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err}
		}),
	)
}

// getAvailableImageEditor returns the first available image editor
func getAvailableImageEditor() string {
	editors := []string{"textual-paint", "durdraw"}
	for _, editor := range editors {
		if editorAvailable(editor) {
			return editor
		}
	}
	return ""
}

// openImageEditor opens an image in a TUI editor
func openImageEditor(path string) tea.Cmd {
	editor := getAvailableImageEditor()
	if editor == "" {
		return func() tea.Msg {
			return editorFinishedMsg{fmt.Errorf("no image editor found (install textual-paint)")}
		}
	}

	c := exec.Command(editor, path)
	return tea.Sequence(
		tea.ClearScreen,
		tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err}
		}),
	)
}
