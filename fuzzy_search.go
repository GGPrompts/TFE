package main

// Module: fuzzy_search.go
// Purpose: Fuzzy file search functionality using external fzf + fd/find
// Responsibilities:
// - Detecting available file finding tools (fd, fdfind, find)
// - Launching external fzf with file list
// - Processing search results

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// getFileFinder returns the best available file finder command
// Preference: fd > fdfind > find
func getFileFinder() (string, []string) {
	// Try fd (modern, fast)
	if _, err := exec.LookPath("fd"); err == nil {
		return "fd", []string{
			"--type", "f",           // Files only
			"--hidden",              // Include hidden files
			"--follow",              // Follow symlinks
			"--exclude", ".git",     // Exclude .git directories
			"--exclude", "node_modules", // Exclude node_modules
			"--color", "never",      // No color codes
		}
	}

	// Try fdfind (Ubuntu's renamed fd)
	if _, err := exec.LookPath("fdfind"); err == nil {
		return "fdfind", []string{
			"--type", "f",
			"--hidden",
			"--follow",
			"--exclude", ".git",
			"--exclude", "node_modules",
			"--color", "never",
		}
	}

	// Fall back to find (always available, slower)
	return "find", []string{
		".",
		"-type", "f",
		"-not", "-path", "*/.git/*",
		"-not", "-path", "*/node_modules/*",
	}
}

// getSearchRoot determines the root directory for fuzzy search
// Priority: git root > home directory
func (m *model) getSearchRoot() (string, string) {
	// Try to find git repository root
	gitRoot := m.findGitRoot(m.currentPath)
	if gitRoot != "" {
		// Return git root with a friendly display name
		displayName := filepath.Base(gitRoot)
		return gitRoot, fmt.Sprintf("Git:%s> ", displayName)
	}

	// Fall back to home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// If can't get home, use current path (original behavior)
		return m.currentPath, "File> "
	}

	return homeDir, "~> "
}

// findGitRoot walks up the directory tree to find .git directory
func (m *model) findGitRoot(startPath string) string {
	currentPath := startPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return currentPath
		}

		// Move up one directory
		parentPath := filepath.Dir(currentPath)

		// Reached root without finding .git
		if parentPath == currentPath {
			return ""
		}

		currentPath = parentPath
	}
}

// launchFuzzySearch uses external fzf + fd/find for blazing fast search
func (m *model) launchFuzzySearch() tea.Cmd {
	return func() tea.Msg {
		// Check if fzf is installed
		_, err := exec.LookPath("fzf")
		if err != nil {
			return fuzzySearchResultMsg{
				selected: "",
				err:      fmt.Errorf("fzf not found. Install: sudo apt install fzf (Linux) or brew install fzf (macOS)"),
			}
		}

		// Get the best file finder
		finder, args := getFileFinder()

		// Determine search root (git root or home directory)
		searchRoot, promptText := m.getSearchRoot()

		// Build shell command that pipes file finder to fzf
		// This is simpler and more reliable than trying to coordinate stdin/tty
		var shellCmd string

		// Optimized fzf options for performance:
		// --no-preview: Disable preview by default (user can toggle with '?' if needed)
		// --no-mouse: Disable mouse for better performance
		// --cycle: Wrap around when reaching end
		// --bind '?:toggle-preview': Press '?' to show/hide preview
		fzfOpts := fmt.Sprintf("--height=100%% --layout=reverse --border --prompt='%s' --no-preview --no-mouse --cycle --bind '?:toggle-preview' --preview='head -50 {}' --preview-window=right:50%%:wrap:hidden", promptText)

		if finder == "find" {
			// find outputs relative paths like "./file.txt", clean them up
			shellCmd = fmt.Sprintf("cd %q && %s %s 2>/dev/null | sed 's|^\\./||' | fzf %s",
				searchRoot, finder, strings.Join(args, " "), fzfOpts)
		} else {
			// fd/fdfind already outputs clean relative paths
			shellCmd = fmt.Sprintf("cd %q && %s %s 2>/dev/null | fzf %s",
				searchRoot, finder, strings.Join(args, " "), fzfOpts)
		}

		// Run the shell command with proper TTY access
		cmd := exec.Command("sh", "-c", shellCmd)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		// Capture output
		output, err := cmd.Output()
		if err != nil {
			// User cancelled (exit code 130) or other error
			// This is normal, not an error
			return fuzzySearchResultMsg{
				selected: "",
				err:      nil,
			}
		}

		// Get the selected file path
		selectedFile := strings.TrimSpace(string(output))
		if selectedFile == "" {
			return fuzzySearchResultMsg{
				selected: "",
				err:      nil,
			}
		}

		// Convert relative path to absolute (using search root, not current path)
		resultSearchRoot, _ := m.getSearchRoot()
		absPath := filepath.Join(resultSearchRoot, selectedFile)

		return fuzzySearchResultMsg{
			selected: absPath,
			err:      nil,
		}
	}
}

// navigateToFuzzyResult navigates to the selected file from fuzzy search
func (m *model) navigateToFuzzyResult(selectedPath string) {
	if selectedPath == "" {
		return
	}

	// Get directory and filename
	dir := filepath.Dir(selectedPath)
	filename := filepath.Base(selectedPath)

	// If the file is in a different directory, navigate there
	if dir != m.currentPath {
		m.currentPath = dir
		m.loadFiles()
	}

	// Find the file in the current file list and move cursor to it
	for i, file := range m.files {
		if strings.EqualFold(file.name, filename) {
			m.cursor = i

			// Load preview if not a directory
			if !file.isDir && m.viewMode == viewDualPane {
				m.loadPreview(file.path)
				m.populatePreviewCache()
			}
			break
		}
	}
}
