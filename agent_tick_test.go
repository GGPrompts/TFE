package main

// Tests for agent auto-watch tick scheduling (tfe-jew): the agentCheckTick
// loop must self-terminate when Auto Changes is disabled (it previously
// re-armed unconditionally and ran for the process lifetime), and enabling
// Auto Changes mid-session must start the loop exactly once (it was only
// created in Init at startup before). Mirrors the gitReposTick pattern
// (tfe-54g) using the agentTickRunning guard flag.

import (
	"testing"
)

// TestAgentCheckTickMsg_StopsWhenDisabled verifies the agent poll tick does
// NOT re-arm (and clears its running flag) once auto-watch is disabled.
func TestAgentCheckTickMsg_StopsWhenDisabled(t *testing.T) {
	m := model{
		agentAutoWatch:         false,
		agentTickRunning:       true,
		lastKnownAgentSessions: make(map[string]string),
	}

	newModel, cmd := m.Update(agentCheckTickMsg{})
	if cmd != nil {
		t.Fatalf("expected nil cmd when auto-watch is disabled, got non-nil")
	}
	nm, ok := newModel.(model)
	if !ok {
		t.Fatalf("expected model back from Update")
	}
	if nm.agentTickRunning {
		t.Fatalf("expected agentTickRunning to be cleared when auto-watch is disabled")
	}
}

// TestAgentCheckTickMsg_RearmsWhileEnabled verifies the agent poll tick keeps
// re-arming while auto-watch stays enabled.
func TestAgentCheckTickMsg_RearmsWhileEnabled(t *testing.T) {
	m := model{
		agentAutoWatch:         true,
		agentTickRunning:       true,
		lastKnownAgentSessions: make(map[string]string),
	}

	_, cmd := m.Update(agentCheckTickMsg{})
	if cmd == nil {
		t.Fatalf("expected non-nil cmd while auto-watch is enabled, got nil")
	}
}

// TestSettingsToggle_StartsAgentTickOnEnable verifies enabling auto_changes
// mid-session starts the poll loop exactly once and sets the running flag, and
// that a stacked enable (loop already live) does NOT start a second loop.
func TestSettingsToggle_StartsAgentTickOnEnable(t *testing.T) {
	m := &model{
		lastKnownAgentSessions: make(map[string]string),
	}

	// Enable mid-session: setConfigBool flips agentAutoWatch, settingsToggleCmd
	// must start the tick and set the guard.
	m.setConfigBool("auto_changes", true)
	if !m.agentAutoWatch {
		t.Fatalf("expected agentAutoWatch to be true after enabling auto_changes")
	}
	cmd := m.settingsToggleCmd("auto_changes", true)
	if cmd == nil {
		t.Fatalf("expected a tick cmd when enabling auto_changes mid-session, got nil")
	}
	if !m.agentTickRunning {
		t.Fatalf("expected agentTickRunning to be set after enabling auto_changes")
	}

	// A second enable while a loop is already live must NOT start another loop
	// (the guard prevents two concurrent ticks).
	cmd = m.settingsToggleCmd("auto_changes", true)
	if cmd != nil {
		t.Fatalf("expected nil cmd when a loop is already running, got non-nil")
	}
}
