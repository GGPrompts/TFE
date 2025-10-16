package main

import (
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
)

// displayMode represents different view modes for displaying files
type displayMode int

const (
	modeList displayMode = iota
	modeGrid
	modeDetail
	modeTree
)

func (d displayMode) String() string {
	switch d {
	case modeList:
		return "List"
	case modeGrid:
		return "Grid"
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
	name    string
	path    string
	isDir   bool
	size    int64
	modTime time.Time
	mode    os.FileMode
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
	tooLarge   bool
	fileSize   int64
	isMarkdown bool // Whether the file is markdown
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
	gridColumns int
	sortBy      string // "name", "size", "modified" for detail view
	sortAsc     bool   // Sort ascending or descending
	// Preview-related fields
	viewMode    viewMode
	preview     previewModel
	leftWidth   int // Width of left pane in dual-pane mode
	rightWidth  int // Width of right pane in dual-pane mode
	focusedPane paneType // Which pane has focus in dual-pane mode
	// Double-click detection
	lastClickTime  time.Time
	lastClickIndex int
	// Command prompt (always visible)
	commandInput   string
	commandHistory []string
	historyPos     int
	commandFocused bool // Whether command prompt has input focus
	// Loading spinner
	spinner spinner.Model
	loading bool
	// Favorites system
	favorites         map[string]bool // Path -> favorited
	showFavoritesOnly bool            // Filter to show only favorites
	// Tree view expansion
	expandedDirs map[string]bool // Path -> expanded state
	treeItems    []treeItem       // Cached tree items for tree view
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
