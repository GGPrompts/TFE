package main

// Tests for dropdown menu geometry (tfe-9aa): the mouse hit-test in
// isInDropdown previously estimated dropdown width with byte len(), while
// renderActiveDropdown used m.visualWidthCompensated(). Emoji-prefixed labels
// (4 bytes vs 2 visual cells) made the hit-test box wider than the drawn
// dropdown, so clicks beside the visible dropdown triggered menu items.
// Both paths now share m.dropdownWidth(menu).

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func newMenuTestModel() model {
	currentTheme = defaultTheme()
	return model{
		width:  200,
		height: 50,
	}
}

// TestDropdownWidth_UsesVisualWidthNotByteLen verifies that dropdownWidth is
// based on visual width, not byte length, for emoji-prefixed labels.
func TestDropdownWidth_UsesVisualWidthNotByteLen(t *testing.T) {
	m := newMenuTestModel()

	// Emoji is 4 bytes but 2 visual cells; label long enough to exceed the
	// 20-column minimum so the item actually determines the width.
	menu := Menu{
		Label: "Test",
		Items: []MenuItem{
			{Label: "📁 New Folder With A Long Name...", Action: "x", Shortcut: "F7"},
			{IsSeparator: true},
			{Label: "🗑  Trash", Action: "y"},
		},
	}

	got := m.dropdownWidth(menu)

	// Expected: same formula as renderActiveDropdown's first pass.
	want := m.visualWidthCompensated("📁 New Folder With A Long Name...") +
		m.visualWidthCompensated("F7") + 3 + 4
	if got != want {
		t.Errorf("dropdownWidth = %d, want %d (visual-width based)", got, want)
	}

	// The old byte-len estimate must be strictly larger for emoji labels —
	// this is the divergence that caused the bug.
	byteEstimate := len("📁 New Folder With A Long Name...") + len("F7") + 3 + 4
	if got >= byteEstimate {
		t.Errorf("dropdownWidth = %d should be less than byte-len estimate %d for emoji labels", got, byteEstimate)
	}
}

// TestDropdownWidth_Minimum verifies the 20-column floor.
func TestDropdownWidth_Minimum(t *testing.T) {
	m := newMenuTestModel()
	menu := Menu{Label: "T", Items: []MenuItem{{Label: "Hi", Action: "x"}}}
	if got := m.dropdownWidth(menu); got != 20 {
		t.Errorf("dropdownWidth = %d, want minimum 20", got)
	}
}

// TestDropdownWidth_CheckableAddsCheckmarkWidth verifies checkable items
// reserve room for the "✓ " prefix, matching renderActiveDropdown.
func TestDropdownWidth_CheckableAddsCheckmarkWidth(t *testing.T) {
	m := newMenuTestModel()
	label := "Show Hidden Files And More Stuff"
	plain := Menu{Items: []MenuItem{{Label: label, Action: "x"}}}
	checkable := Menu{Items: []MenuItem{{Label: label, Action: "x", IsCheckable: true}}}

	diff := m.dropdownWidth(checkable) - m.dropdownWidth(plain)
	if want := m.visualWidthCompensated("✓ "); diff != want {
		t.Errorf("checkable width delta = %d, want %d", diff, want)
	}
}

// TestIsInDropdown_MatchesRenderedGeometry verifies, for every real menu, that
// the hit-test box exactly matches the rendered dropdown's width and position.
func TestIsInDropdown_MatchesRenderedGeometry(t *testing.T) {
	m := newMenuTestModel()
	m.menuOpen = true

	for _, key := range getMenuOrder() {
		m.activeMenu = key
		menus := m.getMenus()
		menu, ok := menus[key]
		if !ok {
			t.Fatalf("menu %q not found", key)
		}

		rendered := m.renderActiveDropdown()
		renderedWidth := lipgloss.Width(rendered)
		hitWidth := m.dropdownWidth(menu) + 2 // +2 border, as used in isInDropdown

		if hitWidth != renderedWidth {
			t.Errorf("menu %q: hit-test width %d != rendered width %d", key, hitWidth, renderedWidth)
		}

		menuX := m.getMenuXPosition(key)
		y := 2 // first item row inside the dropdown

		// Inside the rendered box on both edges.
		if !m.isInDropdown(menuX, y) {
			t.Errorf("menu %q: left edge x=%d should be inside dropdown", key, menuX)
		}
		if !m.isInDropdown(menuX+renderedWidth-1, y) {
			t.Errorf("menu %q: right edge x=%d should be inside dropdown", key, menuX+renderedWidth-1)
		}

		// One column past the rendered right edge must be outside — this was
		// the bug: byte-len estimates made these clicks register as menu hits.
		if m.isInDropdown(menuX+renderedWidth, y) {
			t.Errorf("menu %q: x=%d is beside the rendered dropdown but hit-test claims inside", key, menuX+renderedWidth)
		}
		// One column before the left edge must be outside too.
		if menuX > 0 && m.isInDropdown(menuX-1, y) {
			t.Errorf("menu %q: x=%d is left of the dropdown but hit-test claims inside", key, menuX-1)
		}
	}
}

// TestIsInDropdown_ClampMatchesOverlay verifies that near the right terminal
// edge the hit-test clamps X with the same width the overlay (view.go) uses,
// so the hit box still matches where the dropdown is actually drawn.
func TestIsInDropdown_ClampMatchesOverlay(t *testing.T) {
	m := newMenuTestModel()
	m.menuOpen = true
	m.activeMenu = "help" // rightmost menu, most likely to clamp

	menu := m.getMenus()["help"]
	renderedWidth := lipgloss.Width(m.renderActiveDropdown())

	// Shrink terminal so the dropdown must be clamped against the right edge.
	m.width = m.getMenuXPosition("help") + renderedWidth - 5
	clampedX := m.width - renderedWidth
	if clampedX < 0 {
		clampedX = 0
	}

	if hitWidth := m.dropdownWidth(menu) + 2; hitWidth != renderedWidth {
		t.Fatalf("hit-test width %d != rendered width %d", hitWidth, renderedWidth)
	}
	if !m.isInDropdown(clampedX, 2) {
		t.Errorf("clamped left edge x=%d should be inside dropdown", clampedX)
	}
	if clampedX > 0 && m.isInDropdown(clampedX-1, 2) {
		t.Errorf("x=%d left of clamped dropdown should be outside", clampedX-1)
	}
	if m.isInDropdown(clampedX+renderedWidth, 2) {
		t.Errorf("x=%d right of clamped dropdown should be outside", clampedX+renderedWidth)
	}
}
