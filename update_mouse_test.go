package main

// Tests for mouse hit-testing in update_mouse.go (tfe-108): vertical-split
// pane boundary detection must use verticalSplitHeights (the single source
// of truth used by rendering) so it respects panelsLocked/lockedTopRatio
// instead of always assuming the focus-based 2/3 accordion.

import (
	"testing"
)

// newVerticalSplitTestModel returns a dual-pane model in Detail mode (which
// always uses the vertical split) with a known terminal size.
func newVerticalSplitTestModel() model {
	return model{
		width:       120,
		height:      40, // headerLines(4) + maxVisible(32) + footerLines(4)
		viewMode:    viewDualPane,
		displayMode: modeDetail,
		focusedPane: leftPane,
	}
}

func TestIsClickInFileListArea_AccordionUnlocked(t *testing.T) {
	m := newVerticalSplitTestModel()
	// maxVisible = 32; top pane focused -> topHeight = (32*2)/3 = 21.
	// Pane area starts at mouseY = 4 (headerLines).

	// Last row of the top pane (paneY = 20) is in the file list.
	if !m.isClickInFileListArea(10, 4+20) {
		t.Errorf("click at paneY=20 should be in file list (topHeight=21)")
	}
	// First row of the bottom pane (paneY = 21) is in the preview.
	if m.isClickInFileListArea(10, 4+21) {
		t.Errorf("click at paneY=21 should be in preview (topHeight=21)")
	}

	// Preview focused -> topHeight = 32 - 21 = 11.
	m.focusedPane = rightPane
	if !m.isClickInFileListArea(10, 4+10) {
		t.Errorf("preview focused: click at paneY=10 should be in file list (topHeight=11)")
	}
	if m.isClickInFileListArea(10, 4+11) {
		t.Errorf("preview focused: click at paneY=11 should be in preview (topHeight=11)")
	}
}

func TestIsClickInFileListArea_RespectsPanelLock(t *testing.T) {
	m := newVerticalSplitTestModel()
	m.panelsLocked = true
	m.lockedTopRatio = 1.0 / 3.0
	// maxVisible = 32; locked topHeight = int(32 * 1/3) = 10, regardless of focus.

	for _, pane := range []paneType{leftPane, rightPane} {
		m.focusedPane = pane
		if !m.isClickInFileListArea(10, 4+9) {
			t.Errorf("locked 1/3 (focus=%v): click at paneY=9 should be in file list (topHeight=10)", pane)
		}
		if m.isClickInFileListArea(10, 4+10) {
			t.Errorf("locked 1/3 (focus=%v): click at paneY=10 should be in preview (topHeight=10)", pane)
		}
		// paneY=15 would be inside the hardcoded focused-2/3 top pane (21 rows)
		// but is in the preview under the locked 1/3 ratio.
		if m.isClickInFileListArea(10, 4+15) {
			t.Errorf("locked 1/3 (focus=%v): click at paneY=15 should be in preview, not file list", pane)
		}
	}

	// Hit-testing must agree with the rendered split for any locked ratio.
	m.lockedTopRatio = 2.0 / 3.0
	m.focusedPane = rightPane // would shrink top pane to 1/3 if lock were ignored
	topHeight, _ := m.verticalSplitHeights(m.height - 8)
	if !m.isClickInFileListArea(10, 4+topHeight-1) {
		t.Errorf("locked 2/3: click on last top-pane row (paneY=%d) should be in file list", topHeight-1)
	}
	if m.isClickInFileListArea(10, 4+topHeight) {
		t.Errorf("locked 2/3: click on first bottom-pane row (paneY=%d) should be in preview", topHeight)
	}
}
