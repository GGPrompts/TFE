package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestOpenFileWithBestTool_ViewerDispatch verifies the shared F4 dispatch
// helper (tfe-bpl) routes each special file type to a viewer command.
// The open* viewer helpers always return a non-nil tea.Cmd (they emit an
// editorFinishedMsg error command when no viewer is installed), so a
// non-nil result proves the viewer branch was taken.
func TestOpenFileWithBestTool_ViewerDispatch(t *testing.T) {
	paths := []string{
		"/tmp/data.csv",  // CSV viewer
		"/tmp/movie.mp4", // video player
		"/tmp/song.mp3",  // audio player
		"/tmp/doc.pdf",   // PDF viewer
		"/tmp/app.db",    // database viewer
	}

	for _, path := range paths {
		m := model{}
		cmd := m.openFileWithBestTool(path)
		if cmd == nil {
			t.Errorf("openFileWithBestTool(%q) returned nil, expected viewer command", path)
		}
		if m.statusMessage != "" {
			t.Errorf("openFileWithBestTool(%q) set status %q, expected none", path, m.statusMessage)
		}
	}
}

// TestOpenFileWithBestTool_BinaryFile verifies binary files (containing
// null bytes) are routed to the hex viewer rather than a text editor.
func TestOpenFileWithBestTool_BinaryFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blob.bin")
	if err := os.WriteFile(path, []byte{0x7f, 0x00, 0x01, 0x02}, 0o644); err != nil {
		t.Fatal(err)
	}

	m := model{}
	cmd := m.openFileWithBestTool(path)
	if cmd == nil {
		t.Errorf("openFileWithBestTool(%q) returned nil, expected hex viewer command", path)
	}
}

// TestOpenFileWithBestTool_NoEditor verifies the text-file fallback when no
// editor exists on PATH: nil command plus an error status message.
func TestOpenFileWithBestTool_NoEditor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("plain text"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("PATH", t.TempDir()) // empty dir: no editors available

	m := model{}
	cmd := m.openFileWithBestTool(path)
	if cmd != nil {
		t.Error("openFileWithBestTool returned a command, expected nil with no editor on PATH")
	}
	if !m.statusIsError {
		t.Error("statusIsError = false, expected true")
	}
	if !strings.Contains(m.statusMessage, "No editor available") {
		t.Errorf("statusMessage = %q, expected it to mention no editor available", m.statusMessage)
	}
}

// TestOpenFileWithBestTool_TextEditor verifies a text file opens in an
// editor when one is on PATH (stubbed nano in a temp dir).
func TestOpenFileWithBestTool_TextEditor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notes.txt")
	if err := os.WriteFile(path, []byte("plain text"), 0o644); err != nil {
		t.Fatal(err)
	}

	binDir := t.TempDir()
	stub := filepath.Join(binDir, "nano")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir)

	m := model{}
	cmd := m.openFileWithBestTool(path)
	if cmd == nil {
		t.Error("openFileWithBestTool returned nil, expected editor command for text file")
	}
	if m.statusMessage != "" {
		t.Errorf("statusMessage = %q, expected none", m.statusMessage)
	}
}
