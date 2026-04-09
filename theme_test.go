package main

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestInitialModelAppliesExplicitThemeFlag(t *testing.T) {
	origArgs := os.Args
	origStartPath := startPath
	origSelectFile := selectFile
	origAutoPreview := autoPreview
	origPreviewFile := previewFile

	t.Cleanup(func() {
		os.Args = origArgs
		startPath = origStartPath
		selectFile = origSelectFile
		autoPreview = origAutoPreview
		previewFile = origPreviewFile
	})

	tempHome := t.TempDir()
	tempPath := t.TempDir()
	t.Setenv("HOME", tempHome)
	startPath = tempPath
	selectFile = ""
	autoPreview = false
	previewFile = ""

	tests := []struct {
		name            string
		args            []string
		initialDarkBg   bool
		wantDarkMode    bool
		wantLightTheme  bool
		wantDarkBgFinal bool
	}{
		{
			name:            "light flag overrides dark autodetect",
			args:            []string{"tfe", "--light"},
			initialDarkBg:   true,
			wantDarkMode:    false,
			wantLightTheme:  true,
			wantDarkBgFinal: false,
		},
		{
			name:            "dark flag overrides light autodetect",
			args:            []string{"tfe", "--dark"},
			initialDarkBg:   false,
			wantDarkMode:    true,
			wantLightTheme:  false,
			wantDarkBgFinal: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			lipgloss.SetHasDarkBackground(tc.initialDarkBg)
			os.Args = tc.args

			m := initialModel()
			t.Cleanup(func() {
				m.closeWatcher()
			})

			if m.config.DarkMode != tc.wantDarkMode {
				t.Fatalf("config.DarkMode = %v, want %v", m.config.DarkMode, tc.wantDarkMode)
			}
			if m.forceLightTheme != tc.wantLightTheme {
				t.Fatalf("forceLightTheme = %v, want %v", m.forceLightTheme, tc.wantLightTheme)
			}
			if lipgloss.HasDarkBackground() != tc.wantDarkBgFinal {
				t.Fatalf("lipgloss.HasDarkBackground() = %v, want %v", lipgloss.HasDarkBackground(), tc.wantDarkBgFinal)
			}
		})
	}
}
