package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			PaddingLeft(2)

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(2).
			PaddingBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("39")).
			Bold(true)

	folderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39"))

	fileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			PaddingLeft(2)

	claudeContextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")) // Orange
)

// isClaudeContextFile checks if a file/folder is automatically loaded by Claude Code
func isClaudeContextFile(name string) bool {
	// Files that Claude Code automatically loads into context
	claudeFiles := []string{
		"CLAUDE.md",
		"CLAUDE.local.md",
		".claude",
	}

	for _, cf := range claudeFiles {
		if name == cf {
			return true
		}
	}

	return false
}

// getFileIcon returns the appropriate icon based on file type
func getFileIcon(item fileItem) string {
	if item.isDir {
		if item.name == ".." {
			return "↑"  // Up arrow for parent dir
		}
		return "▸"  // Triangle for folders
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(item.name))

	// Map extensions to simple text markers
	iconMap := map[string]string{
		// Programming languages
		".go":     "[GO]",
		".py":     "[PY]",
		".js":     "[JS]",
		".ts":     "[TS]",
		".jsx":    "[JSX]",
		".tsx":    "[TSX]",
		".rs":     "[RS]",
		".c":      "[C]",
		".cpp":    "[C++]",
		".h":      "[H]",
		".java":   "[JAVA]",
		".rb":     "[RB]",
		".php":    "[PHP]",
		".sh":     "[SH]",
		".bash":   "[SH]",

		// Web
		".html":   "[HTML]",
		".css":    "[CSS]",
		".vue":    "[VUE]",

		// Data/Config
		".json":   "[JSON]",
		".yaml":   "[YAML]",
		".yml":    "[YAML]",
		".toml":   "[TOML]",
		".xml":    "[XML]",

		// Documents
		".md":     "[MD]",
		".txt":    "[TXT]",
		".pdf":    "[PDF]",

		// Archives
		".zip":    "[ZIP]",
		".tar":    "[TAR]",
		".gz":     "[GZ]",
	}

	// Check for icon mapping
	if icon, ok := iconMap[ext]; ok {
		return icon
	}

	// Check for special files without extension
	switch item.name {
	case "CLAUDE.md", "CLAUDE.local.md":
		return "[CLAUDE]"
	case "Makefile", "makefile":
		return "[MAKE]"
	case "Dockerfile":
		return "[DOCKER]"
	case "LICENSE", "LICENSE.txt", "LICENSE.md":
		return "[LIC]"
	case "README", "README.md", "README.txt":
		return "[README]"
	case ".gitignore":
		return "[GIT]"
	case ".claude":
		return "▸[CLAUDE]"
	case "go.mod", "go.sum":
		return "[GO]"
	}

	// Default file marker
	return "•"
}

// formatFileSize returns a human-readable file size
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// formatModTime returns a relative time string
func formatModTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	} else if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1w ago"
		}
		return fmt.Sprintf("%dw ago", weeks)
	} else if diff < 365*24*time.Hour {
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1mo ago"
		}
		return fmt.Sprintf("%dmo ago", months)
	}
	years := int(diff.Hours() / 24 / 365)
	if years == 1 {
		return "1y ago"
	}
	return fmt.Sprintf("%dy ago", years)
}

type fileItem struct {
	name    string
	path    string
	isDir   bool
	size    int64
	modTime time.Time
	mode    os.FileMode
}

type model struct {
	currentPath string
	files       []fileItem
	cursor      int
	height      int
	width       int
	showHidden  bool
}

func initialModel() model {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	m := model{
		currentPath: cwd,
		cursor:      0,
		height:      24,
		width:       80,
		showHidden:  false,
	}

	m.loadFiles()
	return m
}

func (m *model) loadFiles() {
	entries, err := os.ReadDir(m.currentPath)
	if err != nil {
		m.files = []fileItem{}
		return
	}

	// Reset files slice
	m.files = []fileItem{}

	// Add parent directory if not at root
	if m.currentPath != "/" {
		m.files = append(m.files, fileItem{
			name:  "..",
			path:  filepath.Dir(m.currentPath),
			isDir: true,
		})
	}

	// Add directories first, then files
	var dirs, files []fileItem

	for _, entry := range entries {
		// Skip hidden files starting with . (unless showHidden is true)
		if !m.showHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		item := fileItem{
			name:    entry.Name(),
			path:    filepath.Join(m.currentPath, entry.Name()),
			isDir:   entry.IsDir(),
			size:    info.Size(),
			modTime: info.ModTime(),
			mode:    info.Mode(),
		}

		if entry.IsDir() {
			dirs = append(dirs, item)
		} else {
			files = append(files, item)
		}
	}

	// Sort alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
	})

	m.files = append(m.files, dirs...)
	m.files = append(m.files, files...)

	// Reset cursor if out of bounds
	if m.cursor >= len(m.files) {
		m.cursor = 0
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}

		case "enter", " ":
			if len(m.files) > 0 && m.files[m.cursor].isDir {
				m.currentPath = m.files[m.cursor].path
				m.cursor = 0
				m.loadFiles()
			}

		case "h", "left":
			// Go to parent directory
			if m.currentPath != "/" {
				m.currentPath = filepath.Dir(m.currentPath)
				m.cursor = 0
				m.loadFiles()
			}

		case ".", "ctrl+h":
			// Toggle hidden files
			m.showHidden = !m.showHidden
			m.loadFiles()
		}

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Action == tea.MouseActionRelease {
				// Calculate which item was clicked (accounting for header lines)
				clickedIndex := msg.Y - 3 // 3 lines for title and path
				if clickedIndex >= 0 && clickedIndex < len(m.files) {
					m.cursor = clickedIndex
					if m.files[m.cursor].isDir {
						m.currentPath = m.files[m.cursor].path
						m.cursor = 0
						m.loadFiles()
					}
				}
			}

		case tea.MouseButtonWheelUp:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.MouseButtonWheelDown:
			if m.cursor < len(m.files)-1 {
				m.cursor++
			}
		}

	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("TFE - Terminal File Explorer"))
	s.WriteString("\n")

	// Current path
	s.WriteString(pathStyle.Render(m.currentPath))
	s.WriteString("\n")

	// File list
	maxVisible := m.height - 5 // Reserve space for title, path, and help

	// Calculate visible range (simple scrolling)
	start := 0
	end := len(m.files)

	if len(m.files) > maxVisible {
		start = m.cursor - maxVisible/2
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(m.files) {
			end = len(m.files)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		file := m.files[i]

		// Get icon based on file type
		icon := getFileIcon(file)
		style := fileStyle

		if file.isDir {
			style = folderStyle
		}

		// Override with orange color if it's a Claude context file
		if isClaudeContextFile(file.name) {
			style = claudeContextStyle
		}

		// Build the line
		line := fmt.Sprintf("  %s %s", icon, file.name)

		// Apply selection style
		if i == m.cursor {
			line = selectedStyle.Render(line)
		} else {
			line = style.Render(line)
		}

		s.WriteString(line)
		s.WriteString("\n")
	}

	// Status bar
	s.WriteString("\n")

	// Count directories and files
	dirCount, fileCount := 0, 0
	for _, f := range m.files {
		if f.name == ".." {
			continue
		}
		if f.isDir {
			dirCount++
		} else {
			fileCount++
		}
	}

	// Selected file info
	var selectedInfo string
	if len(m.files) > 0 && m.cursor < len(m.files) {
		selected := m.files[m.cursor]
		if selected.isDir {
			selectedInfo = fmt.Sprintf("Selected: %s (folder)", selected.name)
		} else {
			selectedInfo = fmt.Sprintf("Selected: %s (%s, %s)",
				selected.name,
				formatFileSize(selected.size),
				formatModTime(selected.modTime))
		}
	}

	itemsInfo := fmt.Sprintf("%d items", len(m.files))
	if dirCount > 0 || fileCount > 0 {
		itemsInfo = fmt.Sprintf("%d folders, %d files", dirCount, fileCount)
	}

	hiddenIndicator := ""
	if m.showHidden {
		hiddenIndicator = " • showing hidden"
	}

	statusText := fmt.Sprintf("%s%s | %s", itemsInfo, hiddenIndicator, selectedInfo)
	s.WriteString(statusStyle.Render(statusText))

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	s.WriteString(helpStyle.Render("↑/↓: navigate • enter: open • h/←: parent • .: toggle hidden • q: quit"))

	return s.String()
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
