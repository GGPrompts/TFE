package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
)

// Global theme flag (set before model initialization)
var forceLightTheme bool

func main() {
	// Handle command-line flags
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version", "-v":
			fmt.Printf("TFE (Terminal File Explorer) v%s\n", Version)
			os.Exit(0)
		case "--light":
			forceLightTheme = true
		case "--dark":
			forceLightTheme = false // Explicit dark mode (default)
		case "--help", "-h":
			fmt.Println("TFE (Terminal File Explorer)")
			fmt.Println()
			fmt.Println("Usage: tfe [options] [directory]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --light      Use light theme (for light terminal backgrounds)")
			fmt.Println("  --dark       Use dark theme (default)")
			fmt.Println("  --version    Show version information")
			fmt.Println("  --help       Show this help message")
			os.Exit(0)
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
