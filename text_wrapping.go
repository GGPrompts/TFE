package main

// Module: text_wrapping.go
// Purpose: Text wrapping and line counting utilities
// Responsibilities:
// - Wrapping text to fit within width constraints
// - Counting wrapped lines for scrolling calculations
// - Calculating prompt header heights for layout

import (
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

// wrapLine wraps a line of text to fit within the specified width
func wrapLine(line string, width int) []string {
	if width <= 0 {
		return []string{line}
	}

	// Handle empty lines
	if len(line) == 0 {
		return []string{""}
	}

	var wrapped []string
	currentLine := ""
	currentWidth := 0

	words := strings.Fields(line)
	if len(words) == 0 {
		// Line is only whitespace
		return []string{line}
	}

	for i, word := range words {
		wordWidth := visualWidth(word)
		spaceWidth := 1

		// Check if this word fits on the current line
		if currentWidth == 0 {
			// First word on line
			if wordWidth <= width {
				currentLine = word
				currentWidth = wordWidth
			} else {
				// Word is too long, hard-break it across lines without
				// dropping content; the last chunk seeds the current line
				chunks := breakLongWord(word, width)
				wrapped = append(wrapped, chunks[:len(chunks)-1]...)
				currentLine = chunks[len(chunks)-1]
				currentWidth = visualWidth(currentLine)
			}
		} else if currentWidth+spaceWidth+wordWidth <= width {
			// Word fits on current line
			currentLine += " " + word
			currentWidth += spaceWidth + wordWidth
		} else {
			// Word doesn't fit, start new line
			wrapped = append(wrapped, currentLine)
			if wordWidth <= width {
				currentLine = word
				currentWidth = wordWidth
			} else {
				// Word is too long, hard-break it across lines without
				// dropping content; the last chunk seeds the current line
				chunks := breakLongWord(word, width)
				wrapped = append(wrapped, chunks[:len(chunks)-1]...)
				currentLine = chunks[len(chunks)-1]
				currentWidth = visualWidth(currentLine)
			}
		}

		// If this is the last word, add the current line
		if i == len(words)-1 && currentLine != "" {
			wrapped = append(wrapped, currentLine)
		}
	}

	if len(wrapped) == 0 {
		return []string{line}
	}

	return wrapped
}

// breakLongWord hard-breaks a word that is wider than width into visual-width-
// limited chunks. Unlike truncateToWidth it never appends "..." and never
// drops content: every rune of the word appears in exactly one chunk, in
// order. ANSI escape sequences pass through without counting toward width
// (mirroring truncateToWidth). Always returns at least one chunk; each chunk
// contains at least one visible rune even if that rune alone exceeds width
// (e.g. a wide CJK rune at width 1), which guarantees forward progress.
func breakLongWord(word string, width int) []string {
	if width <= 0 || word == "" {
		return []string{word}
	}

	var chunks []string
	var current strings.Builder
	currentWidth := 0
	inAnsi := false

	for _, ch := range word {
		// Handle ANSI escape sequences (don't count toward width)
		if ch == '\033' {
			inAnsi = true
			current.WriteRune(ch)
			continue
		}
		if inAnsi {
			current.WriteRune(ch)
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inAnsi = false
			}
			continue
		}

		// Use runewidth to properly handle wide characters (emojis, CJK)
		charWidth := runewidth.RuneWidth(ch)
		if currentWidth+charWidth > width && currentWidth > 0 {
			chunks = append(chunks, current.String())
			current.Reset()
			currentWidth = 0
		}
		current.WriteRune(ch)
		currentWidth += charWidth
	}

	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	if len(chunks) == 0 {
		return []string{word}
	}
	return chunks
}

// previewBoxContentWidth returns the inner content width of the preview box
// for the current layout. This is the SINGLE source of truth for the preview
// width branching: populatePreviewCache(), renderPreview(), and
// getWrappedLineCount() must all derive their widths from this helper.
// If any call site computes the width ad hoc and disagrees, the wrap/Glamour
// cache silently misses and the whole preview is re-wrapped (or re-rendered
// through Glamour) on every frame/scroll event.
func (m model) previewBoxContentWidth() int {
	if m.viewMode == viewFullPreview {
		return m.width - 6 // Full preview: box is Width(m.width - 6)
	}
	if m.displayMode == modeDetail || m.isNarrowTerminal() {
		return m.width - 6 // Vertical split: box is Width(m.width - 6)
	}
	return m.rightWidth - 2 // Horizontal split: box is Width(m.rightWidth - 2)
}

// previewAvailableWidth returns the width available for preview text content,
// derived from previewBoxContentWidth() based on the current file type.
// Never compute this ad hoc — see previewBoxContentWidth().
func (m model) previewAvailableWidth() int {
	availableWidth := m.previewBoxContentWidth()
	if m.preview.isMarkdown {
		// Markdown: no line numbers or scrollbar, but subtract 2 for left
		// padding (prevents code blocks from touching the border)
		availableWidth -= 2
	} else {
		// Regular text: subtract line nums (6) + scrollbar (1) + space (1) = 8 chars
		availableWidth -= 8
	}
	if availableWidth < 20 {
		availableWidth = 20 // Minimum width
	}
	return availableWidth
}

// diffPreviewAvailableWidth returns the width available for diff preview
// content: scrollbar (1) + space (1) = 2 chars overhead, no line numbers.
// Derived from previewBoxContentWidth() — never compute this ad hoc, or the
// diff cache (refreshDiffPreviewCacheIfStale / renderDiffPreview) will miss.
func (m model) diffPreviewAvailableWidth() int {
	availableWidth := m.previewBoxContentWidth() - 2
	if availableWidth < 20 {
		availableWidth = 20
	}
	return availableWidth
}

// getWrappedLineCount calculates the total number of wrapped lines for the current preview
func (m model) getWrappedLineCount() int {
	if !m.preview.loaded {
		return 0
	}

	// JSONL files: count the cached rendered lines (same cache
	// renderJSONLPreview reads, so the count always matches the render)
	if m.preview.isJSONL && len(m.preview.cachedJSONLMessages) > 0 {
		return len(m.jsonlRenderedLines())
	}

	// Calculate available width via the shared helper so the cache check below
	// agrees with populatePreviewCache() and renderPreview()
	availableWidth := m.previewAvailableWidth()

	// Use the cached line count whenever the cache is valid for the CURRENT
	// width. The cache is refreshed once per layout change by
	// refreshPreviewCacheIfStale() (Update) / populatePreviewCache(), so in the
	// common scroll path (held-down arrow at a fixed width) this returns in O(1)
	// instead of re-wrapping the whole file on every keypress.
	//
	// The width check is what keeps the count correct: a stale cache built for a
	// different width (e.g. dual-pane vs full-screen) fails this guard and falls
	// through to a fresh wrap below. Do NOT add a `cachedLineCount > 0` guard
	// here — an empty file legitimately wraps to 0 lines, and treating 0 as
	// "uncached" would re-wrap it on every call for no benefit.
	if m.preview.cacheValid && m.preview.cachedWidth == availableWidth {
		return m.preview.cachedLineCount
	}

	// For markdown, we need to render it to count lines
	if m.preview.isMarkdown {
		markdownContent := strings.Join(m.preview.content, "\n")
		// Use cached rendering with timeout to prevent hangs
		// Note: renderMarkdownWithTimeout is in file_operations.go
		rendered, err := m.renderMarkdownWithTimeout(markdownContent, availableWidth, 5*time.Second)
		if err == nil {
			renderedLines := strings.Split(strings.TrimRight(rendered, "\n"), "\n")
			contentLineCount := len(renderedLines)

			// For prompt files, return only content line count
			// The header is fixed and doesn't scroll, so it shouldn't be included
			// in scroll calculations
			return contentLineCount
		}
		// Fallback if glamour fails or times out
	}

	// For regular text, count wrapped lines
	totalLines := 0
	for _, line := range m.preview.content {
		wrapped := wrapLine(line, availableWidth)
		totalLines += len(wrapped)
	}

	// For prompt files, return only content line count
	// The header is fixed and doesn't scroll, so it shouldn't be included
	// in scroll calculations
	return totalLines
}

// visualWidth calculates the visual width of a string, accounting for tabs and ANSI codes
// This is important for consistent scrollbar alignment and box borders
// NOTE: This is the non-terminal-aware version - use m.visualWidthCompensated() for layout calculations
func visualWidth(s string) int {
	// Strip ANSI codes first
	stripped := ""
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

		// Keep visible characters
		stripped += string(ch)
	}

	return runewidth.StringWidth(stripped)
}

// visualWidthCompensated calculates visual width with terminal-specific emoji compensation
// Use this for layout calculations that need accurate emoji widths
// STRIPS ANSI escape codes before calculating width
func (m model) visualWidthCompensated(s string) int {
	// Use visualWidth() which strips ANSI codes, not runewidth.StringWidth()
	width := visualWidth(s)

	// Apply variation selector compensation for Windows Terminal ONLY
	// runewidth reports emoji+VS as 1 cell, but Windows Terminal renders as 2 cells
	// We ADD 1 per VS for Windows Terminal to match its wider rendering
	variationSelectorCount := strings.Count(s, "\uFE0F")
	if m.terminalType == terminalWindowsTerminal && variationSelectorCount > 0 {
		width += variationSelectorCount
	}

	// NOTE: xterm.js actually renders emojis as 2 cells (same as runewidth)
	// Do NOT compensate for xterm - trust runewidth's 2-cell calculation

	return width
}

// truncateANSIAware is the shared core for all ANSI-aware truncation. It walks
// s rune by rune, never counting ANSI escape sequences toward visual width, and
// accumulates visible runes until adding the next one would exceed targetWidth.
// Per-rune visible width comes from runeW (so callers choose runewidth.RuneWidth
// vs the terminal-aware m.runeWidth). When the string overflows, ellipsis is
// appended IFF its own visual width still fits in the remaining cells
// (targetWidth-width); the reset code "\033[0m" has visual width 0, so it is
// always appended, while "..." (width 3) is appended only when >= 3 cells
// remain — exactly matching the three legacy implementations.
//
// expandTabs controls whether '\t' is expanded to the next 8-column tab stop
// (truncateToWidth / truncateToWidthCompensated) or treated as an ordinary rune
// measured by runeW (truncateToVisualWidth, which historically did not special
// case tabs — runeW reports '\t' as width 0). This flag is required to keep all
// three wrappers byte-for-byte identical to their originals; it cannot be folded
// into runeW because tab width depends on the running column.
func truncateANSIAware(s string, targetWidth int, runeW func(rune) int, ellipsis string, expandTabs bool) string {
	width := 0
	var result strings.Builder
	inAnsi := false

	for _, ch := range s {
		// Handle ANSI escape sequences (don't count toward width)
		if ch == '\033' {
			inAnsi = true
			result.WriteRune(ch)
			continue
		}

		if inAnsi {
			result.WriteRune(ch)
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inAnsi = false
			}
			continue
		}

		// Calculate character width
		charWidth := 1
		if expandTabs && ch == '\t' {
			charWidth = 8 - (width % 8)
		} else {
			charWidth = runeW(ch)
		}

		if width+charWidth > targetWidth {
			// Can't fit this character. Append ellipsis only if its visual
			// width still fits in the remaining cells.
			if targetWidth-width >= visualWidth(ellipsis) {
				result.WriteString(ellipsis)
			}
			return result.String()
		}

		width += charWidth
		result.WriteString(string(ch))
	}

	return result.String()
}

// truncateToWidth truncates a string to fit within a target visual width
func truncateToWidth(s string, targetWidth int) string {
	return truncateANSIAware(s, targetWidth, runewidth.RuneWidth, "...", true)
}

// truncateTail keeps the trailing portion of s that fits within targetWidth
// visual cells, prefixing "..." when truncation occurs. It walks runes from the
// END summing rune width (mirroring how truncateToWidth walks from the front)
// so a multibyte rune is never split — unlike byte slicing, which can emit
// invalid UTF-8. Used for tail-truncating paths so the most specific trailing
// part (e.g. a symlink target's filename) stays visible.
func truncateTail(s string, targetWidth int) string {
	if visualWidth(s) <= targetWidth {
		return s
	}
	const ellipsis = "..."
	avail := targetWidth - 3 // visual width of "..."
	if avail <= 0 {
		return truncateToWidth(ellipsis, targetWidth)
	}
	runes := []rune(s)
	width := 0
	start := len(runes)
	for i := len(runes) - 1; i >= 0; i-- {
		w := runewidth.RuneWidth(runes[i])
		if width+w > avail {
			break
		}
		width += w
		start = i
	}
	return ellipsis + string(runes[start:])
}

// truncateToWidthCompensated truncates a string to fit within a target visual width
// with terminal-specific emoji width compensation (uses m.runeWidth for accurate widths)
func (m model) truncateToWidthCompensated(s string, targetWidth int) string {
	return truncateANSIAware(s, targetWidth, m.runeWidth, "...", true)
}

// padIconToWidth pads an icon emoji to a fixed width (2 cells) for consistent alignment
// Some terminals render certain emojis as 1 cell, so we pad them to 2 cells
func (m model) padIconToWidth(icon string) string {
	return m.padToVisualWidth(icon, 2)
}

// padToVisualWidth pads a string to a specific visual width using spaces
// This correctly handles emojis, wide characters, AND ANSI escape codes
// Terminal-aware: Variation selector compensation only for WezTerm
func (m model) padToVisualWidth(s string, targetWidth int) string {
	// Use visualWidthCompensated which already handles ANSI codes via m.visualWidthCompensated
	// which delegates to visualWidth for ANSI stripping
	calculatedWidth := m.visualWidthCompensated(s)

	if calculatedWidth >= targetWidth {
		return s
	}
	padding := targetWidth - calculatedWidth
	return s + strings.Repeat(" ", padding)
}

// getPromptHeaderHeight calculates how many lines the prompt header takes up
// This matches the logic in renderPromptPreview() to ensure consistent calculations
func (m model) getPromptHeaderHeight(boxContentWidth int) int {
	if !m.preview.isPrompt || m.preview.promptTemplate == nil {
		return 0
	}

	tmpl := m.preview.promptTemplate
	headerWrapWidth := boxContentWidth - 2 // Leave room for padding
	if headerWrapWidth < 20 {
		headerWrapWidth = 20
	}

	headerLineCount := 0

	// Prompt name (if available)
	if tmpl.name != "" {
		nameLine := "📝 " + tmpl.name
		if visualWidth(nameLine) > headerWrapWidth {
			wrapped := wrapLine(nameLine, headerWrapWidth)
			headerLineCount += len(wrapped)
		} else {
			headerLineCount++ // One line
		}
		headerLineCount++ // Blank line after name
	}

	// Description (if available)
	if tmpl.description != "" {
		if visualWidth(tmpl.description) > headerWrapWidth {
			wrapped := wrapLine(tmpl.description, headerWrapWidth)
			headerLineCount += len(wrapped)
		} else {
			headerLineCount++ // One line
		}
		headerLineCount++ // Blank line after description
	}

	// Source indicator - account for wrapping
	sourceIcon := ""
	sourceLabel := ""
	switch tmpl.source {
	case "global":
		sourceIcon = "🌐"
		sourceLabel = "Global Prompt (~/.prompts/)"
	case "command":
		sourceIcon = "⚙"
		sourceLabel = "Project Command (.claude/commands/)"
	case "agent":
		sourceIcon = "🤖"
		sourceLabel = "Project Agent (.claude/agents/)"
	case "skill":
		sourceIcon = "🎯"
		sourceLabel = "Project Skill (.claude/skills/)"
	case "local":
		sourceIcon = "📁"
		sourceLabel = "Local Prompt"
	}
	sourceLine := sourceIcon + " " + sourceLabel
	if visualWidth(sourceLine) > headerWrapWidth {
		wrapped := wrapLine(sourceLine, headerWrapWidth)
		headerLineCount += len(wrapped)
	} else {
		headerLineCount++ // One line
	}

	// Variables line (if any) - account for wrapping
	if len(tmpl.variables) > 0 {
		// Build the plain variables line to calculate wrapping
		plainVarsLine := "Variables: " + strings.Join(tmpl.variables, ", ")
		if visualWidth(plainVarsLine) > headerWrapWidth {
			wrapped := wrapLine(plainVarsLine, headerWrapWidth)
			headerLineCount += len(wrapped)
		} else {
			headerLineCount++ // One line
		}
	}

	// Separator line - always one line since it's exactly headerWrapWidth characters
	// Each '─' has visual width 1, so total width equals headerWrapWidth
	headerLineCount++

	return headerLineCount
}
