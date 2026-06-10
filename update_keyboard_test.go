package main

import (
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
)

// TestCommandInput_StaleCursorInsert is a regression test for tfe-r3s:
// several code paths cleared m.commandInput without resetting
// m.commandCursorPos, so typing a character afterwards sliced an empty
// string with a stale cursor (`""[:2]`) and panicked. The keyboard handler
// now clamps the cursor before any command-input slicing.
func TestCommandInput_StaleCursorInsert(t *testing.T) {
	m := model{
		commandFocused:   true,
		commandInput:     "", // cleared by e.g. Enter / toolbar click / right-click
		commandCursorPos: 2,  // stale cursor left over from previous input ("ls")
	}

	// Typing a character must not panic and must insert at a clamped position
	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	result := newModel.(model)

	if result.commandInput != "a" {
		t.Errorf("commandInput = %q, expected %q", result.commandInput, "a")
	}
	if result.commandCursorPos != 1 {
		t.Errorf("commandCursorPos = %d, expected 1", result.commandCursorPos)
	}
}

// TestCommandInput_StaleCursorBackspace verifies backspace with a stale
// cursor position (beyond the input length) does not panic (tfe-r3s).
func TestCommandInput_StaleCursorBackspace(t *testing.T) {
	m := model{
		commandFocused:   true,
		commandInput:     "x",
		commandCursorPos: 5, // stale: beyond len(commandInput)
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyBackspace})
	result := newModel.(model)

	if result.commandInput != "" {
		t.Errorf("commandInput = %q, expected empty string", result.commandInput)
	}
	if result.commandCursorPos != 0 {
		t.Errorf("commandCursorPos = %d, expected 0", result.commandCursorPos)
	}
}

// TestCommandInput_NegativeCursorClamped verifies a negative cursor position
// is clamped to 0 before any slicing (tfe-r3s defense in depth).
func TestCommandInput_NegativeCursorClamped(t *testing.T) {
	m := model{
		commandFocused:   true,
		commandInput:     "ab",
		commandCursorPos: -1,
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	result := newModel.(model)

	if result.commandInput != "cab" {
		t.Errorf("commandInput = %q, expected %q", result.commandInput, "cab")
	}
	if result.commandCursorPos != 1 {
		t.Errorf("commandCursorPos = %d, expected 1", result.commandCursorPos)
	}
}

// TestPromptEditMode_JKInsertIntoVariable is a regression test for tfe-36e:
// the prompt edit mode block bound "j"/"k" to preview scrolling, so those
// letters never reached the default rune-insertion handler and values like
// "json" or "kitchen" could not be typed into template variables. In edit
// mode, plain letters must insert; arrows still scroll.
func TestPromptEditMode_JKInsertIntoVariable(t *testing.T) {
	newEditModel := func() model {
		// Loaded preview with ample content and height so the consolidated
		// scroll helper (tfe-o5f) has headroom: maxPreviewScroll() well above
		// the starting scrollPos of 5, so the up-arrow decrement is not
		// clamped away. (In real usage scrollPos never exceeds maxScroll.)
		content := make([]string, 100)
		for i := range content {
			content[i] = "line"
		}
		return model{
			promptEditMode:       true,
			focusedVariableIndex: 0,
			filledVariables:      make(map[string]string),
			width:                120,
			height:               40,
			preview: previewModel{
				loaded:   true,
				content:  content,
				isPrompt: true,
				promptTemplate: &promptTemplate{
					name:      "test",
					variables: []string{"TOPIC"},
					template:  "Write about {{TOPIC}}",
				},
				scrollPos: 5,
			},
		}
	}

	for _, r := range []rune{'j', 'k'} {
		m := newEditModel()
		newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		result := newModel.(model)

		if got := result.filledVariables["TOPIC"]; got != string(r) {
			t.Errorf("typing %q: filledVariables[TOPIC] = %q, expected %q", r, got, string(r))
		}
		if result.preview.scrollPos != 5 {
			t.Errorf("typing %q: scrollPos = %d, expected 5 (must not scroll in edit mode)", r, result.preview.scrollPos)
		}
	}

	// Arrow keys must still scroll while editing
	m := newEditModel()
	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyUp})
	result := newModel.(model)
	if result.preview.scrollPos != 4 {
		t.Errorf("up arrow: scrollPos = %d, expected 4", result.preview.scrollPos)
	}
	if got := result.filledVariables["TOPIC"]; got != "" {
		t.Errorf("up arrow: filledVariables[TOPIC] = %q, expected empty", got)
	}
}

// TestCommandInput_EnterResetsCursor verifies that executing a command via
// Enter resets the cursor position along with the input (tfe-r3s).
func TestCommandInput_EnterResetsCursor(t *testing.T) {
	m := model{
		commandFocused:      true,
		commandInput:        "ls",
		commandCursorPos:    2,
		commandHistoryByDir: make(map[string][]string),
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyEnter})
	result := newModel.(model)

	if result.commandInput != "" {
		t.Errorf("commandInput = %q, expected empty string after Enter", result.commandInput)
	}
	if result.commandCursorPos != 0 {
		t.Errorf("commandCursorPos = %d, expected 0 after Enter", result.commandCursorPos)
	}
	if result.commandFocused {
		t.Error("commandFocused should be false after executing a command")
	}
}

// The following tests are regression tests for tfe-7la: the command-line
// cursor was moved/edited one byte at a time, so after a multibyte UTF-8
// character one left-arrow put the cursor mid-rune. Render-time slicing then
// emitted invalid UTF-8 and backspace/delete corrupted the string. Cursor
// movement and editing are now rune-aware.

// "é" is U+00E9 (2 bytes in UTF-8); "你" is U+4F60 (3 bytes); "😀" is
// U+1F600 (4 bytes). These exercise 2-, 3- and 4-byte rune steps.

func TestCommandInput_LeftArrowRuneAware(t *testing.T) {
	// Cursor at end of a 2-byte rune; one left-arrow must land on the
	// rune boundary (byte 0), not mid-rune (byte 1).
	m := model{
		commandFocused:   true,
		commandInput:     "é",
		commandCursorPos: 2,
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyLeft})
	result := newModel.(model)

	if result.commandCursorPos != 0 {
		t.Errorf("commandCursorPos = %d, expected 0 (start of rune)", result.commandCursorPos)
	}
}

func TestCommandInput_RightArrowRuneAware(t *testing.T) {
	// Cursor at start of a 3-byte rune; one right-arrow must land at byte 3,
	// the boundary after the rune.
	m := model{
		commandFocused:   true,
		commandInput:     "你",
		commandCursorPos: 0,
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyRight})
	result := newModel.(model)

	if result.commandCursorPos != 3 {
		t.Errorf("commandCursorPos = %d, expected 3 (end of rune)", result.commandCursorPos)
	}
}

func TestCommandInput_BackspaceRuneAware(t *testing.T) {
	// Backspace at the end of a 4-byte rune must remove the whole rune and
	// leave a valid UTF-8 string, not a single trailing byte.
	m := model{
		commandFocused:   true,
		commandInput:     "a😀",
		commandCursorPos: 5, // 1 (a) + 4 (😀)
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyBackspace})
	result := newModel.(model)

	if result.commandInput != "a" {
		t.Errorf("commandInput = %q, expected %q", result.commandInput, "a")
	}
	if result.commandCursorPos != 1 {
		t.Errorf("commandCursorPos = %d, expected 1", result.commandCursorPos)
	}
	if !utf8.ValidString(result.commandInput) {
		t.Errorf("commandInput %q is not valid UTF-8", result.commandInput)
	}
}

func TestCommandInput_DeleteRuneAware(t *testing.T) {
	// Forward-delete at the start of a 3-byte rune must remove the whole rune.
	m := model{
		commandFocused:   true,
		commandInput:     "你b",
		commandCursorPos: 0,
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyDelete})
	result := newModel.(model)

	if result.commandInput != "b" {
		t.Errorf("commandInput = %q, expected %q", result.commandInput, "b")
	}
	if !utf8.ValidString(result.commandInput) {
		t.Errorf("commandInput %q is not valid UTF-8", result.commandInput)
	}
}

// TestCommandInput_LeftThenBackspaceMultibyte exercises the full corruption
// scenario from the bug report: type a multibyte char, arrow left over it,
// then backspace. With byte-wise handling this left the cursor mid-rune and
// corrupted the string; rune-aware handling keeps everything valid.
func TestCommandInput_LeftThenBackspaceMultibyte(t *testing.T) {
	m := model{
		commandFocused:   true,
		commandInput:     "你好",
		commandCursorPos: 6, // end (two 3-byte runes)
	}

	// Left over "好" -> boundary at byte 3
	nm, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyLeft})
	m = nm.(model)
	if m.commandCursorPos != 3 {
		t.Fatalf("after left: commandCursorPos = %d, expected 3", m.commandCursorPos)
	}

	// Backspace removes "你" entirely
	nm, _ = m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyBackspace})
	m = nm.(model)
	if m.commandInput != "好" {
		t.Errorf("commandInput = %q, expected %q", m.commandInput, "好")
	}
	if m.commandCursorPos != 0 {
		t.Errorf("commandCursorPos = %d, expected 0", m.commandCursorPos)
	}
	if !utf8.ValidString(m.commandInput) {
		t.Errorf("commandInput %q is not valid UTF-8", m.commandInput)
	}
}

// TestF5Copy_StalePreviewFallsBackToPath is a regression test for tfe-spg:
// the browser-mode F5 handler used to copy m.preview.content whenever a
// preview was loaded, without checking it belonged to the selected file. So
// previewing file A, pressing Esc, moving the cursor to file B, then F5 copied
// A's content while reporting success. The handler now guards the content-copy
// branch with m.preview.filePath == currentFile.path and falls back to copying
// the selected file's path on a mismatch.
func TestF5Copy_StalePreviewFallsBackToPath(t *testing.T) {
	m := model{
		files: []fileItem{
			{name: "a.txt", path: "/tmp/a.txt"},
			{name: "b.txt", path: "/tmp/b.txt"},
		},
		cursor: 1, // selected file is b.txt
		// Stale preview left over from viewing a.txt
		preview: previewModel{
			loaded:   true,
			filePath: "/tmp/a.txt",
			content:  []string{"contents of A"},
		},
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyF5})
	result := newModel.(model)

	// Because the loaded preview belongs to a.txt (not the selected b.txt),
	// F5 must fall through to the path-copy branch.
	if result.statusMessage != "Path copied to clipboard" {
		t.Errorf("statusMessage = %q, expected %q (stale preview must not copy content)",
			result.statusMessage, "Path copied to clipboard")
	}
}

// TestF5Copy_FreshPreviewCopiesContent verifies the complementary case: when
// the loaded preview matches the selected file, F5 still copies its content
// (tfe-spg).
func TestF5Copy_FreshPreviewCopiesContent(t *testing.T) {
	m := model{
		files: []fileItem{
			{name: "b.txt", path: "/tmp/b.txt"},
		},
		cursor: 0,
		preview: previewModel{
			loaded:   true,
			filePath: "/tmp/b.txt", // matches selected file
			content:  []string{"contents of B"},
		},
	}

	newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyF5})
	result := newModel.(model)

	if result.statusMessage != "✓ File content copied to clipboard" {
		t.Errorf("statusMessage = %q, expected %q (fresh preview should copy content)",
			result.statusMessage, "✓ File content copied to clipboard")
	}
}

// TestDialogConfirm_DispatchesByActionNotTitle is a regression test for
// tfe-cqe: dialog-confirm behavior must be routed by the typed
// dialogModel.action field, NOT by string-matching the user-facing title.
// We deliberately give each dialog a bogus title that does not match any
// former magic string, and assert the action still fires.
func TestDialogConfirm_DispatchesByActionNotTitle(t *testing.T) {
	t.Run("create directory input dialog", func(t *testing.T) {
		dir := t.TempDir()
		m := model{
			currentPath: dir,
			showDialog:  true,
			dialog: dialogModel{
				dialogType: dialogInput,
				action:     dialogActionCreateDir,
				title:      "totally different wording", // would have broken old title dispatch
				input:      "newdir",
			},
		}

		newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyEnter})
		result := newModel.(model)

		if _, err := os.Stat(filepath.Join(dir, "newdir")); err != nil {
			t.Errorf("expected directory to be created via action dispatch, stat err: %v", err)
		}
		if result.showDialog {
			t.Errorf("dialog should be dismissed after confirm")
		}
	})

	t.Run("delete-entry confirm dialog", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "victim.txt")
		if err := os.WriteFile(target, []byte("bye"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}

		m := model{
			currentPath: dir,
			showDialog:  true,
			files: []fileItem{
				{name: "victim.txt", path: target, isDir: false},
			},
			cursor: 0,
			dialog: dialogModel{
				dialogType: dialogConfirm,
				action:     dialogActionDeleteEntry,
				// Bogus title proves the old "Delete file" || "Delete directory"
				// compound title match is no longer the dispatch key.
				title: "Remove this thing",
			},
		}

		newModel, _ := m.handleKeyEvent(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
		result := newModel.(model)

		// deleteFileOrDir moves the entry to trash; assert it left its original path.
		if _, err := os.Stat(target); !os.IsNotExist(err) {
			t.Errorf("expected file to be removed from original path via action dispatch, stat err: %v", err)
		}
		if result.showDialog {
			t.Errorf("dialog should be dismissed after confirm")
		}
	})
}
