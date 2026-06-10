package main

import (
	"encoding/json"
	"strings"
	"testing"
	"unicode/utf8"
)

// jsonlCacheTestModel builds a model with a loaded JSONL preview in the given
// layout, with pane widths derived from calculateLayout() (the real layout
// math), and parsed messages already cached.
func jsonlCacheTestModel(width int, vm viewMode, dm displayMode) model {
	m := model{
		width:       width,
		height:      40,
		viewMode:    vm,
		displayMode: dm,
		focusedPane: leftPane,
		preview: previewModel{
			loaded:  true,
			isJSONL: true,
			cachedJSONLMessages: []jsonlMessage{
				{Type: "user", Message: []byte(`{"role":"user","content":"hello there, this is a user message long enough to wrap at narrow preview widths"}`)},
				{Type: "assistant", Message: []byte(`{"role":"assistant","content":[{"type":"text","text":"hi! here is an assistant reply that also has enough text to wrap"}]}`)},
				{Type: "system", Subtype: "stop_hook_summary"},
			},
		},
	}
	m.calculateLayout()
	return m
}

// TestJSONLRenderCache_PopulatedAndHit is a regression test for tfe-p5u:
// renderJSONLPreview and getWrappedLineCount used to re-render the ENTIRE
// transcript (JSON decode + wrap + style) on every View() frame and every
// scroll keystroke. Both must now read the rendered-line cache populated by
// populateJSONLRenderCache().
func TestJSONLRenderCache_PopulatedAndHit(t *testing.T) {
	layouts := []struct {
		name string
		m    model
	}{
		{"full preview", jsonlCacheTestModel(120, viewFullPreview, modeList)},
		{"dual-pane horizontal split", jsonlCacheTestModel(120, viewDualPane, modeList)},
		{"dual-pane vertical split (detail)", jsonlCacheTestModel(120, viewDualPane, modeDetail)},
		{"dual-pane narrow terminal", jsonlCacheTestModel(80, viewDualPane, modeList)},
	}

	for _, tc := range layouts {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.m

			// populatePreviewCache must route JSONL previews to the
			// rendered-line cache at exactly jsonlPreviewWidth().
			m.populatePreviewCache()
			if m.preview.cachedJSONLRenderedWidth != m.jsonlPreviewWidth() {
				t.Errorf("cachedJSONLRenderedWidth = %d, but jsonlPreviewWidth() = %d",
					m.preview.cachedJSONLRenderedWidth, m.jsonlPreviewWidth())
			}
			if len(m.preview.cachedJSONLRenderedLines) == 0 {
				t.Fatal("populatePreviewCache left cachedJSONLRenderedLines empty")
			}

			// getWrappedLineCount must hit the cache: plant a sentinel slice
			// and verify its length is returned (a miss would re-render).
			realLines := m.preview.cachedJSONLRenderedLines
			sentinel := []string{"SENTINEL_1", "SENTINEL_2", "SENTINEL_3", "SENTINEL_4", "SENTINEL_5"}
			m.preview.cachedJSONLRenderedLines = sentinel
			if got := m.getWrappedLineCount(); got != len(sentinel) {
				t.Errorf("getWrappedLineCount() = %d, expected sentinel length %d (cache miss: width formula disagrees with populateJSONLRenderCache)",
					got, len(sentinel))
			}

			// renderJSONLPreview must hit the same cache: the sentinel lines
			// must appear in the output instead of a fresh render.
			out := m.renderJSONLPreview(5)
			if !strings.Contains(out, "SENTINEL_1") {
				t.Errorf("renderJSONLPreview() did not use cached rendered lines\noutput:\n%s", out)
			}
			m.preview.cachedJSONLRenderedLines = realLines
		})
	}
}

// TestJSONLRenderCache_RefreshOnWidthChange verifies refreshPreviewCacheIfStale
// repopulates the JSONL rendered-line cache when a layout change (e.g. toggling
// between horizontal and vertical split) changes the preview width.
func TestJSONLRenderCache_RefreshOnWidthChange(t *testing.T) {
	m := jsonlCacheTestModel(120, viewDualPane, modeList)
	m.populateJSONLRenderCache()
	oldWidth := m.preview.cachedJSONLRenderedWidth

	// Switch to a layout with a different preview width
	m.displayMode = modeDetail // vertical split: width formula changes
	m.calculateLayout()
	if m.jsonlPreviewWidth() == oldWidth {
		t.Fatal("test setup: layout change did not change jsonlPreviewWidth")
	}

	m.refreshPreviewCacheIfStale()
	if m.preview.cachedJSONLRenderedWidth != m.jsonlPreviewWidth() {
		t.Errorf("after refresh, cachedJSONLRenderedWidth = %d, expected %d",
			m.preview.cachedJSONLRenderedWidth, m.jsonlPreviewWidth())
	}

	// And when the width already matches, the cache must be left alone.
	sentinel := []string{"SENTINEL"}
	m.preview.cachedJSONLRenderedLines = sentinel
	m.refreshPreviewCacheIfStale()
	if len(m.preview.cachedJSONLRenderedLines) != 1 || m.preview.cachedJSONLRenderedLines[0] != "SENTINEL" {
		t.Error("refreshPreviewCacheIfStale re-rendered the JSONL cache even though the width already matched")
	}
}

// TestJSONLRenderedLines_FallbackOnStaleWidth verifies the View-path fallback:
// if a View happens before Update refreshed the cache, jsonlRenderedLines must
// render fresh at the current width rather than returning stale lines.
func TestJSONLRenderedLines_FallbackOnStaleWidth(t *testing.T) {
	m := jsonlCacheTestModel(120, viewDualPane, modeList)
	m.populateJSONLRenderCache()

	// Plant stale cache content at a width that cannot match.
	m.preview.cachedJSONLRenderedWidth = m.jsonlPreviewWidth() + 1
	m.preview.cachedJSONLRenderedLines = []string{"STALE"}

	lines := m.jsonlRenderedLines()
	if len(lines) == 1 && lines[0] == "STALE" {
		t.Error("jsonlRenderedLines returned stale cached lines despite width mismatch")
	}
	if len(lines) == 0 {
		t.Error("jsonlRenderedLines fallback render produced no lines")
	}
}

// assertValidUTF8Lines fails if any rendered line contains invalid UTF-8 or a
// replacement character — the symptom of byte-slicing multibyte runes (tfe-7hx).
func assertValidUTF8Lines(t *testing.T, lines []string, context string) {
	t.Helper()
	for i, line := range lines {
		if !utf8.ValidString(line) {
			t.Errorf("%s: line %d is not valid UTF-8 (rune split by byte slicing): %q", context, i, line)
		}
		if strings.ContainsRune(line, utf8.RuneError) {
			t.Errorf("%s: line %d contains U+FFFD replacement character: %q", context, i, line)
		}
	}
}

// TestRenderToolUseSummary_MultibyteDetail is a regression test for tfe-7hx:
// renderToolUseSummary used to chunk detail text at fixed byte offsets, which
// split CJK/emoji runes at every chunk boundary, rendering replacement-char
// pairs on consecutive lines.
func TestRenderToolUseSummary_MultibyteDetail(t *testing.T) {
	// A Bash command full of wide multibyte runes, long enough to wrap several times
	cmd := strings.Repeat("echo \"日本語のテキストと絵文字🎉を含むコマンド\" && ", 6)
	block := jsonlContentBlock{
		Type:  "tool_use",
		Name:  "Bash",
		Input: []byte(`{"command":` + string(mustJSONString(cmd)) + `}`),
	}

	width := 60
	lines := renderToolUseSummary(block, width)
	if len(lines) < 2 {
		t.Fatalf("expected multibyte detail to wrap to multiple lines, got %d", len(lines))
	}
	assertValidUTF8Lines(t, lines, "renderToolUseSummary")

	// Each line must stay within the column budget (width), not just byte budget
	for i, line := range lines {
		if w := visualWidth(line); w > width {
			t.Errorf("line %d exceeds width budget: visualWidth=%d > %d: %q", i, w, width, line)
		}
	}
}

// TestRenderToolUseSummary_SingleLineUsesVisualWidth verifies the single-line
// case compares visual width, not byte length: a CJK detail whose byte length
// exceeds firstLineMax but whose column width fits must stay on one line.
func TestRenderToolUseSummary_SingleLineUsesVisualWidth(t *testing.T) {
	// 15 CJK runes: 45 bytes, but only 30 columns
	detail := strings.Repeat("語", 15)
	block := jsonlContentBlock{
		Type:  "tool_use",
		Name:  "Bash",
		Input: []byte(`{"command":"` + detail + `"}`),
	}
	width := 60 // firstLineMax = 60 - visualWidth("Bash") - 5 = 51 columns
	lines := renderToolUseSummary(block, width)
	if len(lines) != 1 {
		t.Errorf("30-column detail should fit on one line at width %d, got %d lines: %q",
			width, len(lines), lines)
	}
	assertValidUTF8Lines(t, lines, "renderToolUseSummary single-line")
}

// TestRenderJSONLToolResult_MultibyteTruncation is a regression test for
// tfe-7hx: tool results were truncated with resultText[:maxLen], splitting
// multibyte runes at the cut point.
func TestRenderJSONLToolResult_MultibyteTruncation(t *testing.T) {
	result := strings.Repeat("結果テキスト🎉", 20)
	msg := jsonlMessage{
		Type: "user",
		Message: []byte(`{"role":"user","content":[{"type":"tool_result","text":` +
			string(mustJSONString(result)) + `}]}`),
	}
	width := 50
	lines := renderJSONLUserMessage(msg, width)
	if len(lines) == 0 {
		t.Fatal("expected tool result line, got none")
	}
	assertValidUTF8Lines(t, lines, "tool_result")
	for i, line := range lines {
		// "  " prefix + truncated text must stay within width-4+2 columns
		if w := visualWidth(line); w > width {
			t.Errorf("tool result line %d exceeds width: visualWidth=%d > %d: %q", i, w, width, line)
		}
	}
}

// TestExtractToolDetail_AgentPromptMultibyte is a regression test for tfe-7hx:
// Agent prompts were truncated with prompt[:80] by bytes, splitting runes.
func TestExtractToolDetail_AgentPromptMultibyte(t *testing.T) {
	prompt := strings.Repeat("エージェントへの指示🤖", 15) // far over 80 columns, multibyte
	block := jsonlContentBlock{
		Type:  "tool_use",
		Name:  "Agent",
		Input: []byte(`{"prompt":` + string(mustJSONString(prompt)) + `}`),
	}
	detail := extractToolDetail(block)
	if !utf8.ValidString(detail) {
		t.Errorf("Agent prompt detail is not valid UTF-8: %q", detail)
	}
	if strings.ContainsRune(detail, utf8.RuneError) {
		t.Errorf("Agent prompt detail contains U+FFFD: %q", detail)
	}
	if w := visualWidth(detail); w > 80 {
		t.Errorf("Agent prompt detail exceeds 80 columns: visualWidth=%d: %q", w, detail)
	}
}

// TestChunkByVisualWidth covers the rune-aware chunking helper directly.
func TestChunkByVisualWidth(t *testing.T) {
	t.Run("ascii respects first and rest widths", func(t *testing.T) {
		chunks := chunkByVisualWidth(strings.Repeat("a", 25), 10, 8)
		want := []string{strings.Repeat("a", 10), strings.Repeat("a", 8), strings.Repeat("a", 7)}
		if len(chunks) != len(want) {
			t.Fatalf("got %d chunks %q, want %d", len(chunks), chunks, len(want))
		}
		for i := range want {
			if chunks[i] != want[i] {
				t.Errorf("chunk %d = %q, want %q", i, chunks[i], want[i])
			}
		}
	})

	t.Run("wide runes never split and never exceed budget", func(t *testing.T) {
		chunks := chunkByVisualWidth(strings.Repeat("漢", 20), 7, 7)
		for i, c := range chunks {
			if !utf8.ValidString(c) {
				t.Errorf("chunk %d is invalid UTF-8: %q", i, c)
			}
			if w := visualWidth(c); w > 7 {
				t.Errorf("chunk %d width %d exceeds 7: %q", i, w, c)
			}
		}
		if got := strings.Join(chunks, ""); got != strings.Repeat("漢", 20) {
			t.Errorf("chunks do not reassemble to original input: %q", got)
		}
	})

	t.Run("empty input yields single empty chunk", func(t *testing.T) {
		chunks := chunkByVisualWidth("", 10, 10)
		if len(chunks) != 1 || chunks[0] != "" {
			t.Errorf("got %q, want one empty chunk", chunks)
		}
	})

	t.Run("oversized single rune still emits", func(t *testing.T) {
		chunks := chunkByVisualWidth("漢", 1, 1)
		if len(chunks) != 1 || chunks[0] != "漢" {
			t.Errorf("got %q, want [\"漢\"]", chunks)
		}
	})
}

// mustJSONString JSON-encodes a string for embedding in raw test JSON.
func mustJSONString(s string) []byte {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return b
}
