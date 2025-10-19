package main

// Module: trash.go
// Purpose: Trash/Recycle bin functionality for safe file deletion
// Responsibilities:
// - Moving files to trash instead of permanent deletion
// - Tracking trash metadata (original path, deletion time)
// - Restoring files from trash
// - Emptying trash (permanent deletion)
// - Listing trash contents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// trashItem represents a deleted item in the trash
type trashItem struct {
	OriginalPath string    `json:"original_path"` // Full path before deletion
	TrashedPath  string    `json:"trashed_path"`  // Path in trash directory
	DeletedAt    time.Time `json:"deleted_at"`    // When it was deleted
	OriginalName string    `json:"original_name"` // Original filename
	IsDir        bool      `json:"is_dir"`        // Is it a directory?
	Size         int64     `json:"size"`          // File/dir size in bytes
}

// getTrashDir returns the path to the trash directory
func getTrashDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	trashDir := filepath.Join(home, ".config", "tfe", "trash")

	// Create trash directory if it doesn't exist
	if err := os.MkdirAll(trashDir, 0755); err != nil {
		return "", err
	}

	return trashDir, nil
}

// getTrashMetadataPath returns the path to the trash metadata file
func getTrashMetadataPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "tfe")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "trash.json"), nil
}

// loadTrashMetadata loads the trash metadata from disk
func loadTrashMetadata() ([]trashItem, error) {
	metadataPath, err := getTrashMetadataPath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty list
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return []trashItem{}, nil
	}

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, err
	}

	var items []trashItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	return items, nil
}

// saveTrashMetadata saves the trash metadata to disk
func saveTrashMetadata(items []trashItem) error {
	metadataPath, err := getTrashMetadataPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(metadataPath, data, 0644)
}

// moveToTrash moves a file or directory to the trash
func moveToTrash(path string) error {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Get trash directory
	trashDir, err := getTrashDir()
	if err != nil {
		return fmt.Errorf("failed to get trash directory: %w", err)
	}

	// Generate unique name in trash (timestamp + original name)
	timestamp := time.Now().Format("20060102_150405")
	originalName := filepath.Base(path)
	trashedName := fmt.Sprintf("%s_%s", timestamp, originalName)
	trashedPath := filepath.Join(trashDir, trashedName)

	// Handle name collisions (rare but possible if deleting multiple files per second)
	counter := 1
	for {
		if _, err := os.Stat(trashedPath); os.IsNotExist(err) {
			break
		}
		trashedName = fmt.Sprintf("%s_%s_%d", timestamp, originalName, counter)
		trashedPath = filepath.Join(trashDir, trashedName)
		counter++
	}

	// Move the file/directory to trash
	if err := os.Rename(path, trashedPath); err != nil {
		return fmt.Errorf("failed to move to trash: %w", err)
	}

	// Load existing trash metadata
	items, err := loadTrashMetadata()
	if err != nil {
		// If metadata load fails, try to restore the file
		os.Rename(trashedPath, path)
		return fmt.Errorf("failed to load trash metadata: %w", err)
	}

	// Add new item to metadata
	newItem := trashItem{
		OriginalPath: path,
		TrashedPath:  trashedPath,
		DeletedAt:    time.Now(),
		OriginalName: originalName,
		IsDir:        info.IsDir(),
		Size:         info.Size(),
	}
	items = append(items, newItem)

	// Save updated metadata
	if err := saveTrashMetadata(items); err != nil {
		// If metadata save fails, try to restore the file
		os.Rename(trashedPath, path)
		return fmt.Errorf("failed to save trash metadata: %w", err)
	}

	return nil
}

// restoreFromTrash restores a file from trash to its original location
func restoreFromTrash(trashedPath string) error {
	// Load trash metadata
	items, err := loadTrashMetadata()
	if err != nil {
		return fmt.Errorf("failed to load trash metadata: %w", err)
	}

	// Find the item in metadata
	itemIndex := -1
	var item trashItem
	for i, it := range items {
		if it.TrashedPath == trashedPath {
			itemIndex = i
			item = it
			break
		}
	}

	if itemIndex == -1 {
		return fmt.Errorf("item not found in trash metadata")
	}

	// Check if original path still exists
	if _, err := os.Stat(item.OriginalPath); err == nil {
		return fmt.Errorf("cannot restore: file already exists at original location")
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(item.OriginalPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}

	// Restore the file
	if err := os.Rename(trashedPath, item.OriginalPath); err != nil {
		return fmt.Errorf("failed to restore file: %w", err)
	}

	// Remove from metadata
	items = append(items[:itemIndex], items[itemIndex+1:]...)
	if err := saveTrashMetadata(items); err != nil {
		// File is already restored, just log the metadata error
		return fmt.Errorf("file restored but failed to update metadata: %w", err)
	}

	return nil
}

// emptyTrash permanently deletes all items in the trash
func emptyTrash() error {
	// Load trash metadata
	items, err := loadTrashMetadata()
	if err != nil {
		return fmt.Errorf("failed to load trash metadata: %w", err)
	}

	// Delete all trashed files/directories
	var errors []string
	for _, item := range items {
		if err := os.RemoveAll(item.TrashedPath); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", item.OriginalName, err))
		}
	}

	// Clear metadata
	if err := saveTrashMetadata([]trashItem{}); err != nil {
		return fmt.Errorf("failed to clear trash metadata: %w", err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("some items failed to delete: %v", errors)
	}

	return nil
}

// getTrashItems returns all items currently in the trash, sorted by deletion time (newest first)
func getTrashItems() ([]trashItem, error) {
	items, err := loadTrashMetadata()
	if err != nil {
		return nil, err
	}

	// Sort by deletion time (newest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].DeletedAt.After(items[j].DeletedAt)
	})

	return items, nil
}

// getTrashSize returns the total size of all items in trash
func getTrashSize() (int64, error) {
	items, err := loadTrashMetadata()
	if err != nil {
		return 0, err
	}

	var totalSize int64
	for _, item := range items {
		totalSize += item.Size
	}

	return totalSize, nil
}

// cleanupOldTrash removes items from trash older than the specified duration
func cleanupOldTrash(olderThan time.Duration) (int, error) {
	items, err := loadTrashMetadata()
	if err != nil {
		return 0, fmt.Errorf("failed to load trash metadata: %w", err)
	}

	cutoffTime := time.Now().Add(-olderThan)
	var keptItems []trashItem
	removedCount := 0

	for _, item := range items {
		if item.DeletedAt.Before(cutoffTime) {
			// Remove old item
			os.RemoveAll(item.TrashedPath)
			removedCount++
		} else {
			// Keep recent item
			keptItems = append(keptItems, item)
		}
	}

	// Save updated metadata
	if err := saveTrashMetadata(keptItems); err != nil {
		return removedCount, fmt.Errorf("removed %d items but failed to update metadata: %w", removedCount, err)
	}

	return removedCount, nil
}

// convertTrashItemsToFileItems converts trash items to fileItems for display
func convertTrashItemsToFileItems(items []trashItem) []fileItem {
	fileItems := make([]fileItem, 0, len(items))

	for _, item := range items {
		// Get current file info from trashed path
		info, err := os.Stat(item.TrashedPath)
		if err != nil {
			// File no longer exists in trash, skip it
			continue
		}

		// Create fileItem with clean original name (no location suffix)
		// The location will be shown in the "Original Location" column in detail view
		fileItem := fileItem{
			name:    item.OriginalName, // Clean name without "(from ...)"
			path:    item.TrashedPath,  // Use trashed path for operations
			isDir:   item.IsDir,
			size:    item.Size,
			modTime: item.DeletedAt, // Show deletion time instead of modtime
			mode:    info.Mode(),
		}
		fileItems = append(fileItems, fileItem)
	}

	return fileItems
}

// getTrashItemByPath looks up a trash item by its trashed path
// Returns the trash item and true if found, or empty item and false if not found
func getTrashItemByPath(trashItems []trashItem, trashedPath string) (trashItem, bool) {
	for _, item := range trashItems {
		if item.TrashedPath == trashedPath {
			return item, true
		}
	}
	return trashItem{}, false
}

// permanentlyDelete permanently deletes a single item from trash
func permanentlyDeleteFromTrash(trashedPath string) error {
	// Load trash metadata
	items, err := loadTrashMetadata()
	if err != nil {
		return fmt.Errorf("failed to load trash metadata: %w", err)
	}

	// Find the item in metadata
	itemIndex := -1
	for i, item := range items {
		if item.TrashedPath == trashedPath {
			itemIndex = i
			break
		}
	}

	if itemIndex == -1 {
		return fmt.Errorf("item not found in trash metadata")
	}

	// Permanently delete the file/directory
	if err := os.RemoveAll(trashedPath); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	// Remove from metadata
	items = append(items[:itemIndex], items[itemIndex+1:]...)
	if err := saveTrashMetadata(items); err != nil {
		return fmt.Errorf("deleted but failed to update metadata: %w", err)
	}

	return nil
}
