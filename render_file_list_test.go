package main

// Tests for render_file_list.go helpers.
// Focus: rune-aware tail truncation used by detail-view columns
// (location/branch/commit/description/type) — see tfe-mda.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateToWidthFromEnd(t *testing.T) {
	m := model{terminalType: terminalUnknown}

	tests := []struct {
		name        string
		input       string
		targetWidth int
		want        string
	}{
		{
			name:        "fits unchanged",
			input:       "~/projects/tfe",
			targetWidth: 20,
			want:        "~/projects/tfe",
		},
		{
			name:        "exact width unchanged",
			input:       "abcdef",
			targetWidth: 6,
			want:        "abcdef",
		},
		{
			name:        "ascii tail kept",
			input:       "/very/long/path/to/dir",
			targetWidth: 10,
			want:        ".../to/dir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := m.truncateToWidthFromEnd(tt.input, tt.targetWidth)
			if got != tt.want {
				t.Errorf("truncateToWidthFromEnd(%q, %d) = %q, want %q", tt.input, tt.targetWidth, got, tt.want)
			}
		})
	}
}

func TestTruncateToWidthFromEndMultibyte(t *testing.T) {
	m := model{terminalType: terminalUnknown}

	// Non-ASCII path: byte slicing would split runes and produce invalid UTF-8.
	input := "/home/użytkownik/projekty/ważne/dokumenty"
	for width := 4; width < 30; width++ {
		got := m.truncateToWidthFromEnd(input, width)
		if !utf8.ValidString(got) {
			t.Errorf("width %d: result is invalid UTF-8: %q", width, got)
		}
		if !strings.HasPrefix(got, "...") {
			t.Errorf("width %d: expected ... prefix, got %q", width, got)
		}
		if vw := visualWidth(got); vw > width {
			t.Errorf("width %d: visual width %d exceeds target: %q", width, vw, got)
		}
		// The kept tail must be a suffix of the original string.
		if !strings.HasSuffix(input, strings.TrimPrefix(got, "...")) {
			t.Errorf("width %d: %q is not an ellipsis plus a suffix of input", width, got)
		}
	}
}

func TestTruncateToWidthFromEndWideRunes(t *testing.T) {
	m := model{terminalType: terminalUnknown}

	// CJK runes are 2 cells wide; ensure accounting is by visual width, not runes.
	input := "/路径/非常/长的/目录"
	got := m.truncateToWidthFromEnd(input, 8)
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

func TestTruncateToWidthFromEndTinyWidth(t *testing.T) {
	m := model{terminalType: terminalUnknown}

	// Widths too small for "..." plus content must not panic or overflow.
	for width := 0; width <= 3; width++ {
		got := m.truncateToWidthFromEnd("/some/long/path", width)
		if vw := visualWidth(got); vw > width {
			t.Errorf("width %d: visual width %d exceeds target: %q", width, vw, got)
		}
	}
}

// Tests for truncateNameWithEllipsis, the head-truncation helper used by the
// detail view and tree view to shorten file names — see tfe-4oy.

func TestTruncateNameWithEllipsis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{
			name:     "fits unchanged",
			input:    "main.go",
			maxWidth: 20,
			want:     "main.go",
		},
		{
			name:     "exact width unchanged",
			input:    "abcdef",
			maxWidth: 6,
			want:     "abcdef",
		},
		{
			name:     "ascii head kept",
			input:    "a_very_long_file_name.txt",
			maxWidth: 10,
			want:     "a_very_l..",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateNameWithEllipsis(tt.input, tt.maxWidth)
			if got != tt.want {
				t.Errorf("truncateNameWithEllipsis(%q, %d) = %q, want %q", tt.input, tt.maxWidth, got, tt.want)
			}
		})
	}
}

func TestTruncateNameWithEllipsisMultibyte(t *testing.T) {
	// Non-ASCII name: the old byte slicing (name[:maxLen-2]) would split runes
	// and produce invalid UTF-8, and len() over-counted multibyte names.
	input := "ważne_dokumenty_użytkownika.md"
	for width := 5; width < 30; width++ {
		got := truncateNameWithEllipsis(input, width)
		if !utf8.ValidString(got) {
			t.Errorf("width %d: result is invalid UTF-8: %q", width, got)
		}
		if !strings.HasSuffix(got, "..") {
			t.Errorf("width %d: expected .. suffix, got %q", width, got)
		}
		if vw := visualWidth(got); vw > width {
			t.Errorf("width %d: visual width %d exceeds target: %q", width, vw, got)
		}
		// The kept head must be a prefix of the original string.
		if !strings.HasPrefix(input, strings.TrimSuffix(got, "..")) {
			t.Errorf("width %d: %q is not a prefix of input plus ellipsis", width, got)
		}
	}
}

func TestTruncateNameWithEllipsisWideRunes(t *testing.T) {
	// CJK runes are 2 cells wide; ensure accounting is by visual width, not bytes.
	input := "非常长的文件名称记录.txt"
	got := truncateNameWithEllipsis(input, 8)
	if !utf8.ValidString(got) {
		t.Fatalf("result is invalid UTF-8: %q", got)
	}
	if vw := visualWidth(got); vw > 8 {
		t.Errorf("visual width %d exceeds target 8: %q", vw, got)
	}
	if !strings.HasSuffix(got, "..") {
		t.Errorf("expected .. suffix, got %q", got)
	}
}

func TestTruncateNameWithEllipsisVirtualFolderPrefix(t *testing.T) {
	// Tree view extracts the "🌐 "/"🤖 " prefix of virtual folders AFTER
	// truncation. The minimum tree-view width is 5 (clamped in renderTreeView),
	// which leaves 3 cells of content — exactly the emoji + space — so the
	// prefix match must survive truncation at every reachable width.
	for _, prefix := range []string{"🌐 ", "🤖 "} {
		input := prefix + "Global Prompts"
		for width := 5; width < 20; width++ {
			got := truncateNameWithEllipsis(input, width)
			if !utf8.ValidString(got) {
				t.Errorf("prefix %q width %d: result is invalid UTF-8: %q", prefix, width, got)
			}
			if !strings.HasPrefix(got, prefix) {
				t.Errorf("prefix %q width %d: emoji prefix lost: %q", prefix, width, got)
			}
			if vw := visualWidth(got); vw > width {
				t.Errorf("prefix %q width %d: visual width %d exceeds target: %q", prefix, width, vw, got)
			}
		}
	}
}

// --- Event-driven tree cache tests (tfe-xsx) ---
//
// updateTreeItems must be a lazy, event-driven rebuild: a no-op while the
// cache is clean (so tree mode pays no disk I/O per tea.Msg) and a full
// rebuild once markTreeItemsDirty has been called.

// newTreeTestModel builds a minimal tree-mode model with two synthetic files
// and a freshly built (clean) tree cache.
func newTreeTestModel(t *testing.T) *model {
	t.Helper()
	m := &model{
		displayMode:  modeTree,
		expandedDirs: make(map[string]bool),
		files: []fileItem{
			{name: "alpha.txt", path: "/synthetic/alpha.txt"},
			{name: "beta.txt", path: "/synthetic/beta.txt"},
		},
	}
	m.markTreeItemsDirty()
	m.updateTreeItems()
	if len(m.treeItems) != 2 {
		t.Fatalf("Setup: expected 2 tree items after initial rebuild, got %d", len(m.treeItems))
	}
	return m
}

// TestUpdateTreeItemsNoOpWhenClean verifies that updateTreeItems does not
// rebuild the cache when nothing marked it dirty, even if the underlying
// state changed behind its back.
func TestUpdateTreeItemsNoOpWhenClean(t *testing.T) {
	m := newTreeTestModel(t)

	if m.treeItemsDirty {
		t.Fatal("treeItemsDirty should be cleared after a rebuild")
	}

	// Mutate the file list WITHOUT marking the cache dirty
	m.files = append(m.files, fileItem{name: "gamma.txt", path: "/synthetic/gamma.txt"})

	m.updateTreeItems()
	if len(m.treeItems) != 2 {
		t.Errorf("updateTreeItems rebuilt a clean cache: got %d items, want 2 (no-op)", len(m.treeItems))
	}
}

// TestUpdateTreeItemsRebuildsWhenDirty verifies that marking the cache dirty
// causes the next updateTreeItems call to rebuild it from current state and
// clear the flag.
func TestUpdateTreeItemsRebuildsWhenDirty(t *testing.T) {
	m := newTreeTestModel(t)

	m.files = append(m.files, fileItem{name: "gamma.txt", path: "/synthetic/gamma.txt"})
	m.markTreeItemsDirty()

	m.updateTreeItems()
	if len(m.treeItems) != 3 {
		t.Errorf("Expected rebuild to pick up 3 files, got %d items", len(m.treeItems))
	}
	if m.treeItemsDirty {
		t.Error("treeItemsDirty should be cleared after rebuild")
	}
}

// TestCurrentTreeItemsDirtyFallback verifies that getCurrentFile/getMaxCursor
// read the cached treeItems when clean, but fall back to a fresh local build
// when the cache is dirty (mid-event mutation), without touching the cache.
func TestCurrentTreeItemsDirtyFallback(t *testing.T) {
	m := newTreeTestModel(t)

	// Clean cache: reads come straight from m.treeItems
	if got := m.getMaxCursor(); got != 1 {
		t.Fatalf("getMaxCursor (clean) = %d, want 1", got)
	}
	m.cursor = 1
	if f := m.getCurrentFile(); f == nil || f.name != "beta.txt" {
		t.Fatalf("getCurrentFile (clean) = %v, want beta.txt", f)
	}

	// Simulate a mid-event mutation: state changed and dirty flag set, but
	// updateTreeItems has not run yet
	m.files = append(m.files, fileItem{name: "gamma.txt", path: "/synthetic/gamma.txt"})
	m.markTreeItemsDirty()

	if got := m.getMaxCursor(); got != 2 {
		t.Errorf("getMaxCursor (dirty) = %d, want 2 (fresh rebuild fallback)", got)
	}
	m.cursor = 2
	if f := m.getCurrentFile(); f == nil || f.name != "gamma.txt" {
		t.Errorf("getCurrentFile (dirty) = %v, want gamma.txt", f)
	}

	// The fallback must not have mutated the cache (value receivers)
	if len(m.treeItems) != 2 {
		t.Errorf("Dirty-read fallback mutated the cache: %d items, want 2", len(m.treeItems))
	}
	if !m.treeItemsDirty {
		t.Error("Dirty flag should remain set until updateTreeItems runs")
	}
}

// TestLoadFilesMarksTreeDirty verifies the loadFiles integration point: any
// directory (re)load flags the tree cache so the next Update rebuilds it.
func TestLoadFilesMarksTreeDirty(t *testing.T) {
	tmpDir := t.TempDir()
	createTestFileWithContent(t, filepath.Join(tmpDir, "one.txt"), []byte("x"))

	m := &model{
		displayMode:  modeTree,
		expandedDirs: make(map[string]bool),
		currentPath:  tmpDir,
	}
	m.loadFiles()
	if !m.treeItemsDirty {
		t.Fatal("loadFiles should mark the tree cache dirty")
	}

	m.updateTreeItems()
	if m.treeItemsDirty {
		t.Fatal("updateTreeItems should clear the dirty flag")
	}

	// New file appears on disk; reload (as the fsnotify path does) must
	// re-flag the cache
	createTestFileWithContent(t, filepath.Join(tmpDir, "two.txt"), []byte("y"))
	m.loadFiles()
	if !m.treeItemsDirty {
		t.Error("loadFiles after a disk change should mark the tree cache dirty again")
	}
	m.updateTreeItems()

	// Tree should now include ".." + both files
	found := false
	for _, item := range m.treeItems {
		if item.file.name == "two.txt" {
			found = true
		}
	}
	if !found {
		t.Errorf("Rebuilt tree missing new file two.txt (items: %d)", len(m.treeItems))
	}
}

// TestExpandCollapseMarksTreeDirty verifies that toggling a directory's
// expansion plus a dirty-flag rebuild surfaces (and removes) its children.
func TestExpandCollapseMarksTreeDirty(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	createTestFileWithContent(t, filepath.Join(subDir, "child.txt"), []byte("x"))

	m := &model{
		displayMode:  modeTree,
		expandedDirs: make(map[string]bool),
		currentPath:  tmpDir,
	}
	m.loadFiles()
	m.updateTreeItems()
	baseCount := len(m.treeItems)

	// Expand (as the keyboard handler does: mutate + mark dirty)
	m.expandedDirs[subDir] = true
	m.markTreeItemsDirty()
	m.updateTreeItems()
	if len(m.treeItems) != baseCount+1 {
		t.Errorf("Expected %d items after expanding sub/, got %d", baseCount+1, len(m.treeItems))
	}

	// Collapse
	m.expandedDirs[subDir] = false
	m.markTreeItemsDirty()
	m.updateTreeItems()
	if len(m.treeItems) != baseCount {
		t.Errorf("Expected %d items after collapsing sub/, got %d", baseCount, len(m.treeItems))
	}
}

// TestRenderDetailViewChangesModeGitRootRelative verifies that changes-mode
// detail rendering shortens each row's location to a path relative to the git
// root. The git root and home dir are computed once above the row loop (see
// tfe-1fd, hoisting loop-invariant findGitRoot/os.UserHomeDir out of the
// per-row loop); this exercises that path so a regression in the hoist would
// show up as absolute paths in the output.
func TestRenderDetailViewChangesModeGitRootRelative(t *testing.T) {
	repo := t.TempDir()
	if err := os.Mkdir(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}
	subDir := filepath.Join(repo, "pkg", "deep")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	changed := filepath.Join(subDir, "changed.go")
	if err := os.WriteFile(changed, []byte("package deep\n"), 0o644); err != nil {
		t.Fatalf("failed to write changed file: %v", err)
	}

	m := &model{
		width:           120,
		showChangesOnly: true,
		currentPath:     repo,
		changedFiles: []fileItem{
			{name: "changed.go", path: changed, size: 12},
		},
	}

	out := m.renderDetailView(20)

	// Location column should be the git-root-relative dir (pkg/deep), not the
	// absolute path of its parent.
	wantRel := filepath.Join("pkg", "deep")
	if !strings.Contains(out, wantRel) {
		t.Errorf("expected git-root-relative location %q in output, got:\n%s", wantRel, out)
	}
	if strings.Contains(out, subDir) {
		t.Errorf("output should not contain the absolute location %q; expected it relativized to git root:\n%s", subDir, out)
	}
}
