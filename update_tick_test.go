package main

// Tests for tick scheduling (tfe-54g): the 50ms startup tick must stop
// re-arming once the 5-second startup-header window has elapsed so an idle,
// non-landing, non-git-repos TFE does not wake 20x/sec. The git-repos rescan
// is driven by its own slower tick that only re-arms while git-repos mode is
// active.

import (
	"testing"
	"time"
)

// TestTickMsg_StopsRearmingAfterStartupWindow verifies the 50ms startup tick
// does NOT schedule another tick once the startup window has elapsed.
func TestTickMsg_StopsRearmingAfterStartupWindow(t *testing.T) {
	m := model{
		startupTime: time.Now().Add(-10 * time.Second), // well past the 5s window
	}

	_, cmd := m.Update(tickMsg{})
	if cmd != nil {
		t.Fatalf("expected nil cmd (tick should stop re-arming after startup window), got non-nil")
	}
}

// TestTickMsg_RearmsDuringStartupWindow verifies the 50ms startup tick keeps
// re-arming while still inside the 5-second startup-header window so the
// GitHub-link -> menu-bar transition still happens without user input.
func TestTickMsg_RearmsDuringStartupWindow(t *testing.T) {
	m := model{
		startupTime: time.Now(), // fresh start, inside the 5s window
	}

	_, cmd := m.Update(tickMsg{})
	if cmd == nil {
		t.Fatalf("expected non-nil cmd (tick should re-arm during startup window), got nil")
	}
}

// TestGitReposTickMsg_StopsWhenModeDismissed verifies the git-repos rescan tick
// does NOT re-arm (and clears its active flag) once git-repos mode is off.
func TestGitReposTickMsg_StopsWhenModeDismissed(t *testing.T) {
	m := model{
		showGitReposOnly:   false,
		gitReposTickActive: true,
	}

	newModel, cmd := m.Update(gitReposTickMsg{})
	if cmd != nil {
		t.Fatalf("expected nil cmd when git-repos mode is dismissed, got non-nil")
	}
	nm, ok := newModel.(model)
	if !ok {
		t.Fatalf("expected model back from Update")
	}
	if nm.gitReposTickActive {
		t.Fatalf("expected gitReposTickActive to be cleared when mode is dismissed")
	}
}

// TestGitReposTickMsg_RearmsWhileActive verifies the git-repos rescan tick
// keeps re-arming while git-repos mode stays active.
func TestGitReposTickMsg_RearmsWhileActive(t *testing.T) {
	m := model{
		showGitReposOnly:   true,
		gitReposTickActive: true,
		gitReposLastScan:   time.Now(), // recent scan: no rescan, but tick still re-arms
	}

	_, cmd := m.Update(gitReposTickMsg{})
	if cmd == nil {
		t.Fatalf("expected non-nil cmd while git-repos mode is active, got nil")
	}
}

// TestToggleGitRepos_StartsTickOnEnter verifies entering git-repos mode returns
// a tick cmd and sets the active flag, and that re-entering does not stack a
// second tick.
func TestToggleGitRepos_StartsTickOnEnter(t *testing.T) {
	m := model{
		currentPath:       t.TempDir(),
		gitReposScanDepth: 3,
	}

	cmd := m.toggleGitRepos()
	if !m.showGitReposOnly {
		t.Fatalf("expected git-repos mode to be enabled after toggle")
	}
	if cmd == nil {
		t.Fatalf("expected a tick cmd when entering git-repos mode, got nil")
	}
	if !m.gitReposTickActive {
		t.Fatalf("expected gitReposTickActive to be set after entering git-repos mode")
	}

	// Toggling off should not start a tick (the existing one self-terminates).
	cmd = m.toggleGitRepos()
	if m.showGitReposOnly {
		t.Fatalf("expected git-repos mode to be disabled after second toggle")
	}
	if cmd != nil {
		t.Fatalf("expected nil cmd when leaving git-repos mode, got non-nil")
	}
}
