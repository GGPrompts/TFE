package main

import (
	"strings"
	"testing"
	"unicode/utf8"
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
