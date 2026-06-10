package main

import (
	"strings"
	"testing"
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
