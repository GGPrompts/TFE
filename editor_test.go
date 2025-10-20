package main

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

// TestEditorAvailable tests editor detection
func TestEditorAvailable(t *testing.T) {
	tests := []struct {
		name   string
		editor string
	}{
		{"nano", "nano"},
		{"vim", "vim"},
		{"vi", "vi"},
		{"micro", "micro"},
		{"nonexistent", "nonexistent_editor_12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := editorAvailable(tt.editor)
			// Check that result matches actual availability
			_, err := exec.LookPath(tt.editor)
			expected := err == nil

			if result != expected {
				t.Errorf("editorAvailable(%s) = %v, but LookPath says %v",
					tt.editor, result, expected)
			}
		})
	}
}

// TestGetAvailableEditor tests editor priority selection
func TestGetAvailableEditor(t *testing.T) {
	editor := getAvailableEditor()

	// Should return non-empty string (one of micro, nano, vim, vi should exist)
	// OR empty string if none available
	if editor != "" {
		// If we got an editor, verify it's actually available
		if !editorAvailable(editor) {
			t.Errorf("getAvailableEditor() returned %s but it's not available", editor)
		}

		// Verify it's one of the expected editors
		validEditors := []string{"micro", "nano", "vim", "vi"}
		found := false
		for _, valid := range validEditors {
			if editor == valid {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("getAvailableEditor() returned unexpected editor: %s", editor)
		}
	}
}

// TestIsImageFile tests image file extension detection
func TestIsImageFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"photo.png", true},
		{"image.jpg", true},
		{"pic.jpeg", true},
		{"graphic.gif", true},
		{"icon.svg", true},
		{"banner.webp", true},
		{"image.bmp", true},
		{"picture.PNG", true}, // Uppercase
		{"document.pdf", false},
		{"file.txt", false},
		{"main.go", false},
		{"data.json", false},
		{"", false},
		{"noextension", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isImageFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isImageFile(%s) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestIsHTMLFile tests HTML file extension detection
func TestIsHTMLFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"index.html", true},
		{"page.htm", true},
		{"INDEX.HTML", true}, // Uppercase
		{"Page.HTM", true},
		{"index.html.bak", false},
		{"main.go", false},
		{"file.txt", false},
		{"", false},
		{"noextension", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isHTMLFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isHTMLFile(%s) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestIsBrowserFile tests combined browser-openable file detection
func TestIsBrowserFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		// Images
		{"photo.png", true},
		{"image.jpg", true},
		{"graphic.svg", true},
		// HTML
		{"index.html", true},
		{"page.htm", true},
		// Non-browser files
		{"document.pdf", false},
		{"main.go", false},
		{"data.json", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := isBrowserFile(tt.filename)
			if result != tt.expected {
				t.Errorf("isBrowserFile(%s) = %v, expected %v", tt.filename, result, tt.expected)
			}
		})
	}
}

// TestIsWSL tests WSL environment detection
func TestIsWSL(t *testing.T) {
	// This test just ensures the function runs without panic
	// Actual result depends on environment
	result := isWSL()

	// On Linux, check if /proc/version mentions Microsoft (WSL indicator)
	if runtime.GOOS == "linux" {
		content, err := os.ReadFile("/proc/version")
		if err == nil {
			hasMicrosoft := len(content) > 0 &&
				(containsIgnoreCase(string(content), "microsoft") ||
				 containsIgnoreCase(string(content), "WSL"))
			if result != hasMicrosoft {
				t.Logf("isWSL() = %v, but /proc/version check says %v (non-critical)", result, hasMicrosoft)
			}
		}
	} else {
		// Non-Linux should return false
		if result {
			t.Error("isWSL() should return false on non-Linux systems")
		}
	}
}

// Helper function for case-insensitive contains check
func containsIgnoreCase(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// TestGetAvailableBrowser tests browser detection
func TestGetAvailableBrowser(t *testing.T) {
	browser := getAvailableBrowser()

	// Should return a browser command appropriate for the platform
	// OR empty string if none available
	if browser != "" {
		t.Logf("Detected browser: %s", browser)

		// Verify it's one of the expected browsers
		validBrowsers := []string{
			"wslview",  // WSL
			"cmd.exe",  // Windows
			"xdg-open", // Linux
			"open",     // macOS
		}

		found := false
		for _, valid := range validBrowsers {
			if browser == valid {
				found = true
				break
			}
		}

		if !found {
			t.Logf("Unexpected browser: %s (may be valid on this system)", browser)
		}
	}
}

// TestGetAvailableImageViewer tests image viewer detection
func TestGetAvailableImageViewer(t *testing.T) {
	viewer := getAvailableImageViewer()

	// Should return viu, timg, chafa, or empty string
	if viewer != "" {
		t.Logf("Detected image viewer: %s", viewer)

		validViewers := []string{"viu", "timg", "chafa"}
		found := false
		for _, valid := range validViewers {
			if viewer == valid {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("getAvailableImageViewer() returned unexpected viewer: %s", viewer)
		}

		// Verify it's actually available
		_, err := exec.LookPath(viewer)
		if err != nil {
			t.Errorf("getAvailableImageViewer() returned %s but it's not in PATH", viewer)
		}
	}
}

// TestGetAvailableImageEditor tests image editor detection
func TestGetAvailableImageEditor(t *testing.T) {
	editor := getAvailableImageEditor()

	// Should return textual-paint or empty string
	if editor != "" {
		t.Logf("Detected image editor: %s", editor)

		if editor != "textual-paint" {
			t.Errorf("getAvailableImageEditor() returned unexpected editor: %s", editor)
		}

		// Verify it's actually available
		_, err := exec.LookPath(editor)
		if err != nil {
			t.Errorf("getAvailableImageEditor() returned %s but it's not in PATH", editor)
		}
	}
}

// TestCopyToClipboard tests clipboard integration
func TestCopyToClipboard(t *testing.T) {
	testText := "test clipboard content"

	// This test ensures the function runs without panic
	err := copyToClipboard(testText)

	// The function returns an error if no clipboard tool is available
	// OR nil if clipboard copy was attempted
	if err != nil {
		t.Logf("Clipboard not available (expected on systems without clipboard support): %v", err)
	} else {
		t.Log("Clipboard operation attempted (system supports clipboard)")
	}
}

// TestOpenEditor tests editor command creation
func TestOpenEditor(t *testing.T) {
	// Create a temporary file to edit
	tmpFile, err := os.CreateTemp("", "tfe_test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Get available editor
	editor := getAvailableEditor()
	if editor == "" {
		t.Skip("No editor available on this system")
	}

	// Test opening the editor
	cmd := openEditor(editor, tmpFile.Name())

	if cmd == nil {
		t.Error("openEditor() returned nil command")
		return
	}

	// tea.Cmd is just a function, so we can't easily inspect it
	// We just verify it doesn't panic
	t.Log("openEditor() returned a valid tea.Cmd")
}

// TestOpenTUITool tests TUI tool command creation
func TestOpenTUITool(t *testing.T) {
	tests := []struct {
		name string
		tool string
	}{
		{"lazygit", "lazygit"},
		{"htop", "htop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test with current directory
			cmd := openTUITool(tt.tool, ".")

			// Check if tool is actually available
			_, err := exec.LookPath(tt.tool)
			toolAvailable := err == nil

			if toolAvailable {
				if cmd == nil {
					t.Error("openTUITool() returned nil for available tool")
				}
			} else {
				t.Logf("Tool %s not available on this system", tt.tool)
			}
		})
	}
}

// TestOpenInBrowser tests browser opening command creation
func TestOpenInBrowser(t *testing.T) {
	// Create a temporary HTML file
	tmpFile, err := os.CreateTemp("", "tfe_test_*.html")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("<html><body>Test</body></html>")
	tmpFile.Close()

	// Get available browser
	browser := getAvailableBrowser()
	if browser == "" {
		t.Skip("No browser available on this system")
	}

	// Test opening in browser
	cmd := openInBrowser(tmpFile.Name())

	if cmd == nil {
		t.Error("openInBrowser() returned nil command")
		return
	}

	// tea.Cmd is just a function, so we can't easily inspect it
	// We just verify it doesn't panic
	t.Log("openInBrowser() returned a valid tea.Cmd")
}

// TestOpenImageViewer tests image viewer command creation
func TestOpenImageViewer(t *testing.T) {
	// Create a fake image file
	tmpFile, err := os.CreateTemp("", "tfe_test_*.png")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Get available viewer
	viewer := getAvailableImageViewer()
	if viewer == "" {
		t.Skip("No image viewer available on this system")
	}

	// Test opening image viewer
	cmd := openImageViewer(tmpFile.Name())

	if cmd == nil {
		t.Error("openImageViewer() returned nil command")
		return
	}

	// tea.Cmd is just a function, so we can't easily inspect it
	// We just verify it doesn't panic
	t.Log("openImageViewer() returned a valid tea.Cmd")
}
