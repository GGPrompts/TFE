package main

import (
	"strings"
	"testing"
)

// diffPreviewTestModel builds a model in changes mode with diff preview
// enabled and a preview target set, with pane widths from calculateLayout().
func diffPreviewTestModel(width int, vm viewMode, dm displayMode) model {
	m := previewWidthTestModel(width, vm, dm)
	m.showChangesOnly = true
	m.showDiffPreview = true
	m.preview.filePath = "/tmp/example.go"
	m.changedFiles = []fileItem{
		{name: "[ M] example.go", path: "/tmp/example.go"},
	}
	return m
}

// TestDiffPreviewAvailableWidth verifies the diff width helper derives from
// previewBoxContentWidth() (scrollbar + space = 2 chars overhead) and clamps
// to the minimum, so the cache-fill and render paths always agree.
func TestDiffPreviewAvailableWidth(t *testing.T) {
	for _, tc := range previewWidthLayouts() {
		t.Run(tc.name, func(t *testing.T) {
			expected := tc.m.previewBoxContentWidth() - 2
			if expected < 20 {
				expected = 20
			}
			if got := tc.m.diffPreviewAvailableWidth(); got != expected {
				t.Errorf("diffPreviewAvailableWidth() = %d, expected previewBoxContentWidth()-2 = %d", got, expected)
			}
		})
	}

	tiny := previewWidthTestModel(10, viewFullPreview, modeList)
	if got := tiny.diffPreviewAvailableWidth(); got != 20 {
		t.Errorf("diffPreviewAvailableWidth() at tiny terminal = %d, expected clamped minimum 20", got)
	}
}

// TestWrapAndStyleDiffLines verifies the cache-fill helper wraps long diff
// lines to the available width and keeps short lines intact.
func TestWrapAndStyleDiffLines(t *testing.T) {
	diff := strings.Join([]string{
		"diff --git a/example.go b/example.go",
		"@@ -1,2 +1,2 @@",
		"-old line",
		"+new line",
		" context",
	}, "\n") + "\n"

	lines := wrapAndStyleDiffLines(diff, 80)
	if len(lines) != 5 {
		t.Fatalf("wrapAndStyleDiffLines returned %d lines, expected 5", len(lines))
	}
	if !strings.Contains(lines[2], "old line") {
		t.Errorf("removed line missing from output: %q", lines[2])
	}
	if !strings.Contains(lines[3], "new line") {
		t.Errorf("added line missing from output: %q", lines[3])
	}

	// A long added line must wrap into multiple lines, each within width
	longDiff := "+" + strings.Repeat("word ", 30) + "\n"
	wrapped := wrapAndStyleDiffLines(longDiff, 40)
	if len(wrapped) < 2 {
		t.Fatalf("long diff line did not wrap: got %d lines", len(wrapped))
	}
	for i, line := range wrapped {
		if visualWidth(line) > 40 {
			t.Errorf("wrapped diff line %d exceeds width 40: %d (%q)", i, visualWidth(line), line)
		}
	}
}

// TestRefreshDiffPreviewCache_RenderUsesCache is a regression test for
// tfe-4nb: renderDiffPreview used to exec git (rev-parse + up to 3 git diff
// subprocesses) and re-wrap the entire diff on every View() frame. The diff
// must instead come from the cache filled by refreshDiffPreviewCacheIfStale,
// keyed on (filePath, gitStatusCode) for the text and width for the lines.
func TestRefreshDiffPreviewCache_RenderUsesCache(t *testing.T) {
	m := diffPreviewTestModel(120, viewDualPane, modeDetail)

	// Pretend the diff text was already fetched for this (path, status) so
	// the wrap step runs without spawning git
	m.preview.diffLoaded = true
	m.preview.diffForPath = m.preview.filePath
	m.preview.diffForStatus = m.previewDiffStatusCode()
	m.preview.diffContent = "@@ -1 +1 @@\n-old\n+new\n"

	m.refreshDiffPreviewCacheIfStale()
	if !m.preview.diffCacheValid {
		t.Fatal("refreshDiffPreviewCacheIfStale left diff cache invalid")
	}
	if m.preview.cachedDiffWidth != m.diffPreviewAvailableWidth() {
		t.Errorf("cachedDiffWidth = %d, but diffPreviewAvailableWidth() = %d (cache will miss on render)",
			m.preview.cachedDiffWidth, m.diffPreviewAvailableWidth())
	}

	// renderDiffPreview must hit the cache: plant a sentinel line and verify
	// it is rendered instead of re-fetching/re-wrapping the diff
	const sentinelLine = "DIFF_CACHE_SENTINEL"
	m.preview.cachedDiffLines = []string{sentinelLine}
	out := m.renderDiffPreview(5)
	if !strings.Contains(out, sentinelLine) {
		t.Errorf("renderDiffPreview() did not use cached diff lines (cache miss: width or key disagrees with refreshDiffPreviewCacheIfStale)\noutput:\n%s", out)
	}

	// When the cache already matches, refresh must be a no-op
	m.refreshDiffPreviewCacheIfStale()
	if len(m.preview.cachedDiffLines) != 1 || m.preview.cachedDiffLines[0] != sentinelLine {
		t.Error("refreshDiffPreviewCacheIfStale rewrapped the diff even though key and width already matched")
	}

	// Invalidation must drop both the raw diff and the wrapped lines
	m.invalidateDiffPreviewCache()
	if m.preview.diffLoaded || m.preview.diffCacheValid || m.preview.cachedDiffLines != nil {
		t.Error("invalidateDiffPreviewCache did not clear the diff cache")
	}
}

// TestRefreshDiffPreviewCache_WidthChange verifies a layout change refills
// the wrapped lines at the new width without re-fetching the diff text.
func TestRefreshDiffPreviewCache_WidthChange(t *testing.T) {
	m := diffPreviewTestModel(120, viewDualPane, modeList)
	m.preview.diffLoaded = true
	m.preview.diffForPath = m.preview.filePath
	m.preview.diffForStatus = m.previewDiffStatusCode()
	m.preview.diffContent = "+added\n"

	m.refreshDiffPreviewCacheIfStale()
	oldWidth := m.preview.cachedDiffWidth

	// Simulate switching to Detail mode (horizontal -> vertical split)
	m.displayMode = modeDetail
	m.calculateLayout()
	if m.diffPreviewAvailableWidth() == oldWidth {
		t.Fatal("test setup broken: layout change did not change diff preview width")
	}

	m.refreshDiffPreviewCacheIfStale()
	if m.preview.cachedDiffWidth != m.diffPreviewAvailableWidth() {
		t.Errorf("after layout change, cachedDiffWidth = %d, expected %d",
			m.preview.cachedDiffWidth, m.diffPreviewAvailableWidth())
	}
	if !m.preview.diffLoaded || m.preview.diffContent != "+added\n" {
		t.Error("width-only refresh should not have dropped the raw diff text")
	}
}

// TestRefreshDiffPreviewCache_InactiveModes verifies the refresh is a no-op
// outside changes mode / diff preview, so it never spawns git from the
// post-dispatch refresh path during normal browsing.
func TestRefreshDiffPreviewCache_InactiveModes(t *testing.T) {
	m := diffPreviewTestModel(120, viewDualPane, modeDetail)
	m.showDiffPreview = false
	m.refreshDiffPreviewCacheIfStale()
	if m.preview.diffLoaded || m.preview.diffCacheValid {
		t.Error("refreshDiffPreviewCacheIfStale filled the cache with diff preview disabled")
	}

	m.showDiffPreview = true
	m.showChangesOnly = false
	m.refreshDiffPreviewCacheIfStale()
	if m.preview.diffLoaded || m.preview.diffCacheValid {
		t.Error("refreshDiffPreviewCacheIfStale filled the cache outside changes mode")
	}
}
