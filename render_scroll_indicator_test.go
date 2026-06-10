package main

import (
	"strings"
	"testing"
)

// TestComputeScrollPercent exercises the scroll-percentage math shared by all
// preview scroll indicators, including the maxScrollPos<=0 -> 100% special case
// and the >100 clamp.
func TestComputeScrollPercent(t *testing.T) {
	tests := []struct {
		name            string
		scrollPos       int
		totalLines      int
		scrollableLines int
		want            int
	}{
		// maxScrollPos = totalLines - scrollableLines <= 0 -> content fits -> 100%
		{"fits exactly", 0, 10, 10, 100},
		{"fits with extra room", 0, 5, 10, 100},
		{"negative max scroll", 3, 5, 10, 100},
		// normal fractional progress
		{"top of scrollable", 0, 110, 10, 0},
		{"halfway", 50, 110, 10, 50},
		{"bottom", 100, 110, 10, 100},
		// scrollPos beyond bottom must clamp to 100, not overshoot
		{"overshoot clamps to 100", 200, 110, 10, 100},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := model{}
			m.preview.scrollPos = tc.scrollPos
			if got := m.computeScrollPercent(tc.totalLines, tc.scrollableLines); got != tc.want {
				t.Errorf("computeScrollPercent(scrollPos=%d, total=%d, scrollable=%d) = %d, want %d",
					tc.scrollPos, tc.totalLines, tc.scrollableLines, got, tc.want)
			}
		})
	}
}

// TestRenderScrollIndicator verifies the " %d/%d (%d%%)"+suffix body for each
// suffix variant used by the call sites, matching the byte-identical output of
// the previous per-site blocks. (Styling is verified separately in
// styles_test.go; lipgloss strips ANSI under the test color profile.)
func TestRenderScrollIndicator(t *testing.T) {
	initStyles()

	m := model{}
	m.preview.scrollPos = 50

	// suffix variants used by the call sites
	cases := []struct {
		name       string
		suffix     string
		wantSubstr string
	}{
		{"plain", " ", " 60/110 (50%) "},
		{"diff", " [diff]", " 60/110 (50%) [diff]"},
		{"jsonl", " [jsonl]", " 60/110 (50%) [jsonl]"},
		{"jsonl tailed", " [jsonl] | F: load full", " 60/110 (50%) [jsonl] | F: load full"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := m.renderScrollIndicator(60, 110, 10, tc.suffix)
			if !strings.Contains(got, tc.wantSubstr) {
				t.Errorf("renderScrollIndicator body = %q, want it to contain %q", got, tc.wantSubstr)
			}
		})
	}
}

// TestRenderScrollIndicatorFitsCase confirms the 100% special case flows
// through the indicator (scrollableLines >= totalLines).
func TestRenderScrollIndicatorFitsCase(t *testing.T) {
	initStyles()
	m := model{}
	m.preview.scrollPos = 0
	got := m.renderScrollIndicator(5, 5, 10, " ")
	if !strings.Contains(got, " 5/5 (100%) ") {
		t.Errorf("renderScrollIndicator fits-case = %q, want it to contain %q", got, " 5/5 (100%) ")
	}
}
