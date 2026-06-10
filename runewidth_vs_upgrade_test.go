package main

// Tests pinning the go-runewidth >= v0.0.21 behavior for variation selectors.
//
// go-runewidth bug #76 (variation selectors U+FE00-U+FE0F counted as width 1
// instead of 0) was fixed upstream by PR #90, first released in v0.0.21.
// TFE previously pinned v0.0.19 and carried compensation code because of it.
// These tests fail on v0.0.19 and pass on v0.0.21+ (currently v0.0.24),
// guarding against an accidental downgrade.
//
// NOTE: variation selectors appear here ONLY as ️ escapes, never as raw
// bytes, per the CLAUDE.md "no variation selectors in source" rule.

import (
	"testing"

	"github.com/mattn/go-runewidth"
)

// TestRuneWidthVariationSelectorIsZero proves the upstream fix: RuneWidth of
// a variation selector is 0 (it was 1 on v0.0.19).
func TestRuneWidthVariationSelectorIsZero(t *testing.T) {
	if w := runewidth.RuneWidth('️'); w != 0 {
		t.Errorf("RuneWidth(U+FE0F) = %d, want 0 (go-runewidth bug #76 regression; is the dep downgraded below v0.0.21?)", w)
	}
	if w := runewidth.RuneWidth('︎'); w != 0 {
		t.Errorf("RuneWidth(U+FE0E) = %d, want 0 (go-runewidth bug #76 regression)", w)
	}
}

// TestStringWidthEmojiWithVariationSelector verifies that an emoji+VS16
// sequence is one grapheme cluster whose width equals the base emoji alone,
// and that a standalone VS contributes zero width (was 1 on v0.0.19).
func TestStringWidthEmojiWithVariationSelector(t *testing.T) {
	cases := []struct {
		name string
		s    string
		want int
	}{
		{"standalone VS16", "️", 0},
		{"gear U+2699 alone", "⚙", 1},
		{"gear + VS16", "⚙️", 1},
		{"wastebasket U+1F5D1 alone", "\U0001F5D1", 1},
		{"wastebasket + VS16", "\U0001F5D1️", 1},
	}
	for _, c := range cases {
		if got := runewidth.StringWidth(c.s); got != c.want {
			t.Errorf("StringWidth(%s) = %d, want %d", c.name, got, c.want)
		}
	}
}

// TestVisualWidthTruncateConsistencyWithVS verifies that TFE's per-rune
// truncateToWidth now agrees with the grapheme-based visualWidth for
// VS-bearing content. On v0.0.19, truncateToWidth over-counted each VS by 1
// (RuneWidth-based loop, no VS interception) and truncated such strings one
// cell short of what visualWidth computed.
func TestVisualWidthTruncateConsistencyWithVS(t *testing.T) {
	s := "⚙️abc" // gear + VS16 + "abc": visual width 4

	if got := visualWidth(s); got != 4 {
		t.Fatalf("visualWidth(%q) = %d, want 4", s, got)
	}

	// The whole string fits in exactly its visual width.
	if got := truncateToWidth(s, 4); got != s {
		t.Errorf("truncateToWidth(%q, 4) = %q, want the full string (VS over-counted as width 1?)", s, got)
	}

	// Truncating a longer VS-bearing string yields a prefix whose visual
	// width matches the budget exactly (budget < 3 remaining, so no "...").
	long := "⚙️abcdef"
	got := truncateToWidth(long, 4)
	if want := "⚙️abc"; got != want {
		t.Errorf("truncateToWidth(%q, 4) = %q, want %q", long, got, want)
	}
	if w := visualWidth(got); w != 4 {
		t.Errorf("visualWidth(truncated) = %d, want 4 (truncateToWidth disagrees with visualWidth)", w)
	}
}
