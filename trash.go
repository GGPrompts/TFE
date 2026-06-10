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
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"syscall"
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

	// Atomic write: a crash mid-write must never truncate trash.json,
	// because loadTrashMetadata hard-fails on corrupt JSON, which would
	// permanently break all trash operations.
	return atomicWriteFile(metadataPath, data, 0644)
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
	// Try rename first (fast, atomic)
	err = os.Rename(path, trashedPath)
	if err != nil {
		// Check if this is a cross-device error (different mount points)
		if errors.Is(err, syscall.EXDEV) {
			// Fallback to copy+delete for cross-device moves
			// This happens when moving between different filesystems
			// (e.g., /tmp → ~/.config/tfe/trash on different partitions)
			err = copyRecursive(path, trashedPath)
			if err != nil {
				return fmt.Errorf("failed to copy to trash: %w", err)
			}

			// Only delete original after successful copy
			err = os.RemoveAll(path)
			if err != nil {
				// Try to clean up the copy since we couldn't delete the original
				os.RemoveAll(trashedPath)
				return fmt.Errorf("failed to delete original after copy: %w", err)
			}
		} else {
			// Some other error (permissions, etc.)
			return fmt.Errorf("failed to move to trash: %w", err)
		}
	}

	// Load existing trash metadata
	items, err := loadTrashMetadata()
	if err != nil {
		// If metadata load fails, try to restore the file
		if rbErr := rollbackTrashMove(trashedPath, path); rbErr != nil {
			return fmt.Errorf("failed to load trash metadata: %v; rollback also failed: %v; file preserved at %s", err, rbErr, trashedPath)
		}
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
		if rbErr := rollbackTrashMove(trashedPath, path); rbErr != nil {
			return fmt.Errorf("failed to save trash metadata: %v; rollback also failed: %v; file preserved at %s", err, rbErr, trashedPath)
		}
		return fmt.Errorf("failed to save trash metadata: %w", err)
	}

	return nil
}

// rollbackRename is os.Rename, indirected so tests can simulate rename
// failures (e.g. cross-device EXDEV) when exercising the rollback fallback.
var rollbackRename = os.Rename

// restoreRename is os.Rename, indirected so tests can simulate rename
// failures (e.g. cross-device EXDEV) when exercising the copy+delete fallback
// in restoreFromTrash.
var restoreRename = os.Rename

// rollbackTrashMove moves a file back from trash to its original location
// after a metadata failure. os.Rename alone is not enough: if the original
// move used the copy+delete fallback (cross-device), the reverse rename
// fails with EXDEV too, so on rename failure this falls back to copy+delete.
// Returns an error only if the file could not be restored — in that case
// the file still exists at trashedPath.
func rollbackTrashMove(trashedPath, originalPath string) error {
	// Ensure the original parent directory still exists (it may have been
	// removed while the file was being trashed)
	if err := os.MkdirAll(filepath.Dir(originalPath), 0755); err != nil {
		return fmt.Errorf("failed to recreate parent directory: %w", err)
	}

	// Try rename first (fast, atomic)
	if err := rollbackRename(trashedPath, originalPath); err == nil {
		return nil
	}

	// Rename failed (e.g. cross-device EXDEV) — fall back to copy+delete
	if err := copyRecursive(trashedPath, originalPath); err != nil {
		return fmt.Errorf("failed to copy back from trash: %w", err)
	}

	// Original restored; remove the trash copy. Best effort: the restore
	// itself succeeded even if this cleanup fails.
	os.RemoveAll(trashedPath)
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

	// Re-stat the original path immediately before the rename to narrow the
	// TOCTOU window from the existence check above: os.Rename would silently
	// overwrite a file created at OriginalPath between the check and the move.
	if _, err := os.Stat(item.OriginalPath); err == nil {
		return fmt.Errorf("cannot restore: file already exists at original location")
	}

	// Restore the file
	// Try rename first (fast, atomic)
	if err := restoreRename(trashedPath, item.OriginalPath); err != nil {
		// Check if this is a cross-device error (different mount points).
		// Exactly the files that needed the copy fallback when trashed
		// (e.g. anything deleted from /tmp or another mount) hit this on
		// restore too, so mirror moveToTrash's cross-device handling.
		if errors.Is(err, syscall.EXDEV) {
			// Fallback to copy+delete for cross-device moves
			if err := copyRecursive(trashedPath, item.OriginalPath); err != nil {
				// Clean up the partial copy so a failed restore doesn't
				// leave a half-written file at the original location.
				os.RemoveAll(item.OriginalPath)
				return fmt.Errorf("failed to restore file: %w", err)
			}

			// Only delete the trash copy after a successful copy
			os.RemoveAll(trashedPath)
		} else {
			// Some other error (permissions, etc.)
			return fmt.Errorf("failed to restore file: %w", err)
		}
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

	// Delete all trashed files/directories, retaining any that fail to delete
	// so they remain visible in the trash view and can be retried later.
	var errors []string
	var failedItems []trashItem
	for _, item := range items {
		if err := os.RemoveAll(item.TrashedPath); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", item.OriginalName, err))
			failedItems = append(failedItems, item)
		}
	}

	// Persist metadata for items that could not be deleted (empty slice if all succeeded)
	if failedItems == nil {
		failedItems = []trashItem{}
	}
	if err := saveTrashMetadata(failedItems); err != nil {
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
	slices.SortFunc(items, func(a, b trashItem) int {
		return b.DeletedAt.Compare(a.DeletedAt)
	})

	return items, nil
}

// copyRecursive copies a file or directory recursively from src to dst
// This is used as a fallback when os.Rename() fails due to cross-device errors
func copyRecursive(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if info.IsDir() {
		return copyDir(src, dst, info)
	}
	return copyFile(src, dst, info)
}

// copyDir recursively copies a directory
func copyDir(src, dst string, info os.FileInfo) error {
	// Create destination directory with same permissions
	if err := os.MkdirAll(dst, info.Mode()); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Read directory entries
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	// Copy each entry recursively
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if err := copyRecursive(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// copyFile copies a single file, preserving permissions
func copyFile(src, dst string, info os.FileInfo) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	// Create destination file with same permissions
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	// Copy contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy contents: %w", err)
	}

	return nil
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
