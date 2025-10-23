package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/mattn/go-runewidth"
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

// isObsidianVault checks if a directory is an Obsidian vault (contains .obsidian folder)
func isObsidianVault(path string) bool {
	obsidianPath := filepath.Join(path, ".obsidian")
	info, err := os.Stat(obsidianPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// isDirEmpty checks if a directory is empty (no files or subdirectories)
func isDirEmpty(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false // Can't read, assume not empty
	}
	return len(entries) == 0
}

// getDirItemCount returns the number of items in a directory
func getDirItemCount(path string) int {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}
	return len(entries)
}

// getFileIcon returns the appropriate emoji icon based on file type
func getFileIcon(item fileItem) string {
	// Check for symlinks first (takes priority over other icons)
	if item.isSymlink {
		return "🌀" // Portal emoji for symlinks
	}

	if item.isDir {
		if item.name == ".." {
			return "⬆️" // Up arrow for parent dir
		}
		// Check if this is the user's home directory
		if homeDir, err := os.UserHomeDir(); err == nil {
			if item.path == homeDir {
				return "🏠" // Home emoji for home directory
			}
		}
		// Virtual global prompts folder - no icon since name already has 🌐
		if isGlobalPromptsVirtualFolder(item.name) {
			return "" // Name already contains 🌐 emoji
		}
		// Virtual global claude folder - no icon since name already has 🤖
		if isGlobalClaudeVirtualFolder(item.name) {
			return "" // Name already contains 🤖 emoji
		}
		// Special folder icons
		switch item.name {
		case ".claude":
			return "🤖" // Robot for Claude config
		case ".codex":
			return "🤖" // Robot for GitHub Codex
		case ".copilot":
			return "🤖" // Robot for GitHub Copilot
		case ".gemini":
			return "🤖" // Robot for Google Gemini
		case ".opencode":
			return "🤖" // Robot for OpenCode
		case ".git":
			return "📦" // Package for git
		case ".vscode":
			return "💻" // Laptop for VS Code
		case ".github":
			return "🐙" // Octopus for GitHub
		case ".docker":
			return "🐳" // Whale for Docker
		case ".devcontainer":
			return "🐳" // Whale for Dev Containers (Docker-based)
		case ".prompts":
			return "📝" // Memo for prompts library
		case "node_modules":
			return "📚" // Books for dependencies
		case "docs", "documentation":
			return "📖" // Open book
		case "src", "source":
			return "📂" // Open folder
		case "test", "tests", "__tests__":
			return "🧪" // Test tube
		case "build", "dist", "out":
			return "📦" // Package
		case "public", "static", "assets":
			return "🌐" // Globe
		case "config", "configs", ".config":
			return "⚙️" // Gear
		case "scripts":
			return "📜" // Scroll
		default:
			// Check if this is an Obsidian vault
			if isObsidianVault(item.path) {
				return "🧠" // Brain emoji for Obsidian vaults
			}
			// Check if folder is empty
			if isDirEmpty(item.path) {
				return "📂" // Open/empty folder
			}
			return "📁" // Regular closed folder (has content)
		}
	}

	// Check for secrets/credentials files (security awareness for AI usage)
	if isSecretsFile(item.name) {
		return "🔒" // Lock for secrets/credentials files
	}

	// Check for ignore files (.gitignore, etc. - know what's excluded)
	if isIgnoreFile(item.name) {
		return "🚫" // Prohibited sign for ignore files
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(item.name))

	// Map extensions to emoji icons
	iconMap := map[string]string{
		// Programming languages
		".go":   "🐹", // Gopher
		".py":   "🐍", // Python snake
		".js":   "🟨", // JavaScript yellow
		".ts":   "🔷", // TypeScript blue diamond
		".jsx":  "⚛️", // React atom
		".tsx":  "⚛️", // React atom
		".rs":   "🦀", // Rust crab
		".c":    "©️", // C copyright symbol
		".cpp":  "➕", // C++ plus
		".h":    "📋", // Header clipboard
		".java": "☕", // Java coffee
		".rb":   "💎", // Ruby gem
		".php":  "🐘", // PHP elephant
		".sh":   "🐚", // Shell
		".bash": "🐚", // Shell
		".lua":  "🌙", // Lua moon
		".r":    "📊", // R statistics

		// Web
		".html": "🌐", // HTML globe
		".css":  "🎨", // CSS art palette
		".scss": "🎨", // SCSS art palette
		".sass": "🎨", // Sass art palette
		".vue":  "💚", // Vue green heart
		".svelte": "🧡", // Svelte orange heart

		// Data/Config
		".json": "📊", // JSON chart
		".yaml": "📄", // YAML document
		".yml":  "📄", // YAML document
		".toml": "📄", // TOML document
		".xml":  "📰", // XML newspaper
		".csv":  "📈", // CSV chart
		".sql":  "🗄️", // SQL database

		// Documents
		".md":  "📝", // Markdown memo
		".txt": "📄", // Text document
		".pdf": "📕", // PDF red book
		".doc": "📘", // DOC blue book
		".docx": "📘", // DOCX blue book

		// Archives
		".zip": "🗜️", // ZIP compression
		".tar": "📦", // TAR package
		".gz":  "🗜️", // GZ compression
		".7z":  "🗜️", // 7Z compression
		".rar": "🗜️", // RAR compression

		// Images
		".png": "🖼️", // PNG frame
		".jpg": "🖼️", // JPG frame
		".jpeg": "🖼️", // JPEG frame
		".gif": "🎞️", // GIF film
		".svg": "🎨", // SVG palette
		".ico": "🖼️", // ICO frame
		".webp": "🖼️", // WebP frame

		// Audio/Video
		".mp3": "🎵", // MP3 music
		".mp4": "🎬", // MP4 movie
		".wav": "🎵", // WAV music
		".avi": "🎬", // AVI movie
		".mkv": "🎬", // MKV movie

		// System/Config
		".env":  "🔐", // ENV lock
		".ini":  "⚙️", // INI gear
		".conf": "⚙️", // CONF gear
		".cfg":  "⚙️", // CFG gear
		".lock": "🔒", // LOCK locked

		// Build/Package
		".gradle": "🐘", // Gradle elephant
		".maven":  "📦", // Maven package
		".npm":    "📦", // NPM package
	}

	// Check for icon mapping
	if icon, ok := iconMap[ext]; ok {
		return icon
	}

	// Check for special files without extension
	switch item.name {
	case "CLAUDE.md", "CLAUDE.local.md":
		return "🤖" // Claude AI
	case "Makefile", "makefile", "GNUmakefile":
		return "🔨" // Build hammer
	case "Dockerfile":
		return "🐳" // Docker whale
	case "docker-compose.yml", "docker-compose.yaml":
		return "🐳" // Docker whale
	case "LICENSE", "LICENSE.txt", "LICENSE.md":
		return "📜" // License scroll
	case "README", "README.md", "README.txt":
		return "📖" // README book
	case ".gitignore", ".gitattributes", ".gitmodules":
		return "🔀" // Git branch
	case "package.json":
		return "📦" // NPM package
	case "package-lock.json":
		return "🔒" // Lock
	case "tsconfig.json":
		return "🔷" // TypeScript
	case "go.mod", "go.sum":
		return "🐹" // Go gopher
	case "Cargo.toml", "Cargo.lock":
		return "🦀" // Rust crab
	case "requirements.txt":
		return "🐍" // Python
	case "Gemfile", "Gemfile.lock":
		return "💎" // Ruby gem
	}

	// Default file marker
	return "📄" // Generic document
}

// formatFileSize returns a human-readable file size
func formatFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%dB", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// formatModTime returns a relative time string
func formatModTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", days)
	} else if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1w ago"
		}
		return fmt.Sprintf("%dw ago", weeks)
	} else if diff < 365*24*time.Hour {
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1mo ago"
		}
		return fmt.Sprintf("%dmo ago", months)
	}
	years := int(diff.Hours() / 24 / 365)
	if years == 1 {
		return "1y ago"
	}
	return fmt.Sprintf("%dy ago", years)
}

// isMarkdownFile checks if a file is markdown
func isMarkdownFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

// isBinaryFile checks if a file is likely binary
func isBinaryFile(path string) bool {
	// Read first 512 bytes to check for binary content
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return false
	}

	// Check for null bytes (common in binary files)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}

	return false
}

// highlightCode applies syntax highlighting to code files using Chroma
// Returns highlighted content and success status
func highlightCode(content, filepath string) (string, bool) {
	var buf bytes.Buffer

	// Try to determine lexer from filename
	lexer := lexers.Match(filepath)
	if lexer == nil {
		// Fallback: analyze content
		lexer = lexers.Analyse(content)
	}
	if lexer == nil {
		// Still nothing, use fallback plain text
		return "", false
	}

	// Configure lexer
	lexer = chroma.Coalesce(lexer)

	// Use terminal256 formatter for better color support
	formatter := formatters.Get("terminal256")
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Use monokai style (works well in dark terminals)
	// Alternative styles: dracula, vim, github, solarized-dark
	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	// Tokenize and format
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return "", false
	}

	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return "", false
	}

	return buf.String(), true
}

// visualWidth calculates the visual width of a string, accounting for tabs and ANSI codes
// This is important for consistent scrollbar alignment and box borders
func visualWidth(s string) int {
	width := 0
	inAnsi := false

	for _, ch := range s {
		// Detect start of ANSI escape sequence
		if ch == '\033' {
			inAnsi = true
			continue
		}

		// Skip characters inside ANSI sequences
		if inAnsi {
			// ANSI sequences end with a letter (A-Z, a-z)
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inAnsi = false
			}
			continue
		}

		// Count visible characters with proper wide character support
		if ch == '\t' {
			// Tabs typically expand to next multiple of 8
			width += 8 - (width % 8)
		} else {
			// Use runewidth to properly handle wide characters (emojis, CJK)
			// Most emojis are width 2, regular ASCII is width 1
			width += runewidth.RuneWidth(ch)
		}
	}
	return width
}

// truncateToWidth truncates a string to fit within a target visual width
func truncateToWidth(s string, targetWidth int) string {
	width := 0
	result := ""
	inAnsi := false

	for _, ch := range s {
		// Handle ANSI escape sequences (don't count toward width)
		if ch == '\033' {
			inAnsi = true
			result += string(ch)
			continue
		}

		if inAnsi {
			result += string(ch)
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inAnsi = false
			}
			continue
		}

		// Calculate character width
		charWidth := 1
		if ch == '\t' {
			charWidth = 8 - (width % 8)
		} else {
			// Use runewidth to properly handle wide characters (emojis, CJK)
			charWidth = runewidth.RuneWidth(ch)
		}

		if width+charWidth > targetWidth {
			// Can't fit this character
			if targetWidth-width >= 3 {
				return result + "..."
			}
			return result
		}

		width += charWidth
		result += string(ch)
	}

	return result
}

// getFileType returns a descriptive file type string based on file extension
func getFileType(item fileItem) string {
	// Check for symlinks first
	if item.isSymlink {
		if item.symlinkTarget != "" {
			// Don't truncate here - let the rendering code handle it based on available width
			return "Link → " + item.symlinkTarget
		}
		return "Symlink"
	}

	if item.isDir {
		return "Folder"
	}

	// Get file extension
	ext := strings.ToLower(filepath.Ext(item.name))

	// Map extensions to descriptive types
	typeMap := map[string]string{
		// Programming languages
		".go":     "Go Source",
		".py":     "Python",
		".js":     "JavaScript",
		".ts":     "TypeScript",
		".jsx":    "React (JSX)",
		".tsx":    "React (TSX)",
		".rs":     "Rust",
		".c":      "C Source",
		".cpp":    "C++",
		".cc":     "C++",
		".cxx":    "C++",
		".h":      "C Header",
		".hpp":    "C++ Header",
		".java":   "Java",
		".rb":     "Ruby",
		".php":    "PHP",
		".sh":     "Shell Script",
		".bash":   "Bash Script",
		".zsh":    "ZSH Script",
		".fish":   "Fish Script",
		".lua":    "Lua",
		".r":      "R Script",
		".swift":  "Swift",
		".kt":     "Kotlin",
		".scala":  "Scala",
		".cs":     "C#",
		".vb":     "Visual Basic",
		".pl":     "Perl",

		// Web
		".html":   "HTML",
		".htm":    "HTML",
		".css":    "CSS",
		".scss":   "SCSS",
		".sass":   "Sass",
		".less":   "Less",
		".vue":    "Vue Component",
		".svelte": "Svelte",

		// Data/Config
		".json":  "JSON",
		".yaml":  "YAML",
		".yml":   "YAML",
		".toml":  "TOML",
		".xml":   "XML",
		".csv":   "CSV",
		".sql":   "SQL",
		".env":   "Environment",
		".ini":   "INI Config",
		".conf":  "Config",
		".cfg":   "Config",
		".properties": "Properties",

		// Documents
		".md":       "Markdown",
		".markdown": "Markdown",
		".txt":      "Text",
		".pdf":      "PDF Document",
		".doc":      "Word Doc",
		".docx":     "Word Doc",
		".rtf":      "Rich Text",
		".odt":      "OpenDocument",

		// Archives
		".zip":  "ZIP Archive",
		".tar":  "TAR Archive",
		".gz":   "GZip Archive",
		".bz2":  "BZip2 Archive",
		".xz":   "XZ Archive",
		".7z":   "7-Zip Archive",
		".rar":  "RAR Archive",
		".tgz":  "TAR.GZ Archive",

		// Images
		".png":  "PNG Image",
		".jpg":  "JPEG Image",
		".jpeg": "JPEG Image",
		".gif":  "GIF Image",
		".svg":  "SVG Image",
		".ico":  "Icon",
		".webp": "WebP Image",
		".bmp":  "Bitmap Image",
		".tiff": "TIFF Image",
		".tif":  "TIFF Image",

		// Audio/Video
		".mp3":  "MP3 Audio",
		".mp4":  "MP4 Video",
		".wav":  "WAV Audio",
		".flac": "FLAC Audio",
		".ogg":  "OGG Audio",
		".avi":  "AVI Video",
		".mkv":  "MKV Video",
		".mov":  "MOV Video",
		".wmv":  "WMV Video",

		// System/Build
		".exe":    "Executable",
		".dll":    "DLL Library",
		".so":     "Shared Library",
		".dylib":  "Dynamic Library",
		".a":      "Static Library",
		".o":      "Object File",
		".lock":   "Lock File",
		".log":    "Log File",
		".tmp":    "Temporary",
		".bak":    "Backup",
		".swp":    "Swap File",

		// Build/Package files
		".gradle": "Gradle",
		".maven":  "Maven",
		".npm":    "NPM",
		".mod":    "Go Module",
		".sum":    "Go Checksum",
		".gem":    "Ruby Gem",
		".whl":    "Python Wheel",
		".deb":    "Debian Package",
		".rpm":    "RPM Package",
	}

	// Check for extension mapping
	if fileType, ok := typeMap[ext]; ok {
		return fileType
	}

	// Check for special files without extension or specific names
	switch item.name {
	case "Makefile", "makefile", "GNUmakefile":
		return "Makefile"
	case "Dockerfile":
		return "Dockerfile"
	case "docker-compose.yml", "docker-compose.yaml":
		return "Docker Compose"
	case "LICENSE", "LICENSE.txt", "LICENSE.md":
		return "License"
	case "README", "README.md", "README.txt":
		return "ReadMe"
	case ".gitignore":
		return "Git Ignore"
	case ".gitattributes":
		return "Git Attributes"
	case ".gitmodules":
		return "Git Modules"
	case "package.json":
		return "NPM Package"
	case "package-lock.json":
		return "NPM Lock"
	case "tsconfig.json":
		return "TS Config"
	case "go.mod":
		return "Go Module"
	case "go.sum":
		return "Go Checksum"
	case "Cargo.toml":
		return "Cargo Config"
	case "Cargo.lock":
		return "Cargo Lock"
	case "requirements.txt":
		return "Python Deps"
	case "Gemfile":
		return "Ruby Gemfile"
	case "Gemfile.lock":
		return "Ruby Lock"
	case "CLAUDE.md", "CLAUDE.local.md":
		return "Claude Context"
	}

	// If extension exists but not mapped, return it
	if ext != "" {
		return strings.TrimPrefix(ext, ".") + " File"
	}

	// No extension - return generic "File"
	return "File"
}

// padToVisualWidth pads a string to a specific visual width using spaces
// This correctly handles emojis and wide characters that take up more than 1 cell
func padToVisualWidth(s string, targetWidth int) string {
	visualWidth := runewidth.StringWidth(s)

	// Fix for variation selector emojis (U+FE0F) which runewidth miscalculates
	// These emojis render as width=2 but runewidth reports width=1
	// Count variation selectors and add them to the width
	variationSelectorCount := strings.Count(s, "\uFE0F")
	visualWidth += variationSelectorCount

	if visualWidth >= targetWidth {
		return s
	}
	padding := targetWidth - visualWidth
	return s + strings.Repeat(" ", padding)
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

	m.files = append(m.files, dirs...)
	m.files = append(m.files, files...)

	// Reset cursor if out of bounds
	if m.cursor >= len(m.files) {
		m.cursor = 0
	}

	// Apply sorting based on sortBy and sortAsc settings
	m.sortFiles()

	// Rebuild combined command history for current directory
	// This ensures Up/Down arrows show directory-specific commands first
	m.rebuildCombinedHistory()
}

// sortFiles sorts the file list based on sortBy and sortAsc settings
// Always keeps ".." parent directory at the top
// When sorting by name: keeps folders grouped before files (traditional behavior)
// When sorting by other criteria: mixes folders and files
func (m *model) sortFiles() {
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
	m.preview.isPrompt = false
	m.preview.promptTemplate = nil
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
			content = append(content, "Points to: " + target)

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
				content = append(content, "Absolute path: " + absTarget)
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
					content = append(content, "Size: " + formatFileSize(targetInfo.Size()))
					content = append(content, "Modified: " + formatModTime(targetInfo.ModTime()))
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

	// Check if binary
	if isBinaryFile(path) {
		m.preview.isBinary = true
		content := []string{
			"Binary file detected",
			fmt.Sprintf("Size: %s", formatFileSize(info.Size())),
			"",
			"Cannot preview binary files",
		}
		// Add hint for image files
		if isImageFile(path) {
			content = append(content, "", "Press 'V' to view image in terminal viewer")
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
			rendered, err := renderMarkdownWithTimeout(markdownContent, availableWidth, timeout)

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
func renderMarkdownWithTimeout(content string, width int, timeout time.Duration) (string, error) {
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

		// Render markdown with custom TFE style
		// First try custom style file, fall back to "dark" if not found
		exePath, _ := os.Executable()
		exeDir := filepath.Dir(exePath)
		customStylePath := filepath.Join(exeDir, "styles", "tfe.json")

		// Check if custom style exists, otherwise use built-in dark style
		var renderer *glamour.TermRenderer
		var err error
		if _, statErr := os.Stat(customStylePath); statErr == nil {
			renderer, err = glamour.NewTermRenderer(
				glamour.WithStylePath(customStylePath),
				glamour.WithWordWrap(width),
			)
		} else {
			// Fall back to auto style (detects light/dark terminal theme)
			renderer, err = glamour.NewTermRenderer(
				glamour.WithAutoStyle(),
				glamour.WithWordWrap(width),
			)
		}
		if err != nil {
			resultChan <- renderResult{rendered: "", err: err}
			return
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
