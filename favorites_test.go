package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// setupTestFavorites creates a temporary favorites environment
func setupTestFavorites(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}

	return tmpDir, cleanup
}

// TestGetFavoritesPath tests favorites file path retrieval
func TestGetFavoritesPath(t *testing.T) {
	_, cleanup := setupTestFavorites(t)
	defer cleanup()

	path := getFavoritesPath()
	if path == "" {
		t.Error("getFavoritesPath returned empty string")
	}

	// Verify path structure
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".config", "tfe", "favorites.json")
	if path != expected {
		t.Errorf("Expected path %s, got %s", expected, path)
	}

	// Verify config directory was created
	configDir := filepath.Dir(path)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}
}

// TestLoadFavorites_Empty tests loading when no favorites exist
func TestLoadFavorites_Empty(t *testing.T) {
	_, cleanup := setupTestFavorites(t)
	defer cleanup()

	favorites := loadFavorites()
	if len(favorites) != 0 {
		t.Errorf("Expected empty favorites, got %d items", len(favorites))
	}
}

// TestSaveAndLoadFavorites tests favorites persistence
func TestSaveAndLoadFavorites(t *testing.T) {
	tmpHome, cleanup := setupTestFavorites(t)
	defer cleanup()

	// Create test favorites
	testFavorites := map[string]bool{
		filepath.Join(tmpHome, "file1.txt"): true,
		filepath.Join(tmpHome, "file2.txt"): true,
		filepath.Join(tmpHome, "dir1"):      true,
	}

	// Save favorites
	if err := saveFavorites(testFavorites); err != nil {
		t.Fatalf("saveFavorites failed: %v", err)
	}

	// Load favorites
	loaded := loadFavorites()

	// Verify all favorites were loaded
	if len(loaded) != len(testFavorites) {
		t.Fatalf("Expected %d favorites, got %d", len(testFavorites), len(loaded))
	}

	for path := range testFavorites {
		if !loaded[path] {
			t.Errorf("Favorite %s was not loaded", path)
		}
	}
}

// TestSaveAndLoadFavorites_FileFormat tests JSON format
func TestSaveAndLoadFavorites_FileFormat(t *testing.T) {
	tmpHome, cleanup := setupTestFavorites(t)
	defer cleanup()

	testFavorites := map[string]bool{
		filepath.Join(tmpHome, "test.txt"): true,
		filepath.Join(tmpHome, "dir"):      true,
	}

	// Save favorites
	if err := saveFavorites(testFavorites); err != nil {
		t.Fatalf("saveFavorites failed: %v", err)
	}

	// Read the file directly
	favPath := getFavoritesPath()
	data, err := os.ReadFile(favPath)
	if err != nil {
		t.Fatalf("Failed to read favorites file: %v", err)
	}

	// Verify it's valid JSON array
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		t.Fatalf("Invalid JSON format: %v", err)
	}

	if len(paths) != 2 {
		t.Errorf("Expected 2 paths in JSON, got %d", len(paths))
	}
}

// TestToggleFavorite tests adding and removing favorites
func TestToggleFavorite(t *testing.T) {
	tmpHome, cleanup := setupTestFavorites(t)
	defer cleanup()

	testPath := filepath.Join(tmpHome, "test.txt")

	// Create a model with empty favorites
	m := &model{
		favorites: make(map[string]bool),
	}

	// Initially not favorited
	if m.isFavorite(testPath) {
		t.Error("Path should not be favorited initially")
	}

	// Toggle to add
	m.toggleFavorite(testPath)

	// Should now be favorited
	if !m.isFavorite(testPath) {
		t.Error("Path should be favorited after toggle")
	}

	// Verify it was saved to disk
	loaded := loadFavorites()
	if !loaded[testPath] {
		t.Error("Favorite was not persisted to disk")
	}

	// Toggle to remove
	m.toggleFavorite(testPath)

	// Should not be favorited anymore
	if m.isFavorite(testPath) {
		t.Error("Path should not be favorited after second toggle")
	}

	// Verify removal was saved to disk
	loaded = loadFavorites()
	if loaded[testPath] {
		t.Error("Favorite removal was not persisted to disk")
	}
}

// TestIsFavorite tests favorite checking
func TestIsFavorite(t *testing.T) {
	tmpHome, cleanup := setupTestFavorites(t)
	defer cleanup()

	favoritePath := filepath.Join(tmpHome, "favorite.txt")
	normalPath := filepath.Join(tmpHome, "normal.txt")

	m := &model{
		favorites: map[string]bool{
			favoritePath: true,
		},
	}

	// Check favorited path
	if !m.isFavorite(favoritePath) {
		t.Error("Favorited path should return true")
	}

	// Check non-favorited path
	if m.isFavorite(normalPath) {
		t.Error("Non-favorited path should return false")
	}

	// Check empty string
	if m.isFavorite("") {
		t.Error("Empty string should return false")
	}
}

// TestDirectoryContainsPrompts tests prompt detection in directories
func TestDirectoryContainsPrompts(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a minimal model for testing
	m := &model{
		promptDirsCache: make(map[string]bool),
	}

	// Create a directory with a .claude subfolder containing prompt files
	claudeDir := filepath.Join(tmpDir, "project", ".claude", "commands")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Add a prompt file
	promptFile := filepath.Join(claudeDir, "test.md")
	createTestFileWithContent(t, promptFile, []byte("# Test Command"))

	// Test directory containing prompts
	projectDir := filepath.Join(tmpDir, "project")
	if !m.directoryContainsPrompts(projectDir) {
		t.Error("Directory should contain prompts")
	}

	// Create a directory without prompts
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.Mkdir(emptyDir, 0755); err != nil {
		t.Fatalf("Failed to create empty directory: %v", err)
	}

	if m.directoryContainsPrompts(emptyDir) {
		t.Error("Empty directory should not contain prompts")
	}

	// Test with non-existent directory
	if m.directoryContainsPrompts(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("Non-existent directory should not contain prompts")
	}
}

// TestDirectoryContainsPrompts_Depth tests depth limitation
func TestDirectoryContainsPrompts_Depth(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a minimal model for testing
	m := &model{
		promptDirsCache: make(map[string]bool),
	}

	// Create nested directories beyond max depth (2)
	deepDir := filepath.Join(tmpDir, "level1", "level2", "level3", ".claude")
	if err := os.MkdirAll(deepDir, 0755); err != nil {
		t.Fatalf("Failed to create deep directory: %v", err)
	}

	// Add a prompt file at level 3
	promptFile := filepath.Join(deepDir, "command.md")
	createTestFileWithContent(t, promptFile, []byte("# Command"))

	// Should not find it (beyond max depth of 2)
	if m.directoryContainsPrompts(tmpDir) {
		t.Error("Should not find prompts beyond max depth")
	}

	// Create a prompt at level 1 (within depth limit)
	// tmpDir = depth 0, l1 = depth 1, .claude = depth 2 (max depth)
	level1Dir := filepath.Join(tmpDir, "l1", ".claude")
	if err := os.MkdirAll(level1Dir, 0755); err != nil {
		t.Fatalf("Failed to create level1 directory: %v", err)
	}
	promptFile2 := filepath.Join(level1Dir, "cmd.md")
	createTestFileWithContent(t, promptFile2, []byte("# Cmd"))

	// Clear cache so it re-scans after we added new files
	m.promptDirsCache = make(map[string]bool)

	// Should find it (within depth limit)
	if !m.directoryContainsPrompts(tmpDir) {
		t.Error("Should find prompts within max depth")
	}
}

// TestDirectoryContainsPrompts_HiddenFiles tests that hidden files are skipped
func TestDirectoryContainsPrompts_HiddenFiles(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a minimal model for testing
	m := &model{
		promptDirsCache: make(map[string]bool),
	}

	// Create .hidden directory (not in important list)
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	if err := os.Mkdir(hiddenDir, 0755); err != nil {
		t.Fatalf("Failed to create hidden directory: %v", err)
	}

	// Add a .md file in hidden directory
	hiddenFile := filepath.Join(hiddenDir, "test.md")
	createTestFileWithContent(t, hiddenFile, []byte("# Test"))

	// Should not find prompts in hidden directories (unless it's .claude or .prompts)
	if m.directoryContainsPrompts(tmpDir) {
		t.Error("Should not search in hidden directories")
	}

	// But .claude should be searched
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.Mkdir(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}
	claudeFile := filepath.Join(claudeDir, "cmd.md")
	createTestFileWithContent(t, claudeFile, []byte("# Cmd"))

	// Clear cache so it re-scans after we added new files
	m.promptDirsCache = make(map[string]bool)

	// Should find prompts in .claude
	if !m.directoryContainsPrompts(tmpDir) {
		t.Error("Should search in .claude directory")
	}
}

// TestGetFilteredFiles_NoFilter tests filtering with no filters active
func TestGetFilteredFiles_NoFilter(t *testing.T) {
	files := []fileItem{
		{name: "file1.txt", path: "/path/file1.txt"},
		{name: "file2.txt", path: "/path/file2.txt"},
		{name: "dir1", path: "/path/dir1", isDir: true},
	}

	m := model{
		files:             files,
		showFavoritesOnly: false,
		showPromptsOnly:   false,
		filteredIndices:   []int{},
	}

	result := m.getFilteredFiles()
	if len(result) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(result))
	}
}

// TestGetFilteredFiles_FavoritesOnly tests favorites filtering
func TestGetFilteredFiles_FavoritesOnly(t *testing.T) {
	tmpDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create test files
	file1 := filepath.Join(tmpDir, "favorite.txt")
	file2 := filepath.Join(tmpDir, "normal.txt")
	createTestFileWithContent(t, file1, []byte("content"))
	createTestFileWithContent(t, file2, []byte("content"))

	files := []fileItem{
		{name: "favorite.txt", path: file1, isDir: false},
		{name: "normal.txt", path: file2, isDir: false},
	}

	m := model{
		files: files,
		favorites: map[string]bool{
			file1: true, // Only file1 is favorited
		},
		showFavoritesOnly: true,
		showPromptsOnly:   false,
		filteredIndices:   []int{},
	}

	result := m.getFilteredFiles()

	// Should only return the favorited file
	if len(result) != 1 {
		t.Errorf("Expected 1 favorite file, got %d", len(result))
	}

	if len(result) > 0 && result[0].name != "favorite.txt" {
		t.Errorf("Expected favorite.txt, got %s", result[0].name)
	}
}

// TestLoadFavorites_CorruptedFile tests handling of corrupted favorites file
func TestLoadFavorites_CorruptedFile(t *testing.T) {
	_, cleanup := setupTestFavorites(t)
	defer cleanup()

	// Write invalid JSON to favorites file
	favPath := getFavoritesPath()
	if err := os.WriteFile(favPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Should return empty map instead of crashing
	favorites := loadFavorites()
	if len(favorites) != 0 {
		t.Error("Corrupted file should return empty favorites")
	}
}

// BenchmarkLoadFavorites benchmarks favorites loading
func BenchmarkLoadFavorites(b *testing.B) {
	tmpDir := b.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create favorites with 100 paths
	favorites := make(map[string]bool)
	for i := 0; i < 100; i++ {
		favorites[filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+".txt")] = true
	}
	saveFavorites(favorites)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loadFavorites()
	}
}

// BenchmarkSaveFavorites benchmarks favorites saving
func BenchmarkSaveFavorites(b *testing.B) {
	tmpDir := b.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	favorites := make(map[string]bool)
	for i := 0; i < 100; i++ {
		favorites[filepath.Join(tmpDir, "file"+string(rune('0'+i%10))+".txt")] = true
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		saveFavorites(favorites)
	}
}
