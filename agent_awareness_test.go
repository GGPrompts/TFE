package main

import (
	"testing"
)

func TestIsUnderDir(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		dir      string
		want     bool
	}{
		{"exact match", "/home/user/projects/TFE", "/home/user/projects/TFE", true},
		{"child file", "/home/user/projects/TFE/main.go", "/home/user/projects/TFE", true},
		{"deep child", "/home/user/projects/TFE/pkg/foo/bar.go", "/home/user/projects/TFE", true},
		{"different project", "/home/user/projects/other/main.go", "/home/user/projects/TFE", false},
		{"prefix but not subdir", "/home/user/projects/TFE2/main.go", "/home/user/projects/TFE", false},
		{"parent dir", "/home/user/projects", "/home/user/projects/TFE", false},
		{"empty paths", "", "", true}, // filepath.Clean("") returns "."
		{"empty file", "", "/home/user", false},
		{"empty dir", "/home/user/file.go", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isUnderDir(tt.filePath, tt.dir)
			if got != tt.want {
				t.Errorf("isUnderDir(%q, %q) = %v, want %v", tt.filePath, tt.dir, got, tt.want)
			}
		})
	}
}

func TestAgentLabel(t *testing.T) {
	tests := []struct {
		name    string
		session AgentSession
		want    string
	}{
		{"empty agent type", AgentSession{AgentType: ""}, "CC"},
		{"explore type", AgentSession{AgentType: "Explore"}, "CC:Explore"},
		{"general-purpose type", AgentSession{AgentType: "general-purpose"}, "CC:Agent"},
		{"unknown type", AgentSession{AgentType: "custom"}, "CC:custom"},
		{"long type truncated", AgentSession{AgentType: "very-long-agent-type-name"}, "CC:very-long-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := agentLabel(tt.session)
			if got != tt.want {
				t.Errorf("agentLabel(%v) = %q, want %q", tt.session.AgentType, got, tt.want)
			}
		})
	}
}

func TestMatchFileToAgent(t *testing.T) {
	sessions := []AgentSession{
		{SessionID: "abc", WorkingDir: "/home/user/projects/TFE", AgentType: ""},
		{SessionID: "def", WorkingDir: "/home/user/projects/other", AgentType: "Explore"},
		// Subagent should be skipped
		{SessionID: "abc.agent.123", WorkingDir: "/home/user/projects/TFE", AgentType: "general-purpose", ParentSessionID: "abc"},
	}

	tests := []struct {
		name     string
		filePath string
		want     string
	}{
		{"file in TFE project", "/home/user/projects/TFE/main.go", "CC"},
		{"file in other project", "/home/user/projects/other/foo.go", "CC:Explore"},
		{"file in unrelated project", "/home/user/projects/unrelated/bar.go", ""},
		{"empty path", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchFileToAgent(tt.filePath, sessions)
			if got != tt.want {
				t.Errorf("matchFileToAgent(%q) = %q, want %q", tt.filePath, got, tt.want)
			}
		})
	}
}

func TestMatchFileToAgentEmptySessions(t *testing.T) {
	got := matchFileToAgent("/some/file.go", nil)
	if got != "" {
		t.Errorf("matchFileToAgent with nil sessions = %q, want empty", got)
	}

	got = matchFileToAgent("/some/file.go", []AgentSession{})
	if got != "" {
		t.Errorf("matchFileToAgent with empty sessions = %q, want empty", got)
	}
}

func TestBuildAgentFileMap(t *testing.T) {
	sessions := []AgentSession{
		{SessionID: "abc", WorkingDir: "/home/user/projects/TFE"},
	}

	changedFiles := []fileItem{
		{name: "[M ] main.go", path: "/home/user/projects/TFE/main.go"},
		{name: "[??] new.go", path: "/home/user/projects/TFE/new.go"},
	}

	result := buildAgentFileMap(changedFiles, sessions)
	if result == nil {
		t.Fatal("expected non-nil map")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result))
	}
	if result["/home/user/projects/TFE/main.go"] != "CC" {
		t.Errorf("expected CC label for main.go, got %q", result["/home/user/projects/TFE/main.go"])
	}
}

func TestBuildAgentFileMapNilInputs(t *testing.T) {
	// No sessions
	result := buildAgentFileMap([]fileItem{{path: "/foo"}}, nil)
	if result != nil {
		t.Error("expected nil map with no sessions")
	}

	// No changed files
	result = buildAgentFileMap(nil, []AgentSession{{WorkingDir: "/foo"}})
	if result != nil {
		t.Error("expected nil map with no changed files")
	}

	// No matches
	result = buildAgentFileMap(
		[]fileItem{{path: "/unrelated/file.go"}},
		[]AgentSession{{WorkingDir: "/other/project"}},
	)
	if result != nil {
		t.Error("expected nil map when no files match")
	}
}

func TestGetAgentSessionsGraceful(t *testing.T) {
	// This test verifies that getAgentSessions doesn't panic or error
	// when the state directory doesn't exist or has issues.
	// It reads the real /tmp/claude-code-state/ if available, or returns nil gracefully.
	sessions := getAgentSessions()
	// We can't assert the count since it depends on the environment,
	// but we verify it doesn't panic.
	_ = sessions
}
