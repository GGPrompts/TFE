package main

import (
	"github.com/charmbracelet/lipgloss"
)

// Style variables — initialized by initStyles() in theme.go from the active theme.
// Do NOT hardcode colors here; all colors come from currentTheme.
var (
	// Title bar styling (left-aligned, no padding)
	titleStyle lipgloss.Style

	// Path display styling
	pathStyle lipgloss.Style

	// Status bar styling
	statusStyle lipgloss.Style

	// Selected item styling
	selectedStyle lipgloss.Style

	// Narrow terminal selection styling (matrix green - no background to prevent wrapping)
	narrowSelectedStyle lipgloss.Style

	// Folder styling
	folderStyle lipgloss.Style

	// File styling
	fileStyle lipgloss.Style

	// Claude context file styling (orange)
	claudeContextStyle lipgloss.Style

	// Agents file styling (purple)
	agentsStyle lipgloss.Style

	// Prompts folder styling (bright magenta/pink)
	promptsFolderStyle lipgloss.Style

	// Obsidian vault styling (teal/cyan)
	obsidianVaultStyle lipgloss.Style

	// Diff preview styles (git diff coloring in changes mode)
	diffAddedStyle      lipgloss.Style // Green for added lines (+)
	diffRemovedStyle    lipgloss.Style // Red for removed lines (-)
	diffHunkHeaderStyle lipgloss.Style // Cyan for @@ hunk headers
	diffMetaStyle       lipgloss.Style // Dim for diff/index/---/+++ headers

	// JSONL conversation preview styles (render_jsonl.go)
	jsonlUserStyle       lipgloss.Style // Bold title color for USER messages
	jsonlAssistantStyle  lipgloss.Style // Body text for ASSISTANT messages
	jsonlToolNameStyle   lipgloss.Style // Bold accent for tool names
	jsonlToolInputStyle  lipgloss.Style // Muted text for tool input summaries
	jsonlThinkingStyle   lipgloss.Style // Dim italic for thinking blocks
	jsonlSystemStyle     lipgloss.Style // Muted text for system messages
	jsonlSeparatorStyle  lipgloss.Style // Subtle separator lines
	jsonlToolResultStyle lipgloss.Style // Muted text for tool results

	// Preview pane styles (render_preview.go) — rebuilt per line before hoisting
	lineNumStyle         lipgloss.Style // Subtle line-number gutter
	scrollbarTrackStyle  lipgloss.Style // Dim scrollbar track (│)
	scrollbarThumbStyle  lipgloss.Style // Bright scrollbar thumb (│)
	scrollIndicatorStyle lipgloss.Style // Subtle italic scroll-position indicator

	// File-list styles (render_file_list.go)
	detailHeaderStyle lipgloss.Style // Bold title-colored detail-view column header

	// Alternate-row background variants of each file-type style.
	// Indexed by base style via alternateRowStyle(); built once in initStyles().
	fileAltStyle          lipgloss.Style
	folderAltStyle        lipgloss.Style
	claudeContextAltStyle lipgloss.Style
	agentsAltStyle        lipgloss.Style
	promptsFolderAltStyle lipgloss.Style
	obsidianVaultAltStyle lipgloss.Style

	// Toolbar button styles (helpers.go) — hardcoded colors, rebuilt per frame
	toolbarButtonStyle       lipgloss.Style // Inactive emoji button (blue, bold)
	toolbarButtonActiveStyle lipgloss.Style // Active emoji button (blue, bold, gray bg)
	toolbarTermStyle         lipgloss.Style // Command ">_" glyph (green, bold)
	toolbarTermActiveStyle   lipgloss.Style // Active ">_" glyph (green, bold, gray bg)

	// Tab bar styles (render_layout.go)
	activeTabStyle    lipgloss.Style // Active tab (selection colors, bold)
	inactiveTabStyle  lipgloss.Style // Inactive tab (body text on panel bg)
	tabModifiedStyle  lipgloss.Style // Git "M"/"R" indicator
	tabAddedStyle     lipgloss.Style // Git "+" indicator
	tabDeletedStyle   lipgloss.Style // Git "-" indicator
	tabUntrackedStyle lipgloss.Style // Git "?" indicator
	tabCloseStyle     lipgloss.Style // "x" close glyph on active tab
	tabOverflowStyle  lipgloss.Style // "+N more" overflow indicator

	// Command-line styles (render_layout.go renderCommandLine)
	cmdPromptStyle lipgloss.Style // "$ " prompt + path (title, bold)
	cmdInputStyle  lipgloss.Style // Typed command text (body)
	cmdHelperStyle lipgloss.Style // Contextual hint text (muted italic)
	cmdCursorStyle lipgloss.Style // Block cursor (title, bold)
	cmdBangStyle   lipgloss.Style // "!" run-and-exit prefix (red, bold)
	cmdGhostStyle  lipgloss.Style // Ghost-text suggestion (muted italic)
)

// alternateRowStyle returns the alternate-row background variant for a given
// file-type base style. Falls back to the file (default) variant.
func alternateRowStyle(base lipgloss.Style) lipgloss.Style {
	switch {
	case base.GetForeground() == folderStyle.GetForeground():
		return folderAltStyle
	case base.GetForeground() == claudeContextStyle.GetForeground():
		return claudeContextAltStyle
	case base.GetForeground() == agentsStyle.GetForeground():
		return agentsAltStyle
	case base.GetForeground() == promptsFolderStyle.GetForeground():
		return promptsFolderAltStyle
	case base.GetForeground() == obsidianVaultStyle.GetForeground():
		return obsidianVaultAltStyle
	default:
		return fileAltStyle
	}
}
