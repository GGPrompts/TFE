package main

// Tests for the shared pane chrome helpers extracted from renderSinglePane /
// renderDualPane (tfe-wkz): renderHeader, renderCommandLine,
// statusMessageVisible, and renderStatusMessageBar.

import (
	"strings"
	"testing"
	"time"
)

// newChromeTestModel returns a model with just enough state to render the
// header and command-line chrome.
func newChromeTestModel() model {
	return model{
		width:       120,
		height:      40,
		currentPath: "/tmp",
		startupTime: time.Now(), // within 5s window -> GitHub title shown
	}
}

func TestRenderHeader_TitleSuffix(t *testing.T) {
	m := newChromeTestModel()

	single := m.renderHeader("")
	if strings.Contains(single, "[Dual-Pane]") {
		t.Errorf("renderHeader(\"\") should not contain dual-pane suffix")
	}
	if !strings.Contains(single, "(T)erminal (F)ile (E)xplorer") {
		t.Errorf("renderHeader(\"\") missing base title")
	}

	dual := m.renderHeader(" [Dual-Pane] (xterm)")
	if !strings.Contains(dual, "[Dual-Pane] (xterm)") {
		t.Errorf("renderHeader with suffix missing suffix in output")
	}
}

func TestRenderHeader_ModeIndicators(t *testing.T) {
	m := newChromeTestModel()
	m.commandFocused = true
	out := m.renderHeader("")
	if !strings.Contains(out, "[Command Mode]") {
		t.Errorf("renderHeader missing [Command Mode] indicator when commandFocused")
	}

	m = newChromeTestModel()
	m.filePickerMode = true
	out = m.renderHeader("")
	if !strings.Contains(out, "File Picker") {
		t.Errorf("renderHeader missing file picker indicator")
	}
}

// TestRenderHeader_EmojiRightAlignment is a regression test for tfe-ifj: the
// right-align spacer was computed with byte len() on titleText/displayText.
// Emoji like 📋/📁/🎉 are 4 bytes but 2 visual columns, so len() over-counted
// by 2 per emoji and the right-side link landed short of the right edge. The
// spacer must be sized from visual width so the title row fills m.width.
func TestRenderHeader_EmojiRightAlignment(t *testing.T) {
	m := newChromeTestModel()
	m.filePickerMode = true // titleText gains " [📁 File Picker]"
	m.updateAvailable = true
	m.updateVersion = "v9.9.9" // displayText gains "🎉 Update Available: ..."

	titleText := "(T)erminal (F)ile (E)xplorer [📁 File Picker]"
	displayText := "🎉 Update Available: v9.9.9 (click for details)"
	wantSpacing := m.width - m.visualWidthCompensated(titleText) - m.visualWidthCompensated(displayText) - 2

	out := m.renderHeader("")
	firstLine := strings.SplitN(out, "\n", 2)[0]

	// The spacer is the longest run of spaces on the title row.
	gotSpacing := 0
	run := 0
	for _, ch := range firstLine {
		if ch == ' ' {
			run++
			if run > gotSpacing {
				gotSpacing = run
			}
		} else {
			run = 0
		}
	}

	if gotSpacing != wantSpacing {
		t.Errorf("right-align spacer = %d spaces, want %d (byte len() vs visual width regression)", gotSpacing, wantSpacing)
	}
}

func TestRenderHeader_MenuBarAfterStartupWindow(t *testing.T) {
	m := newChromeTestModel()
	m.startupTime = time.Now().Add(-10 * time.Second) // past the 5s window
	out := m.renderHeader("")
	if strings.Contains(out, "github.com/GGPrompts/TFE") {
		t.Errorf("renderHeader should show menu bar (not GitHub link) after 5 seconds")
	}
}

// TestRenderCommandLine_GhostText is a regression test for the drift fixed
// in tfe-wkz: ghost text was rendered only by the single-pane copy of the
// command prompt, so suggestions were invisible (yet tab-acceptable) in
// dual-pane command mode. The shared helper must render the ghost suffix.
func TestRenderCommandLine_GhostText(t *testing.T) {
	m := newChromeTestModel()
	m.commandFocused = true
	m.commandInput = "git s"
	m.commandCursorPos = len(m.commandInput) // cursor at end -> afterCursor == ""
	m.ghostText = "git status"

	out := m.renderCommandLine()
	if !strings.Contains(out, "tatus") {
		t.Errorf("renderCommandLine missing ghost text suffix %q in output: %q", "tatus", out)
	}
}

// TestRenderDualPane_GhostTextShared verifies the dual-pane view actually
// uses the shared command line (the original bug: ghost text missing in
// dual-pane command mode).
func TestRenderDualPane_GhostTextShared(t *testing.T) {
	m := newChromeTestModel()
	m.leftWidth = 60
	m.rightWidth = 60
	m.commandFocused = true
	m.commandInput = "git s"
	m.commandCursorPos = len(m.commandInput)
	m.ghostText = "git status"

	out := m.renderDualPane()
	if !strings.Contains(out, m.renderCommandLine()) {
		t.Errorf("renderDualPane does not embed the shared command line (ghost text drift regression)")
	}
}

// TestRenderSinglePane_SharedChrome verifies the single-pane view embeds the
// shared header and command line verbatim (behavior preserved after the
// extraction).
func TestRenderSinglePane_SharedChrome(t *testing.T) {
	m := newChromeTestModel()
	m.commandFocused = true
	m.commandInput = "!ls"
	m.commandCursorPos = len(m.commandInput)

	out := m.renderSinglePane()
	if !strings.Contains(out, m.renderHeader("")) {
		t.Errorf("renderSinglePane does not embed the shared header")
	}
	if !strings.Contains(out, m.renderCommandLine()) {
		t.Errorf("renderSinglePane does not embed the shared command line")
	}
}

// TestRenderCommandLine_NarrowDetailHint verifies the single-pane-originated
// scroll hint is preserved behind its existing condition (detail view on a
// narrow terminal), and absent otherwise.
func TestRenderCommandLine_NarrowDetailHint(t *testing.T) {
	m := newChromeTestModel()
	m.width = 60 // narrow terminal
	m.displayMode = modeDetail

	out := m.renderCommandLine()
	if !strings.Contains(out, "←→ scroll") {
		t.Errorf("renderCommandLine missing narrow-terminal detail scroll hint")
	}

	m.width = 120 // wide terminal -> normal hint
	out = m.renderCommandLine()
	if strings.Contains(out, "←→ scroll") {
		t.Errorf("renderCommandLine should not show scroll hint on wide terminal")
	}
	if !strings.Contains(out, ": to focus") {
		t.Errorf("renderCommandLine missing ': to focus' hint")
	}
}

func TestStatusMessageVisible(t *testing.T) {
	m := newChromeTestModel()

	// No message -> not visible
	if m.statusMessageVisible() {
		t.Errorf("statusMessageVisible should be false with empty message")
	}

	// Fresh message -> visible
	m.statusMessage = "Copied"
	m.statusTime = time.Now()
	if !m.statusMessageVisible() {
		t.Errorf("statusMessageVisible should be true for fresh message")
	}

	// Stale message -> auto-dismissed
	m.statusTime = time.Now().Add(-5 * time.Second)
	if m.statusMessageVisible() {
		t.Errorf("statusMessageVisible should be false after 3s auto-dismiss")
	}

	// Stale message but file picker mode -> still visible
	m.filePickerMode = true
	if !m.statusMessageVisible() {
		t.Errorf("statusMessageVisible should be true in file picker mode")
	}

	// Stale message but prompt edit mode -> still visible
	m.filePickerMode = false
	m.promptEditMode = true
	if !m.statusMessageVisible() {
		t.Errorf("statusMessageVisible should be true in prompt edit mode")
	}
}

func TestRenderStatusMessageBar_Truncation(t *testing.T) {
	m := newChromeTestModel()
	m.width = 30
	m.statusMessage = strings.Repeat("x", 100)
	m.statusTime = time.Now()

	out := m.renderStatusMessageBar()
	if !strings.HasSuffix(out, "\033[0m") {
		t.Errorf("renderStatusMessageBar must end with ANSI reset")
	}
	// Visible width must not exceed the terminal width budget (m.width-4
	// content + 2 cells of padding from the style).
	if w := visualWidth(out); w > m.width-2 {
		t.Errorf("renderStatusMessageBar width = %d, want <= %d", w, m.width-2)
	}
}
