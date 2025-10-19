package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGetDisplayPath tests path display with home directory replacement
func TestGetDisplayPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "Home directory",
			path:     homeDir,
			expected: "~",
		},
		{
			name:     "Subdirectory of home",
			path:     filepath.Join(homeDir, "Documents"),
			expected: "~/Documents",
		},
		{
			name:     "Nested subdirectory",
			path:     filepath.Join(homeDir, "Documents", "Projects", "test"),
			expected: "~/Documents/Projects/test",
		},
		{
			name:     "Root path",
			path:     "/",
			expected: "/",
		},
		{
			name:     "Absolute path outside home",
			path:     "/usr/local/bin",
			expected: "/usr/local/bin",
		},
		{
			name:     "Relative-like path",
			path:     "./test",
			expected: "./test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDisplayPath(tt.path)
			if result != tt.expected {
				t.Errorf("getDisplayPath(%s) = %s, expected %s", tt.path, result, tt.expected)
			}
		})
	}
}

// TestIsDualPaneCompatible tests dual-pane view compatibility
func TestIsDualPaneCompatible(t *testing.T) {
	tests := []struct {
		name        string
		displayMode displayMode
		expected    bool
	}{
		{
			name:        "List mode (compatible)",
			displayMode: modeList,
			expected:    true,
		},
		{
			name:        "Tree mode (compatible)",
			displayMode: modeTree,
			expected:    true,
		},
		{
			name:        "Detail mode (incompatible)",
			displayMode: modeDetail,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{displayMode: tt.displayMode}
			result := m.isDualPaneCompatible()
			if result != tt.expected {
				t.Errorf("isDualPaneCompatible() for %s = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestIsPromptFile tests prompt file detection
func TestIsPromptFile(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	tests := []struct {
		name     string
		fileItem fileItem
		expected bool
	}{
		{
			name: "Directory (not a prompt)",
			fileItem: fileItem{
				name:  "test",
				path:  "/path/to/test",
				isDir: true,
			},
			expected: false,
		},
		{
			name: ".prompty file (always a prompt)",
			fileItem: fileItem{
				name:  "template.prompty",
				path:  "/anywhere/template.prompty",
				isDir: false,
			},
			expected: true,
		},
		{
			name: ".prompty in random location",
			fileItem: fileItem{
				name:  "test.prompty",
				path:  "/random/path/test.prompty",
				isDir: false,
			},
			expected: true,
		},
		{
			name: ".md in .claude directory",
			fileItem: fileItem{
				name:  "command.md",
				path:  "/project/.claude/commands/command.md",
				isDir: false,
			},
			expected: true,
		},
		{
			name: ".md in ~/.prompts directory",
			fileItem: fileItem{
				name:  "prompt.md",
				path:  filepath.Join(homeDir, ".prompts", "prompt.md"),
				isDir: false,
			},
			expected: true,
		},
		{
			name: ".yaml in .claude directory",
			fileItem: fileItem{
				name:  "config.yaml",
				path:  "/project/.claude/agents/config.yaml",
				isDir: false,
			},
			expected: true,
		},
		{
			name: ".yml in ~/.prompts",
			fileItem: fileItem{
				name:  "template.yml",
				path:  filepath.Join(homeDir, ".prompts", "library", "template.yml"),
				isDir: false,
			},
			expected: true,
		},
		{
			name: ".txt in .claude",
			fileItem: fileItem{
				name:  "notes.txt",
				path:  "/project/.claude/notes.txt",
				isDir: false,
			},
			expected: true,
		},
		{
			name: ".md outside special directories",
			fileItem: fileItem{
				name:  "README.md",
				path:  "/project/README.md",
				isDir: false,
			},
			expected: false,
		},
		{
			name: ".yaml outside special directories",
			fileItem: fileItem{
				name:  "config.yaml",
				path:  "/project/config.yaml",
				isDir: false,
			},
			expected: false,
		},
		{
			name: ".txt outside special directories",
			fileItem: fileItem{
				name:  "notes.txt",
				path:  "/project/notes.txt",
				isDir: false,
			},
			expected: false,
		},
		{
			name: ".go file (not a prompt extension)",
			fileItem: fileItem{
				name:  "main.go",
				path:  "/project/.claude/main.go",
				isDir: false,
			},
			expected: false,
		},
		{
			name: ".js file in .claude",
			fileItem: fileItem{
				name:  "script.js",
				path:  "/project/.claude/script.js",
				isDir: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPromptFile(tt.fileItem)
			if result != tt.expected {
				t.Errorf("isPromptFile(%s at %s) = %v, expected %v",
					tt.fileItem.name, tt.fileItem.path, result, tt.expected)
			}
		})
	}
}

// TestGetCurrentFile tests getting the current file from model
func TestGetCurrentFile(t *testing.T) {
	tests := []struct {
		name        string
		model       model
		expected    *fileItem
		expectNil   bool
	}{
		{
			name: "Empty file list",
			model: model{
				files:  []fileItem{},
				cursor: 0,
			},
			expectNil: true,
		},
		{
			name: "Negative cursor",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
				},
				cursor: -1,
			},
			expectNil: true,
		},
		{
			name: "Valid cursor in list mode",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
					{name: "file2.txt", path: "/path/file2.txt"},
					{name: "file3.txt", path: "/path/file3.txt"},
				},
				cursor:      1,
				displayMode: modeList,
			},
			expected: &fileItem{name: "file2.txt", path: "/path/file2.txt"},
		},
		{
			name: "Cursor at first position",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
					{name: "file2.txt", path: "/path/file2.txt"},
				},
				cursor:      0,
				displayMode: modeList,
			},
			expected: &fileItem{name: "file1.txt", path: "/path/file1.txt"},
		},
		{
			name: "Cursor at last position",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
					{name: "file2.txt", path: "/path/file2.txt"},
					{name: "file3.txt", path: "/path/file3.txt"},
				},
				cursor:      2,
				displayMode: modeList,
			},
			expected: &fileItem{name: "file3.txt", path: "/path/file3.txt"},
		},
		{
			name: "Cursor out of bounds",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
				},
				cursor:      5,
				displayMode: modeList,
			},
			expectNil: true,
		},
		{
			name: "Detail mode",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
					{name: "file2.txt", path: "/path/file2.txt"},
				},
				cursor:      0,
				displayMode: modeDetail,
			},
			expected: &fileItem{name: "file1.txt", path: "/path/file1.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.getCurrentFile()

			if tt.expectNil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			if result.name != tt.expected.name || result.path != tt.expected.path {
				t.Errorf("getCurrentFile() = {name: %s, path: %s}, expected {name: %s, path: %s}",
					result.name, result.path, tt.expected.name, tt.expected.path)
			}
		})
	}
}

// TestGetMaxCursor tests maximum cursor position calculation
func TestGetMaxCursor(t *testing.T) {
	tests := []struct {
		name        string
		model       model
		expected    int
	}{
		{
			name: "Empty file list",
			model: model{
				files:       []fileItem{},
				displayMode: modeList,
			},
			expected: -1,
		},
		{
			name: "Single file",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
				},
				displayMode: modeList,
			},
			expected: 0,
		},
		{
			name: "Three files",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
					{name: "file2.txt", path: "/path/file2.txt"},
					{name: "file3.txt", path: "/path/file3.txt"},
				},
				displayMode: modeList,
			},
			expected: 2,
		},
		{
			name: "Detail mode with files",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
					{name: "file2.txt", path: "/path/file2.txt"},
				},
				displayMode: modeDetail,
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.getMaxCursor()
			if result != tt.expected {
				t.Errorf("getMaxCursor() = %d, expected %d", result, tt.expected)
			}
		})
	}
}

// TestGetFilteredFiles tests file filtering functionality
func TestGetFilteredFiles(t *testing.T) {
	tests := []struct {
		name           string
		model          model
		expectedCount  int
	}{
		{
			name: "No filter",
			model: model{
				files: []fileItem{
					{name: "file1.txt", path: "/path/file1.txt"},
					{name: "file2.txt", path: "/path/file2.txt"},
					{name: "dir1", path: "/path/dir1", isDir: true},
				},
				showFavoritesOnly: false,
				showPromptsOnly:   false,
			},
			expectedCount: 3,
		},
		{
			name: "All files no filter",
			model: model{
				files: []fileItem{
					{name: "test.go", path: "/path/test.go"},
					{name: "main.go", path: "/path/main.go"},
					{name: "README.md", path: "/path/README.md"},
				},
				showFavoritesOnly: false,
				showPromptsOnly:   false,
			},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.model.getFilteredFiles()
			if len(result) != tt.expectedCount {
				t.Errorf("getFilteredFiles() returned %d files, expected %d", len(result), tt.expectedCount)
			}
		})
	}
}

// BenchmarkGetCurrentFile benchmarks getCurrentFile performance
func BenchmarkGetCurrentFile(b *testing.B) {
	// Create a model with 100 files
	files := make([]fileItem, 100)
	for i := 0; i < 100; i++ {
		files[i] = fileItem{
			name:  "file" + string(rune('0'+i%10)) + ".txt",
			path:  "/path/file" + string(rune('0'+i%10)) + ".txt",
			isDir: false,
		}
	}

	m := model{
		files:       files,
		cursor:      50,
		displayMode: modeList,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.getCurrentFile()
	}
}

// BenchmarkGetMaxCursor benchmarks getMaxCursor performance
func BenchmarkGetMaxCursor(b *testing.B) {
	files := make([]fileItem, 1000)
	for i := 0; i < 1000; i++ {
		files[i] = fileItem{
			name:  "file.txt",
			path:  "/path/file.txt",
			isDir: false,
		}
	}

	m := model{
		files:       files,
		displayMode: modeList,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.getMaxCursor()
	}
}

// BenchmarkIsPromptFile benchmarks prompt file detection
func BenchmarkIsPromptFile(b *testing.B) {
	testItems := []fileItem{
		{name: "template.prompty", path: "/path/template.prompty", isDir: false},
		{name: "README.md", path: "/project/README.md", isDir: false},
		{name: "command.md", path: "/project/.claude/commands/command.md", isDir: false},
		{name: "config.yaml", path: "/project/config.yaml", isDir: false},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isPromptFile(testItems[i%len(testItems)])
	}
}
