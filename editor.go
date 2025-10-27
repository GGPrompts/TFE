package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	// SECURITY: Validate filename to prevent argument injection
	// Filenames starting with '-' could be interpreted as editor flags
	cleanPath := filepath.Clean(path)
	filename := filepath.Base(cleanPath)

	if strings.HasPrefix(filename, "-") {
		return func() tea.Msg {
			return editorFinishedMsg{
				err: fmt.Errorf("invalid filename: cannot open files starting with '-' (potential argument injection)"),
			}
		}
	}

	// Use absolute path to avoid ambiguity
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return func() tea.Msg {
			return editorFinishedMsg{err: err}
		}
	}

	c := exec.Command(editor, absPath)
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
		// Termux (Android) - use shell wrapper because termux-clipboard-set
		// fails with exit status 2 when using StdinPipe directly
		cmd = exec.Command("bash", "-c", fmt.Sprintf("termux-clipboard-set <<'CLIPBOARD_EOF'\n%s\nCLIPBOARD_EOF", text))
		return cmd.Run()
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

	// For non-Termux platforms, use the standard StdinPipe approach
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
	return isImageFile(path) || isHTMLFile(path) || isPDFFile(path)
}

// isCSVFile checks if a file is a CSV/TSV file based on extension (case-insensitive)
func isCSVFile(path string) bool {
	csvExts := []string{".csv", ".tsv"}
	lowerPath := ""
	for _, ch := range path {
		if ch >= 'A' && ch <= 'Z' {
			lowerPath += string(ch + 32)
		} else {
			lowerPath += string(ch)
		}
	}
	for _, ext := range csvExts {
		if len(lowerPath) >= len(ext) && lowerPath[len(lowerPath)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// isPDFFile checks if a file is a PDF based on extension (case-insensitive)
func isPDFFile(path string) bool {
	lowerPath := ""
	for _, ch := range path {
		if ch >= 'A' && ch <= 'Z' {
			lowerPath += string(ch + 32)
		} else {
			lowerPath += string(ch)
		}
	}
	return len(lowerPath) >= 4 && lowerPath[len(lowerPath)-4:] == ".pdf"
}

// isVideoFile checks if a file is a video based on extension (case-insensitive)
func isVideoFile(path string) bool {
	videoExts := []string{".mp4", ".mkv", ".avi", ".mov", ".webm", ".flv", ".wmv", ".m4v"}
	lowerPath := ""
	for _, ch := range path {
		if ch >= 'A' && ch <= 'Z' {
			lowerPath += string(ch + 32)
		} else {
			lowerPath += string(ch)
		}
	}
	for _, ext := range videoExts {
		if len(lowerPath) >= len(ext) && lowerPath[len(lowerPath)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// isAudioFile checks if a file is an audio file based on extension (case-insensitive)
func isAudioFile(path string) bool {
	audioExts := []string{".mp3", ".wav", ".flac", ".ogg", ".m4a", ".aac", ".wma", ".opus", ".ape"}
	lowerPath := ""
	for _, ch := range path {
		if ch >= 'A' && ch <= 'Z' {
			lowerPath += string(ch + 32)
		} else {
			lowerPath += string(ch)
		}
	}
	for _, ext := range audioExts {
		if len(lowerPath) >= len(ext) && lowerPath[len(lowerPath)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// isDatabaseFile checks if a file is a database file based on extension (case-insensitive)
func isDatabaseFile(path string) bool {
	dbExts := []string{".db", ".sqlite", ".sqlite3"}
	lowerPath := ""
	for _, ch := range path {
		if ch >= 'A' && ch <= 'Z' {
			lowerPath += string(ch + 32)
		} else {
			lowerPath += string(ch)
		}
	}
	for _, ext := range dbExts {
		if len(lowerPath) >= len(ext) && lowerPath[len(lowerPath)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// isArchiveFile checks if a file is an archive based on extension (case-insensitive)
func isArchiveFile(path string) bool {
	archiveExts := []string{".zip", ".tar", ".gz", ".7z", ".rar", ".bz2", ".xz", ".tar.gz", ".tgz"}
	lowerPath := ""
	for _, ch := range path {
		if ch >= 'A' && ch <= 'Z' {
			lowerPath += string(ch + 32)
		} else {
			lowerPath += string(ch)
		}
	}
	for _, ext := range archiveExts {
		if len(lowerPath) >= len(ext) && lowerPath[len(lowerPath)-len(ext):] == ext {
			return true
		}
	}
	return false
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
			// Windows via WSL - use powershell.exe Start-Process (handles UNC paths)
			// cmd.exe /c start doesn't support UNC paths (\\wsl.localhost\...)
			c = exec.Command("powershell.exe", "-Command", "Start-Process", "'"+browserPath+"'")
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

// getAvailableCSVViewer returns the first available CSV/spreadsheet viewer
func getAvailableCSVViewer() string {
	viewers := []string{"visidata", "sc-im"}
	for _, viewer := range viewers {
		if editorAvailable(viewer) {
			return viewer
		}
	}
	return ""
}

// openCSVViewer opens a CSV file in a spreadsheet viewer
func openCSVViewer(path string) tea.Cmd {
	viewer := getAvailableCSVViewer()
	if viewer == "" {
		// Fallback to text editor
		editor := getAvailableEditor()
		if editor == "" {
			return func() tea.Msg {
				return editorFinishedMsg{fmt.Errorf("no CSV viewer or text editor found")}
			}
		}
		return openEditor(editor, path)
	}

	c := exec.Command(viewer, path)
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

// getAvailableVideoPlayer returns the first available video player
func getAvailableVideoPlayer() string {
	players := []string{"mpv", "vlc", "mplayer"}
	for _, player := range players {
		if editorAvailable(player) {
			return player
		}
	}
	return ""
}

// openVideoPlayer opens a video file in a media player
func openVideoPlayer(path string) tea.Cmd {
	player := getAvailableVideoPlayer()
	if player == "" {
		return func() tea.Msg {
			return editorFinishedMsg{fmt.Errorf("no video player found (install mpv)")}
		}
	}

	// mpv with terminal output
	c := exec.Command(player, path)
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

// getAvailableAudioPlayer returns the first available audio player
func getAvailableAudioPlayer() string {
	players := []string{"mpv", "cmus", "moc"}
	for _, player := range players {
		if editorAvailable(player) {
			return player
		}
	}
	return ""
}

// openAudioPlayer opens an audio file in a media player
func openAudioPlayer(path string) tea.Cmd {
	player := getAvailableAudioPlayer()
	if player == "" {
		return func() tea.Msg {
			return editorFinishedMsg{fmt.Errorf("no audio player found (install mpv)")}
		}
	}

	// mpv with audio-only mode
	c := exec.Command(player, path)
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

// getAvailableHexViewer returns the first available hex viewer
func getAvailableHexViewer() string {
	viewers := []string{"hexyl", "hexpatch", "hexabyte", "xxd"}
	for _, viewer := range viewers {
		if editorAvailable(viewer) {
			return viewer
		}
	}
	return ""
}

// openHexViewer opens a binary file in a hex viewer
func openHexViewer(path string) tea.Cmd {
	viewer := getAvailableHexViewer()
	if viewer == "" {
		return func() tea.Msg {
			return editorFinishedMsg{fmt.Errorf("no hex viewer found (install hexyl)")}
		}
	}

	var c *exec.Cmd
	if viewer == "hexyl" {
		// hexyl displays directly with built-in paging on terminals
		// Add a pause wrapper to return to TFE cleanly
		script := fmt.Sprintf(`hexyl '%s'
echo ""
echo "Press any key to return to TFE..."
read -n 1 -s -r`, path)
		c = exec.Command("bash", "-c", script)
	} else {
		// Other hex viewers
		c = exec.Command(viewer, path)
	}

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

// getAvailableDatabaseViewer returns the first available database viewer
func getAvailableDatabaseViewer() string {
	viewers := []string{"harlequin", "litecli", "sqlite3"}
	for _, viewer := range viewers {
		if editorAvailable(viewer) {
			return viewer
		}
	}
	return ""
}

// openDatabaseViewer opens a database file in a viewer
func openDatabaseViewer(path string) tea.Cmd {
	viewer := getAvailableDatabaseViewer()
	if viewer == "" {
		return func() tea.Msg {
			return editorFinishedMsg{fmt.Errorf("no database viewer found (install harlequin)")}
		}
	}

	c := exec.Command(viewer, path)
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

// getAvailablePDFViewer returns the first available PDF viewer
func getAvailablePDFViewer() string {
	viewers := []string{"timg", "termpdf.py"}
	for _, viewer := range viewers {
		if editorAvailable(viewer) {
			return viewer
		}
	}
	return ""
}

// openPDFViewer opens a PDF file in a terminal viewer
func openPDFViewer(path string) tea.Cmd {
	viewer := getAvailablePDFViewer()
	if viewer == "" {
		// Fallback to browser
		return openInBrowser(path)
	}

	var c *exec.Cmd
	if viewer == "timg" {
		script := fmt.Sprintf(`timg '%s'
echo ""
echo "Press any key to continue..."
read -n 1 -s -r`, path)
		c = exec.Command("bash", "-c", script)
	} else {
		c = exec.Command(viewer, path)
	}

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

// openImageEditorNew launches an image editor with a blank canvas (no file required)
func openImageEditorNew(currentDir string) tea.Cmd {
	editor := getAvailableImageEditor()
	if editor == "" {
		return func() tea.Msg {
			return editorFinishedMsg{fmt.Errorf("no image editor found (install textual-paint)")}
		}
	}

	// Launch textual-paint without a file argument - it will start with a blank canvas
	// Set the working directory so when user saves, it defaults to current directory
	c := exec.Command(editor)
	c.Dir = currentDir
	return tea.Sequence(
		tea.ClearScreen,
		tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err}
		}),
	)
}

// fileExplorerOpenedMsg is returned when a folder is opened in file explorer
type fileExplorerOpenedMsg struct {
	success bool
	err     error
}

// openInFileExplorer opens the current directory in the system file explorer
func openInFileExplorer(path string) tea.Cmd {
	return func() tea.Msg {
		var c *exec.Cmd

		if isWSL() {
			// WSL: Use explorer.exe to open Windows Explorer
			// Convert Linux path to Windows path for better compatibility
			cmd := exec.Command("wslpath", "-w", path)
			output, err := cmd.Output()
			var winPath string
			if err == nil {
				winPath = strings.TrimSpace(string(output))
			} else {
				// If conversion fails, use the Linux path (explorer.exe can handle some WSL paths)
				winPath = path
			}
			c = exec.Command("explorer.exe", winPath)
		} else if editorAvailable("xdg-open") {
			// Linux: Use xdg-open to open default file manager
			c = exec.Command("xdg-open", path)
		} else if editorAvailable("open") {
			// macOS: Use open to open Finder
			c = exec.Command("open", path)
		} else {
			return fileExplorerOpenedMsg{
				success: false,
				err:     fmt.Errorf("no file explorer command found"),
			}
		}

		// Start the file explorer without blocking
		err := c.Start()
		if err != nil {
			return fileExplorerOpenedMsg{
				success: false,
				err:     err,
			}
		}

		// Success
		return fileExplorerOpenedMsg{
			success: true,
			err:     nil,
		}
	}
}
