package main

import (
	"strings"
	"testing"
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
