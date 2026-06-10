package main

// Tests for tfe-cw9: recurring lipgloss styles hoisted into initStyles() as
// package vars (scrollbar, line-number, toolbar buttons, tab bar, command line,
// alternate-row variants) instead of being rebuilt every frame/line. These
// tests pin the wiring: initStyles() must populate the new vars from the active
// theme, and alternateRowStyle() must map each file-type base style to its
// alternate-row variant.

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func initTestStyles() {
	currentTheme = defaultTheme()
	initStyles()
}

// TestInitStyles_PopulatesHoistedStyles verifies the foreground colors of the
// newly hoisted styles match the theme values they are built from. A zero-value
// (un-initialized) lipgloss.Style returns a nil foreground, so a non-nil match
// confirms initStyles() actually built each var.
func TestInitStyles_PopulatesHoistedStyles(t *testing.T) {
	initTestStyles()

	cases := []struct {
		name string
		got  lipgloss.TerminalColor
		want lipgloss.TerminalColor
	}{
		{"lineNumStyle", lineNumStyle.GetForeground(), uiSubtleText()},
		{"scrollbarTrackStyle", scrollbarTrackStyle.GetForeground(), uiMutedText()},
		{"scrollbarThumbStyle", scrollbarThumbStyle.GetForeground(), currentTheme.Title.adaptiveColor()},
		{"scrollIndicatorStyle", scrollIndicatorStyle.GetForeground(), uiSubtleText()},
		{"detailHeaderStyle", detailHeaderStyle.GetForeground(), currentTheme.Title.adaptiveColor()},
		{"toolbarButtonStyle", toolbarButtonStyle.GetForeground(), lipgloss.Color("39")},
		{"toolbarTermStyle", toolbarTermStyle.GetForeground(), lipgloss.Color("46")},
		{"activeTabStyle", activeTabStyle.GetForeground(), currentTheme.SelectionFg.adaptiveColor()},
		{"inactiveTabStyle", inactiveTabStyle.GetForeground(), uiBodyText()},
		{"cmdPromptStyle", cmdPromptStyle.GetForeground(), currentTheme.Title.adaptiveColor()},
		{"cmdBangStyle", cmdBangStyle.GetForeground(), currentTheme.DiffRemoved.adaptiveColor()},
		{"cmdGhostStyle", cmdGhostStyle.GetForeground(), uiMutedText()},
	}

	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s foreground = %v, want %v", c.name, c.got, c.want)
		}
	}
}

// TestInitStyles_ActiveToolbarStylesHaveBackground verifies the active toolbar
// button variants carry the gray (237) background that distinguishes them from
// the inactive variants.
func TestInitStyles_ActiveToolbarStylesHaveBackground(t *testing.T) {
	initTestStyles()

	if got := toolbarButtonActiveStyle.GetBackground(); got != lipgloss.Color("237") {
		t.Errorf("toolbarButtonActiveStyle background = %v, want 237", got)
	}
	if got := toolbarTermActiveStyle.GetBackground(); got != lipgloss.Color("237") {
		t.Errorf("toolbarTermActiveStyle background = %v, want 237", got)
	}
	// Inactive variants must NOT have the gray background.
	if got := toolbarButtonStyle.GetBackground(); got == lipgloss.Color("237") {
		t.Errorf("toolbarButtonStyle should not carry the active background")
	}
}

// TestAlternateRowStyle_MapsBaseToVariant verifies that each file-type base
// style maps to its pre-built alternate-row variant, that the variant preserves
// the base foreground, and that it adds the theme's alternate-row background.
func TestAlternateRowStyle_MapsBaseToVariant(t *testing.T) {
	initTestStyles()

	altBg := currentTheme.AlternateRow.adaptiveColor()

	cases := []struct {
		name string
		base lipgloss.Style
		want lipgloss.Style
	}{
		{"file", fileStyle, fileAltStyle},
		{"folder", folderStyle, folderAltStyle},
		{"claudeContext", claudeContextStyle, claudeContextAltStyle},
		{"agents", agentsStyle, agentsAltStyle},
		{"promptsFolder", promptsFolderStyle, promptsFolderAltStyle},
		{"obsidianVault", obsidianVaultStyle, obsidianVaultAltStyle},
	}

	for _, c := range cases {
		got := alternateRowStyle(c.base)
		if got.GetForeground() != c.want.GetForeground() {
			t.Errorf("%s: alternateRowStyle foreground = %v, want %v",
				c.name, got.GetForeground(), c.want.GetForeground())
		}
		if got.GetForeground() != c.base.GetForeground() {
			t.Errorf("%s: alternate variant dropped base foreground (%v != %v)",
				c.name, got.GetForeground(), c.base.GetForeground())
		}
		if got.GetBackground() != altBg {
			t.Errorf("%s: alternate variant background = %v, want %v",
				c.name, got.GetBackground(), altBg)
		}
	}
}

// TestAlternateRowStyle_EquivalentToLegacyCopy verifies the hoisted variant is
// byte-for-byte identical to the previous per-row construction
// (style.Copy().Background(altBg)) that it replaced, guarding behavior parity.
func TestAlternateRowStyle_EquivalentToLegacyCopy(t *testing.T) {
	initTestStyles()

	altBg := currentTheme.AlternateRow.adaptiveColor()
	sample := "example.go"

	for _, base := range []lipgloss.Style{
		fileStyle, folderStyle, claudeContextStyle, agentsStyle,
		promptsFolderStyle, obsidianVaultStyle,
	} {
		legacy := base.Background(altBg).Render(sample)
		hoisted := alternateRowStyle(base).Render(sample)
		if legacy != hoisted {
			t.Errorf("rendered output differs:\n legacy = %q\nhoisted = %q", legacy, hoisted)
		}
	}
}
