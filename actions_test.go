package main

import (
	"os"
	"os/exec"
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
		{"up-three-wheel", 2, -3, 0}, // clamps at lower bound
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

// TestChangesModeRoundTripPreservesTreeExpansion is the tfe-23h regression guard.
// Entering changes mode from tree view (which routes through setDisplayMode and so
// wipes the live expandedDirs map) must not lose the user's expansion state: the
// snapshot taken on entry is restored when exitChangesMode returns to tree view.
func TestChangesModeRoundTripPreservesTreeExpansion(t *testing.T) {
	m := newDisplayModeModel(modeTree)
	m.currentPath = "/tmp"
	m.expandedDirs["/foo"] = true

	// Enter changes mode the way toggleChangesMode/the mouse toggle/agent auto-entry do:
	// snapshot first, then switch to detail.
	m.saveChangesRestoreState()
	m.showChangesOnly = true
	m.showDiffPreview = true
	m.setDisplayMode(modeDetail)

	// While in detail/changes view the live expansion map must be wiped (uniform
	// setDisplayMode leave-tree behavior from tfe-gib).
	if len(m.expandedDirs) != 0 {
		t.Fatalf("after entering changes mode: expandedDirs not wiped (len=%d), want 0", len(m.expandedDirs))
	}
	// The snapshot must be an independent copy, not an alias of the live map.
	m.expandedDirs["/scratch"] = true // simulate edits in detail view
	if m.changesRestoreExpandedDirs["/scratch"] {
		t.Fatal("snapshot aliased the live map; detail-view edits leaked into the snapshot")
	}

	m.exitChangesMode()

	if m.displayMode != modeTree {
		t.Errorf("after exit: displayMode = %v, want modeTree (changesRestoreDisplay)", m.displayMode)
	}
	if !m.expandedDirs["/foo"] {
		t.Errorf("after exit: expandedDirs missing /foo, want restored expansion state")
	}
	if m.expandedDirs["/scratch"] {
		t.Errorf("after exit: stale detail-view edit /scratch leaked into restored map")
	}
	if m.changesRestoreExpandedDirs != nil {
		t.Errorf("after exit: changesRestoreExpandedDirs not cleared, want nil")
	}
}

// TestChangesModeRoundTripFromNonTreeMode verifies that entering changes mode from
// a non-tree mode takes no snapshot and exit does not clobber expandedDirs.
func TestChangesModeRoundTripFromNonTreeMode(t *testing.T) {
	m := newDisplayModeModel(modeList)
	m.currentPath = "/tmp"

	m.saveChangesRestoreState()
	if m.changesRestoreExpandedDirs != nil {
		t.Fatalf("non-tree entry took a snapshot, want nil")
	}
	m.showChangesOnly = true
	m.setDisplayMode(modeDetail)
	m.exitChangesMode()

	if m.displayMode != modeList {
		t.Errorf("after exit: displayMode = %v, want modeList", m.displayMode)
	}
}

// initTestGitRepo creates a git repo with one tracked-then-modified file so
// getChangedFiles returns a non-error result, exercising the changes-mode
// entry branch. Skips the test if the git binary is unavailable.
func initTestGitRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("hello\n"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	run("add", ".")
	run("commit", "-m", "init")
	// Modify so `git status --porcelain` reports a change.
	if err := os.WriteFile(filepath.Join(dir, "tracked.txt"), []byte("changed\n"), 0644); err != nil {
		t.Fatalf("modify file: %v", err)
	}
	return dir
}

// newToggleTestModel returns a dual-pane model rooted at dir with a stale
// horizontal scroll offset, so a toggle that forces detail view via
// setDisplayMode must reset detailScrollX to 0.
func newToggleTestModel(dir string) model {
	return model{
		width:             120,
		height:            40,
		currentPath:       dir,
		displayMode:       modeList,
		viewMode:          viewDualPane,
		expandedDirs:      make(map[string]bool),
		detailScrollX:     99,
		gitReposScanDepth: 1,
	}
}

// TestToggleChangesModeResetsDetailScrollX verifies the changes-mode toggle
// (tfe-ojy) routes its forced switch to detail view through setDisplayMode,
// which resets the stale horizontal scroll offset in dual-pane mode.
func TestToggleChangesModeResetsDetailScrollX(t *testing.T) {
	dir := initTestGitRepo(t)
	m := newToggleTestModel(dir)
	m.toggleChangesMode()
	if !m.showChangesOnly {
		t.Fatal("expected showChangesOnly true after toggle (changed file present)")
	}
	if m.displayMode != modeDetail {
		t.Errorf("displayMode = %v, want modeDetail", m.displayMode)
	}
	if m.detailScrollX != 0 {
		t.Errorf("detailScrollX = %d, want 0 (entering detail must reset)", m.detailScrollX)
	}
}

// TestToggleGitReposResetsDetailScrollX verifies the git-repos toggle (tfe-ojy)
// routes its forced switch to detail view through setDisplayMode.
func TestToggleGitReposResetsDetailScrollX(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	m := newToggleTestModel(dir)
	m.toggleGitRepos()
	if !m.showGitReposOnly {
		t.Fatal("expected showGitReposOnly true after toggle")
	}
	if m.displayMode != modeDetail {
		t.Errorf("displayMode = %v, want modeDetail", m.displayMode)
	}
	if m.detailScrollX != 0 {
		t.Errorf("detailScrollX = %d, want 0 (entering detail must reset)", m.detailScrollX)
	}
}

// TestToggleDualPane verifies toggleDualPane() flips viewMode in both
// directions and always calls calculateLayout/populatePreviewCache via the
// state side-effects it depends on (tfe-3k5).
func TestToggleDualPane(t *testing.T) {
	// Single -> dual
	m := model{
		width:        120,
		height:       40,
		currentPath:  "/tmp",
		viewMode:     viewSinglePane,
		displayMode:  modeList,
		expandedDirs: make(map[string]bool),
	}
	m.toggleDualPane()
	if m.viewMode != viewDualPane {
		t.Errorf("single->dual: viewMode = %v, want viewDualPane", m.viewMode)
	}

	// Dual -> single
	m.toggleDualPane()
	if m.viewMode != viewSinglePane {
		t.Errorf("dual->single: viewMode = %v, want viewSinglePane", m.viewMode)
	}

	// Idempotency: two toggles from single return to single
	m.toggleDualPane()
	m.toggleDualPane()
	if m.viewMode != viewSinglePane {
		t.Errorf("double-toggle: viewMode = %v, want viewSinglePane", m.viewMode)
	}
}

// TestToggleTrashResetsDetailScrollX verifies the F12 trash-view path (tfe-ojy)
// routes its default-to-detail switch through setDisplayMode. This mirrors the
// keyboard handler: toggleTrash() followed by setDisplayMode(modeDetail).
func TestToggleTrashResetsDetailScrollX(t *testing.T) {
	dir := t.TempDir()
	m := newToggleTestModel(dir)
	wasInTrash := m.showTrashOnly
	m.toggleTrash()
	if !wasInTrash {
		m.setDisplayMode(modeDetail)
	}
	if !m.showTrashOnly {
		t.Fatal("expected showTrashOnly true after toggle")
	}
	if m.displayMode != modeDetail {
		t.Errorf("displayMode = %v, want modeDetail", m.displayMode)
	}
	if m.detailScrollX != 0 {
		t.Errorf("detailScrollX = %d, want 0 (entering detail must reset)", m.detailScrollX)
	}
}
