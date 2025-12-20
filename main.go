package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

// Global flags (set before model initialization)
var (
	forceLightTheme bool
	startPath       string // Directory to open
	selectFile      string // File to select (basename)
	autoPreview     bool   // Auto-open preview pane
)

func main() {
	// Handle command-line flags
	for _, arg := range os.Args[1:] {
		switch {
		case arg == "--version" || arg == "-v":
			fmt.Printf("TFE (Terminal File Explorer) v%s\n", Version)
			os.Exit(0)
		case arg == "--light":
			forceLightTheme = true
		case arg == "--dark":
			forceLightTheme = false // Explicit dark mode (default)
		case arg == "--preview" || arg == "-p":
			autoPreview = true
		case arg == "--help" || arg == "-h":
			fmt.Println("TFE (Terminal File Explorer)")
			fmt.Println()
			fmt.Println("Usage: tfe [options] [path]")
			fmt.Println()
			fmt.Println("Arguments:")
			fmt.Println("  path         Directory to open, or file to select")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --preview    Auto-open preview pane (useful with file path)")
			fmt.Println("  --light      Use light theme (for light terminal backgrounds)")
			fmt.Println("  --dark       Use dark theme (default)")
			fmt.Println("  --version    Show version information")
			fmt.Println("  --help       Show this help message")
			fmt.Println()
			fmt.Println("Examples:")
			fmt.Println("  tfe                        Open current directory")
			fmt.Println("  tfe ~/projects             Open ~/projects directory")
			fmt.Println("  tfe ~/projects/main.go     Open ~/projects with main.go selected")
			fmt.Println("  tfe --preview src/app.ts   Open with app.ts selected and previewed")
			os.Exit(0)
		case !strings.HasPrefix(arg, "-"):
			// Non-flag argument is the path
			targetPath := arg

			// Expand ~ to home directory
			if strings.HasPrefix(targetPath, "~") {
				home, err := os.UserHomeDir()
				if err == nil {
					targetPath = filepath.Join(home, targetPath[1:])
				}
			}

			// Make absolute
			absPath, err := filepath.Abs(targetPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: invalid path %q: %v\n", arg, err)
				os.Exit(1)
			}

			// Check if path exists
			info, err := os.Stat(absPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: path not found %q: %v\n", arg, err)
				os.Exit(1)
			}

			if info.IsDir() {
				// It's a directory - open it directly
				startPath = absPath
			} else {
				// It's a file - open parent dir and select the file
				startPath = filepath.Dir(absPath)
				selectFile = filepath.Base(absPath)
			}
		}
	}

	// Ensure terminal cleanup on exit (defer runs even if panic/interrupt)
	defer cleanupTerminal()

	// Set up signal catching to handle Ctrl+C gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Handle signals in a goroutine
	go func() {
		<-sigChan
		p.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

// cleanupTerminal resets terminal state to prevent formatting bleed
func cleanupTerminal() {
	// Exit alt screen (in case Bubbletea didn't clean up)
	fmt.Print("\033[?1049l")

	// Disable mouse tracking (in case it was left on)
	fmt.Print("\033[?1000l") // Disable X10 mouse
	fmt.Print("\033[?1002l") // Disable cell motion mouse tracking
	fmt.Print("\033[?1003l") // Disable all motion mouse tracking
	fmt.Print("\033[?1006l") // Disable SGR mouse mode

	// Reset all ANSI formatting
	fmt.Print("\033[0m")

	// Show cursor (in case it was hidden)
	fmt.Print("\033[?25h")

	// Reset scrolling region
	fmt.Print("\033[r")

	// Clear from cursor to end of screen (clean up any leftover artifacts)
	fmt.Print("\033[J")

	// Move cursor to start of line
	fmt.Print("\r")
}
