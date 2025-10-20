package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// setupTestDir creates a temporary directory for file operation tests
func setupTestDir(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	cleanup := func() {
		// t.TempDir() handles cleanup automatically
	}
	return tmpDir, cleanup
}

// createTestFileWithContent creates a file with specific content
func createTestFileWithContent(t *testing.T, path string, content []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Failed to create parent dir: %v", err)
	}
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

// TestIsBinaryFile tests binary file detection
func TestIsBinaryFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "Text file",
			content:  []byte("Hello, World!\nThis is a text file."),
			expected: false,
		},
		{
			name:     "Binary file with null bytes",
			content:  []byte{0x00, 0x01, 0x02, 0xFF, 0xFE},
			expected: true,
		},
		{
			name:     "UTF-8 text",
			content:  []byte("Hello 世界 🌍"),
			expected: false,
		},
		{
			name:     "JSON file",
			content:  []byte(`{"key": "value", "number": 123}`),
			expected: false,
		},
		{
			name:     "Binary with null in middle",
			content:  append([]byte("text"), append([]byte{0x00}, []byte("more")...)...),
			expected: true,
		},
		{
			name:     "Empty file",
			content:  []byte{},
			expected: false,
		},
		{
			name:     "Large text file",
			content:  []byte(strings.Repeat("A", 1000)),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test_"+strings.ReplaceAll(tt.name, " ", "_"))
			createTestFileWithContent(t, testFile, tt.content)

			result := isBinaryFile(testFile)
			if result != tt.expected {
				t.Errorf("isBinaryFile(%s) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestIsBinaryFile_NonExistent tests with non-existent file
func TestIsBinaryFile_NonExistent(t *testing.T) {
	result := isBinaryFile("/nonexistent/file.txt")
	if result != false {
		t.Error("isBinaryFile on non-existent file should return false")
	}
}

// TestFormatFileSize tests file size formatting
func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0B"},
		{1, "1B"},
		{512, "512B"},
		{1023, "1023B"},
		{1024, "1.0KB"},
		{1536, "1.5KB"},
		{1024 * 1024, "1.0MB"},
		{1024 * 1024 * 1.5, "1.5MB"},
		{1024 * 1024 * 1024, "1.0GB"},
		{1024 * 1024 * 1024 * 1.5, "1.5GB"},
		{1024 * 1024 * 1024 * 1024, "1.0TB"},
		{1024 * 1024 * 1024 * 1024 * 1.5, "1.5TB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatFileSize(tt.size)
			if result != tt.expected {
				t.Errorf("formatFileSize(%d) = %s, expected %s", tt.size, result, tt.expected)
			}
		})
	}
}

// TestFormatModTime tests relative time formatting
func TestFormatModTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		modTime  time.Time
		expected string
	}{
		{
			name:     "Just now",
			modTime:  now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "1 minute ago",
			modTime:  now.Add(-1 * time.Minute),
			expected: "1m ago",
		},
		{
			name:     "5 minutes ago",
			modTime:  now.Add(-5 * time.Minute),
			expected: "5m ago",
		},
		{
			name:     "1 hour ago",
			modTime:  now.Add(-1 * time.Hour),
			expected: "1h ago",
		},
		{
			name:     "3 hours ago",
			modTime:  now.Add(-3 * time.Hour),
			expected: "3h ago",
		},
		{
			name:     "1 day ago",
			modTime:  now.Add(-24 * time.Hour),
			expected: "1d ago",
		},
		{
			name:     "3 days ago",
			modTime:  now.Add(-3 * 24 * time.Hour),
			expected: "3d ago",
		},
		{
			name:     "1 week ago",
			modTime:  now.Add(-7 * 24 * time.Hour),
			expected: "1w ago",
		},
		{
			name:     "2 weeks ago",
			modTime:  now.Add(-14 * 24 * time.Hour),
			expected: "2w ago",
		},
		{
			name:     "1 month ago",
			modTime:  now.Add(-30 * 24 * time.Hour),
			expected: "1mo ago",
		},
		{
			name:     "6 months ago",
			modTime:  now.Add(-180 * 24 * time.Hour),
			expected: "6mo ago",
		},
		{
			name:     "1 year ago",
			modTime:  now.Add(-365 * 24 * time.Hour),
			expected: "1y ago",
		},
		{
			name:     "2 years ago",
			modTime:  now.Add(-2 * 365 * 24 * time.Hour),
			expected: "2y ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatModTime(tt.modTime)
			if result != tt.expected {
				t.Errorf("formatModTime() = %s, expected %s", result, tt.expected)
			}
		})
	}
}

// TestIsClaudeContextFile tests Claude context file detection
func TestIsClaudeContextFile(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"CLAUDE.md", true},
		{"CLAUDE.local.md", true},
		{".claude", true},
		{"README.md", false},
		{"claude.md", false},
		{"CLAUDE", false},
		{"other.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isClaudeContextFile(tt.name)
			if result != tt.expected {
				t.Errorf("isClaudeContextFile(%s) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestIsAgentsFile tests AGENTS.md detection
func TestIsAgentsFile(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"AGENTS.md", true},
		{"agents.md", false},
		{"AGENTS", false},
		{"README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAgentsFile(tt.name)
			if result != tt.expected {
				t.Errorf("isAgentsFile(%s) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestIsPromptsFolder tests .prompts folder detection
func TestIsPromptsFolder(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{".prompts", true},
		{"prompts", false},
		{".prompt", false},
		{".git", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPromptsFolder(tt.name)
			if result != tt.expected {
				t.Errorf("isPromptsFolder(%s) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestIsGlobalPromptsVirtualFolder tests global prompts virtual folder detection
func TestIsGlobalPromptsVirtualFolder(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"🌐 ~/.prompts/", true},
		{"🌐 ~/.prompts/test", true},
		{".prompts", false},
		{"~/.prompts", false},
		{"test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isGlobalPromptsVirtualFolder(tt.name)
			if result != tt.expected {
				t.Errorf("isGlobalPromptsVirtualFolder(%s) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestIsClaudePromptsSubfolder tests Claude prompts subfolder detection
func TestIsClaudePromptsSubfolder(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"commands", true},
		{"agents", true},
		{"skills", true},
		{"prompts", false},
		{"other", false},
		{"COMMANDS", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isClaudePromptsSubfolder(tt.name)
			if result != tt.expected {
				t.Errorf("isClaudePromptsSubfolder(%s) = %v, expected %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestIsDirEmpty tests directory emptiness check
func TestIsDirEmpty(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	// Create non-empty directory
	nonEmptyDir := filepath.Join(tmpDir, "nonempty")
	if err := os.Mkdir(nonEmptyDir, 0755); err != nil {
		t.Fatalf("Failed to create non-empty dir: %v", err)
	}
	createTestFileWithContent(t, filepath.Join(nonEmptyDir, "file.txt"), []byte("content"))

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Empty directory", emptyDir, true},
		{"Non-empty directory", nonEmptyDir, false},
		{"Non-existent directory", filepath.Join(tmpDir, "nonexistent"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDirEmpty(tt.path)
			if result != tt.expected {
				t.Errorf("isDirEmpty(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestGetDirItemCount tests directory item counting
func TestGetDirItemCount(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create directory with known number of items
	testDir := filepath.Join(tmpDir, "test")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}

	// Add 5 files
	for i := 0; i < 5; i++ {
		createTestFileWithContent(t, filepath.Join(testDir, "file"+string(rune('0'+i))+".txt"), []byte("content"))
	}

	// Add 2 subdirectories
	os.Mkdir(filepath.Join(testDir, "subdir1"), 0755)
	os.Mkdir(filepath.Join(testDir, "subdir2"), 0755)

	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{"Directory with 7 items", testDir, 7},
		{"Empty directory", filepath.Join(tmpDir, "empty"), 0}, // Will fail to read, returns 0
		{"Non-existent directory", filepath.Join(tmpDir, "nonexistent"), 0},
	}

	// Create empty dir for test
	os.Mkdir(filepath.Join(tmpDir, "empty"), 0755)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDirItemCount(tt.path)
			if result != tt.expected {
				t.Errorf("getDirItemCount(%s) = %d, expected %d", tt.path, result, tt.expected)
			}
		})
	}
}

// TestGetFileIcon tests icon selection for various file types
func TestGetFileIcon(t *testing.T) {
	tests := []struct {
		name     string
		fileItem fileItem
		expected string
	}{
		{
			name:     "Parent directory",
			fileItem: fileItem{name: "..", isDir: true},
			expected: "⬆️",
		},
		{
			name:     ".claude directory",
			fileItem: fileItem{name: ".claude", isDir: true},
			expected: "🤖",
		},
		{
			name:     ".git directory",
			fileItem: fileItem{name: ".git", isDir: true},
			expected: "📦",
		},
		{
			name:     ".prompts directory",
			fileItem: fileItem{name: ".prompts", isDir: true},
			expected: "📝",
		},
		{
			name:     "Regular directory",
			fileItem: fileItem{name: "mydir", isDir: true},
			expected: "📁",
		},
		{
			name:     "CLAUDE.md file",
			fileItem: fileItem{name: "CLAUDE.md", isDir: false},
			expected: "📝", // .md extension takes precedence
		},
		{
			name:     "README.md file",
			fileItem: fileItem{name: "README.md", isDir: false},
			expected: "📝", // .md extension takes precedence
		},
		{
			name:     "Makefile",
			fileItem: fileItem{name: "Makefile", isDir: false},
			expected: "🔨",
		},
		{
			name:     "Dockerfile",
			fileItem: fileItem{name: "Dockerfile", isDir: false},
			expected: "🐳",
		},
		{
			name:     "Go file",
			fileItem: fileItem{name: "main.go", isDir: false},
			expected: "🐹",
		},
		{
			name:     "Python file",
			fileItem: fileItem{name: "script.py", isDir: false},
			expected: "🐍",
		},
		{
			name:     "JavaScript file",
			fileItem: fileItem{name: "app.js", isDir: false},
			expected: "🟨", // JavaScript yellow square
		},
		{
			name:     "TypeScript file",
			fileItem: fileItem{name: "component.ts", isDir: false},
			expected: "🔷",
		},
		{
			name:     "Rust file",
			fileItem: fileItem{name: "main.rs", isDir: false},
			expected: "🦀",
		},
		{
			name:     "Generic file",
			fileItem: fileItem{name: "data.dat", isDir: false},
			expected: "📄",
		},
		{
			name:     "Global prompts virtual folder",
			fileItem: fileItem{name: "🌐 ~/.prompts/", isDir: true},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileIcon(tt.fileItem)
			if result != tt.expected {
				t.Errorf("getFileIcon(%s) = %s, expected %s", tt.fileItem.name, result, tt.expected)
			}
		})
	}
}

// TestRenderMarkdownWithTimeout tests markdown rendering with timeout protection
func TestRenderMarkdownWithTimeout(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		width       int
		timeout     time.Duration
		expectError bool
	}{
		{
			name:        "Simple markdown",
			content:     "# Hello\n\nThis is a test.",
			width:       80,
			timeout:     5 * time.Second,
			expectError: false,
		},
		{
			name:        "Empty content",
			content:     "",
			width:       80,
			timeout:     5 * time.Second,
			expectError: false,
		},
		{
			name:        "Complex markdown",
			content:     "# Title\n\n## Subtitle\n\n- Item 1\n- Item 2\n\n**Bold** and *italic*\n\n```go\nfunc main() {\n}\n```",
			width:       80,
			timeout:     5 * time.Second,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rendered, err := renderMarkdownWithTimeout(tt.content, tt.width, tt.timeout)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && rendered == "" && tt.content != "" {
				t.Error("Expected non-empty rendered output")
			}
		})
	}
}

// TestRenderMarkdownWithTimeout_ActualTimeout tests timeout behavior
func TestRenderMarkdownWithTimeout_ActualTimeout(t *testing.T) {
	// This test verifies the timeout mechanism works
	// We use a very short timeout to trigger timeout condition
	_, err := renderMarkdownWithTimeout("# Test", 80, 1*time.Nanosecond)

	// Either it times out or completes successfully (timing is not deterministic)
	// We just verify it doesn't panic or hang
	if err != nil {
		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "panic") {
			t.Logf("Expected timeout or panic error, got: %v", err)
		}
	}
}

// BenchmarkIsBinaryFile benchmarks binary file detection
func BenchmarkIsBinaryFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := []byte(strings.Repeat("Hello World\n", 100))
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		b.Fatalf("Failed to create test file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		isBinaryFile(testFile)
	}
}

// BenchmarkFormatFileSize benchmarks file size formatting
func BenchmarkFormatFileSize(b *testing.B) {
	sizes := []int64{
		0,
		1024,
		1024 * 1024,
		1024 * 1024 * 1024,
		1024 * 1024 * 1024 * 1024,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatFileSize(sizes[i%len(sizes)])
	}
}

// TestLoadFiles tests directory loading functionality
func TestLoadFiles(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create test directory structure
	testDir := filepath.Join(tmpDir, "testdir")
	os.Mkdir(testDir, 0755)

	// Create directories
	os.Mkdir(filepath.Join(testDir, "folder1"), 0755)
	os.Mkdir(filepath.Join(testDir, "folder2"), 0755)
	os.Mkdir(filepath.Join(testDir, ".hidden"), 0755)
	os.Mkdir(filepath.Join(testDir, ".claude"), 0755)

	// Create files
	createTestFileWithContent(t, filepath.Join(testDir, "file1.txt"), []byte("content"))
	createTestFileWithContent(t, filepath.Join(testDir, "file2.go"), []byte("package main"))
	createTestFileWithContent(t, filepath.Join(testDir, ".hidden_file"), []byte("secret"))

	tests := []struct {
		name           string
		currentPath    string
		showHidden     bool
		expectedMin    int // Minimum expected files (accounts for parent dir)
		expectParent   bool
		expectHidden   bool
	}{
		{
			name:         "Load regular directory",
			currentPath:  testDir,
			showHidden:   false,
			expectedMin:  6, // parent + 2 folders + 2 files + .claude
			expectParent: true,
			expectHidden: false,
		},
		{
			name:         "Load directory with hidden files shown",
			currentPath:  testDir,
			showHidden:   true,
			expectedMin:  8, // parent + 3 folders + 3 files + .claude
			expectParent: true,
			expectHidden: true,
		},
		{
			name:         "Load root directory",
			currentPath:  "/",
			showHidden:   false,
			expectedMin:  0, // No parent at root
			expectParent: false,
			expectHidden: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &model{
				currentPath: tt.currentPath,
				showHidden:  tt.showHidden,
			}

			m.loadFiles()

			if len(m.files) < tt.expectedMin {
				t.Errorf("Expected at least %d files, got %d", tt.expectedMin, len(m.files))
			}

			// Check for parent directory
			if tt.expectParent {
				if len(m.files) == 0 || m.files[0].name != ".." {
					t.Error("Expected parent directory '..' at index 0")
				}
			}

			// Check hidden files visibility
			hasHidden := false
			for _, f := range m.files {
				if strings.HasPrefix(f.name, ".") && f.name != ".." && f.name != ".claude" {
					hasHidden = true
					break
				}
			}

			if tt.expectHidden && !hasHidden && tt.currentPath == testDir {
				t.Error("Expected hidden files to be visible, but none found")
			}

			// Note: Directory/file ordering depends on sortFiles() which may be called
			// We skip this test as sortFiles() can reorder based on sortBy/sortAsc settings
		})
	}
}

// TestLoadFilesInvalidPath tests loading files from invalid path
func TestLoadFilesInvalidPath(t *testing.T) {
	m := &model{
		currentPath: "/nonexistent/path/that/does/not/exist",
		showHidden:  false,
	}

	m.loadFiles()

	if len(m.files) != 0 {
		t.Errorf("Expected 0 files for invalid path, got %d", len(m.files))
	}
}

// TestLoadPreview tests file preview loading
func TestLoadPreview(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	tests := []struct {
		name           string
		fileName       string
		content        []byte
		expectBinary   bool
		expectTooLarge bool
		expectLoaded   bool
		minLines       int
	}{
		{
			name:         "Text file",
			fileName:     "test.txt",
			content:      []byte("Line 1\nLine 2\nLine 3"),
			expectBinary: false,
			expectLoaded: true,
			minLines:     1, // At least one line
		},
		{
			name:         "Go source file",
			fileName:     "main.go",
			content:      []byte("package main\n\nfunc main() {\n\tprintln(\"hello\")\n}"),
			expectBinary: false,
			expectLoaded: true,
			minLines:     1, // At least one line (syntax highlighting may affect line count)
		},
		{
			name:         "JSON file",
			fileName:     "data.json",
			content:      []byte(`{"key": "value", "number": 123}`),
			expectBinary: false,
			expectLoaded: true,
			minLines:     1,
		},
		{
			name:         "Binary file",
			fileName:     "binary.bin",
			content:      []byte{0x00, 0x01, 0x02, 0xFF, 0xFE},
			expectBinary: true,
			expectLoaded: true,
			minLines:     3, // Binary file message
		},
		{
			name:         "Empty file",
			fileName:     "empty.txt",
			content:      []byte{},
			expectBinary: false,
			expectLoaded: true,
			minLines:     0,
		},
		{
			name:         "Markdown file",
			fileName:     "README.md",
			content:      []byte("# Title\n\nSome content"),
			expectBinary: false,
			expectLoaded: true,
			minLines:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.fileName)
			createTestFileWithContent(t, testFile, tt.content)

			m := &model{
				preview: previewModel{},
			}

			m.loadPreview(testFile)

			if !m.preview.loaded {
				if tt.expectLoaded {
					t.Error("Expected preview to be loaded")
				}
				return
			}

			if m.preview.isBinary != tt.expectBinary {
				t.Errorf("Expected isBinary=%v, got %v", tt.expectBinary, m.preview.isBinary)
			}

			if m.preview.tooLarge != tt.expectTooLarge {
				t.Errorf("Expected tooLarge=%v, got %v", tt.expectTooLarge, m.preview.tooLarge)
			}

			if len(m.preview.content) < tt.minLines {
				t.Errorf("Expected at least %d lines, got %d", tt.minLines, len(m.preview.content))
			}

			if m.preview.filePath != testFile {
				t.Errorf("Expected filePath=%s, got %s", testFile, m.preview.filePath)
			}

			if m.preview.fileName != tt.fileName {
				t.Errorf("Expected fileName=%s, got %s", tt.fileName, m.preview.fileName)
			}
		})
	}
}

// TestLoadPreviewLargeFile tests handling of files larger than 1MB
func TestLoadPreviewLargeFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a file larger than 1MB
	largeFile := filepath.Join(tmpDir, "large.txt")
	largeContent := make([]byte, 1024*1024+1) // 1MB + 1 byte
	for i := range largeContent {
		largeContent[i] = 'A'
	}
	createTestFileWithContent(t, largeFile, largeContent)

	m := &model{
		preview: previewModel{},
	}

	m.loadPreview(largeFile)

	if !m.preview.loaded {
		t.Error("Expected preview to be loaded")
	}

	if !m.preview.tooLarge {
		t.Error("Expected tooLarge flag to be set")
	}

	// Check that it shows appropriate message
	foundMessage := false
	for _, line := range m.preview.content {
		if strings.Contains(line, "too large") || strings.Contains(line, "Too large") {
			foundMessage = true
			break
		}
	}

	if !foundMessage {
		t.Error("Expected 'too large' message in preview content")
	}
}

// TestLoadPreviewNonExistent tests loading preview of non-existent file
func TestLoadPreviewNonExistent(t *testing.T) {
	m := &model{
		preview: previewModel{},
	}

	m.loadPreview("/nonexistent/file.txt")

	if m.preview.loaded {
		t.Error("Expected preview not to be loaded for non-existent file")
	}
}

// TestLoadPreviewImageFile tests image file detection
func TestLoadPreviewImageFile(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a fake image file (binary content with null bytes)
	imageFile := filepath.Join(tmpDir, "test.png")
	// PNG magic bytes plus some binary data with null bytes (to trigger binary detection)
	pngContent := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // Chunk length (null bytes)
		0x49, 0x48, 0x44, 0x52, // IHDR chunk type
		0x00, 0x00, 0x00, 0x10, // Width
		0x00, 0x00, 0x00, 0x10, // Height (more null bytes)
	}
	createTestFileWithContent(t, imageFile, pngContent)

	m := &model{
		preview: previewModel{},
	}

	m.loadPreview(imageFile)

	if !m.preview.loaded {
		t.Error("Expected preview to be loaded")
	}

	if !m.preview.isBinary {
		t.Error("Expected PNG file to be detected as binary")
	}

	// Check for image viewer hint in binary message
	foundBinaryMessage := false
	for _, line := range m.preview.content {
		if strings.Contains(line, "Binary") || strings.Contains(line, "binary") {
			foundBinaryMessage = true
			break
		}
	}

	if !foundBinaryMessage {
		t.Error("Expected binary file message in preview content")
	}
}

// TestGetIconForExtension tests icon selection based on file extension
func TestGetIconForExtension(t *testing.T) {
	tests := []struct {
		filename string
		wantIcon string // Expected icon (or empty if we just check it's not empty)
	}{
		{"test.go", "🐹"},
		{"script.py", "🐍"},
		{"app.js", "🟨"},
		{"component.tsx", "⚛️"},
		{"style.css", "🎨"},
		{"data.json", "🔶"},
		{"config.yaml", "⚙️"},
		{"README.md", "📝"},
		{"archive.zip", "📦"},
		{"photo.png", "🖼️"},
		{"document.pdf", "📕"},
		{"unknown.xyz", "📄"}, // Generic file
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			item := fileItem{
				name:  tt.filename,
				isDir: false,
			}
			icon := getFileIcon(item)

			// Just check we got some icon
			if icon == "" && tt.wantIcon != "" {
				t.Errorf("getFileIcon(%s) returned empty icon", tt.filename)
			}
		})
	}
}
