package main

// Tests for render_file_list.go helpers.
// Focus: rune-aware tail truncation used by detail-view columns
// (location/branch/commit/description/type) — see tfe-mda.

import (
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
