package main

// Module: file_icons.go
// Purpose: File type classification, icons, and metadata utilities
// Responsibilities:
// - File type detection (text, image, video, audio, etc.)
// - File icon assignment
// - File metadata formatting (size, mod time)
// - Syntax highlighting support

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

// isPromptFile checks if a file is a prompt file (.prompty, .yaml, .md, .txt)
// Only files in special directories (.claude/, ~/.prompts/) are considered prompts
// Exception: .prompty files are always prompts (Microsoft Prompty format)
func isPromptFile(item fileItem) bool {
	if item.isDir {
		return false
	}

	ext := strings.ToLower(filepath.Ext(item.name))

	// .prompty is always a prompt file (Microsoft Prompty format)
	if ext == ".prompty" {
		return true
	}

	// For other extensions, only consider them prompts if in special directories
	if ext == ".md" || ext == ".yaml" || ext == ".yml" || ext == ".txt" {
		// Exclude .claude/agents/ - those are documentation files, not prompt templates
		if strings.Contains(item.path, "/.claude/agents/") {
			return false
		}

		// Check if in .claude/ or any subfolder
		if strings.Contains(item.path, "/.claude/") || strings.HasSuffix(item.path, "/.claude") {
			return true
		}
		// Check if in ~/.prompts/ or any subfolder
		homeDir, _ := os.UserHomeDir()
		promptsDir := filepath.Join(homeDir, ".prompts")
		if strings.HasPrefix(item.path, promptsDir) {
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

// isGitRepo checks if a directory is a git repository (contains .git folder)
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
			return "⬆" // Up arrow for parent dir
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
			return "⚙" // Gear
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

	// Check for prompt files - differentiate templates with variables from plain prompts
	// Uses cached value from loadFiles() to avoid per-frame file reads (performance optimization)
	if ext == ".prompty" || ext == ".md" || ext == ".txt" || ext == ".yaml" || ext == ".yml" {
		// Use cached value if available (populated when showPromptsOnly is true)
		if item.hasVariables != nil {
			if *item.hasVariables {
				return "📝" // Memo with pencil = editable template
			}
			// Has been checked and has no variables
			if ext == ".prompty" || ext == ".yaml" || ext == ".yml" || isInPromptsDirectory(item.path) {
				return "📄" // Document = plain prompt without variables
			}
			// Otherwise fall through to default .md/.txt icons
		}
		// If not cached (showPromptsOnly is false), fall through to default icons
	}

	// Map extensions to emoji icons
	iconMap := map[string]string{
		// Programming languages
		".go":   "🐹", // Gopher
		".py":   "🐍", // Python snake
		".js":   "🟨", // JavaScript yellow
		".ts":   "🔷", // TypeScript blue diamond
		".jsx":  "⚛", // React atom
		".tsx":  "⚛", // React atom
		".rs":   "🦀", // Rust crab
		".c":    "©", // C copyright symbol
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
		".sql":  "🗄", // SQL database

		// Documents
		".md":  "📝", // Markdown memo
		".txt": "📄", // Text document
		".pdf": "📕", // PDF red book
		".doc": "📘", // DOC blue book
		".docx": "📘", // DOCX blue book

		// Archives
		".zip": "🗜", // ZIP compression
		".tar": "📦", // TAR package
		".gz":  "🗜", // GZ compression
		".7z":  "🗜", // 7Z compression
		".rar": "🗜", // RAR compression

		// Images
		".png": "🖼", // PNG frame
		".jpg": "🖼", // JPG frame
		".jpeg": "🖼", // JPEG frame
		".gif": "🎞", // GIF film
		".svg": "🎨", // SVG palette
		".ico": "🖼", // ICO frame
		".webp": "🖼", // WebP frame

		// Audio/Video
		".mp3": "🎵", // MP3 music
		".mp4": "🎬", // MP4 movie
		".wav": "🎵", // WAV music
		".avi": "🎬", // AVI movie
		".mkv": "🎬", // MKV movie

		// System/Config
		".env":  "🔐", // ENV lock
		".ini":  "⚙", // INI gear
		".conf": "⚙", // CONF gear
		".cfg":  "⚙", // CFG gear
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
