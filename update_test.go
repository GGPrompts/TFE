package main

import (
	"strings"
	"testing"
)

// TestStripANSI verifies that stripANSI removes ANSI escape sequences and
// terminal response sequences while preserving plain text. The regexes were
// hoisted to package level (ansiStripRegex, terminalResponseRegex) for
// performance, so these tests also guard against regressions in that move.
func TestStripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text unchanged",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "CSI color sequence",
			input: "\x1b[31mred\x1b[0m",
			want:  "red",
		},
		{
			name:  "CSI with multiple params",
			input: "\x1b[1;32;40mstyled\x1b[0m text",
			want:  "styled text",
		},
		{
			name:  "CSI cursor movement",
			input: "\x1b[2Aup two lines",
			want:  "up two lines",
		},
		{
			name:  "OSC sequence terminated by BEL",
			input: "\x1b]0;window title\x07after",
			want:  "after",
		},
		{
			name:  "OSC sequence terminated by ST",
			input: "\x1b]8;;http://example.com\x1b\\link",
			want:  "link",
		},
		{
			name:  "other escape sequences",
			input: "\x1b=keypad\x1b>mode",
			want:  "keypadmode",
		},
		{
			name:  "terminal rgb response without ESC prefix",
			input: ";rgb:ffff/ffff/ffff plain",
			want:  " plain",
		},
		{
			name:  "rgb response with numeric prefix",
			input: "11;rgb:1e1e/1e1e/2e2etext",
			want:  "text",
		},
		{
			name:  "numeric response codes",
			input: "1;2;3leftover",
			want:  "leftover",
		},
		{
			name:  "mixed ANSI and plain text",
			input: "\x1b[1mfile.txt\x1b[0m  \x1b[36m1.2K\x1b[0m",
			want:  "file.txt  1.2K",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripANSI(tt.input)
			if got != tt.want {
				t.Errorf("stripANSI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestStripANSIPackageLevelRegexes verifies the hoisted regexes are
// initialized and usable directly (concurrent-safe compiled patterns).
func TestStripANSIPackageLevelRegexes(t *testing.T) {
	if ansiStripRegex == nil {
		t.Fatal("ansiStripRegex is nil; expected package-level compiled regex")
	}
	if terminalResponseRegex == nil {
		t.Fatal("terminalResponseRegex is nil; expected package-level compiled regex")
	}
	if !ansiStripRegex.MatchString("\x1b[0m") {
		t.Error("ansiStripRegex should match a CSI reset sequence")
	}
	if !terminalResponseRegex.MatchString("rgb:ffff/ffff/ffff") {
		t.Error("terminalResponseRegex should match an rgb response")
	}
}

// BenchmarkStripANSI exercises the hot path (per line per frame in
// narrow-terminal detail view) to make the package-level compilation
// win measurable.
func BenchmarkStripANSI(b *testing.B) {
	line := strings.Repeat("\x1b[36m📁 folder\x1b[0m some text ", 4)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stripANSI(line)
	}
}

// TestFooterTickGeneration verifies that footer scroll ticks carry a
// generation stamp and that the Update handler drops ticks belonging to a
// stale chain (gen mismatch) while re-arming only the current generation.
// This guards against overlapping in-flight tick chains accumulating when
// the user rapidly toggles footer scrolling off and back on (tfe-pjx).
func TestFooterTickGeneration(t *testing.T) {
	// Current generation, scrolling active: should advance offset and re-arm.
	m := model{footerScrolling: true, footerTickGen: 3, footerOffset: 0}
	updated, cmd := m.Update(footerTickMsg{gen: 3})
	um := updated.(model)
	if cmd == nil {
		t.Fatalf("current-gen tick while scrolling must re-arm (non-nil cmd)")
	}
	if um.footerOffset != 1 {
		t.Errorf("current-gen tick should advance footerOffset: got %d, want 1", um.footerOffset)
	}

	// Stale generation: must be dropped with no re-arm, even while scrolling.
	m2 := model{footerScrolling: true, footerTickGen: 5, footerOffset: 7}
	updated2, cmd2 := m2.Update(footerTickMsg{gen: 4})
	um2 := updated2.(model)
	if cmd2 != nil {
		t.Errorf("stale-gen tick must not re-arm (expected nil cmd)")
	}
	if um2.footerOffset != 7 {
		t.Errorf("stale-gen tick must not advance footerOffset: got %d, want 7", um2.footerOffset)
	}

	// Current generation but scrolling stopped: no re-arm.
	m3 := model{footerScrolling: false, footerTickGen: 2}
	_, cmd3 := m3.Update(footerTickMsg{gen: 2})
	if cmd3 != nil {
		t.Errorf("current-gen tick with scrolling off must not re-arm (expected nil cmd)")
	}
}
