package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestOpenFileWithBestTool_ViewerDispatch verifies the shared F4 dispatch
// helper (tfe-bpl) routes each special file type to a viewer command.
// The open* viewer helpers always return a non-nil tea.Cmd (they emit an
// editorFinishedMsg error command when no viewer is installed), so a
// non-nil result proves the viewer branch was taken.
func TestOpenFileWithBestTool_ViewerDispatch(t *testing.T) {
	paths := []string{
		"/tmp/data.csv",  // CSV viewer
		"/tmp/movie.mp4", // video player
		"/tmp/song.mp3",  // audio player
		"/tmp/doc.pdf",   // PDF viewer
		"/tmp/app.db",    // database viewer
	}

	for _, path := range paths {
		m := model{}
		cmd := m.openFileWithBestTool(path)
		if cmd == nil {
			t.Errorf("openFileWithBestTool(%q) returned nil, expected viewer command", path)
		}
		if m.statusMessage != "" {
			t.Errorf("openFileWithBestTool(%q) set status %q, expected none", path, m.statusMessage)
		}
	}
}

// TestOpenFileWithBestTool_BinaryFile verifies binary files (containing
// null bytes) are routed to the hex viewer rather than a text editor.
func TestOpenFileWithBestTool_BinaryFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blob.bin")
	if err := os.WriteFile(path, []byte{0x7f, 0x00, 0x01, 0x02}, 0o644); err != nil {
		t.Fatal(err)
	}

	m := model{}
	cmd := m.openFileWithBestTool(path)
	if cmd == nil {
		t.Errorf("openFileWithBestTool(%q) returned nil, expected hex viewer command", path)
	}
}

// TestOpenFileWithBestTool_NoEditor verifies the text-file fallback when no
// editor exists on PATH: nil command plus an error status message.
func TestOpenFileWithBestTool_NoEditor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("plain text"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", t.TempDir()) // empty dir: no editors available

	m := model{}
	cmd := m.openFileWithBestTool(path)
	if cmd != nil {
		t.Error("openFileWithBestTool returned a command, expected nil with no editor on PATH")
	}
	if !m.statusIsError {
		t.Error("statusIsError = false, expected true")
	}
	if !strings.Contains(m.statusMessage, "No editor available") {
		t.Errorf("statusMessage = %q, expected it to mention no editor available", m.statusMessage)
	}
}

// TestOpenFileWithBestTool_TextEditor verifies a text file opens in an
// editor when one is on PATH (stubbed nano in a temp dir).
func TestOpenFileWithBestTool_TextEditor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("plain text"), 0o644); err != nil {
		t.Fatal(err)
	}

	binDir := t.TempDir()
	stub := filepath.Join(binDir, "nano")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	m := model{}
	cmd := m.openFileWithBestTool(path)
	if cmd == nil {
		t.Error("openFileWithBestTool returned nil, expected editor command for text file")
	}
	if m.statusMessage != "" {
		t.Errorf("statusMessage = %q, expected none", m.statusMessage)
	}
}

// newPreviewScrollModel builds a model in standalone-preview mode whose
// getWrappedLineCount() and getPreviewVisibleLines() return deterministic
// values, so the scroll-clamp helpers (tfe-o5f) can be tested in isolation.
//
// Content is numLines short single-character lines; with a wide width each
// wraps to exactly one display line, so getWrappedLineCount() == numLines.
// In previewOnly mode getPreviewVisibleLines() == max(3, height-4).
func newPreviewScrollModel(numLines, height int) model {
	content := make([]string, numLines)
	for i := range content {
		content[i] = "x"
	}
	m := model{
		previewOnly: true,
		width:       200,
		height:      height,
	}
	m.preview.loaded = true
	m.preview.content = content
	return m
}

// TestMaxPreviewScroll verifies maxPreviewScroll() == max(0, total-visible)
// and never returns a negative value.
func TestMaxPreviewScroll(t *testing.T) {
	cases := []struct {
		name      string
		numLines  int
		height    int
		wantTotal int
		wantVis   int
		want      int
	}{
		// height-4 visible lines (>=3); content overflows the viewport.
		{"overflow", 100, 24, 100, 20, 80},
		// Exactly fills the viewport: no scrolling possible.
		{"exact", 20, 24, 20, 20, 0},
		// Fewer lines than the viewport: clamps to 0, never negative.
		{"underflow", 5, 24, 5, 20, 0},
		// visibleLines floors at 3 in previewOnly mode.
		{"min-visible", 10, 5, 10, 3, 7},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newPreviewScrollModel(tc.numLines, tc.height)
			if got := m.getWrappedLineCount(); got != tc.wantTotal {
				t.Fatalf("getWrappedLineCount() = %d, want %d (setup invalid)", got, tc.wantTotal)
			}
			if got := m.getPreviewVisibleLines(); got != tc.wantVis {
				t.Fatalf("getPreviewVisibleLines() = %d, want %d (setup invalid)", got, tc.wantVis)
			}
			if got := m.maxPreviewScroll(); got != tc.want {
				t.Errorf("maxPreviewScroll() = %d, want %d", got, tc.want)
			}
			if m.maxPreviewScroll() < 0 {
				t.Errorf("maxPreviewScroll() = %d, must never be negative", m.maxPreviewScroll())
			}
		})
	}
}

// TestScrollPreviewBy verifies add-then-clamp behavior: the result is always
// clamped into [0, maxPreviewScroll()], matching every call site it replaced.
func TestScrollPreviewBy(t *testing.T) {
	// 100 lines, height 24 => maxScroll 80.
	cases := []struct {
		name  string
		start int
		delta int
		want  int
	}{
		{"down-one", 0, 1, 1},
		{"up-one-from-mid", 10, -1, 9},
		{"down-three-wheel", 0, 3, 3},
		{"up-three-wheel", 2, -3, 0},     // clamps at lower bound
		{"page-down", 0, 20, 20},
		{"down-past-end-clamps", 70, 20, 80},
		{"up-below-zero-clamps", 5, -20, 0},
		{"at-max-stays", 80, 5, 80},
		{"at-zero-up-stays", 0, -1, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newPreviewScrollModel(100, 24)
			m.preview.scrollPos = tc.start
			m.scrollPreviewBy(tc.delta)
			if m.preview.scrollPos != tc.want {
				t.Errorf("scrollPreviewBy(%d) from %d = %d, want %d", tc.delta, tc.start, m.preview.scrollPos, tc.want)
			}
		})
	}
}

// TestScrollPreviewByPage verifies page scrolling moves by getPreviewVisibleLines()
// in the requested direction, clamped to bounds.
func TestScrollPreviewByPage(t *testing.T) {
	m := newPreviewScrollModel(100, 24) // visible 20, maxScroll 80
	if v := m.getPreviewVisibleLines(); v != 20 {
		t.Fatalf("setup: getPreviewVisibleLines() = %d, want 20", v)
	}

	m.preview.scrollPos = 0
	m.scrollPreviewByPage(1)
	if m.preview.scrollPos != 20 {
		t.Errorf("page down from 0 = %d, want 20", m.preview.scrollPos)
	}

	m.scrollPreviewByPage(-1)
	if m.preview.scrollPos != 0 {
		t.Errorf("page up back to 0 = %d, want 0", m.preview.scrollPos)
	}

	// Page up at the top stays at 0.
	m.scrollPreviewByPage(-1)
	if m.preview.scrollPos != 0 {
		t.Errorf("page up at top = %d, want 0", m.preview.scrollPos)
	}

	// Page down near the end clamps to maxScroll.
	m.preview.scrollPos = 75
	m.scrollPreviewByPage(1)
	if m.preview.scrollPos != 80 {
		t.Errorf("page down near end = %d, want 80 (clamped)", m.preview.scrollPos)
	}
}

// TestScrollPreviewToBottom verifies it lands exactly on maxPreviewScroll().
func TestScrollPreviewToBottom(t *testing.T) {
	m := newPreviewScrollModel(100, 24)
	m.preview.scrollPos = 0
	m.scrollPreviewToBottom()
	if want := m.maxPreviewScroll(); m.preview.scrollPos != want {
		t.Errorf("scrollPreviewToBottom() = %d, want %d", m.preview.scrollPos, want)
	}

	// When content fits entirely, bottom is 0.
	m2 := newPreviewScrollModel(5, 24)
	m2.preview.scrollPos = 3
	m2.scrollPreviewToBottom()
	if m2.preview.scrollPos != 0 {
		t.Errorf("scrollPreviewToBottom() with fitting content = %d, want 0", m2.preview.scrollPos)
	}
}

// newDisplayModeModel returns a model with enough state for setDisplayMode
// (calculateLayout needs width/height) starting in the given mode.
func newDisplayModeModel(start displayMode) model {
	return model{
		width:        120,
		height:       40,
		currentPath:  "/tmp",
		displayMode:  start,
		expandedDirs: make(map[string]bool),
	}
}

// TestSetDisplayMode_DetailScrollXReset verifies the drift fix at the heart of
// tfe-gib: entering detail view (from any mode, including via the keyboard '2'
// path that previously skipped it) always resets the horizontal scroll offset.
func TestSetDisplayMode_DetailScrollXReset(t *testing.T) {
	for _, start := range []displayMode{modeList, modeDetail, modeTree} {
		m := newDisplayModeModel(start)
		m.detailScrollX = 42 // stale offset left by detail-mode arrow keys
		m.setDisplayMode(modeDetail)
		if m.displayMode != modeDetail {
			t.Errorf("from %v: displayMode = %v, want modeDetail", start, m.displayMode)
		}
		if m.detailScrollX != 0 {
			t.Errorf("from %v: detailScrollX = %d, want 0 (entering detail must reset)", start, m.detailScrollX)
		}
	}
}

// TestSetDisplayMode_DetailScrollXPreservedWithinDetail verifies that switching
// to a non-detail mode does not reset detailScrollX (only ENTERING detail does),
// and that a stale offset on a non-detail mode is left untouched.
func TestSetDisplayMode_NonDetailLeavesScrollUntouched(t *testing.T) {
	m := newDisplayModeModel(modeList)
	m.detailScrollX = 7
	m.setDisplayMode(modeTree)
	if m.detailScrollX != 7 {
		t.Errorf("switching to tree changed detailScrollX to %d, want 7 (unchanged)", m.detailScrollX)
	}
}

// TestSetDisplayMode_ExpandedDirsResetOnLeavingTree verifies tree expansion is
// cleared only when LEAVING tree view, and preserved otherwise.
func TestSetDisplayMode_ExpandedDirsResetOnLeavingTree(t *testing.T) {
	// Leaving tree -> expansion reset.
	for _, target := range []displayMode{modeList, modeDetail} {
		m := newDisplayModeModel(modeTree)
		m.expandedDirs["/tmp/foo"] = true
		m.setDisplayMode(target)
		if len(m.expandedDirs) != 0 {
			t.Errorf("tree -> %v: expandedDirs not reset (len=%d)", target, len(m.expandedDirs))
		}
		if !m.treeItemsDirty {
			t.Errorf("tree -> %v: treeItemsDirty not set after resetting expansion", target)
		}
	}

	// Not leaving tree (list -> tree) -> expansion preserved.
	m := newDisplayModeModel(modeList)
	m.expandedDirs["/tmp/bar"] = true
	m.setDisplayMode(modeTree)
	if !m.expandedDirs["/tmp/bar"] {
		t.Errorf("list -> tree: expandedDirs was reset, want preserved")
	}

	// Tree -> tree (no-op target) -> expansion preserved.
	m2 := newDisplayModeModel(modeTree)
	m2.expandedDirs["/tmp/baz"] = true
	m2.setDisplayMode(modeTree)
	if !m2.expandedDirs["/tmp/baz"] {
		t.Errorf("tree -> tree: expandedDirs was reset, want preserved")
	}
}
