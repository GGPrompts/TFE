package main

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// previewWidthTestModel builds a model with a loaded plain-text preview in the
// given layout, with pane widths derived from calculateLayout() (the real
// layout math), not hand-set values.
func previewWidthTestModel(width int, vm viewMode, dm displayMode) model {
	m := model{
		width:       width,
		height:      40,
		viewMode:    vm,
		displayMode: dm,
		focusedPane: leftPane,
		preview: previewModel{
			loaded:     true,
			maxPreview: 1000,
			content: []string{
				"first line of preview content",
				"second line that is somewhat longer so wrapping can occur at narrow widths",
				"third line",
			},
		},
	}
	m.calculateLayout()
	return m
}

// previewWidthLayouts returns one case per preview layout branch:
// full preview, dual-pane horizontal split, dual-pane vertical split
// (Detail mode), and dual-pane on a narrow terminal (forced vertical).
func previewWidthLayouts() []struct {
	name string
	m    model
} {
	return []struct {
		name string
		m    model
	}{
		{"full preview", previewWidthTestModel(120, viewFullPreview, modeList)},
		{"dual-pane horizontal split", previewWidthTestModel(120, viewDualPane, modeList)},
		{"dual-pane vertical split (detail)", previewWidthTestModel(120, viewDualPane, modeDetail)},
		{"dual-pane narrow terminal", previewWidthTestModel(80, viewDualPane, modeList)},
	}
}

// TestPreviewWidth_CacheAgreesAcrossCallSites is a regression test for tfe-n13:
// populatePreviewCache, renderPreview, and getWrappedLineCount used three
// different width formulas, so in dual-pane layouts the cachedWidth checks
// never matched and the entire preview was re-wrapped (or re-rendered through
// Glamour) on every frame/scroll event. All three call sites must derive the
// width from previewAvailableWidth() so the cache populated once always hits.
func TestPreviewWidth_CacheAgreesAcrossCallSites(t *testing.T) {
	for _, tc := range previewWidthLayouts() {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.m

			// populatePreviewCache must store the cache at exactly the width
			// the render/count paths will look it up with.
			m.populatePreviewCache()
			if !m.preview.cacheValid {
				t.Fatal("populatePreviewCache did not set cacheValid")
			}
			if m.preview.cachedWidth != m.previewAvailableWidth() {
				t.Errorf("populatePreviewCache cached width %d, but previewAvailableWidth() = %d",
					m.preview.cachedWidth, m.previewAvailableWidth())
			}

			// getWrappedLineCount must hit the cache: plant a sentinel count
			// and verify it is returned (a width mismatch would recount).
			const sentinelCount = 424242
			m.preview.cachedLineCount = sentinelCount
			if got := m.getWrappedLineCount(); got != sentinelCount {
				t.Errorf("getWrappedLineCount() = %d, expected sentinel %d (cache miss: width formula disagrees with populatePreviewCache)",
					got, sentinelCount)
			}
			m.populatePreviewCache() // restore real cache

			// renderPreview must hit the cache: plant a sentinel wrapped line
			// and verify it is rendered instead of re-wrapping m.preview.content.
			const sentinelLine = "CACHE_SENTINEL_LINE"
			m.preview.cachedWrappedLines = []string{sentinelLine}
			out := m.renderPreview(5)
			if !strings.Contains(out, sentinelLine) {
				t.Errorf("renderPreview() did not use cached wrapped lines (cache miss: width formula disagrees with populatePreviewCache)\noutput:\n%s", out)
			}
		})
	}
}

// freshWrapCount wraps every preview content line at the given width and
// returns the total, mirroring populatePreviewCache()'s plain-text path. It is
// the ground truth getWrappedLineCount()'s cached value must agree with.
func freshWrapCount(content []string, width int) int {
	total := 0
	for _, line := range content {
		total += len(wrapLine(line, width))
	}
	return total
}

// TestGetWrappedLineCount_MatchesFreshWrap is the tfe-b5w regression test:
// getWrappedLineCount must return the cached count (O(1), no per-call re-wrap),
// and that count must equal a freshly-computed wrap at the CURRENT width. When
// the layout width changes (dual-pane vs full-screen), the count must update
// rather than returning a stale value for the wrong width.
func TestGetWrappedLineCount_MatchesFreshWrap(t *testing.T) {
	// Width 120 (>= 100) keeps List in a horizontal split (narrower preview)
	// while Detail forces a vertical split (wider preview), so the layout
	// change below actually moves the wrap width. Below 100 both modes share
	// the vertical-split width (isNarrowTerminal) and the count wouldn't change.
	m := previewWidthTestModel(120, viewDualPane, modeList)
	m.populatePreviewCache()

	want := freshWrapCount(m.preview.content, m.previewAvailableWidth())
	if got := m.getWrappedLineCount(); got != want {
		t.Errorf("getWrappedLineCount() = %d, want fresh-wrap count %d at width %d",
			got, want, m.previewAvailableWidth())
	}

	// The returned count must equal the cached count (proving it read the cache
	// rather than recomputing a possibly-divergent value).
	if m.getWrappedLineCount() != m.preview.cachedLineCount {
		t.Errorf("getWrappedLineCount() = %d did not match cachedLineCount %d",
			m.getWrappedLineCount(), m.preview.cachedLineCount)
	}

	// Change the layout so the preview width changes, then refresh the cache as
	// Update would. The count must follow the new width (no stale value).
	narrowWidth := m.previewAvailableWidth()
	m.displayMode = modeDetail // List (horizontal split) -> Detail (vertical split, wider)
	m.calculateLayout()
	wideWidth := m.previewAvailableWidth()
	if wideWidth == narrowWidth {
		t.Fatal("test setup broken: layout change did not change preview width")
	}
	m.refreshPreviewCacheIfStale()

	wantWide := freshWrapCount(m.preview.content, wideWidth)
	if got := m.getWrappedLineCount(); got != wantWide {
		t.Errorf("after width change getWrappedLineCount() = %d, want fresh-wrap count %d at width %d (stale count for wrong width?)",
			got, wantWide, wideWidth)
	}

	// Sanity: a wider preview wraps to fewer-or-equal lines than the narrow one.
	if wantWide > want {
		t.Errorf("wider layout produced MORE wrapped lines (%d) than narrow (%d) — width plumbing is wrong", wantWide, want)
	}
}

// TestGetWrappedLineCount_StaleWidthRecounts verifies the width guard: a cache
// left valid but populated for a DIFFERENT width must not be returned as-is;
// getWrappedLineCount must fall through and re-wrap at the current width so the
// scroll-bounds count is never stale for the wrong layout.
func TestGetWrappedLineCount_StaleWidthRecounts(t *testing.T) {
	m := previewWidthTestModel(80, viewDualPane, modeList)
	m.populatePreviewCache()

	// Corrupt the cache so it claims a count for a width that no longer matches.
	m.preview.cachedWidth = m.previewAvailableWidth() - 7
	m.preview.cachedLineCount = 999999
	m.preview.cacheValid = true

	want := freshWrapCount(m.preview.content, m.previewAvailableWidth())
	if got := m.getWrappedLineCount(); got != want {
		t.Errorf("getWrappedLineCount() returned stale count %d for mismatched width; want fresh-wrap %d", got, want)
	}
}

// TestGetWrappedLineCount_EmptyContentCached verifies the removed
// `cachedLineCount > 0` guard: an empty preview legitimately wraps to 0 lines,
// and a valid cache holding 0 must be trusted (returned in O(1)) rather than
// re-wrapping on every call.
func TestGetWrappedLineCount_EmptyContentCached(t *testing.T) {
	m := previewWidthTestModel(80, viewDualPane, modeList)
	m.preview.content = nil // empty file
	m.populatePreviewCache()

	if m.preview.cachedLineCount != 0 {
		t.Fatalf("expected empty content to cache 0 lines, got %d", m.preview.cachedLineCount)
	}
	if !m.preview.cacheValid {
		t.Fatal("expected cacheValid for empty content")
	}
	// Must hit the cache (return 0) at the matching width, not fall through.
	if got := m.getWrappedLineCount(); got != 0 {
		t.Errorf("getWrappedLineCount() = %d for empty cached content, want 0", got)
	}
}

// TestPreviewAvailableWidth_Formulas verifies the single-source-of-truth
// helpers encode the expected per-layout and per-file-type formulas.
func TestPreviewAvailableWidth_Formulas(t *testing.T) {
	tests := []struct {
		name        string
		m           model
		expectedBox int
	}{
		{"full preview", previewWidthTestModel(120, viewFullPreview, modeList), 120 - 6},
		{"dual-pane vertical split (detail)", previewWidthTestModel(120, viewDualPane, modeDetail), 120 - 6},
		{"dual-pane narrow terminal", previewWidthTestModel(80, viewDualPane, modeList), 80 - 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.previewBoxContentWidth(); got != tt.expectedBox {
				t.Errorf("previewBoxContentWidth() = %d, expected %d", got, tt.expectedBox)
			}
		})
	}

	// Horizontal split derives from rightWidth (accordion ratio), so assert
	// the relationship rather than a hardcoded value.
	hm := previewWidthTestModel(120, viewDualPane, modeList)
	if got := hm.previewBoxContentWidth(); got != hm.rightWidth-2 {
		t.Errorf("horizontal split previewBoxContentWidth() = %d, expected rightWidth-2 = %d", got, hm.rightWidth-2)
	}

	// File-type adjustments: markdown subtracts 2 (padding), text subtracts 8
	// (line numbers + scrollbar + space).
	m := previewWidthTestModel(120, viewFullPreview, modeList)
	box := m.previewBoxContentWidth()
	if got := m.previewAvailableWidth(); got != box-8 {
		t.Errorf("plain text previewAvailableWidth() = %d, expected %d", got, box-8)
	}
	m.preview.isMarkdown = true
	if got := m.previewAvailableWidth(); got != box-2 {
		t.Errorf("markdown previewAvailableWidth() = %d, expected %d", got, box-2)
	}

	// Minimum width clamp
	tiny := previewWidthTestModel(10, viewFullPreview, modeList)
	if got := tiny.previewAvailableWidth(); got != 20 {
		t.Errorf("previewAvailableWidth() at tiny terminal = %d, expected clamped minimum 20", got)
	}
}

// TestRefreshPreviewCacheIfStale verifies the Update-driven cache refresh:
// after a layout change (e.g. List -> Detail flips horizontal split to
// vertical split), refreshPreviewCacheIfStale must repopulate the cache at
// the new width exactly once, instead of renderPreview falling back to a
// per-frame re-wrap it can never store back (value receiver).
func TestRefreshPreviewCacheIfStale(t *testing.T) {
	m := previewWidthTestModel(120, viewDualPane, modeList)
	m.populatePreviewCache()
	oldWidth := m.preview.cachedWidth

	// Simulate a keyboard handler switching to Detail mode (vertical split)
	m.displayMode = modeDetail
	m.calculateLayout()

	if m.previewAvailableWidth() == oldWidth {
		t.Fatal("test setup broken: layout change did not change preview width")
	}

	m.refreshPreviewCacheIfStale()
	if !m.preview.cacheValid {
		t.Fatal("refreshPreviewCacheIfStale left cache invalid")
	}
	if m.preview.cachedWidth != m.previewAvailableWidth() {
		t.Errorf("after refresh, cachedWidth = %d, expected %d", m.preview.cachedWidth, m.previewAvailableWidth())
	}

	// When the cache already matches, refresh must be a no-op (no rewrap)
	m.preview.cachedLineCount = 424242 // sentinel survives only if no repopulate
	m.refreshPreviewCacheIfStale()
	if m.preview.cachedLineCount != 424242 {
		t.Error("refreshPreviewCacheIfStale repopulated the cache even though the width already matched")
	}
}

// joinWrappedContent concatenates wrapped lines and strips spaces so the
// result can be compared against the original line's non-whitespace content.
// wrapLine collapses whitespace (strings.Fields), so this is the invariant a
// lossless wrap must preserve: no visible character may vanish.
func joinWrappedContent(wrapped []string) string {
	return strings.ReplaceAll(strings.Join(wrapped, ""), " ", "")
}

// TestWrapLine_HardBreaksLongWords is a regression test for tfe-esu: wrapLine
// used truncateToWidth on words wider than the wrap width, which kept only the
// first `width` columns and silently discarded the rest of the word (long
// URLs, file paths, base64 blobs lost everything past the first line).
func TestWrapLine_HardBreaksLongWords(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		width int
	}{
		{
			name:  "long URL",
			line:  "https://example.com/some/very/long/path?query=abcdefghijklmnopqrstuvwxyz0123456789&token=ZYXWVUTSRQPONMLKJIHGFEDCBA9876543210",
			width: 30,
		},
		{
			name:  "long file path",
			line:  "/home/marci/projects/TFE/node_modules/@scope/very-long-package-name/dist/esm/internal/generated/index.min.js",
			width: 24,
		},
		{
			name:  "long word mid-sentence",
			line:  "see https://example.com/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa for details",
			width: 20,
		},
		{
			name:  "wide CJK run",
			line:  "日本語のとても長い単語をハードブレークするテストです漢字漢字漢字漢字漢字漢字",
			width: 10,
		},
		{
			name:  "CJK run at odd width",
			line:  "漢字漢字漢字漢字漢字漢字漢字",
			width: 7, // wide runes (2 cols) never fit evenly; lines must stay <= 7
		},
		{
			name:  "base64 blob",
			line:  strings.Repeat("QmFzZTY0", 25),
			width: 16,
		},
		{
			name:  "single wide rune wider than width",
			line:  "漢",
			width: 1, // rune width 2 > wrap width; must still be emitted, not dropped
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := wrapLine(tc.line, tc.width)

			// No content may be lost
			want := strings.Join(strings.Fields(tc.line), "")
			got := joinWrappedContent(wrapped)
			if got != want {
				t.Errorf("wrapLine dropped content:\n got %q\nwant %q", got, want)
			}

			// No line may exceed the wrap width (except a single rune wider
			// than the width itself, which cannot be split further)
			for i, l := range wrapped {
				w := visualWidth(l)
				if w > tc.width && len([]rune(l)) > 1 {
					t.Errorf("line %d %q has visual width %d > %d", i, l, w, tc.width)
				}
			}
		})
	}
}

// TestWrapLine_LongWordRemainderJoinsNextWords verifies that after
// hard-breaking an overlong word, the remainder seeds the current line so
// following short words pack onto it instead of starting a fresh line.
func TestWrapLine_LongWordRemainderJoinsNextWords(t *testing.T) {
	// "aaaaaaaaaab" (11 wide) at width 10 breaks into "aaaaaaaaaa" + "b";
	// "cc dd" should join the "b" remainder on one line.
	wrapped := wrapLine("aaaaaaaaaab cc dd", 10)
	expected := []string{"aaaaaaaaaa", "b cc dd"}
	if len(wrapped) != len(expected) {
		t.Fatalf("wrapLine returned %d lines %q, expected %d %q", len(wrapped), wrapped, len(expected), expected)
	}
	for i := range expected {
		if wrapped[i] != expected[i] {
			t.Errorf("line %d = %q, expected %q", i, wrapped[i], expected[i])
		}
	}
}

// TestBreakLongWord covers the visual-width chunking helper directly.
func TestBreakLongWord(t *testing.T) {
	t.Run("ascii exact chunks", func(t *testing.T) {
		chunks := breakLongWord("abcdefghij", 4)
		expected := []string{"abcd", "efgh", "ij"}
		if len(chunks) != len(expected) {
			t.Fatalf("got %q, expected %q", chunks, expected)
		}
		for i := range expected {
			if chunks[i] != expected[i] {
				t.Errorf("chunk %d = %q, expected %q", i, chunks[i], expected[i])
			}
		}
	})

	t.Run("cjk never splits a wide rune across the boundary", func(t *testing.T) {
		// Each rune is 2 columns; at width 5 only 2 runes (4 cols) fit per chunk
		chunks := breakLongWord("漢字漢字漢", 5)
		expected := []string{"漢字", "漢字", "漢"}
		if len(chunks) != len(expected) {
			t.Fatalf("got %q, expected %q", chunks, expected)
		}
		for i := range expected {
			if chunks[i] != expected[i] {
				t.Errorf("chunk %d = %q, expected %q", i, chunks[i], expected[i])
			}
		}
	})

	t.Run("ansi codes pass through without counting", func(t *testing.T) {
		word := "\033[38;5;220mabcd\033[0mef"
		chunks := breakLongWord(word, 4)
		if strings.Join(chunks, "") != word {
			t.Errorf("ANSI content lost: %q", chunks)
		}
		if len(chunks) != 2 {
			t.Fatalf("expected 2 chunks, got %q", chunks)
		}
		if visualWidth(chunks[0]) != 4 || visualWidth(chunks[1]) != 2 {
			t.Errorf("chunk visual widths = %d,%d, expected 4,2", visualWidth(chunks[0]), visualWidth(chunks[1]))
		}
	})

	t.Run("zero width returns word unchanged", func(t *testing.T) {
		chunks := breakLongWord("abc", 0)
		if len(chunks) != 1 || chunks[0] != "abc" {
			t.Errorf("got %q, expected [abc]", chunks)
		}
	})
}

// Tests for truncateTail, the package-level rune-aware tail-truncation helper
// used by the status bar to shorten long symlink targets — see tfe-c8x.
func TestTruncateTail(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		targetWidth int
		want        string
	}{
		{"fits unchanged", "/short", 10, "/short"},
		{"exact fit", "/short", 6, "/short"},
	}
	for _, tt := range tests {
		got := truncateTail(tt.input, tt.targetWidth)
		if got != tt.want {
			t.Errorf("truncateTail(%q, %d) = %q, want %q", tt.input, tt.targetWidth, got, tt.want)
		}
	}

	t.Run("ascii truncates with ellipsis prefix", func(t *testing.T) {
		input := "/very/long/path/to/some/file.txt"
		got := truncateTail(input, 12)
		if !strings.HasPrefix(got, "...") {
			t.Errorf("expected ... prefix, got %q", got)
		}
		if vw := visualWidth(got); vw > 12 {
			t.Errorf("visual width %d exceeds target 12: %q", vw, got)
		}
		if !strings.HasSuffix(input, strings.TrimPrefix(got, "...")) {
			t.Errorf("%q is not an ellipsis plus a suffix of input", got)
		}
	})
}

func TestTruncateTailMultibyte(t *testing.T) {
	// Non-ASCII path: byte slicing would split runes and produce invalid UTF-8.
	input := "/home/użytkownik/projekty/ważne/dokumenty"
	for width := 4; width < 30; width++ {
		got := truncateTail(input, width)
		if !utf8.ValidString(got) {
			t.Errorf("width %d: result is invalid UTF-8: %q", width, got)
		}
		if !strings.HasPrefix(got, "...") {
			t.Errorf("width %d: expected ... prefix, got %q", width, got)
		}
		if vw := visualWidth(got); vw > width {
			t.Errorf("width %d: visual width %d exceeds target: %q", width, vw, got)
		}
		if !strings.HasSuffix(input, strings.TrimPrefix(got, "...")) {
			t.Errorf("width %d: %q is not an ellipsis plus a suffix of input", width, got)
		}
	}
}

func TestTruncateTailWideRunes(t *testing.T) {
	// CJK runes are 2 cells wide; ensure accounting is by visual width, not runes.
	input := "/路径/非常/长的/目录"
	got := truncateTail(input, 8)
	if !utf8.ValidString(got) {
		t.Fatalf("result is invalid UTF-8: %q", got)
	}
	if vw := visualWidth(got); vw > 8 {
		t.Errorf("visual width %d exceeds target 8: %q", vw, got)
	}
	if !strings.HasPrefix(got, "...") {
		t.Errorf("expected ... prefix, got %q", got)
	}
}

func TestTruncateTailEmojiTail(t *testing.T) {
	// Emoji are wide (2 cells); a trailing emoji must never be split mid-rune.
	input := "/projects/release-🚀-final/notes-📝.md"
	for width := 5; width < 25; width++ {
		got := truncateTail(input, width)
		if !utf8.ValidString(got) {
			t.Errorf("width %d: result is invalid UTF-8: %q", width, got)
		}
		if vw := visualWidth(got); vw > width {
			t.Errorf("width %d: visual width %d exceeds target: %q", width, vw, got)
		}
		if !strings.HasSuffix(input, strings.TrimPrefix(got, "...")) {
			t.Errorf("width %d: %q is not an ellipsis plus a suffix of input", width, got)
		}
	}
}

func TestTruncateTailTinyWidth(t *testing.T) {
	// Widths too small for "..." plus content must not panic or overflow.
	for width := 0; width <= 3; width++ {
		got := truncateTail("/some/long/path", width)
		if vw := visualWidth(got); vw > width {
			t.Errorf("width %d: visual width %d exceeds target: %q", width, vw, got)
		}
	}
}

// --- truncateANSIAware (shared core) ---------------------------------------

// legacyTruncateToWidth is the pre-refactor implementation of truncateToWidth,
// reproduced verbatim so the consolidated core can be proven byte-for-byte
// identical against it across a matrix of inputs.
func legacyTruncateToWidth(s string, targetWidth int, runeW func(rune) int) string {
	width := 0
	result := ""
	inAnsi := false
	for _, ch := range s {
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
		charWidth := 1
		if ch == '\t' {
			charWidth = 8 - (width % 8)
		} else {
			charWidth = runeW(ch)
		}
		if width+charWidth > targetWidth {
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

// legacyTruncateToVisualWidth is the pre-refactor implementation of
// truncateToVisualWidth (reset-on-overflow, no tab special-casing).
func legacyTruncateToVisualWidth(s string, targetWidth int, runeW func(rune) int) string {
	var result strings.Builder
	visualWidth := 0
	inEscape := false
	escapeSeq := strings.Builder{}
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '\x1b' {
			inEscape = true
			escapeSeq.Reset()
			escapeSeq.WriteRune(r)
			continue
		}
		if inEscape {
			escapeSeq.WriteRune(r)
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
				result.WriteString(escapeSeq.String())
			}
			continue
		}
		charWidth := runeW(r)
		if visualWidth+charWidth > targetWidth {
			result.WriteString("\033[0m")
			break
		}
		result.WriteRune(r)
		visualWidth += charWidth
	}
	return result.String()
}

func TestTruncateANSIAware(t *testing.T) {
	plainRuneW := func(r rune) int { return runeWidthASCIIish(r) }

	tests := []struct {
		name        string
		input       string
		targetWidth int
		ellipsis    string
		expandTabs  bool
		want        string
	}{
		{"empty", "", 5, "...", true, ""},
		{"fits exactly no ellipsis", "hello", 5, "...", true, "hello"},
		{"fits under", "hi", 5, "...", true, "hi"},
		// Ellipsis is appended only when the overflowing rune leaves >= 3 free
		// cells (visualWidth("...") == 3). With ASCII/CJK runes (width 1-2) an
		// overflow never leaves 3+ free cells, so the "..." path is exercised
		// via a wide-rune runeW in the ellipsisGuard sub-test below; here we
		// assert the common cases where nothing is appended.
		{"no ellipsis when budget exactly filled", "hello world", 8, "...", true, "hello wo"},
		{"ellipsis dropped when under 3 cells remain", "hello", 4, "...", true, "hell"},
		// Reset ellipsis (visual width 0) is appended on any overflow.
		{"reset ellipsis appended on overflow", "hello", 4, "\033[0m", false, "hell\033[0m"},
		{"reset ellipsis at zero width", "hi", 0, "\033[0m", false, "\033[0m"},
		{"reset no ellipsis when whole string fits", "hi", 5, "\033[0m", false, "hi"},
		// ANSI escapes are preserved verbatim and never counted toward width.
		{"ansi preserved within budget", "\033[31mred\033[0m", 3, "...", true, "\033[31mred\033[0m"},
		{"ansi preserved then truncated no ellipsis", "\033[31mredtext\033[0m", 4, "...", true, "\033[31mredt"},
		// Multibyte / wide runes: a wide rune that would overflow is dropped whole.
		{"wide rune fits", "中文", 4, "...", true, "中文"},
		{"wide rune dropped at boundary", "中文x", 3, "...", true, "中"},
		{"multibyte ascii-width", "café", 4, "...", true, "café"},
		// Tab expansion: 8-col tab stop when expandTabs, else measured by runeW.
		{"tab expands to stop then b overflows", "a\tb", 8, "...", true, "a\t"},
		{"tab fits with following char", "a\tb", 9, "...", true, "a\tb"},
		{"tab not expanded counts as runeW width 0", "a\tb", 2, "...", false, "a\tb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateANSIAware(tt.input, tt.targetWidth, plainRuneW, tt.ellipsis, tt.expandTabs)
			if got != tt.want {
				t.Errorf("truncateANSIAware(%q, %d, ellipsis=%q, expandTabs=%v) = %q, want %q",
					tt.input, tt.targetWidth, tt.ellipsis, tt.expandTabs, got, tt.want)
			}
		})
	}

	// targetWidth boundary sweep: result visual width must never exceed target.
	for _, in := range []string{"hello world", "\033[31mred\033[0m text", "中文字符串", "café résumé"} {
		for w := 0; w <= 14; w++ {
			got := truncateANSIAware(in, w, plainRuneW, "...", true)
			if vw := visualWidth(got); vw > w {
				t.Errorf("boundary: input %q width %d -> %q has visual width %d > %d", in, w, got, vw, w)
			}
		}
	}

	// ellipsisGuard: drive a runeW that reports width-4 runes so an overflow
	// can leave >= 3 free cells, exercising the "append ellipsis" branch that
	// ASCII/CJK widths can't reach.
	t.Run("ellipsisGuard", func(t *testing.T) {
		wideRuneW := func(r rune) int { return 4 } // every rune is 4 cells
		// "ABC" at width 7: 'A' -> w=4, 'B' would make 8 > 7, remaining 7-4=3
		// >= 3 so "..." is appended.
		if got := truncateANSIAware("ABC", 7, wideRuneW, "...", false); got != "A..." {
			t.Errorf("ellipsis append: got %q, want %q", got, "A...")
		}
		// At width 6: 'B' overflow leaves 6-4=2 < 3, so no ellipsis.
		if got := truncateANSIAware("ABC", 6, wideRuneW, "...", false); got != "A" {
			t.Errorf("ellipsis drop: got %q, want %q", got, "A")
		}
	})
}

// runeWidthASCIIish is a deterministic, model-free per-rune width used by the
// core test so results don't depend on terminal detection. It delegates to the
// runewidth library directly (control chars 0, wide CJK 2, everything else 1).
func runeWidthASCIIish(r rune) int {
	return runewidth.RuneWidth(r)
}

// TestTruncateWrappersMatchLegacy proves the three thin wrappers produce
// byte-for-byte identical output to their pre-refactor implementations across a
// matrix of well-formed inputs and target widths.
func TestTruncateWrappersMatchLegacy(t *testing.T) {
	m := model{terminalType: terminalWezTerm}

	inputs := []string{
		"",
		"plain ascii string",
		"\033[31mcolored\033[0m text",
		"\033[1m\033[38;5;220mbold yellow\033[0m",
		"中文字符串测试",
		"café résumé naïve",
		"a\tb\tc",
		"emoji 🐹 and 📦 icons",
		"mixed \033[32m中文\033[0m text",
	}

	for _, in := range inputs {
		for w := 0; w <= 20; w++ {
			if got, want := truncateToWidth(in, w), legacyTruncateToWidth(in, w, runewidth.RuneWidth); got != want {
				t.Errorf("truncateToWidth(%q, %d) = %q, legacy = %q", in, w, got, want)
			}
			if got, want := m.truncateToWidthCompensated(in, w), legacyTruncateToWidth(in, w, m.runeWidth); got != want {
				t.Errorf("truncateToWidthCompensated(%q, %d) = %q, legacy = %q", in, w, got, want)
			}
			if got, want := m.truncateToVisualWidth(in, w), legacyTruncateToVisualWidth(in, w, m.runeWidth); got != want {
				t.Errorf("truncateToVisualWidth(%q, %d) = %q, legacy = %q", in, w, got, want)
			}
		}
	}
}
