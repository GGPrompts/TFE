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

// getFilteredFiles returns files filtered by current filter mode
func (m *model) getFilteredFiles() []fileItem {
	if !m.showFavoritesOnly {
		return m.files
	}

	// Filter to only show favorites
	filtered := make([]fileItem, 0)
	for _, file := range m.files {
		// Always include ".." parent directory
		if file.name == ".." {
			filtered = append(filtered, file)
			continue
		}

		// Include if favorited
		if m.isFavorite(file.path) {
			filtered = append(filtered, file)
		}
	}

	return filtered
}
