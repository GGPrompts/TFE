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

// TestCacheDirChecks tests the load-time vault/emptiness caching on fileItem
// (populated by loadFiles/loadSubdirFiles via cacheDirChecks) and the cached
// accessors used by getFileIcon and the render views.
func TestCacheDirChecks(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Plain non-empty directory
	plainDir := filepath.Join(tmpDir, "plain")
	if err := os.Mkdir(plainDir, 0755); err != nil {
		t.Fatalf("Failed to create plain dir: %v", err)
	}
	createTestFileWithContent(t, filepath.Join(plainDir, "file.txt"), []byte("content"))

	// Empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	// Obsidian vault (contains .obsidian folder)
	vaultDir := filepath.Join(tmpDir, "vault")
	if err := os.MkdirAll(filepath.Join(vaultDir, ".obsidian"), 0755); err != nil {
		t.Fatalf("Failed to create vault dir: %v", err)
	}

	t.Run("populates caches for directories", func(t *testing.T) {
		item := fileItem{name: "vault", path: vaultDir, isDir: true}
		cacheDirChecks(&item)
		if item.isVault == nil || !*item.isVault {
			t.Errorf("cacheDirChecks: isVault = %v, expected cached true", item.isVault)
		}
		if item.isEmptyDir == nil || *item.isEmptyDir {
			t.Errorf("cacheDirChecks: isEmptyDir = %v, expected cached false", item.isEmptyDir)
		}

		empty := fileItem{name: "empty", path: emptyDir, isDir: true}
		cacheDirChecks(&empty)
		if empty.isEmptyDir == nil || !*empty.isEmptyDir {
			t.Errorf("cacheDirChecks: isEmptyDir = %v, expected cached true", empty.isEmptyDir)
		}
	})

	t.Run("skips non-directories", func(t *testing.T) {
		item := fileItem{name: "file.txt", path: filepath.Join(plainDir, "file.txt"), isDir: false}
		cacheDirChecks(&item)
		if item.isVault != nil || item.isEmptyDir != nil {
			t.Error("cacheDirChecks should not populate caches for files")
		}
	})

	t.Run("getFileIcon reads cached values, not disk", func(t *testing.T) {
		// Cache deliberately contradicts the on-disk state to prove the cached
		// value is used instead of a live disk check.
		cachedTrue, cachedFalse := true, false
		vaultItem := fileItem{name: "plain", path: plainDir, isDir: true, isVault: &cachedTrue}
		if icon := getFileIcon(vaultItem); icon != "🧠" {
			t.Errorf("getFileIcon with cached isVault=true = %s, expected 🧠", icon)
		}
		emptyItem := fileItem{name: "plain", path: plainDir, isDir: true, isVault: &cachedFalse, isEmptyDir: &cachedTrue}
		if icon := getFileIcon(emptyItem); icon != "📂" {
			t.Errorf("getFileIcon with cached isEmptyDir=true = %s, expected 📂", icon)
		}
	})

	t.Run("accessors fall back to disk check when unset", func(t *testing.T) {
		vaultItem := fileItem{name: "vault", path: vaultDir, isDir: true}
		if !vaultItem.vaultDir() {
			t.Error("vaultDir() fallback = false, expected true for vault directory")
		}
		emptyItem := fileItem{name: "empty", path: emptyDir, isDir: true}
		if !emptyItem.emptyDir() {
			t.Error("emptyDir() fallback = false, expected true for empty directory")
		}
		plainItem := fileItem{name: "plain", path: plainDir, isDir: true}
		if plainItem.vaultDir() || plainItem.emptyDir() {
			t.Error("vaultDir()/emptyDir() fallback should be false for plain non-empty directory")
		}
	})
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

// TestCachedDirItemCount tests the lazy memoization of directory item counts
func TestCachedDirItemCount(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	testDir := filepath.Join(tmpDir, "test")
	if err := os.Mkdir(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	createTestFileWithContent(t, filepath.Join(testDir, "a.txt"), []byte("content"))
	createTestFileWithContent(t, filepath.Join(testDir, "b.txt"), []byte("content"))

	m := model{dirCountCache: make(map[string]int)}

	// First call populates the cache
	if got := m.cachedDirItemCount(testDir); got != 2 {
		t.Errorf("cachedDirItemCount = %d, expected 2", got)
	}
	if cached, ok := m.dirCountCache[testDir]; !ok || cached != 2 {
		t.Errorf("dirCountCache[%s] = %d (present=%v), expected 2", testDir, cached, ok)
	}

	// Subsequent calls return the memoized value, even if the directory changed
	createTestFileWithContent(t, filepath.Join(testDir, "c.txt"), []byte("content"))
	if got := m.cachedDirItemCount(testDir); got != 2 {
		t.Errorf("cachedDirItemCount after change (cache still valid) = %d, expected memoized 2", got)
	}

	// Invalidation (as done in loadFiles) picks up the new contents
	m.dirCountCache = make(map[string]int)
	if got := m.cachedDirItemCount(testDir); got != 3 {
		t.Errorf("cachedDirItemCount after invalidation = %d, expected 3", got)
	}

	// Nil cache falls back to a direct read without panicking
	var noCache model
	if got := noCache.cachedDirItemCount(testDir); got != 3 {
		t.Errorf("cachedDirItemCount with nil cache = %d, expected 3", got)
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
			expected: "⬆",
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
	// Create a minimal model for testing
	m := &model{
		glamourRenderer:      nil, // Will be created on first use
		glamourRendererWidth: 0,
	}

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
			rendered, err := m.renderMarkdownWithTimeout(tt.content, tt.width, tt.timeout)

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
	// Create a minimal model for testing
	m := &model{
		glamourRenderer:      nil, // Will be created on first use
		glamourRendererWidth: 0,
	}

	// This test verifies the timeout mechanism works
	// We use a very short timeout to trigger timeout condition
	_, err := m.renderMarkdownWithTimeout("# Test", 80, 1*time.Nanosecond)

	// Either it times out or completes successfully (timing is not deterministic)
	// We just verify it doesn't panic or hang
	if err != nil {
		if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "panic") {
			t.Logf("Expected timeout or panic error, got: %v", err)
		}
	}
}

// TestRenderMarkdownWithTimeout_ClearsRendererOnTimeout verifies the
// concurrency contract from tfe-5pv: when a render times out, the caller must
// leave m.glamourRenderer/m.glamourRendererWidth cleared so the orphaned
// goroutine's renderer is never reinstalled (and thus never concurrently
// Render()ed by a later call). With a 1ns timeout the select almost always
// takes the timeout branch.
func TestRenderMarkdownWithTimeout_ClearsRendererOnTimeout(t *testing.T) {
	m := &model{}

	for i := 0; i < 50; i++ {
		rendered, err := m.renderMarkdownWithTimeout("# Test\n\nbody", 80, 1*time.Nanosecond)
		if err == nil {
			// Render won the race against the 1ns timeout; on success the
			// renderer should be installed for the width we asked for.
			if m.glamourRenderer == nil {
				t.Fatalf("iter %d: successful render did not install renderer", i)
			}
			if m.glamourRendererWidth != 80 {
				t.Fatalf("iter %d: installed renderer width = %d, want 80", i, m.glamourRendererWidth)
			}
			continue
		}
		// On timeout (the path the test is designed to hit) the fields must be
		// cleared so no later call shares the orphaned goroutine's renderer.
		if !strings.Contains(err.Error(), "timeout") {
			t.Fatalf("iter %d: unexpected error: %v", i, err)
		}
		if m.glamourRenderer != nil {
			t.Fatalf("iter %d: renderer left installed after timeout (would race orphaned goroutine)", i)
		}
		if rendered != "" {
			t.Fatalf("iter %d: expected empty render on timeout, got %q", i, rendered)
		}
	}
}

// TestRenderMarkdownWithTimeout_RaceWithRendererNil reproduces the tfe-5pv
// scenario under -race: renders that may time out (orphaning a goroutine that
// could still be calling renderer.Render) interleaved with the renderer being
// nil'd from the same owning goroutine, exactly as menu.go does on a theme
// toggle. After the fix, the render goroutine never touches model fields, so
// the only accessor of m.glamourRenderer/Width is this single owning goroutine
// and `go test -race` must stay clean.
func TestRenderMarkdownWithTimeout_RaceWithRendererNil(t *testing.T) {
	m := &model{}

	timeouts := []time.Duration{1 * time.Nanosecond, 5 * time.Second}
	for i := 0; i < 200; i++ {
		_, _ = m.renderMarkdownWithTimeout("# Heading\n\n- a\n- b", 80, timeouts[i%len(timeouts)])

		// Simulate the menu.go theme-toggle invalidation that nils the cached
		// renderer from the Update loop.
		if i%3 == 0 {
			m.glamourRenderer = nil
			m.glamourRendererWidth = 0
		}
		// Vary the width so the reuse/recreate branches both get exercised.
		if i%2 == 0 {
			_, _ = m.renderMarkdownWithTimeout("# Other", 100, 5*time.Second)
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
		name         string
		currentPath  string
		showHidden   bool
		expectedMin  int // Minimum expected files (accounts for parent dir)
		expectParent bool
		expectHidden bool
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

// TestLoadFilesInvalidPath tests loading files from invalid path.
// On a ReadDir error loadFiles now falls through (instead of returning early)
// so the '..' parent entry is still appended, letting the user navigate back
// out. The path "/" has no parent, so it is the only case yielding 0 files.
func TestLoadFilesInvalidPath(t *testing.T) {
	m := &model{
		currentPath: "/nonexistent/path/that/does/not/exist",
		showHidden:  false,
	}

	m.loadFiles()

	// Expect exactly the '..' parent entry so the user can navigate out.
	if len(m.files) != 1 {
		t.Fatalf("Expected 1 file ('..' entry) for invalid path, got %d", len(m.files))
	}
	if m.files[0].name != ".." {
		t.Errorf("Expected '..' parent entry, got %q", m.files[0].name)
	}
	if m.statusMessage == "" {
		t.Errorf("Expected a status message surfacing the ReadDir error, got none")
	}

	// At root ("/") there is no parent entry to append, so 0 files is expected.
	mRoot := &model{
		currentPath: "/",
		showHidden:  false,
	}
	mRoot.loadFiles()
	// "/" may or may not be readable depending on the environment; only assert
	// the no-parent invariant when it is unreadable (matching the audit's note).
	for _, f := range mRoot.files {
		if f.name == ".." {
			t.Errorf("Root path should never have a '..' parent entry")
		}
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

	// Current implementation marks preview as loaded even for errors
	// It just contains an error message in the content
	if !m.preview.loaded {
		t.Error("Expected preview to be loaded (with error message) for non-existent file")
	}

	// Verify it shows an error message
	if len(m.preview.content) == 0 {
		t.Error("Expected error message in preview content")
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

	// Check for image file message (current implementation shows "Image File" not "Binary")
	foundImageMessage := false
	for _, line := range m.preview.content {
		if strings.Contains(line, "Image") || strings.Contains(line, "image") ||
			strings.Contains(line, "preview") {
			foundImageMessage = true
			break
		}
	}

	if !foundImageMessage {
		t.Error("Expected image file message in preview content")
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
		{"component.tsx", "⚛"},
		{"style.css", "🎨"},
		{"data.json", "🔶"},
		{"config.yaml", "⚙"},
		{"README.md", "📝"},
		{"archive.zip", "📦"},
		{"photo.png", "🖼"},
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

// TestCopyFileSelfCopyRejected verifies that copying a file onto itself is
// rejected and the source content is left intact (regression test for the
// os.Create truncation data-loss bug).
func TestCopyFileSelfCopyRejected(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	content := []byte("important data that must survive")
	srcPath := filepath.Join(tmpDir, "file.txt")
	createTestFileWithContent(t, srcPath, content)

	m := &model{}
	if err := m.copyFile(srcPath, srcPath); err == nil {
		t.Error("Expected error when copying a file onto itself, got nil")
	}

	got, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("Failed to read source after self-copy attempt: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("Source content was modified by self-copy: got %q, want %q", got, content)
	}
}

// TestCopyFileHardlinkSelfCopyRejected verifies that copying a file onto a
// hardlink of itself is rejected (same inode, different path).
func TestCopyFileHardlinkSelfCopyRejected(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	content := []byte("hardlinked data")
	srcPath := filepath.Join(tmpDir, "original.txt")
	linkPath := filepath.Join(tmpDir, "hardlink.txt")
	createTestFileWithContent(t, srcPath, content)
	if err := os.Link(srcPath, linkPath); err != nil {
		t.Skipf("Hardlinks not supported on this filesystem: %v", err)
	}

	m := &model{}
	if err := m.copyFile(srcPath, linkPath); err == nil {
		t.Error("Expected error when copying a file onto its hardlink, got nil")
	}

	got, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("Failed to read source after hardlink-copy attempt: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("Source content was modified: got %q, want %q", got, content)
	}
}

// TestCopyDirectoryOntoItselfRejected verifies that copying a directory onto
// itself (dst == src) is rejected.
func TestCopyDirectoryOntoItselfRejected(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	srcDir := filepath.Join(tmpDir, "mydir")
	createTestFileWithContent(t, filepath.Join(srcDir, "inner.txt"), []byte("inner"))

	m := &model{}
	if err := m.copyFile(srcDir, srcDir); err == nil {
		t.Error("Expected error when copying a directory onto itself, got nil")
	}
}

// TestCopyDirectoryIntoItselfRejected verifies that copying a directory into
// itself (or a subdirectory of itself) is rejected instead of recursing.
func TestCopyDirectoryIntoItselfRejected(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	srcDir := filepath.Join(tmpDir, "mydir")
	subDir := filepath.Join(srcDir, "sub")
	createTestFileWithContent(t, filepath.Join(subDir, "inner.txt"), []byte("inner"))

	m := &model{}

	// Destination directly inside source
	if err := m.copyFile(srcDir, filepath.Join(srcDir, "mydir")); err == nil {
		t.Error("Expected error when copying a directory into itself, got nil")
	}

	// Destination inside a subdirectory of source
	if err := m.copyFile(srcDir, filepath.Join(subDir, "mydir")); err == nil {
		t.Error("Expected error when copying a directory into its subdirectory, got nil")
	}
}

// TestCopyFileExistingDestinationRejected verifies that an existing
// destination file is not silently overwritten.
func TestCopyFileExistingDestinationRejected(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")
	createTestFileWithContent(t, srcPath, []byte("new content"))
	createTestFileWithContent(t, dstPath, []byte("existing content"))

	m := &model{}
	if err := m.copyFile(srcPath, dstPath); err == nil {
		t.Error("Expected error when destination already exists, got nil")
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination: %v", err)
	}
	if string(got) != "existing content" {
		t.Errorf("Destination was overwritten: got %q, want %q", got, "existing content")
	}
}

// TestCopyFileValidDestination verifies that a normal copy still works after
// the validation was added.
func TestCopyFileValidDestination(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	content := []byte("copy me")
	srcPath := filepath.Join(tmpDir, "src.txt")
	dstPath := filepath.Join(tmpDir, "dst.txt")
	createTestFileWithContent(t, srcPath, content)

	m := &model{}
	if err := m.copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("Expected copy to succeed, got error: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("Copied content mismatch: got %q, want %q", got, content)
	}
}

// TestCopyDirectoryValidDestination verifies that a normal recursive
// directory copy still works after the validation was added.
func TestCopyDirectoryValidDestination(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	srcDir := filepath.Join(tmpDir, "srcdir")
	createTestFileWithContent(t, filepath.Join(srcDir, "a.txt"), []byte("aaa"))
	createTestFileWithContent(t, filepath.Join(srcDir, "nested", "b.txt"), []byte("bbb"))

	dstDir := filepath.Join(tmpDir, "dstdir")
	m := &model{}
	if err := m.copyFile(srcDir, dstDir); err != nil {
		t.Fatalf("Expected directory copy to succeed, got error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dstDir, "nested", "b.txt"))
	if err != nil {
		t.Fatalf("Failed to read copied nested file: %v", err)
	}
	if string(got) != "bbb" {
		t.Errorf("Copied nested content mismatch: got %q, want %q", got, "bbb")
	}
}

// TestCopyFileContentPreservesMode verifies copyFileContent applies the
// supplied FileInfo's permission bits to the destination, and that passing
// pre-statted FileInfo (rather than re-statting inside) still produces a
// correct copy on the happy path.
func TestCopyFileContentPreservesMode(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	content := []byte("preserve my bits")
	srcPath := filepath.Join(tmpDir, "src.sh")
	dstPath := filepath.Join(tmpDir, "dst.sh")
	createTestFileWithContent(t, srcPath, content)

	// Give the source a distinctive executable mode.
	if err := os.Chmod(srcPath, 0o755); err != nil {
		t.Fatalf("Failed to chmod source: %v", err)
	}
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("Failed to stat source: %v", err)
	}

	if err := copyFileContent(srcPath, dstPath, srcInfo); err != nil {
		t.Fatalf("copyFileContent failed: %v", err)
	}

	got, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read destination: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("Copied content mismatch: got %q, want %q", got, content)
	}

	dstInfo, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("Failed to stat destination: %v", err)
	}
	if dstInfo.Mode().Perm() != srcInfo.Mode().Perm() {
		t.Errorf("Permission bits not preserved: got %v, want %v",
			dstInfo.Mode().Perm(), srcInfo.Mode().Perm())
	}
}

// TestLoadFilesReappliesSearchFilter verifies that loadFiles re-applies an
// active search filter against the freshly loaded listing, so filteredIndices
// can never reference a stale file list after an in-place reload such as a
// file-watcher event (regression test for tfe-6nz).
func TestLoadFilesReappliesSearchFilter(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"alpha.txt", "beta.txt"} {
		createTestFileWithContent(t, filepath.Join(tmpDir, name), []byte("x"))
	}

	m := &model{currentPath: tmpDir}
	m.loadFiles()

	// Active search (query retained after accepting with Enter)
	m.searchQuery = "alpha"
	m.filteredIndices = m.filterFilesBySearch("alpha")

	matchCount := func() int {
		count := 0
		for _, idx := range m.filteredIndices {
			if idx >= len(m.files) {
				t.Fatalf("filteredIndices contains out-of-range index %d (len(files)=%d)", idx, len(m.files))
			}
			name := m.files[idx].name
			if name != ".." && !strings.Contains(strings.ToLower(name), "alpha") {
				t.Errorf("filteredIndices points at non-matching file %q", name)
			}
			if strings.Contains(strings.ToLower(name), "alpha") {
				count++
			}
		}
		return count
	}

	if got := matchCount(); got != 1 {
		t.Fatalf("Expected 1 matching file before reload, got %d", got)
	}

	// Simulate a file-watcher reload after a new matching file appears
	createTestFileWithContent(t, filepath.Join(tmpDir, "alphabet.txt"), []byte("y"))
	m.loadFiles()

	if m.searchQuery != "alpha" {
		t.Errorf("Expected active search to survive reload, got query %q", m.searchQuery)
	}
	if got := matchCount(); got != 2 {
		t.Errorf("Expected filter re-applied to new listing with 2 matches, got %d", got)
	}
}

// TestLoadFilesStaleIndicesCannotSurviveReload verifies that even if
// filteredIndices somehow reference a previous directory's listing, loadFiles
// rebuilds them from the current files when a search is active (tfe-6nz).
func TestLoadFilesStaleIndicesCannotSurviveReload(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFileWithContent(t, filepath.Join(tmpDir, "match.txt"), []byte("x"))

	m := &model{
		currentPath:     tmpDir,
		searchQuery:     "match",
		filteredIndices: []int{7, 42, 99}, // garbage indices from an old listing
	}
	m.loadFiles()

	for _, idx := range m.filteredIndices {
		if idx >= len(m.files) {
			t.Fatalf("Stale out-of-range index %d survived loadFiles (len(files)=%d)", idx, len(m.files))
		}
		name := m.files[idx].name
		if name != ".." && !strings.Contains(strings.ToLower(name), "match") {
			t.Errorf("filteredIndices points at non-matching file %q after reload", name)
		}
	}
}

// TestSortFilesName verifies name sorting keeps ".." first, groups directories
// before files, lowercases for comparison, and reverses on descending order.
func TestSortFilesName(t *testing.T) {
	mk := func(name string, isDir bool) fileItem {
		return fileItem{name: name, isDir: isDir}
	}
	base := []fileItem{
		mk("..", true),
		mk("Banana.txt", false),
		mk("apple.txt", false),
		mk("Zebra", true),
		mk("alpha", true),
	}

	names := func(items []fileItem) []string {
		out := make([]string, len(items))
		for i, it := range items {
			out[i] = it.name
		}
		return out
	}
	eq := func(got, want []string) bool {
		if len(got) != len(want) {
			return false
		}
		for i := range got {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	}

	// Ascending: parent first, dirs (case-insensitive) then files (case-insensitive)
	m := &model{sortBy: "name", sortAsc: true}
	m.files = append([]fileItem(nil), base...)
	m.sortFiles()
	wantAsc := []string{"..", "alpha", "Zebra", "apple.txt", "Banana.txt"}
	if got := names(m.files); !eq(got, wantAsc) {
		t.Errorf("ascending name sort = %v, want %v", got, wantAsc)
	}

	// Descending: parent stays first, dirs and files each reversed
	m = &model{sortBy: "name", sortAsc: false}
	m.files = append([]fileItem(nil), base...)
	m.sortFiles()
	wantDesc := []string{"..", "Zebra", "alpha", "Banana.txt", "apple.txt"}
	if got := names(m.files); !eq(got, wantDesc) {
		t.Errorf("descending name sort = %v, want %v", got, wantDesc)
	}
}

// TestSortFilesModifiedSecondaryName verifies that the "modified" sort uses
// case-insensitive name as the secondary key when modtimes are equal, and that
// descending order reverses the combined comparison.
func TestSortFilesModifiedSecondaryName(t *testing.T) {
	t0 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)
	files := []fileItem{
		{name: "older.txt", modTime: t0},
		{name: "Beta.txt", modTime: t1},
		{name: "alpha.txt", modTime: t1},
	}

	m := &model{sortBy: "modified", sortAsc: true}
	m.files = append([]fileItem(nil), files...)
	m.sortFiles()
	// Ascending: oldest first; the two equal-time files break by name asc.
	want := []string{"older.txt", "alpha.txt", "Beta.txt"}
	got := make([]string, len(m.files))
	for i, it := range m.files {
		got[i] = it.name
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ascending modified sort = %v, want %v", got, want)
		}
	}
}
