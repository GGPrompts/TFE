package main

// Module: favorites.go
// Purpose: Favorites/bookmarks functionality for files and directories
// Responsibilities:
// - Loading and saving favorites from/to disk
// - Adding and removing favorites
// - Checking if a path is favorited

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// getFavoritesPath returns the path to the favorites file
func getFavoritesPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configDir := filepath.Join(home, ".config", "tfe")
	// Create config directory if it doesn't exist
	os.MkdirAll(configDir, 0755)

	return filepath.Join(configDir, "favorites.json")
}

// loadFavorites loads favorites from disk
func loadFavorites() map[string]bool {
	favorites := make(map[string]bool)

	path := getFavoritesPath()
	if path == "" {
		return favorites
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist yet or can't be read - return empty map
		return favorites
	}

	// Unmarshal JSON array of paths
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return favorites
	}

	// Convert to map for faster lookups
	for _, p := range paths {
		favorites[p] = true
	}

	return favorites
}

// saveFavorites saves favorites to disk
func saveFavorites(favorites map[string]bool) error {
	path := getFavoritesPath()
	if path == "" {
		return nil
	}

	// Convert map to array for JSON
	paths := make([]string, 0, len(favorites))
	for p := range favorites {
		paths = append(paths, p)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(paths, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0644)
}

// toggleFavorite adds or removes a path from favorites
func (m *model) toggleFavorite(path string) {
	if m.favorites[path] {
		// Remove from favorites
		delete(m.favorites, path)
	} else {
		// Add to favorites
		m.favorites[path] = true
	}

	// Save to disk
	saveFavorites(m.favorites)
}

// isFavorite checks if a path is favorited
func (m *model) isFavorite(path string) bool {
	return m.favorites[path]
}

// directoryContainsPrompts checks if a directory contains any prompt files (recursively, up to 2 levels deep)
func directoryContainsPrompts(dirPath string) bool {
	return checkForPromptsRecursive(dirPath, 0, 2)
}

// checkForPromptsRecursive recursively checks for prompt files up to maxDepth levels
func checkForPromptsRecursive(dirPath string, currentDepth, maxDepth int) bool {
	if currentDepth > maxDepth {
		return false
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		// Skip hidden files/folders (except important ones)
		if strings.HasPrefix(entry.Name(), ".") {
			importantFolders := []string{".claude", ".prompts"}
			isImportant := false
			for _, folder := range importantFolders {
				if entry.Name() == folder {
					isImportant = true
					break
				}
			}
			if !isImportant {
				continue
			}
		}

		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// Recursively check subdirectories
			if checkForPromptsRecursive(fullPath, currentDepth+1, maxDepth) {
				return true
			}
		} else {
			// Check if this file is a prompt file
			item := fileItem{
				name:  entry.Name(),
				path:  fullPath,
				isDir: false,
			}
			if isPromptFile(item) {
				return true
			}
		}
	}

	return false
}

// getFilteredFiles returns files filtered by current filter mode
// Search filtering takes precedence over favorites and prompts filtering
func (m *model) getFilteredFiles() []fileItem {
	// If search is active, use filtered indices
	if len(m.filteredIndices) > 0 {
		filtered := make([]fileItem, 0, len(m.filteredIndices))
		for _, idx := range m.filteredIndices {
			if idx < len(m.files) {
				filtered = append(filtered, m.files[idx])
			}
		}
		return filtered
	}

	// Apply prompts filtering (show only .yaml, .md, .txt files)
	if m.showPromptsOnly {
		filtered := make([]fileItem, 0)

		// Add global prompts section at the top (if not already in ~/.prompts/)
		homeDir, err := os.UserHomeDir()
		if err == nil {
			globalPromptsDir := filepath.Join(homeDir, ".prompts")
			// Only show if we're not already in ~/.prompts/ and it exists
			if m.currentPath != globalPromptsDir && !strings.HasPrefix(m.currentPath, globalPromptsDir+string(filepath.Separator)) {
				// Check if ~/.prompts/ exists
				if info, err := os.Stat(globalPromptsDir); err == nil && info.IsDir() {
					// Create a virtual folder item for ~/.prompts/
					globalPromptsItem := fileItem{
						name:    "🌐 ~/.prompts/ (Global Prompts)",
						path:    globalPromptsDir,
						isDir:   true,
						size:    info.Size(),
						modTime: info.ModTime(),
						mode:    info.Mode(),
					}
					filtered = append(filtered, globalPromptsItem)
				}
			}
		}

		for _, item := range m.files {
			if item.isDir {
				// Always include ".." for navigation
				if item.name == ".." {
					filtered = append(filtered, item)
					continue
				}

				// Always include important dev folders
				importantFolders := []string{".claude", ".prompts", ".config"}
				isImportant := false
				for _, folder := range importantFolders {
					if item.name == folder {
						isImportant = true
						break
					}
				}
				if isImportant {
					filtered = append(filtered, item)
					continue
				}

				// Include directory if it contains prompt files
				if directoryContainsPrompts(item.path) {
					filtered = append(filtered, item)
				}
			} else if isPromptFile(item) {
				filtered = append(filtered, item)
			}
		}
		return filtered
	}

	// Otherwise, apply favorites filtering
	if !m.showFavoritesOnly {
		return m.files
	}

	// Show ALL favorites from anywhere in filesystem
	filtered := make([]fileItem, 0)

	// Don't include ".." when viewing favorites from multiple locations
	// (it doesn't make sense since favorites can be from anywhere)

	// Collect all favorite paths into a slice for sorting
	favPaths := make([]string, 0, len(m.favorites))
	for favPath := range m.favorites {
		favPaths = append(favPaths, favPath)
	}

	// Sort paths alphabetically for consistent ordering
	sort.Strings(favPaths)

	// Add all favorited paths as fileItems in sorted order
	for _, favPath := range favPaths {
		// Get file info for this favorite
		info, err := os.Stat(favPath)
		if err != nil {
			// File no longer exists, skip it
			continue
		}

		// Create fileItem from the favorite path
		item := fileItem{
			name:    filepath.Base(favPath),
			path:    favPath,
			isDir:   info.IsDir(),
			size:    info.Size(),
			modTime: info.ModTime(),
			mode:    info.Mode(),
		}
		filtered = append(filtered, item)
	}

	return filtered
}
