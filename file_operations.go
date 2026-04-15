package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// isClaudeContextFile checks if a file/folder is automatically loaded by Claude Code
func isClaudeContextFile(name string) bool {
	// Files that Claude Code automatically loads into context
	claudeFiles := []string{
		"CLAUDE.md",
		"CLAUDE.local.md",
		".claude",
	}

	for _, cf := range claudeFiles {
		if name == cf {
			return true
		}
	}

	return false
}

// isAgentsFile checks if a file is AGENTS.md (agent configuration/documentation)
func isAgentsFile(name string) bool {
	return name == "AGENTS.md"
}

// isPromptsFolder checks if a folder is the .prompts directory
func isPromptsFolder(name string) bool {
	return name == ".prompts"
}

// isGlobalPromptsVirtualFolder checks if this is the virtual "🌐 ~/.prompts/" folder
func isGlobalPromptsVirtualFolder(name string) bool {
	return strings.HasPrefix(name, "🌐 ~/.prompts/")
}

// isGlobalClaudeVirtualFolder checks if this is the virtual "🤖 ~/.claude/" folder
func isGlobalClaudeVirtualFolder(name string) bool {
	return strings.HasPrefix(name, "🤖 ~/.claude/")
}

// isClaudePromptsSubfolder checks if a folder is a .claude subfolder (commands, agents, skills)
func isClaudePromptsSubfolder(name string) bool {
	return name == "commands" || name == "agents" || name == "skills"
}

// isSecretsFile checks if a file contains secrets/credentials (always show for security awareness)
func isSecretsFile(name string) bool {
	// .env files and variants
	if name == ".env" || strings.HasPrefix(name, ".env.") {
		return true
	}
	// Common secrets file patterns
	secretPatterns := []string{
		"secrets",
		"credentials",
		"secret",
		"credential",
		".key",
		".pem",
		".p12",
		".pfx",
		"private",
	}
	lowerName := strings.ToLower(name)
	for _, pattern := range secretPatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}
	return false
}

// isIgnoreFile checks if a file is a .gitignore or similar ignore file (always show for awareness)
func isIgnoreFile(name string) bool {
	ignoreFiles := []string{
		".gitignore",
		".dockerignore",
		".npmignore",
		".eslintignore",
		".prettierignore",
		".claudeignore",
		".gitattributes",
	}
	for _, ignoreFile := range ignoreFiles {
		if name == ignoreFile {
			return true
		}
	}
	return false
}

// hasPromptVariables checks if a file contains {{variable}} placeholders
// Uses a quick string search without full parsing for performance
func hasPromptVariables(path string) bool {
	// Read first 8KB of file (enough to detect variables in most cases)
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	defer file.Close()

	buffer := make([]byte, 8192)
	n, _ := file.Read(buffer)
	content := string(buffer[:n])

	// Quick check for {{variable}} pattern
	return strings.Contains(content, "{{") && strings.Contains(content, "}}")
}

// isInPromptsDirectory checks if a file is in a prompts-related directory
func isInPromptsDirectory(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Check if in ~/.prompts/ (global prompts)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		globalPromptsDir := filepath.Join(homeDir, ".prompts")
		if strings.HasPrefix(absPath, globalPromptsDir) {
			return true
		}
	}

	// Check if in .claude/commands/, .claude/agents/, or .claude/skills/
	if strings.Contains(absPath, "/.claude/commands/") ||
		strings.Contains(absPath, "/.claude/agents/") ||
		strings.Contains(absPath, "/.claude/skills/") {
		return true
	}

	// Check if in any directory named .prompts or prompts
	parts := strings.Split(absPath, string(filepath.Separator))
	for _, part := range parts {
		if part == ".prompts" || part == "prompts" {
			return true
		}
	}

	return false
}

// loadSubdirFiles loads files from a specific directory (for tree view expansion)
func (m *model) loadSubdirFiles(dirPath string) []fileItem {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return []fileItem{}
	}

	var dirs, files []fileItem

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		// Use Lstat to detect symlinks (doesn't follow them)
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}

		// Detect if this is a symlink
		isSymlink := info.Mode()&os.ModeSymlink != 0

		// Skip hidden files unless showHidden is true
		// Exception: Always show symlinks, even if they start with .
		if !m.showHidden && strings.HasPrefix(entry.Name(), ".") && !isSymlink {
			// Exception: Always show important development folders
			importantFolders := []string{".claude", ".git", ".vscode", ".github", ".config", ".docker", ".prompts", ".codex", ".copilot", ".devcontainer", ".gemini", ".opencode"}
			isImportantFolder := false
			for _, folder := range importantFolders {
				if entry.Name() == folder {
					isImportantFolder = true
					break
				}
			}

			// Exception: Always show secrets/credentials files (security awareness for AI usage)
			isSecretsFileFlag := isSecretsFile(entry.Name())

			// Exception: Always show ignore files (.gitignore, etc. - know what's excluded)
			isIgnoreFileFlag := isIgnoreFile(entry.Name())

			// Exception: If we're inside these folders, show all files
			inImportantFolder := strings.Contains(dirPath, "/.claude") ||
				strings.Contains(dirPath, "/.git") ||
				strings.Contains(dirPath, "/.vscode") ||
				strings.Contains(dirPath, "/.github") ||
				strings.Contains(dirPath, "/.config") ||
				strings.Contains(dirPath, "/.docker") ||
				strings.Contains(dirPath, "/.prompts") ||
				strings.Contains(dirPath, "/.codex") ||
				strings.Contains(dirPath, "/.copilot") ||
				strings.Contains(dirPath, "/.devcontainer") ||
				strings.Contains(dirPath, "/.gemini") ||
				strings.Contains(dirPath, "/.opencode")

			if !isImportantFolder && !inImportantFolder && !isSecretsFileFlag && !isIgnoreFileFlag {
				continue
			}
		}
		var symlinkTarget string
		var targetIsDir bool

		if isSymlink {
			// Read symlink target
			target, err := os.Readlink(fullPath)
			if err == nil {
				symlinkTarget = target
				// Check if target is a directory by following the symlink
				targetInfo, err := os.Stat(fullPath)
				if err == nil {
					targetIsDir = targetInfo.IsDir()
				}
			}
		}

		item := fileItem{
			name:          entry.Name(),
			path:          fullPath,
			isDir:         targetIsDir || entry.IsDir(),
			size:          info.Size(),
			modTime:       info.ModTime(),
			mode:          info.Mode(),
			isSymlink:     isSymlink,
			symlinkTarget: symlinkTarget,
		}

		if item.isDir {
			dirs = append(dirs, item)
		} else {
			files = append(files, item)
		}
	}

	// Sort alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
	})

	result := append(dirs, files...)
	return result
}

// loadFiles loads the files from the current directory
func (m *model) loadFiles() {
	// Update file watcher to track the current directory
	m.switchWatchPath(m.currentPath)

	// Auto-exit agent view if user navigated outside the .claude/projects/ directory
	if m.showAgentView {
		homeDir, _ := os.UserHomeDir()
		claudeProjectsDir := filepath.Join(homeDir, ".claude", "projects")
		if !strings.HasPrefix(m.currentPath, claudeProjectsDir) {
			m.showAgentView = false
			m.sortBy = m.agentViewRestoreSort
			m.sortAsc = m.agentViewRestoreAsc
			m.displayMode = m.agentViewRestoreMode
			m.agentViewRestore = ""
			m.calculateLayout()
		}
	}

	// Clear prompts directory cache when reloading files (performance optimization)
	// This ensures cache stays fresh when files change
	m.promptDirsCache = make(map[string]bool)

	// Special handling for trash view
	if m.showTrashOnly {
		trashItems, err := getTrashItems()
		if err != nil {
			m.files = []fileItem{}
			m.setStatusMessage(fmt.Sprintf("Error loading trash: %v", err), true)
			return
		}

		// Convert trash items to file items for display
		m.files = convertTrashItemsToFileItems(trashItems)
		m.trashItems = trashItems // Cache for later use
		return
	}

	// SECURITY: Validate and clean the path to prevent directory traversal attacks
	// This prevents malicious navigation to sensitive system directories
	cleanPath, err := filepath.Abs(m.currentPath)
	if err != nil {
		m.files = []fileItem{}
		m.setStatusMessage(fmt.Sprintf("Error: Invalid path: %v", err), true)
		return
	}

	// Clean the path to resolve .. and . components
	cleanPath = filepath.Clean(cleanPath)

	// Optional: Restrict navigation to home directory or initial working directory
	// This can be made configurable via a --allow-full-access flag if needed
	homeDir, _ := os.UserHomeDir()
	initialWD, _ := os.Getwd()

	// Allow access if path is under home directory OR under initial working directory
	allowedByHome := homeDir != "" && strings.HasPrefix(cleanPath, homeDir)
	allowedByWD := initialWD != "" && strings.HasPrefix(cleanPath, initialWD)

	if !allowedByHome && !allowedByWD {
		// Check if we're trying to access system directories
		restrictedPrefixes := []string{"/etc", "/root", "/boot", "/sys", "/proc"}
		for _, prefix := range restrictedPrefixes {
			if strings.HasPrefix(cleanPath, prefix) {
				m.files = []fileItem{}
				m.setStatusMessage(fmt.Sprintf("Access denied: %s (restricted system directory)", cleanPath), true)
				// Revert to home directory for safety
				if homeDir != "" {
					m.currentPath = homeDir
					m.loadFiles() // Recursive call with safe path
				}
				return
			}
		}
	}

	// Update to the cleaned path
	m.currentPath = cleanPath

	entries, err := os.ReadDir(m.currentPath)
	if err != nil {
		m.files = []fileItem{}
		return
	}

	// Reset files slice
	m.files = []fileItem{}

	// Add parent directory if not at root
	if m.currentPath != "/" {
		parentPath := filepath.Dir(m.currentPath)
		parentItem := fileItem{
			name:  "..",
			path:  parentPath,
			isDir: true,
		}

		// Get parent directory's actual modification time
		if info, err := os.Stat(parentPath); err == nil {
			parentItem.modTime = info.ModTime()
			parentItem.size = info.Size()
			parentItem.mode = info.Mode()
		}

		m.files = append(m.files, parentItem)
	}

	// Add directories first, then files
	var dirs, files []fileItem

	for _, entry := range entries {
		fullPath := filepath.Join(m.currentPath, entry.Name())

		// Use Lstat to detect symlinks (doesn't follow them)
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue // Skip files we can't stat
		}

		// Detect if this is a symlink
		isSymlink := info.Mode()&os.ModeSymlink != 0

		// Skip hidden files starting with . (unless showHidden is true)
		// Exception: Always show symlinks, even if they start with .
		if !m.showHidden && strings.HasPrefix(entry.Name(), ".") && !isSymlink {
			// Exception: Always show important development folders
			importantFolders := []string{".claude", ".git", ".vscode", ".github", ".config", ".docker", ".prompts", ".codex", ".copilot", ".devcontainer", ".gemini", ".opencode"}
			isImportantFolder := false
			for _, folder := range importantFolders {
				if entry.Name() == folder {
					isImportantFolder = true
					break
				}
			}

			// Exception: Always show secrets/credentials files (security awareness for AI usage)
			isSecretsFileFlag := isSecretsFile(entry.Name())

			// Exception: Always show ignore files (.gitignore, etc. - know what's excluded)
			isIgnoreFileFlag := isIgnoreFile(entry.Name())

			// Exception: If we're inside these folders, show all files
			inImportantFolder := strings.Contains(m.currentPath, "/.claude") ||
				strings.Contains(m.currentPath, "/.git") ||
				strings.Contains(m.currentPath, "/.vscode") ||
				strings.Contains(m.currentPath, "/.github") ||
				strings.Contains(m.currentPath, "/.config") ||
				strings.Contains(m.currentPath, "/.docker") ||
				strings.Contains(m.currentPath, "/.prompts") ||
				strings.Contains(m.currentPath, "/.codex") ||
				strings.Contains(m.currentPath, "/.copilot") ||
				strings.Contains(m.currentPath, "/.devcontainer") ||
				strings.Contains(m.currentPath, "/.gemini") ||
				strings.Contains(m.currentPath, "/.opencode")

			if !isImportantFolder && !inImportantFolder && !isSecretsFileFlag && !isIgnoreFileFlag {
				continue
			}
		}
		var symlinkTarget string
		var targetIsDir bool

		if isSymlink {
			// Read symlink target
			target, err := os.Readlink(fullPath)
			if err == nil {
				symlinkTarget = target
				// Check if target is a directory by following the symlink
				targetInfo, err := os.Stat(fullPath)
				if err == nil {
					targetIsDir = targetInfo.IsDir()
				}
			}
		}

		item := fileItem{
			name:          entry.Name(),
			path:          fullPath,
			isDir:         targetIsDir || entry.IsDir(), // Use target's directory status if symlink
			size:          info.Size(),
			modTime:       info.ModTime(),
			mode:          info.Mode(),
			isSymlink:     isSymlink,
			symlinkTarget: symlinkTarget,
		}

		// Cache variable check for prompt files (performance optimization)
		// Only check if: not a directory, is a prompt-type file, and we're viewing prompts
		if !item.isDir && m.showPromptsOnly {
			ext := strings.ToLower(filepath.Ext(item.name))
			if ext == ".prompty" || ext == ".md" || ext == ".txt" {
				hasVars := hasPromptVariables(fullPath)
				item.hasVariables = &hasVars
			}
		}

		if item.isDir {
			dirs = append(dirs, item)
		} else {
			files = append(files, item)
		}
	}

	// Sort alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
	})

	// In agent view, filter to only show session files (.jsonl) and session directories
	if m.showAgentView {
		var filteredDirs []fileItem
		for _, d := range dirs {
			// Keep session UUID directories, skip memory/ and other non-session dirs
			if d.name == "memory" || d.name == "subagents" {
				continue
			}
			filteredDirs = append(filteredDirs, d)
		}
		var filteredFiles []fileItem
		for _, f := range files {
			// Only show .jsonl session files
			if strings.HasSuffix(f.name, ".jsonl") {
				filteredFiles = append(filteredFiles, f)
			}
		}
		dirs = filteredDirs
		files = filteredFiles
	}

	m.files = append(m.files, dirs...)
	m.files = append(m.files, files...)

	// Reset cursor if out of bounds
	if m.cursor >= len(m.files) {
		m.cursor = 0
	}

	// Apply sorting based on sortBy and sortAsc settings
	m.sortFiles()

	// Populate agent metadata for agent view
	if m.showAgentView {
		m.populateAgentMetadata()
	}

	// Rebuild combined command history for current directory
	// This ensures Up/Down arrows show directory-specific commands first
	m.rebuildCombinedHistory()
}

// sortFiles sorts the file list based on sortBy and sortAsc settings
// Always keeps ".." parent directory at the top
// When sorting by name: keeps folders grouped before files (traditional behavior)
// When sorting by other criteria: mixes folders and files
func (m *model) sortFiles() {
	// When in git repos mode, sort gitReposList instead of files
	if m.showGitReposOnly {
		m.sortGitReposList()
		return
	}

	if len(m.files) <= 1 {
		return
	}

	// Separate parent directory (..) from other files
	var parentDir *fileItem
	var otherFiles []fileItem

	for i := range m.files {
		if m.files[i].name == ".." {
			parentDir = &m.files[i]
		} else {
			otherFiles = append(otherFiles, m.files[i])
		}
	}

	// When sorting by name, keep folders grouped before files (traditional behavior)
	if m.sortBy == "name" {
		var dirs, files []fileItem

		// Separate directories from files
		for _, item := range otherFiles {
			if item.isDir {
				dirs = append(dirs, item)
			} else {
				files = append(files, item)
			}
		}

		// Sort directories alphabetically
		sort.Slice(dirs, func(i, j int) bool {
			less := strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name)
			if !m.sortAsc {
				less = !less
			}
			return less
		})

		// Sort files alphabetically
		sort.Slice(files, func(i, j int) bool {
			less := strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
			if !m.sortAsc {
				less = !less
			}
			return less
		})

		// Reconstruct: parent dir, then folders, then files
		m.files = make([]fileItem, 0, len(m.files))
		if parentDir != nil {
			m.files = append(m.files, *parentDir)
		}
		m.files = append(m.files, dirs...)
		m.files = append(m.files, files...)
		return
	}

	// For other sort criteria (size, modified, type): mix folders and files
	sort.Slice(otherFiles, func(i, j int) bool {
		a, b := otherFiles[i], otherFiles[j]

		// Determine sort result based on sortBy
		var less bool
		switch m.sortBy {
		case "size":
			// For directories, compare by item count
			// For files, compare by size
			aSize := a.size
			bSize := b.size
			if a.isDir {
				aSize = int64(getDirItemCount(a.path))
			}
			if b.isDir {
				bSize = int64(getDirItemCount(b.path))
			}
			if aSize == bSize {
				// If same size, sort by name as secondary
				less = strings.ToLower(a.name) < strings.ToLower(b.name)
			} else {
				less = aSize < bSize
			}

		case "modified":
			if a.modTime.Equal(b.modTime) {
				// If same time, sort by name as secondary
				less = strings.ToLower(a.name) < strings.ToLower(b.name)
			} else {
				less = a.modTime.Before(b.modTime)
			}

		case "type":
			aType := getFileType(a)
			bType := getFileType(b)
			if aType == bType {
				// If same type, sort by name as secondary
				less = strings.ToLower(a.name) < strings.ToLower(b.name)
			} else {
				less = aType < bType
			}

		case "branch":
			// Sort by git branch name (for git repos view)
			aBranch := a.gitBranch
			bBranch := b.gitBranch
			if aBranch == bBranch {
				// If same branch, sort by name as secondary
				less = strings.ToLower(a.name) < strings.ToLower(b.name)
			} else {
				less = strings.ToLower(aBranch) < strings.ToLower(bBranch)
			}

		case "status":
			// Sort by git status (for git repos view)
			// Priority: dirty repos first, then ahead/behind status, then clean
			aStatus := getGitStatusSortValue(a)
			bStatus := getGitStatusSortValue(b)
			if aStatus == bStatus {
				// If same status, sort by name as secondary
				less = strings.ToLower(a.name) < strings.ToLower(b.name)
			} else {
				less = aStatus < bStatus
			}

		default:
			// Fallback to name sorting
			less = strings.ToLower(a.name) < strings.ToLower(b.name)
		}

		// Apply sort direction (ascending vs descending)
		if !m.sortAsc {
			less = !less
		}

		return less
	})

	// Reconstruct files slice with parent directory at top (if present)
	m.files = make([]fileItem, 0, len(m.files))
	if parentDir != nil {
		m.files = append(m.files, *parentDir)
	}
	m.files = append(m.files, otherFiles...)
}

// loadPreview loads the content of a file for preview
func (m *model) loadPreview(path string) {
	m.preview.filePath = path
	m.preview.fileName = filepath.Base(path)
	m.preview.scrollPos = 0
	m.preview.loaded = false
	m.preview.isBinary = false
	m.preview.tooLarge = false
	m.preview.isMarkdown = false
	m.preview.isSyntaxHighlighted = false
	// Clear any existing kitty graphics from the terminal before loading new preview
	if m.preview.hasGraphicsProtocol {
		fmt.Print(clearKittyGraphics())
	}
	m.preview.hasGraphicsProtocol = false
	m.preview.isPrompt = false
	m.preview.promptTemplate = nil
	m.preview.isJSONL = false
	m.preview.cachedJSONLMessages = nil
	// Invalidate cache when loading new file
	m.preview.cacheValid = false
	m.preview.cachedWrappedLines = nil
	m.preview.cachedRenderedContent = ""
	m.preview.cachedLineCount = 0

	// Check if this is a symlink using Lstat (doesn't follow the link)
	linfo, err := os.Lstat(path)
	if err == nil && linfo.Mode()&os.ModeSymlink != 0 {
		// This is a symlink - show symlink information
		target, err := os.Readlink(path)
		content := []string{
			"🌀 Symbolic Link (Shortcut)",
			"",
			"Link Name: " + filepath.Base(path),
			"",
		}

		if err != nil {
			// Can't read symlink target
			content = append(content, "❌ Error: Cannot read symlink target")
			content = append(content, fmt.Sprintf("   %v", err))
		} else {
			// Successfully read target
			content = append(content, "Points to: "+target)

			// Resolve relative paths to absolute for clarity
			absTarget := target
			if !filepath.IsAbs(target) {
				absTarget = filepath.Join(filepath.Dir(path), target)
			}

			// Check if target exists and get its info
			targetInfo, statErr := os.Stat(path) // Stat follows the link

			if statErr != nil {
				// Broken symlink
				content = append(content, "")
				content = append(content, "❌ Status: BROKEN LINK")
				content = append(content, "   Target does not exist or is not accessible")
				content = append(content, "")
				content = append(content, "Absolute path: "+absTarget)
			} else {
				// Valid symlink
				content = append(content, "")
				content = append(content, "✅ Status: Valid")
				content = append(content, "")

				if targetInfo.IsDir() {
					content = append(content, "Target Type: Directory")
					content = append(content, "")
					content = append(content, "💡 Tip: Press Enter to navigate into this directory")
				} else {
					content = append(content, "Target Type: File")
					content = append(content, "Size: "+formatFileSize(targetInfo.Size()))
					content = append(content, "Modified: "+formatModTime(targetInfo.ModTime()))
					content = append(content, "")
					content = append(content, "💡 Tip: Press Enter to view the target file's contents")
				}
			}
		}

		m.preview.content = content
		m.preview.loaded = true
		m.preview.fileSize = 0
		return
	}

	// Get file info (follows symlinks if any)
	info, err := os.Stat(path)
	if err != nil {
		m.preview.content = []string{
			fmt.Sprintf("Error reading file: %v", err),
		}
		m.preview.loaded = true
		return
	}

	m.preview.fileSize = info.Size()

	// JSONL conversation files: read last N lines (tail) regardless of size
	if isJSONLFile(path) {
		m.loadJSONLPreview(path, info.Size())
		return
	}

	// Check if file is too large (>1MB)
	const maxSize = 1024 * 1024 // 1MB
	if info.Size() > maxSize {
		m.preview.tooLarge = true
		m.preview.content = []string{
			"File too large to preview",
			fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
			"",
			"Press 'E' to edit in external editor",
		}
		m.preview.loaded = true
		return
	}

	// Check for CSV files (text-based, but better viewed with specialized tools)
	if isCSVFile(path) {
		// Still read as text, but add a helpful hint
		content, err := os.ReadFile(path)
		if err != nil {
			m.preview.content = []string{
				fmt.Sprintf("Error reading file: %v", err),
			}
			m.preview.loaded = true
			return
		}

		lines := strings.Split(string(content), "\n")
		// Add header with hint based on tool availability
		var hintLines []string
		if getAvailableCSVViewer() != "" {
			hintLines = []string{
				"📈 CSV/Spreadsheet file",
				fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
				"",
				"✨ Press F4 to open in VisiData (interactive spreadsheet viewer)",
				"",
				"─────────────────────────────────────",
				"",
			}
		} else {
			hintLines = []string{
				"📈 CSV/Spreadsheet file",
				fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
				"",
				"💡 Tip: Install VisiData for better viewing:",
				"   sudo apt install visidata",
				"   or: pipx install visidata",
				"",
				"─────────────────────────────────────",
				"",
			}
		}
		m.preview.content = append(hintLines, lines...)
		m.preview.loaded = true
		return
	}

	// Check if binary
	if isBinaryFile(path) {
		m.preview.isBinary = true
		var content []string

		// Specific file type detection with helpful hints
		if isPDFFile(path) {
			if getAvailablePDFViewer() != "" {
				content = []string{
					"📕 PDF Document",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview PDF files in text mode",
					"",
					"Options:",
					"  ✨ Press F4 to view in timg (terminal PDF viewer)",
					"  • Press F3 to open in browser",
				}
			} else {
				content = []string{
					"📕 PDF Document",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview PDF files in text mode",
					"",
					"Options:",
					"  • Press F3 to open in browser",
					"",
					"💡 Or install a terminal PDF viewer:",
					"   sudo apt install timg",
					"   or: brew install timg",
				}
			}
		} else if isVideoFile(path) {
			if getAvailableVideoPlayer() != "" {
				content = []string{
					"🎬 Video File",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview video files",
					"",
					"✨ Press F4 to play in mpv",
				}
			} else {
				content = []string{
					"🎬 Video File",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview video files",
					"",
					"💡 Install mpv to play videos:",
					"   sudo apt install mpv",
					"   or: brew install mpv",
				}
			}
		} else if isAudioFile(path) {
			if getAvailableAudioPlayer() != "" {
				content = []string{
					"🎵 Audio File",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview audio files",
					"",
					"✨ Press F4 to play in mpv",
				}
			} else {
				content = []string{
					"🎵 Audio File",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview audio files",
					"",
					"💡 Install mpv to play audio:",
					"   sudo apt install mpv",
					"   or: brew install mpv",
				}
			}
		} else if isDatabaseFile(path) {
			if getAvailableDatabaseViewer() != "" {
				content = []string{
					"🗄 SQLite Database",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview database files",
					"",
					"✨ Press F4 to open in harlequin (database viewer)",
				}
			} else {
				content = []string{
					"🗄 SQLite Database",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview database files",
					"",
					"💡 Install a database viewer:",
					"   pipx install harlequin",
					"   or: pip install litecli",
				}
			}
		} else if isArchiveFile(path) {
			content = []string{
				"🗜 Archive File",
				fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
				"",
				"Cannot preview archive contents",
				"",
				"💡 Extract to view contents:",
				"   unzip filename.zip",
				"   tar -xzf filename.tar.gz",
				"   7z x filename.7z",
			}
		} else if isImageFile(path) {
			// Try HD terminal graphics rendering first (Kitty/iTerm2/Sixel)
			// Calculate dimensions for preview pane (approximate)
			// These will be adjusted based on actual pane size when rendering
			maxWidth := 80  // Will be refined based on preview pane width
			maxHeight := 30 // Will be refined based on preview pane height

			hdImageData, success := renderImageWithProtocol(path, maxWidth, maxHeight)
			if success {
				// HD image rendering succeeded - store the rendered data
				// Split into lines for preview rendering
				imageLines := strings.Split(strings.TrimRight(hdImageData, "\n"), "\n")

				// Add header info with fallback options
				protocolName := getProtocolName()
				header := []string{
					fmt.Sprintf("🖼 Image File (HD Preview via %s)", protocolName),
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
				}
				content = append(header, imageLines...)

				// Add footer with fallback viewing options
				// Useful if protocol doesn't work (e.g., WezTerm in WSL) or user prefers external viewer
				imageViewer := getAvailableImageViewer()
				footer := []string{""}
				if imageViewer != "" {
					footer = append(footer, fmt.Sprintf("💡 Alternative: Press V to view in %s", imageViewer))
				}
				footer = append(footer, "   Press F3 to open in browser")
				content = append(content, footer...)

				// Set flag to prevent wrapping of graphics protocol escape sequences
				m.preview.hasGraphicsProtocol = true
			} else {
				// Fall back to message if no protocol support
				imageViewer := getAvailableImageViewer()
				if imageViewer != "" {
					content = []string{
						"🖼 Image File",
						fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
						"",
						"Cannot preview in text mode",
						"",
						"Options:",
						fmt.Sprintf("  ✨ Press V to view in %s (terminal image viewer)", imageViewer),
						"  • Press F3 to open in browser",
						"",
						"💡 For HD inline previews, use WezTerm or Kitty terminal",
					}
				} else {
					content = []string{
						"🖼 Image File",
						fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
						"",
						"Cannot preview in text mode",
						"",
						"Options:",
						"  • Press F3 to open in browser",
						"",
						"💡 Install a terminal image viewer:",
						"   sudo apt install timg (or: brew install timg)",
						"   or: cargo install viu",
						"",
						"💡 For HD inline previews, use WezTerm or Kitty terminal",
					}
				}
			}
		} else {
			// Generic binary file
			if getAvailableHexViewer() != "" {
				content = []string{
					"⚙ Binary File",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview binary files",
					"",
					"✨ Press F4 to view in hexyl (hex viewer)",
				}
			} else {
				content = []string{
					"⚙ Binary File",
					fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
					"",
					"Cannot preview binary files",
					"",
					"💡 Install hexyl to view as hex:",
					"   cargo install hexyl",
					"   or: sudo apt install hexyl",
				}
			}
		}

		m.preview.content = content
		m.preview.loaded = true
		return
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		m.preview.content = []string{
			fmt.Sprintf("Error reading file: %v", err),
		}
		m.preview.loaded = true
		return
	}

	// Check if this is a prompt file and parse it
	fileItem := fileItem{
		name:  filepath.Base(path),
		path:  path,
		isDir: false,
	}
	if isPromptFile(fileItem) {
		tmpl, err := parsePromptFile(path)
		if err == nil {
			// Successfully parsed as prompt
			m.preview.isPrompt = true
			m.preview.promptTemplate = tmpl

			// Get context variables
			contextVars := getContextVariables(m)

			// Render the template with variables substituted
			rendered := renderPromptTemplate(tmpl, contextVars)

			// Split into lines for display
			lines := strings.Split(rendered, "\n")
			m.preview.content = lines
			m.preview.loaded = true

			// Mark as markdown if it's a .md file, so it can be rendered with Glamour
			// when prompts mode is off (and inputFieldsActive == false)
			if isMarkdownFile(path) {
				m.preview.isMarkdown = true
			}

			// Initialize inline editing state for prompt variables
			m.filledVariables = make(map[string]string)
			m.promptEditMode = false
			m.focusedVariableIndex = 0

			// Populate cache for better scroll performance
			m.populatePreviewCache()
			return
		}
		// If parsing failed, fall through to regular preview
	}

	// Check if markdown and render with Glamour
	if isMarkdownFile(path) {
		m.preview.isMarkdown = true
		// Use Glamour to render markdown with appropriate width
		// We'll set width in the render function based on available space
		// For now, just store the raw content and mark as markdown
		lines := strings.Split(string(content), "\n")
		m.preview.content = lines
		m.preview.loaded = true

		// DON'T populate cache here - let caller do it after setting view mode
		// This prevents rendering with wrong width (e.g., m.rightWidth=0 in single-pane)
		// Caller should call populatePreviewCache() after setting viewMode correctly
		return
	}

	// Try syntax highlighting for code files
	highlighted, ok := highlightCode(string(content), path)
	var lines []string

	if ok {
		// Syntax highlighting succeeded
		lines = strings.Split(highlighted, "\n")
		m.preview.isSyntaxHighlighted = true
	} else {
		// Fallback to plain text
		lines = strings.Split(string(content), "\n")
		m.preview.isSyntaxHighlighted = false
	}

	// Limit number of lines
	if len(lines) > m.preview.maxPreview {
		lines = lines[:m.preview.maxPreview]
		lines = append(lines, "", fmt.Sprintf("... (truncated after %d lines)", m.preview.maxPreview))
	}

	m.preview.content = lines
	m.preview.loaded = true

	// Populate cache for better scroll performance
	m.populatePreviewCache()
}

// populatePreviewCache pre-computes and caches wrapped/rendered content for better scroll performance
func (m *model) populatePreviewCache() {
	if !m.preview.loaded {
		return
	}

	// Calculate available width - must match renderPreview() logic exactly
	var availableWidth int
	var boxContentWidth int

	if m.viewMode == viewFullPreview {
		boxContentWidth = m.width - 6 // Box content width in full preview
	} else {
		boxContentWidth = m.rightWidth - 2 // Box content width in dual-pane (accounting for borders)
	}

	if m.preview.isMarkdown {
		// Markdown: no line numbers or scrollbar, but add left padding for readability
		// Subtract 2 for left padding (prevents code blocks from touching border)
		availableWidth = boxContentWidth - 2
	} else {
		// Regular text: subtract line nums (6) + scrollbar (1) + space (1) = 8 chars
		availableWidth = boxContentWidth - 8
	}

	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}

	// Cache markdown rendering
	if m.preview.isMarkdown {
		// Safety: skip Glamour for very large markdown files (can cause hangs with complex content)
		// For files > 2000 lines, treat as plain text to avoid performance issues
		const maxMarkdownLines = 2000

		if len(m.preview.content) > maxMarkdownLines {
			// Too large for Glamour - treat as plain text
			m.preview.isMarkdown = false
			// Fall through to regular text wrapping below
		} else {
			markdownContent := strings.Join(m.preview.content, "\n")

			// Render with timeout to prevent hangs
			// Use longer timeout on mobile (Termux) where rendering is slower
			timeout := 5 * time.Second
			if editorAvailable("termux-clipboard-set") {
				timeout = 15 * time.Second // Mobile devices need more time
			}
			rendered, err := m.renderMarkdownWithTimeout(markdownContent, availableWidth, timeout)

			if err == nil {
				// Store rendered content even if empty (Glamour might return empty for some valid markdown)
				m.preview.cachedRenderedContent = rendered
				renderedLines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
				m.preview.cachedLineCount = len(renderedLines)
				m.preview.cachedWidth = availableWidth
				m.preview.cacheValid = true
				return
			} else {
				// Glamour render failed (error or timeout) - treat as plain text
				m.preview.isMarkdown = false
				// Log the error for debugging (appears in status message)
				if strings.Contains(err.Error(), "timeout") {
					m.setStatusMessage("Markdown rendering timed out, showing as plain text", true)
				}
			}
			// Fall through to regular text wrapping
		}
	}

	// Cache wrapped text lines
	var wrappedLines []string
	for _, line := range m.preview.content {
		wrapped := wrapLine(line, availableWidth)
		wrappedLines = append(wrappedLines, wrapped...)
	}
	m.preview.cachedWrappedLines = wrappedLines
	m.preview.cachedLineCount = len(wrappedLines)
	m.preview.cachedWidth = availableWidth
	m.preview.cacheValid = true
}

// renderMarkdownWithTimeout renders markdown with a timeout to prevent hangs
// Returns rendered content and any error (including timeout)
func (m *model) renderMarkdownWithTimeout(content string, width int, timeout time.Duration) (string, error) {
	type renderResult struct {
		rendered string
		err      error
	}

	// Use buffered channel to prevent goroutine leak
	resultChan := make(chan renderResult, 1)

	go func() {
		// Recover from panics in glamour rendering
		defer func() {
			if r := recover(); r != nil {
				resultChan <- renderResult{
					rendered: "",
					err:      fmt.Errorf("markdown rendering panicked: %v", r),
				}
			}
		}()

		// Check if we have a cached renderer for this width
		var renderer *glamour.TermRenderer
		var err error

		if m.glamourRenderer != nil && m.glamourRendererWidth == width {
			// Reuse cached renderer (avoids expensive terminal probing!)
			renderer = m.glamourRenderer.(*glamour.TermRenderer)
		} else {
			// Create new renderer and cache it
			// First try custom style file, fall back to "dark" if not found
			exePath, _ := os.Executable()
			exeDir := filepath.Dir(exePath)
			customStylePath := filepath.Join(exeDir, "styles", "tfe.json")

			// Check if custom style exists, otherwise use built-in dark style
			if _, statErr := os.Stat(customStylePath); statErr == nil {
				renderer, err = glamour.NewTermRenderer(
					glamour.WithStylePath(customStylePath),
					glamour.WithWordWrap(width),
				)
			} else {
				// Match markdown rendering to the active theme mode. applyThemeMode()
				// already synchronizes Lip Gloss background detection with CLI/config.
				glamourStyle := "dark"
				if !lipgloss.HasDarkBackground() {
					glamourStyle = "light"
				}
				renderer, err = glamour.NewTermRenderer(
					glamour.WithStandardStyle(glamourStyle),
					glamour.WithWordWrap(width),
				)
			}
			if err != nil {
				resultChan <- renderResult{rendered: "", err: err}
				return
			}

			// Cache the renderer for future use
			m.glamourRenderer = renderer
			m.glamourRendererWidth = width
		}

		rendered, err := renderer.Render(content)
		resultChan <- renderResult{rendered: rendered, err: err}
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result.rendered, result.err
	case <-time.After(timeout):
		return "", fmt.Errorf("markdown rendering timeout after %v", timeout)
	}
}

// renderMarkdownAsync renders markdown in a background goroutine
func renderMarkdownAsync(m *model) tea.Cmd {
	return func() tea.Msg {
		// Populate cache (includes Glamour rendering)
		m.populatePreviewCache()
		return markdownRenderedMsg{}
	}
}

// setStatusMessage sets a temporary status message with auto-dismiss
func (m *model) setStatusMessage(message string, isError bool) {
	m.statusMessage = message
	m.statusIsError = isError
	m.statusTime = time.Now()
}

// statusTimeoutCmd returns a command that triggers a redraw after the status message expires
func statusTimeoutCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return statusTimeoutMsg{}
	})
}

// statusTimeoutMsg is sent when the status message should be cleared
type statusTimeoutMsg struct{}

// createDirectory creates a new directory in the current path
func (m *model) createDirectory(name string) error {
	// Validate name (no /, \, special chars)
	if strings.ContainsAny(name, "/\\:*?\"<>|") {
		return fmt.Errorf("invalid characters in directory name")
	}
	if name == "" || name == "." || name == ".." {
		return fmt.Errorf("invalid directory name")
	}

	// Create directory in current path
	path := filepath.Join(m.currentPath, name)
	if err := os.Mkdir(path, 0755); err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("directory already exists")
		}
		return err
	}

	return nil
}

// deleteFileOrDir deletes a file or directory
func (m *model) deleteFileOrDir(path string, isDir bool) error {
	// Check if exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found")
		}
		return err
	}

	// Move to trash instead of permanent delete for safety
	if err := moveToTrash(path); err != nil {
		return fmt.Errorf("failed to move to trash: %w", err)
	}

	return nil
}

// permanentDeleteFileOrDir permanently deletes a file without moving to trash
// Used for emptying trash or when explicitly requested
func (m *model) permanentDeleteFileOrDir(path string, isDir bool) error {
	// Check if exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found")
		}
		return err
	}

	if isDir {
		// For directories, use RemoveAll to handle non-empty directories
		return os.RemoveAll(path)
	}

	// Delete file
	// Check if file is writable
	if info.Mode()&0200 == 0 {
		return fmt.Errorf("file is read-only")
	}

	return os.Remove(path)
}

// filterFilesBySearch returns indices of files matching the search query
// Case-insensitive substring matching on file names
func (m *model) filterFilesBySearch(query string) []int {
	if query == "" {
		// Empty query - return all indices
		indices := make([]int, len(m.files))
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	queryLower := strings.ToLower(query)
	var matchingIndices []int

	for i, file := range m.files {
		// Skip parent directory (..) - always show it
		if file.name == ".." {
			matchingIndices = append(matchingIndices, i)
			continue
		}

		// Case-insensitive substring match
		if strings.Contains(strings.ToLower(file.name), queryLower) {
			matchingIndices = append(matchingIndices, i)
		}
	}

	return matchingIndices
}

// copyFile copies a file from src to dst
// Handles both files and directories (recursive copy)
func (m *model) copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if srcInfo.IsDir() {
		return copyDirectory(src, dst)
	}

	return copyFileContent(src, dst)
}

// copyFileContent copies a single file
func copyFileContent(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy: %w", err)
	}

	// Preserve permissions
	srcInfo, _ := os.Stat(src)
	return os.Chmod(dst, srcInfo.Mode())
}

// copyDirectory recursively copies a directory
func copyDirectory(src, dst string) error {
	// Create destination directory
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFileContent(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
