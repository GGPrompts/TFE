package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestTrash creates a temporary trash directory for testing
func setupTestTrash(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)

	cleanup := func() {
		os.Setenv("HOME", originalHome)
	}

	return tmpDir, cleanup
}

// createTestFile creates a test file with given content
func createTestFile(t *testing.T, path, content string) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("Failed to create parent dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

// TestGetTrashDir tests trash directory creation and retrieval
func TestGetTrashDir(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	trashDir, err := getTrashDir()
	if err != nil {
		t.Fatalf("getTrashDir failed: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(trashDir)
	if err != nil {
		t.Fatalf("Trash directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Trash path is not a directory")
	}

	// Verify path structure
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".config", "tfe", "trash")
	if trashDir != expected {
		t.Errorf("Expected trash dir %s, got %s", expected, trashDir)
	}
}

// TestGetTrashMetadataPath tests metadata file path retrieval
func TestGetTrashMetadataPath(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	metadataPath, err := getTrashMetadataPath()
	if err != nil {
		t.Fatalf("getTrashMetadataPath failed: %v", err)
	}

	// Verify config directory was created
	configDir := filepath.Dir(metadataPath)
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path is not a directory")
	}

	// Verify path structure
	home := os.Getenv("HOME")
	expected := filepath.Join(home, ".config", "tfe", "trash.json")
	if metadataPath != expected {
		t.Errorf("Expected metadata path %s, got %s", expected, metadataPath)
	}
}

// TestLoadTrashMetadata_Empty tests loading when no metadata exists
func TestLoadTrashMetadata_Empty(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	items, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("loadTrashMetadata failed: %v", err)
	}

	if len(items) != 0 {
		t.Errorf("Expected empty trash, got %d items", len(items))
	}
}

// TestSaveAndLoadTrashMetadata tests metadata persistence
func TestSaveAndLoadTrashMetadata(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create test items
	testItems := []trashItem{
		{
			OriginalPath: "/tmp/test1.txt",
			TrashedPath:  "/tmp/trash/20240101_120000_test1.txt",
			DeletedAt:    time.Now(),
			OriginalName: "test1.txt",
			IsDir:        false,
			Size:         1024,
		},
		{
			OriginalPath: "/tmp/dir1",
			TrashedPath:  "/tmp/trash/20240101_120001_dir1",
			DeletedAt:    time.Now(),
			OriginalName: "dir1",
			IsDir:        true,
			Size:         4096,
		},
	}

	// Save metadata
	if err := saveTrashMetadata(testItems); err != nil {
		t.Fatalf("saveTrashMetadata failed: %v", err)
	}

	// Load metadata
	loadedItems, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("loadTrashMetadata failed: %v", err)
	}

	// Verify items match
	if len(loadedItems) != len(testItems) {
		t.Fatalf("Expected %d items, got %d", len(testItems), len(loadedItems))
	}

	for i := range testItems {
		if loadedItems[i].OriginalPath != testItems[i].OriginalPath {
			t.Errorf("Item %d: OriginalPath mismatch", i)
		}
		if loadedItems[i].TrashedPath != testItems[i].TrashedPath {
			t.Errorf("Item %d: TrashedPath mismatch", i)
		}
		if loadedItems[i].OriginalName != testItems[i].OriginalName {
			t.Errorf("Item %d: OriginalName mismatch", i)
		}
		if loadedItems[i].IsDir != testItems[i].IsDir {
			t.Errorf("Item %d: IsDir mismatch", i)
		}
		if loadedItems[i].Size != testItems[i].Size {
			t.Errorf("Item %d: Size mismatch", i)
		}
	}
}

// TestMoveToTrash_File tests moving a file to trash
func TestMoveToTrash_File(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create a test file
	testFile := filepath.Join(tmpHome, "test.txt")
	createTestFile(t, testFile, "test content")

	// Move to trash
	if err := moveToTrash(testFile); err != nil {
		t.Fatalf("moveToTrash failed: %v", err)
	}

	// Verify original file is gone
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Original file still exists after moving to trash")
	}

	// Verify metadata was created
	items, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item in trash, got %d", len(items))
	}

	item := items[0]
	if item.OriginalPath != testFile {
		t.Errorf("Expected OriginalPath %s, got %s", testFile, item.OriginalPath)
	}
	if item.OriginalName != "test.txt" {
		t.Errorf("Expected OriginalName test.txt, got %s", item.OriginalName)
	}
	if item.IsDir {
		t.Error("Expected IsDir to be false for file")
	}

	// Verify file exists in trash
	if _, err := os.Stat(item.TrashedPath); err != nil {
		t.Errorf("Trashed file does not exist: %v", err)
	}
}

// TestMoveToTrash_Directory tests moving a directory to trash
func TestMoveToTrash_Directory(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create a test directory with files
	testDir := filepath.Join(tmpHome, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test dir: %v", err)
	}
	createTestFile(t, filepath.Join(testDir, "file1.txt"), "content1")
	createTestFile(t, filepath.Join(testDir, "file2.txt"), "content2")

	// Move to trash
	if err := moveToTrash(testDir); err != nil {
		t.Fatalf("moveToTrash failed: %v", err)
	}

	// Verify original directory is gone
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("Original directory still exists after moving to trash")
	}

	// Verify metadata
	items, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("Expected 1 item in trash, got %d", len(items))
	}

	item := items[0]
	if !item.IsDir {
		t.Error("Expected IsDir to be true for directory")
	}

	// Verify directory and contents exist in trash
	trashedFiles, err := os.ReadDir(item.TrashedPath)
	if err != nil {
		t.Fatalf("Failed to read trashed directory: %v", err)
	}
	if len(trashedFiles) != 2 {
		t.Errorf("Expected 2 files in trashed directory, got %d", len(trashedFiles))
	}
}

// TestMoveToTrash_NameCollision tests handling of name collisions
func TestMoveToTrash_NameCollision(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create multiple files with same name in different directories
	file1 := filepath.Join(tmpHome, "dir1", "test.txt")
	file2 := filepath.Join(tmpHome, "dir2", "test.txt")
	file3 := filepath.Join(tmpHome, "dir3", "test.txt")

	createTestFile(t, file1, "content1")
	createTestFile(t, file2, "content2")
	createTestFile(t, file3, "content3")

	// Move all to trash rapidly (within same second)
	if err := moveToTrash(file1); err != nil {
		t.Fatalf("moveToTrash file1 failed: %v", err)
	}
	if err := moveToTrash(file2); err != nil {
		t.Fatalf("moveToTrash file2 failed: %v", err)
	}
	if err := moveToTrash(file3); err != nil {
		t.Fatalf("moveToTrash file3 failed: %v", err)
	}

	// Verify all 3 items in metadata
	items, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Expected 3 items in trash, got %d", len(items))
	}

	// Verify all have unique trashed paths
	pathSet := make(map[string]bool)
	for _, item := range items {
		if pathSet[item.TrashedPath] {
			t.Errorf("Duplicate trashed path: %s", item.TrashedPath)
		}
		pathSet[item.TrashedPath] = true
	}
}

// TestRestoreFromTrash tests restoring a file from trash
func TestRestoreFromTrash(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create and trash a file
	testFile := filepath.Join(tmpHome, "test.txt")
	createTestFile(t, testFile, "test content")

	if err := moveToTrash(testFile); err != nil {
		t.Fatalf("moveToTrash failed: %v", err)
	}

	// Get trashed path
	items, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	trashedPath := items[0].TrashedPath

	// Restore from trash
	if err := restoreFromTrash(trashedPath); err != nil {
		t.Fatalf("restoreFromTrash failed: %v", err)
	}

	// Verify file is restored
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Restored file does not exist: %v", err)
	}
	if string(content) != "test content" {
		t.Errorf("Restored file content mismatch: got %s", string(content))
	}

	// Verify trashed file is gone
	if _, err := os.Stat(trashedPath); !os.IsNotExist(err) {
		t.Error("Trashed file still exists after restore")
	}

	// Verify metadata is updated
	items, err = loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected empty trash after restore, got %d items", len(items))
	}
}

// TestRestoreFromTrash_Conflict tests restore when file already exists
func TestRestoreFromTrash_Conflict(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create and trash a file
	testFile := filepath.Join(tmpHome, "test.txt")
	createTestFile(t, testFile, "original")

	if err := moveToTrash(testFile); err != nil {
		t.Fatalf("moveToTrash failed: %v", err)
	}

	// Create a new file at the same location
	createTestFile(t, testFile, "new file")

	// Get trashed path
	items, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	trashedPath := items[0].TrashedPath

	// Try to restore (should fail)
	err = restoreFromTrash(trashedPath)
	if err == nil {
		t.Error("Expected restore to fail when file exists, but it succeeded")
	}

	// Verify the new file is unchanged
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != "new file" {
		t.Error("File was overwritten during failed restore")
	}
}

// TestEmptyTrash tests permanently deleting all trash
func TestEmptyTrash(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create and trash multiple files
	for i := 1; i <= 3; i++ {
		testFile := filepath.Join(tmpHome, "test"+string(rune('0'+i))+".txt")
		createTestFile(t, testFile, "content")
		if err := moveToTrash(testFile); err != nil {
			t.Fatalf("moveToTrash failed: %v", err)
		}
	}

	// Verify trash has items
	items, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("Expected 3 items in trash before empty, got %d", len(items))
	}

	// Empty trash
	if err := emptyTrash(); err != nil {
		t.Fatalf("emptyTrash failed: %v", err)
	}

	// Verify trash is empty
	items, err = loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected empty trash, got %d items", len(items))
	}

	// Verify trash directory is empty
	trashDir, _ := getTrashDir()
	entries, err := os.ReadDir(trashDir)
	if err != nil {
		t.Fatalf("Failed to read trash dir: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected empty trash directory, got %d entries", len(entries))
	}
}

// TestGetTrashItems tests retrieving and sorting trash items
func TestGetTrashItems(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create files with deliberate delays to ensure different timestamps
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, name := range files {
		testFile := filepath.Join(tmpHome, name)
		createTestFile(t, testFile, "content")
		if err := moveToTrash(testFile); err != nil {
			t.Fatalf("moveToTrash failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Get trash items
	items, err := getTrashItems()
	if err != nil {
		t.Fatalf("getTrashItems failed: %v", err)
	}

	if len(items) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(items))
	}

	// Verify they are sorted by deletion time (newest first)
	for i := 0; i < len(items)-1; i++ {
		if items[i].DeletedAt.Before(items[i+1].DeletedAt) {
			t.Error("Items not sorted by deletion time (newest first)")
		}
	}

	// Verify newest item is file3.txt
	if items[0].OriginalName != "file3.txt" {
		t.Errorf("Expected newest item to be file3.txt, got %s", items[0].OriginalName)
	}
}

// TestGetTrashSize tests calculating total trash size
func TestGetTrashSize(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create files with known sizes
	file1 := filepath.Join(tmpHome, "small.txt")
	file2 := filepath.Join(tmpHome, "large.txt")

	createTestFile(t, file1, "12345") // 5 bytes
	createTestFile(t, file2, "1234567890") // 10 bytes

	if err := moveToTrash(file1); err != nil {
		t.Fatalf("moveToTrash file1 failed: %v", err)
	}
	if err := moveToTrash(file2); err != nil {
		t.Fatalf("moveToTrash file2 failed: %v", err)
	}

	// Get trash size
	size, err := getTrashSize()
	if err != nil {
		t.Fatalf("getTrashSize failed: %v", err)
	}

	expectedSize := int64(15) // 5 + 10
	if size != expectedSize {
		t.Errorf("Expected trash size %d, got %d", expectedSize, size)
	}
}

// TestCleanupOldTrash tests removing old items from trash
func TestCleanupOldTrash(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create metadata with old and new items
	oldTime := time.Now().Add(-48 * time.Hour)
	newTime := time.Now()

	// Manually create trash items with specific timestamps
	trashDir, _ := getTrashDir()

	items := []trashItem{
		{
			OriginalPath: filepath.Join(tmpHome, "old1.txt"),
			TrashedPath:  filepath.Join(trashDir, "old1.txt"),
			DeletedAt:    oldTime,
			OriginalName: "old1.txt",
			IsDir:        false,
			Size:         100,
		},
		{
			OriginalPath: filepath.Join(tmpHome, "old2.txt"),
			TrashedPath:  filepath.Join(trashDir, "old2.txt"),
			DeletedAt:    oldTime,
			OriginalName: "old2.txt",
			IsDir:        false,
			Size:         100,
		},
		{
			OriginalPath: filepath.Join(tmpHome, "new.txt"),
			TrashedPath:  filepath.Join(trashDir, "new.txt"),
			DeletedAt:    newTime,
			OriginalName: "new.txt",
			IsDir:        false,
			Size:         100,
		},
	}

	// Create actual files in trash
	for _, item := range items {
		createTestFile(t, item.TrashedPath, "content")
	}

	// Save metadata
	if err := saveTrashMetadata(items); err != nil {
		t.Fatalf("saveTrashMetadata failed: %v", err)
	}

	// Cleanup items older than 24 hours
	count, err := cleanupOldTrash(24 * time.Hour)
	if err != nil {
		t.Fatalf("cleanupOldTrash failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected to remove 2 items, removed %d", count)
	}

	// Verify only new item remains
	remainingItems, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if len(remainingItems) != 1 {
		t.Fatalf("Expected 1 remaining item, got %d", len(remainingItems))
	}

	if remainingItems[0].OriginalName != "new.txt" {
		t.Errorf("Expected remaining item to be new.txt, got %s", remainingItems[0].OriginalName)
	}
}

// TestPermanentlyDeleteFromTrash tests deleting a single item permanently
func TestPermanentlyDeleteFromTrash(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create and trash multiple files
	file1 := filepath.Join(tmpHome, "file1.txt")
	file2 := filepath.Join(tmpHome, "file2.txt")

	createTestFile(t, file1, "content1")
	createTestFile(t, file2, "content2")

	if err := moveToTrash(file1); err != nil {
		t.Fatalf("moveToTrash file1 failed: %v", err)
	}
	if err := moveToTrash(file2); err != nil {
		t.Fatalf("moveToTrash file2 failed: %v", err)
	}

	// Get first item's trashed path
	items, err := loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}
	trashedPath := items[0].TrashedPath

	// Permanently delete first item
	if err := permanentlyDeleteFromTrash(trashedPath); err != nil {
		t.Fatalf("permanentlyDeleteFromTrash failed: %v", err)
	}

	// Verify file is gone
	if _, err := os.Stat(trashedPath); !os.IsNotExist(err) {
		t.Error("Permanently deleted file still exists")
	}

	// Verify only 1 item remains in metadata
	items, err = loadTrashMetadata()
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	if len(items) != 1 {
		t.Errorf("Expected 1 remaining item, got %d", len(items))
	}
}

// TestConvertTrashItemsToFileItems tests conversion for display
func TestConvertTrashItemsToFileItems(t *testing.T) {
	tmpHome, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create and trash a file
	testFile := filepath.Join(tmpHome, "subdir", "test.txt")
	createTestFile(t, testFile, "content")

	if err := moveToTrash(testFile); err != nil {
		t.Fatalf("moveToTrash failed: %v", err)
	}

	// Get trash items
	items, err := getTrashItems()
	if err != nil {
		t.Fatalf("getTrashItems failed: %v", err)
	}

	// Convert to file items
	fileItems := convertTrashItemsToFileItems(items)

	if len(fileItems) != 1 {
		t.Fatalf("Expected 1 file item, got %d", len(fileItems))
	}

	item := fileItems[0]

	// Verify name is clean (without location suffix)
	// Location is now shown in the "Original Location" column in detail view
	if item.name != "test.txt" {
		t.Errorf("Expected clean name 'test.txt', got: %s", item.name)
	}

	// Verify path is the trashed path
	if item.path != items[0].TrashedPath {
		t.Error("Path should be trashed path")
	}

	// Verify other fields
	if item.isDir {
		t.Error("File should not be marked as directory")
	}
	if item.size != 7 { // "content" = 7 bytes
		t.Errorf("Expected size 7, got %d", item.size)
	}
}

// TestTrashMetadataJSON tests metadata JSON format
func TestTrashMetadataJSON(t *testing.T) {
	_, cleanup := setupTestTrash(t)
	defer cleanup()

	// Create test item
	testTime := time.Date(2024, 10, 15, 12, 0, 0, 0, time.UTC)
	items := []trashItem{
		{
			OriginalPath: "/home/user/test.txt",
			TrashedPath:  "/home/user/.config/tfe/trash/20241015_120000_test.txt",
			DeletedAt:    testTime,
			OriginalName: "test.txt",
			IsDir:        false,
			Size:         1024,
		},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Verify JSON contains expected fields
	jsonStr := string(data)
	expectedFields := []string{
		"original_path",
		"trashed_path",
		"deleted_at",
		"original_name",
		"is_dir",
		"size",
	}

	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("JSON missing expected field: %s", field)
		}
	}

	// Unmarshal and verify
	var decoded []trashItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(decoded) != 1 {
		t.Fatalf("Expected 1 decoded item, got %d", len(decoded))
	}

	if decoded[0].OriginalPath != items[0].OriginalPath {
		t.Error("Decoded OriginalPath mismatch")
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
