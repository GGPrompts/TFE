package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
)

type fileItem struct {
	name  string
	path  string
	isDir bool
}

type model struct {
	currentPath string
	files       []fileItem
	cursor      int
	height      int
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
		// Skip hidden files starting with .
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		item := fileItem{
			name:  entry.Name(),
			path:  filepath.Join(m.currentPath, entry.Name()),
			isDir: entry.IsDir(),
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

		// Choose icon based on file type
		icon := " "
		style := fileStyle

		if file.isDir {
			icon = " "
			style = folderStyle
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

	// Help text
	s.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).PaddingLeft(2)
	s.WriteString(helpStyle.Render("↑/↓: navigate • enter: open • h/←: parent • q: quit"))

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
