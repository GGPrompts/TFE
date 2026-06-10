package main

// Tests for git_operations.go
// Regression coverage for tfe-6r2: getLastCommitInfo must not panic on
// short/empty commit hashes read from .git/HEAD or ref files.

import (
	"os"
	"os/exec"
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

// --- getAheadBehindCounts (tfe-1jq) ---------------------------------------

// runGit runs a git command in dir and fails the test on error.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	// Keep commits deterministic and independent of the developer's git config.
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
		"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
}

// commitFile writes a file and commits it, returning nothing.
func commitFile(t *testing.T, dir, name, content, msg string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	runGit(t, dir, "add", name)
	runGit(t, dir, "commit", "-m", msg)
}

// setupRepoWithUpstream creates an "origin" bare repo and a clone tracking it,
// returning the clone path and its branch name. Skips the test if git is absent.
func setupRepoWithUpstream(t *testing.T) (clonePath, branch string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	base := t.TempDir()
	origin := filepath.Join(base, "origin.git")
	runGit(t, base, "init", "--bare", "-b", "main", origin)

	clonePath = filepath.Join(base, "clone")
	runGit(t, base, "clone", origin, clonePath)

	commitFile(t, clonePath, "a.txt", "1", "initial")
	runGit(t, clonePath, "push", "origin", "HEAD:main")
	// Ensure the local branch tracks origin/main.
	runGit(t, clonePath, "branch", "--set-upstream-to=origin/main", "main")

	return clonePath, getGitBranch(clonePath)
}

func TestGetAheadBehindInSync(t *testing.T) {
	clone, branch := setupRepoWithUpstream(t)
	ahead, behind := getAheadBehindCounts(clone, branch)
	if ahead != 0 || behind != 0 {
		t.Errorf("in-sync: got ahead=%d behind=%d, want 0/0", ahead, behind)
	}
}

func TestGetAheadBehindAhead(t *testing.T) {
	clone, branch := setupRepoWithUpstream(t)
	commitFile(t, clone, "b.txt", "2", "local only")
	commitFile(t, clone, "c.txt", "3", "local only 2")

	ahead, behind := getAheadBehindCounts(clone, branch)
	if ahead != 2 || behind != 0 {
		t.Errorf("ahead case: got ahead=%d behind=%d, want 2/0", ahead, behind)
	}
}

func TestGetAheadBehindBehind(t *testing.T) {
	clone, branch := setupRepoWithUpstream(t)
	// Advance origin via a second clone, then fetch into the first clone so it
	// is strictly behind.
	base := filepath.Dir(clone)
	origin := filepath.Join(base, "origin.git")
	other := filepath.Join(base, "other")
	runGit(t, base, "clone", origin, other)
	commitFile(t, other, "d.txt", "4", "remote only")
	runGit(t, other, "push", "origin", "HEAD:main")

	runGit(t, clone, "fetch", "origin")

	ahead, behind := getAheadBehindCounts(clone, branch)
	if ahead != 0 || behind != 1 {
		t.Errorf("behind case: got ahead=%d behind=%d, want 0/1", ahead, behind)
	}
}

func TestGetAheadBehindDiverged(t *testing.T) {
	clone, branch := setupRepoWithUpstream(t)
	// Local commit.
	commitFile(t, clone, "b.txt", "2", "local only")

	// Remote commit via second clone.
	base := filepath.Dir(clone)
	origin := filepath.Join(base, "origin.git")
	other := filepath.Join(base, "other")
	runGit(t, base, "clone", origin, other)
	commitFile(t, other, "d.txt", "4", "remote only")
	runGit(t, other, "push", "origin", "HEAD:main")

	runGit(t, clone, "fetch", "origin")

	ahead, behind := getAheadBehindCounts(clone, branch)
	if ahead != 1 || behind != 1 {
		t.Errorf("diverged case: got ahead=%d behind=%d, want 1/1", ahead, behind)
	}
}

func TestGetAheadBehindNoUpstream(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	repo := t.TempDir()
	runGit(t, repo, "init", "-b", "main", repo)
	commitFile(t, repo, "a.txt", "1", "initial")

	// No origin remote -> rev-list errors -> graceful (0, 0), no fabricated dir.
	ahead, behind := getAheadBehindCounts(repo, getGitBranch(repo))
	if ahead != 0 || behind != 0 {
		t.Errorf("no-upstream: got ahead=%d behind=%d, want 0/0", ahead, behind)
	}
}

func TestGetAheadBehindEmptyBranch(t *testing.T) {
	ahead, behind := getAheadBehindCounts(t.TempDir(), "")
	if ahead != 0 || behind != 0 {
		t.Errorf("empty branch: got ahead=%d behind=%d, want 0/0", ahead, behind)
	}
}

func TestFormatGitStatusBehind(t *testing.T) {
	// Regression for tfe-1jq: formatGitStatus must surface a real Behind count.
	got := formatGitStatus(gitStatus{behind: 3})
	if got != "↓3 Behind" {
		t.Errorf("formatGitStatus behind: got %q, want %q", got, "↓3 Behind")
	}
}
