package main

// Tests for git_operations.go
// Regression coverage for tfe-6r2: getLastCommitInfo must not panic on
// short/empty commit hashes read from .git/HEAD or ref files.

import (
	"os"
	"path/filepath"
	"testing"
)

// makeRepoWithHead creates a fake repo dir whose .git/HEAD has the given content.
func makeRepoWithHead(t *testing.T, headContent string) string {
	t.Helper()
	repoPath := t.TempDir()
	gitDir := filepath.Join(repoPath, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte(headContent), 0644); err != nil {
		t.Fatalf("failed to write HEAD: %v", err)
	}
	return repoPath
}

// TestGetLastCommitInfoShortHashNoPanic verifies that malformed HEAD/ref
// contents (empty, 1-char, 2-char) return zero values instead of panicking
// on commitHash[:2] / commitHash[2:] slicing (tfe-6r2).
func TestGetLastCommitInfoShortHashNoPanic(t *testing.T) {
	cases := []struct {
		name string
		head string
	}{
		{"empty HEAD", ""},
		{"whitespace-only HEAD", "\n"},
		{"one-char detached HEAD", "a"},
		{"two-char detached HEAD", "ab"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repoPath := makeRepoWithHead(t, tc.head)
			msg, commitTime := getLastCommitInfo(repoPath)
			if msg != "" {
				t.Errorf("expected empty message, got %q", msg)
			}
			if !commitTime.IsZero() {
				t.Errorf("expected zero time, got %v", commitTime)
			}
		})
	}
}

// TestGetLastCommitInfoEmptyRefFile covers HEAD pointing at a branch ref
// whose file exists but is empty/truncated (interrupted git write).
func TestGetLastCommitInfoEmptyRefFile(t *testing.T) {
	repoPath := makeRepoWithHead(t, "ref: refs/heads/main\n")
	refDir := filepath.Join(repoPath, ".git", "refs", "heads")
	if err := os.MkdirAll(refDir, 0755); err != nil {
		t.Fatalf("failed to create refs dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(refDir, "main"), []byte("\n"), 0644); err != nil {
		t.Fatalf("failed to write ref file: %v", err)
	}

	msg, commitTime := getLastCommitInfo(repoPath)
	if msg != "" {
		t.Errorf("expected empty message, got %q", msg)
	}
	if !commitTime.IsZero() {
		t.Errorf("expected zero time, got %v", commitTime)
	}
}

// TestGetLastCommitInfoMissingHead verifies the existing error path still works.
func TestGetLastCommitInfoMissingHead(t *testing.T) {
	msg, commitTime := getLastCommitInfo(t.TempDir())
	if msg != "" || !commitTime.IsZero() {
		t.Errorf("expected zero values for repo without .git/HEAD, got %q, %v", msg, commitTime)
	}
}

// TestGetLastCommitInfoValidHash verifies a well-formed hash still resolves
// the loose object path and returns its modtime.
func TestGetLastCommitInfoValidHash(t *testing.T) {
	hash := "0123456789abcdef0123456789abcdef01234567"
	repoPath := makeRepoWithHead(t, hash+"\n")
	objDir := filepath.Join(repoPath, ".git", "objects", hash[:2])
	if err := os.MkdirAll(objDir, 0755); err != nil {
		t.Fatalf("failed to create objects dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(objDir, hash[2:]), []byte("dummy"), 0644); err != nil {
		t.Fatalf("failed to write object file: %v", err)
	}

	_, commitTime := getLastCommitInfo(repoPath)
	if commitTime.IsZero() {
		t.Errorf("expected non-zero time for valid loose object")
	}
}
