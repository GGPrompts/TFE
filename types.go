package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

// Version is the current version of TFE
const Version = "1.0.0"

// displayMode represents different view modes for displaying files
type displayMode int

const (
	modeList displayMode = iota
	modeDetail
	modeTree
)

func (d displayMode) String() string {
	switch d {
	case modeList:
		return "List"
	case modeDetail:
		return "Detail"
	case modeTree:
		return "Tree"
	default:
		return "Unknown"
	}
}

// viewMode represents the layout mode (single, dual-pane, or full preview)
type viewMode int

const (
	viewSinglePane viewMode = iota
	viewDualPane
	viewFullPreview
)

func (v viewMode) String() string {
	switch v {
	case viewSinglePane:
		return "Single"
	case viewDualPane:
		return "Dual-Pane"
	case viewFullPreview:
		return "Full Preview"
	default:
		return "Unknown"
	}
}

// paneType represents which pane is focused in dual-pane mode
type paneType int

const (
	leftPane paneType = iota
	rightPane
)

// fileItem represents a file or directory in the file browser
type fileItem struct {
	name          string
	path          string
	isDir         bool
	size          int64
	modTime       time.Time
	mode          os.FileMode
	isSymlink     bool   // Whether this is a symbolic link
	symlinkTarget string // Target path if this is a symlink
	hasVariables  *bool  // Cached: whether prompt file has {{variables}} (nil = not checked yet)
	// Git status (populated for git repositories)
	isGitRepo      bool      // Whether this directory is a git repository
	gitBranch      string    // Current branch name
	gitAhead       int       // Commits ahead of remote
	gitBehind      int       // Commits behind remote
	gitDirty       bool      // Has uncommitted changes
	gitLastCommit  time.Time // Time of last commit
}

// previewModel holds preview pane state
type previewModel struct {
	filePath   string
	fileName   string
	content    []string // Lines of the file
	scrollPos  int      // Current scroll position
	maxPreview int      // Max lines to load (prevent huge files)
	loaded     bool
	isBinary   bool
	tooLarge            bool
	fileSize            int64
	isMarkdown          bool // Whether the file is markdown
	isSyntaxHighlighted bool // Whether syntax highlighting was applied
	// Caching for performance
	cachedWrappedLines    []string // Cached wrapped text lines
	cachedRenderedContent string   // Cached Glamour-rendered markdown
	cachedLineCount       int      // Cached total line count after wrapping
	cachedWidth           int      // Width the cache was computed for
	cacheValid            bool     // Whether cache is valid
	// Prompt template (for prompt files)
	isPrompt       bool            // Whether the file is a prompt template
	promptTemplate *promptTemplate // Parsed prompt template
	// Search within preview
	searchActive  bool   // Whether search mode is active in preview
	searchQuery   string // Current search query
	searchMatches []int  // Line numbers with matches
	currentMatch  int    // Index in searchMatches array
}

// promptTemplate represents a parsed prompt with metadata and template
type promptTemplate struct {
	name        string
	description string
	source      string   // "global", "command", "agent", "local"
	variables   []string // List of {{VAR}} placeholders found
	template    string   // The template text with {{placeholders}}
	raw         string   // Original file content
}

// inputFieldType represents the type of input field for variable entry
type inputFieldType int

const (
	fieldTypeShort inputFieldType = iota // Short text input (single line)
	fieldTypeLong                         // Long text input (shows truncated)
	fieldTypeFile                         // File path (supports file picker)
)

// promptInputField represents a fillable field for a prompt template variable
type promptInputField struct {
	name         string         // Variable name (e.g., "file", "priority")
	value        string         // User's input (full content, no limit)
	defaultValue string         // Auto-filled default value
	fieldType    inputFieldType // Type of field (short/long/file)
	displayWidth int            // Available width for display
	color        string         // Color for highlighting in preview (e.g., "39", "220")
}

// getDisplayValue returns the value to display in the input field
// For long content, shows trailing end with [...] prefix and char count
func (f *promptInputField) getDisplayValue() string {
	// Use current value if filled, otherwise use default
	content := f.value
	if content == "" {
		content = f.defaultValue
	}

	// Calculate max display width (reserve space for brackets and char count)
	maxDisplay := f.displayWidth - 20
	if maxDisplay < 20 {
		maxDisplay = 20
	}

	// Check if content is multi-line
	isMultiLine := strings.Contains(content, "\n")

	if isMultiLine {
		// Multi-line content - just show summary indicator
		lines := strings.Split(content, "\n")
		lineCount := len(lines)
		charCount := len(content)

		// Format character count
		charDisplay := ""
		if charCount < 1000 {
			charDisplay = fmt.Sprintf("%d chars", charCount)
		} else if charCount < 10000 {
			charDisplay = fmt.Sprintf("%.1fk chars", float64(charCount)/1000)
		} else {
			charDisplay = fmt.Sprintf("%dk chars", charCount/1000)
		}

		return fmt.Sprintf("[Pasted: %d lines, %s]", lineCount, charDisplay)
	}

	// Single-line content
	if len(content) <= maxDisplay {
		return content
	}

	// Long single-line content - show trailing end with ellipsis
	suffix := content[len(content)-maxDisplay:]
	return suffix // We'll add [...] and (X chars) in the rendering code
}

// getCharCountDisplay returns a formatted character count string
func (f *promptInputField) getCharCountDisplay() string {
	length := len(f.value)
	if length == 0 {
		return ""
	}

	formatted := formatCharCount(length)
	if formatted == "" {
		return ""
	}
	return " (" + formatted + ")"
}

// formatCharCount formats character count in human-readable form
func formatCharCount(count int) string {
	if count < 1000 {
		return ""
	} else if count < 10000 {
		// Show as "1.2k chars"
		major := count / 1000
		minor := (count % 1000) / 100
		return string(rune('0'+major)) + "." + string(rune('0'+minor)) + "k chars"
	}
	// Show as "12k chars"
	return string(rune('0'+count/1000)) + "k chars"
}

// hasContent returns whether the field has user-entered content
func (f *promptInputField) hasContent() bool {
	return f.value != ""
}

// model represents the main application state
type model struct {
	currentPath string
	files       []fileItem
	cursor      int
	height      int
	width       int
	showHidden  bool
	displayMode displayMode
	sortBy      string // "name", "size", "modified" for detail view
	sortAsc     bool   // Sort ascending or descending
	detailScrollX int  // Horizontal scroll offset for detail view (narrow terminals)
	// Preview-related fields
	viewMode    viewMode
	preview     previewModel
	leftWidth   int // Width of left pane in dual-pane mode
	rightWidth  int // Width of right pane in dual-pane mode
	focusedPane paneType // Which pane has focus in dual-pane mode
	// Glamour renderer cache (avoid recreating on every render)
	glamourRenderer      interface{} // *glamour.TermRenderer
	glamourRendererWidth int         // Width renderer was created for
	// Mouse state for preview mode
	previewMouseEnabled bool // Whether mouse is enabled in preview mode (default: true)
	// Double-click detection
	lastClickTime  time.Time
	lastClickIndex int
	// Command prompt (always visible)
	commandInput         string
	commandCursorPos     int                       // Cursor position in command input (0 = start, len = end)
	commandHistory       []string                  // Combined history (directory + global) for navigation
	commandHistoryByDir  map[string][]string       // Per-directory command history
	commandHistoryGlobal []string                  // Global command history (cross-directory)
	historyPos           int
	commandFocused       bool // Whether command prompt has input focus
	// Loading spinner
	spinner spinner.Model
	loading bool
	// Favorites system
	favorites         map[string]bool // Path -> favorited
	showFavoritesOnly bool            // Filter to show only favorites
	// Prompts system
	showPromptsOnly bool // Filter to show only prompt files (.yaml, .md, .txt)
	// Git repositories filter
	showGitReposOnly  bool        // Filter to show only git repositories
	gitReposList      []fileItem  // Cached list of discovered git repos (recursive scan)
	gitReposLastScan  time.Time   // When we last scanned for git repos
	gitReposScanRoot  string      // Root directory of last scan
	gitReposScanDepth int         // Max depth to scan (default: 5)
	// Trash/Recycle bin system
	showTrashOnly bool        // Filter to show trash contents
	trashItems    []trashItem // Cached trash items when viewing trash
	// Prompt inline editing (fillable variables)
	promptEditMode         bool              // Whether prompt edit mode is active (Tab to activate)
	focusedVariableIndex   int               // Index of currently focused variable in template
	filledVariables        map[string]string // Map of variable name -> filled value
	filePickerMode         bool              // Whether file picker mode is active (F3)
	filePickerRestorePath  string            // Path to restore preview after file picker
	filePickerRestorePrompts bool            // Whether to restore prompts filter after file picker
	filePickerCopySource   string            // Source path when picking copy destination (context menu)
	// Tree view expansion
	expandedDirs map[string]bool // Path -> expanded state
	treeItems    []treeItem       // Cached tree items for tree view
	// Context menu (right-click menu)
	contextMenuOpen   bool
	contextMenuX      int
	contextMenuY      int
	contextMenuFile   *fileItem
	contextMenuCursor int
	// Dialog system
	dialog        dialogModel
	showDialog    bool
	statusMessage string    // Temporary status message
	statusIsError bool      // Whether status message is an error
	statusTime    time.Time // When status was shown
	// Fuzzy search
	fuzzySearchActive bool // Whether fuzzy search is active
	// Directory search (/ key)
	searchMode       bool   // Whether search mode is active
	searchQuery      string // Current search query
	filteredIndices  []int  // Indices of files matching search
	// Menu system (dropdown menus in title bar)
	startupTime      time.Time // When app started (for 5s GitHub link display)
	menuOpen         bool      // Whether any menu is currently open
	activeMenu       string    // Which menu is active ("navigate", "view", "tools", "help")
	selectedMenuItem int       // Index of selected item in active menu (-1 = none)
	menuBarFocused   bool      // Whether menu bar has keyboard focus (Alt/F9 pressed)
	highlightedMenu  string    // Which menu is highlighted in menu bar ("file", "edit", etc.)
	// Menu caching (performance optimization - avoids repeated filesystem checks)
	cachedMenus    map[string]Menu  // Cached menu structure (built once)
	toolsAvailable map[string]bool // Cached tool availability (lazygit, htop, etc.)
	// Performance: Cache for directoryContainsPrompts() to avoid repeated file I/O
	promptDirsCache map[string]bool // Path -> contains prompts (cleared on loadFiles)
	// Update notification
	updateAvailable bool   // Whether an update is available
	updateVersion   string // Version string of available update (e.g., "v0.6.1")
	updateChangelog string // Changelog/release notes from GitHub
	updateURL       string // URL to the release page
}

// treeItem represents an item in the tree view with depth information
type treeItem struct {
	file        fileItem
	depth       int
	isLast      bool
	parentLasts []bool // Track which parent levels are last items
}

// editorFinishedMsg is sent when external editor exits
type editorFinishedMsg struct{ err error }

// markdownRenderedMsg is sent when markdown rendering completes
type markdownRenderedMsg struct{}

// fuzzySearchResultMsg is sent when fuzzy search completes
type fuzzySearchResultMsg struct {
	selected string // Selected file path
	err      error
}

// updateAvailableMsg is sent when a new release is detected
type updateAvailableMsg struct {
	version   string // Version tag (e.g., "v0.6.1")
	changelog string // Release notes/changelog
	url       string // GitHub release URL
}

// dialogType represents different types of dialogs
type dialogType int

const (
	dialogNone dialogType = iota
	dialogInput   // Text input dialog (F7 directory name)
	dialogConfirm // Yes/No confirmation (F8 delete)
	dialogMessage // Status messages (success/error)
)

// dialogModel holds dialog state
type dialogModel struct {
	dialogType dialogType
	title      string
	message    string
	input      string // For text input dialogs
	confirmed  bool   // User confirmed action
	isError    bool   // For message dialogs (red vs green)
}

// MenuItem represents a single menu item
type MenuItem struct {
	Label       string // Display text
	Action      string // Action identifier (e.g., "toggle-favorites", "home")
	Shortcut    string // Keyboard shortcut display (e.g., "F6", "Ctrl+P")
	Disabled    bool   // Whether item is disabled
	IsSeparator bool   // Whether this is a separator line
	IsCheckable bool   // Whether this item shows a checkmark when active
	IsChecked   bool   // Whether checkmark is shown (for toggles)
}

// Menu represents a dropdown menu
type Menu struct {
	Label string     // Menu label in menu bar
	Items []MenuItem // Menu items
}
