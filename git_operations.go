package main

// Module: git_operations.go
// Purpose: Git repository operations (pull, push, sync, fetch)
// Responsibilities:
// - Execute git commands in repository directories
// - Handle git errors and conflicts
// - Provide user feedback for git operations

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// gitOperationFinishedMsg is sent when a git operation completes
type gitOperationFinishedMsg struct {
	operation string // "pull", "push", "sync", "fetch"
	err       error
	output    string
}

// gitPull executes git pull in the specified directory
func gitPull(repoPath string) tea.Cmd {
	script := fmt.Sprintf(`
echo "$ git pull"
cd %s || exit 1
git pull
exitCode=$?
echo ""
if [ $exitCode -eq 0 ]; then
    echo "✓ Pull completed successfully"
else
    echo "✗ Pull failed with exit code: $exitCode"
fi
echo ""
echo "Press any key to continue..."
read -n 1 -s -r
exit $exitCode
`, shellQuote(repoPath))

	c := exec.Command("bash", "-c", script)
	return tea.Sequence(
		tea.ClearScreen,
		tea.ExecProcess(c, func(err error) tea.Msg {
			return gitOperationFinishedMsg{operation: "pull", err: err}
		}),
	)
}

// gitPush executes git push in the specified directory
func gitPush(repoPath string) tea.Cmd {
	script := fmt.Sprintf(`
echo "$ git push"
cd %s || exit 1
git push
exitCode=$?
echo ""
if [ $exitCode -eq 0 ]; then
    echo "✓ Push completed successfully"
else
    echo "✗ Push failed with exit code: $exitCode"
fi
echo ""
echo "Press any key to continue..."
read -n 1 -s -r
exit $exitCode
`, shellQuote(repoPath))

	c := exec.Command("bash", "-c", script)
	return tea.Sequence(
		tea.ClearScreen,
		tea.ExecProcess(c, func(err error) tea.Msg {
			return gitOperationFinishedMsg{operation: "push", err: err}
		}),
	)
}

// gitSync executes git pull followed by git push (smart sync)
func gitSync(repoPath string) tea.Cmd {
	script := fmt.Sprintf(`
echo "$ git sync (pull + push)"
cd %s || exit 1

echo "Step 1: Pulling changes..."
git pull
pullCode=$?

if [ $pullCode -ne 0 ]; then
    echo ""
    echo "✗ Pull failed with exit code: $pullCode"
    echo "Cannot proceed with push."
    echo ""
    echo "Press any key to continue..."
    read -n 1 -s -r
    exit $pullCode
fi

echo ""
echo "✓ Pull completed successfully"
echo ""
echo "Step 2: Pushing changes..."
git push
pushCode=$?

echo ""
if [ $pushCode -eq 0 ]; then
    echo "✓ Sync completed successfully (pulled and pushed)"
else
    echo "✗ Push failed with exit code: $pushCode"
    echo "Note: Pull succeeded, but push failed."
fi
echo ""
echo "Press any key to continue..."
read -n 1 -s -r
exit $pushCode
`, shellQuote(repoPath))

	c := exec.Command("bash", "-c", script)
	return tea.Sequence(
		tea.ClearScreen,
		tea.ExecProcess(c, func(err error) tea.Msg {
			return gitOperationFinishedMsg{operation: "sync", err: err}
		}),
	)
}

// gitFetch executes git fetch to update remote tracking branches
func gitFetch(repoPath string) tea.Cmd {
	script := fmt.Sprintf(`
echo "$ git fetch"
cd %s || exit 1
git fetch
exitCode=$?
echo ""
if [ $exitCode -eq 0 ]; then
    echo "✓ Fetch completed successfully"
    echo ""
    echo "Remote tracking branches updated."
    echo "Use 'git status' to see if your branch is behind/ahead."
else
    echo "✗ Fetch failed with exit code: $exitCode"
fi
echo ""
echo "Press any key to continue..."
read -n 1 -s -r
exit $exitCode
`, shellQuote(repoPath))

	c := exec.Command("bash", "-c", script)
	return tea.Sequence(
		tea.ClearScreen,
		tea.ExecProcess(c, func(err error) tea.Msg {
			return gitOperationFinishedMsg{operation: "fetch", err: err}
		}),
	)
}

func isGitRepo(path string) bool {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// getGitBranch returns the current git branch name for a repository
// Returns empty string if not a git repo or error occurs
func getGitBranch(repoPath string) string {
	// Read .git/HEAD to get current branch
	headPath := filepath.Join(repoPath, ".git", "HEAD")
	content, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}

	// HEAD format: "ref: refs/heads/main\n"
	head := strings.TrimSpace(string(content))
	if strings.HasPrefix(head, "ref: refs/heads/") {
		return strings.TrimPrefix(head, "ref: refs/heads/")
	}

	// Detached HEAD state - show short hash
	if len(head) >= 7 {
		return head[:7]
	}

	return ""
}

// hasUncommittedChanges checks if a git repo has uncommitted changes
// Returns false if not a git repo or error occurs
func hasUncommittedChanges(repoPath string) bool {
	// Use git status --porcelain to check for uncommitted changes
	// This is accurate but slower than file mtime checks
	cmd := exec.Command("git", "-C", repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If output is empty, working directory is clean
	// If output has any lines, there are uncommitted changes
	return len(strings.TrimSpace(string(output))) > 0
}

// gitStatus represents the git status of a repository
type gitStatus struct {
	branch        string    // Current branch name
	ahead         int       // Commits ahead of remote
	behind        int       // Commits behind remote
	dirty         bool      // Has uncommitted changes
	lastCommitMsg string    // Last commit message (first line)
	lastCommitTime time.Time // Time of last commit
}

// getGitStatus returns comprehensive git status for a repository
// Returns empty gitStatus if not a git repo or error occurs
func getGitStatus(repoPath string) gitStatus {
	status := gitStatus{}

	// Check if it's a git repo
	if !isGitRepo(repoPath) {
		return status
	}

	// Get branch
	status.branch = getGitBranch(repoPath)

	// Check for uncommitted changes
	status.dirty = hasUncommittedChanges(repoPath)

	// Get ahead/behind counts by reading refs
	ahead, behind := getAheadBehindCounts(repoPath, status.branch)
	status.ahead = ahead
	status.behind = behind

	// Get last commit info
	commitMsg, commitTime := getLastCommitInfo(repoPath)
	status.lastCommitMsg = commitMsg
	status.lastCommitTime = commitTime

	return status
}

// getAheadBehindCounts returns how many commits ahead/behind the remote
// Returns (0, 0) if no remote or error occurs
func getAheadBehindCounts(repoPath, branch string) (int, int) {
	if branch == "" {
		return 0, 0
	}

	// Read local branch ref
	localRef := filepath.Join(repoPath, ".git", "refs", "heads", branch)
	localHash, err := os.ReadFile(localRef)
	if err != nil {
		return 0, 0
	}
	localCommit := strings.TrimSpace(string(localHash))

	// Read remote branch ref (assuming origin)
	remoteRef := filepath.Join(repoPath, ".git", "refs", "remotes", "origin", branch)
	remoteHash, err := os.ReadFile(remoteRef)
	if err != nil {
		// No remote tracking branch
		return 0, 0
	}
	remoteCommit := strings.TrimSpace(string(remoteHash))

	// If commits are the same, we're in sync
	if localCommit == remoteCommit {
		return 0, 0
	}

	// Count commits ahead and behind using git log
	// This is a simplified check - real implementation would parse git objects
	// For now, we'll use a heuristic based on commit hash comparison
	// If local != remote, we're either ahead or behind (or diverged)

	// Try to determine ahead/behind by checking packed-refs as fallback
	ahead, behind := checkPackedRefs(repoPath, branch, localCommit, remoteCommit)

	return ahead, behind
}

// checkPackedRefs checks packed-refs file for commit history
// This is a simplified heuristic - not 100% accurate
func checkPackedRefs(repoPath, branch, localCommit, remoteCommit string) (int, int) {
	// For now, if commits differ, assume we're ahead by 1
	// A proper implementation would parse the git object database
	// This is a placeholder for the real git log parsing

	// Simple heuristic: if local and remote differ, mark as diverged (1 ahead, 1 behind)
	// Real implementation would use: git rev-list --count local..remote
	if localCommit != remoteCommit {
		return 1, 0 // Assume ahead for now
	}

	return 0, 0
}

// getLastCommitInfo returns the last commit message and time
// Returns empty string and zero time if error occurs
func getLastCommitInfo(repoPath string) (string, time.Time) {
	// Read HEAD to get current commit hash
	headPath := filepath.Join(repoPath, ".git", "HEAD")
	headContent, err := os.ReadFile(headPath)
	if err != nil {
		return "", time.Time{}
	}

	head := strings.TrimSpace(string(headContent))
	var commitHash string

	// Parse HEAD reference
	if strings.HasPrefix(head, "ref: ") {
		// HEAD points to a branch ref
		refPath := strings.TrimPrefix(head, "ref: ")
		refPath = filepath.Join(repoPath, ".git", refPath)
		refContent, err := os.ReadFile(refPath)
		if err != nil {
			return "", time.Time{}
		}
		commitHash = strings.TrimSpace(string(refContent))
	} else {
		// Detached HEAD - head is the commit hash
		commitHash = head
	}

	// Read commit object (simplified - just get timestamp from file modtime)
	// A proper implementation would parse the git commit object
	commitObjectPath := filepath.Join(repoPath, ".git", "objects", commitHash[:2], commitHash[2:])
	commitInfo, err := os.Stat(commitObjectPath)
	if err != nil {
		// Try packed objects
		indexPath := filepath.Join(repoPath, ".git", "index")
		if indexInfo, err := os.Stat(indexPath); err == nil {
			return "", indexInfo.ModTime()
		}
		return "", time.Time{}
	}

	// Use commit object file's modification time as approximate commit time
	return "", commitInfo.ModTime()
}

// formatGitStatus formats git status into a human-readable string with emoji
func formatGitStatus(status gitStatus) string {
	if status.dirty {
		return "⚡ Dirty"
	}

	if status.ahead > 0 && status.behind > 0 {
		return fmt.Sprintf("↑%d↓%d Diverged", status.ahead, status.behind)
	}

	if status.ahead > 0 {
		return fmt.Sprintf("↑%d Ahead", status.ahead)
	}

	if status.behind > 0 {
		return fmt.Sprintf("↓%d Behind", status.behind)
	}

	return "✓ Clean"
}

// formatLastCommitTime formats commit time as relative time (e.g., "2 hours ago")
func formatLastCommitTime(t time.Time) string {
	if t.IsZero() {
		return "Unknown"
	}

	duration := time.Since(t)

	if duration < time.Minute {
		return "Just now"
	}
	if duration < time.Hour {
		mins := int(duration.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", mins)
	}
	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}
	if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
	if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
	if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	years := int(duration.Hours() / 24 / 365)
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// scanGitReposRecursive recursively scans for git repositories
// Returns list of discovered repos, limited by maxDepth and maxRepos
func (m *model) scanGitReposRecursive(startPath string, maxDepth int, maxRepos int) []fileItem {
	repos := make([]fileItem, 0)
	visited := make(map[string]bool) // Prevent infinite loops with symlinks

	var scan func(path string, currentDepth int)
	scan = func(path string, currentDepth int) {
		// Stop if we've hit depth limit or repo limit
		if currentDepth > maxDepth || len(repos) >= maxRepos {
			return
		}

		// Prevent infinite loops
		absPath, err := filepath.Abs(path)
		if err != nil {
			return
		}
		if visited[absPath] {
			return
		}
		visited[absPath] = true

		// Check if this directory is a git repo
		if isGitRepo(path) {
			info, err := os.Stat(path)
			if err == nil {
				// Get comprehensive git status
				gitStat := getGitStatus(path)

				repos = append(repos, fileItem{
					name:          filepath.Base(path),
					path:          path,
					isDir:         true,
					size:          info.Size(),
					modTime:       info.ModTime(),
					mode:          info.Mode(),
					isGitRepo:     true,
					gitBranch:     gitStat.branch,
					gitAhead:      gitStat.ahead,
					gitBehind:     gitStat.behind,
					gitDirty:      gitStat.dirty,
					gitLastCommit: gitStat.lastCommitTime,
				})
			}
			// Don't scan inside git repos (skip .git and subdirs)
			return
		}

		// Read directory entries
		entries, err := os.ReadDir(path)
		if err != nil {
			return // Permission denied or invalid directory
		}

		// Recursively scan subdirectories
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			fullPath := filepath.Join(path, entry.Name())

			// CRITICAL: Skip symlinks entirely - don't follow them
			// This prevents scanning into /usr, /bin, etc. via symlinks
			fileInfo, err := os.Lstat(fullPath)
			if err != nil {
				continue
			}
			if fileInfo.Mode()&os.ModeSymlink != 0 {
				continue // Skip all symlinks
			}

			// Skip hidden directories (except important ones)
			if strings.HasPrefix(entry.Name(), ".") {
				// Exception: scan these important hidden directories
				importantDirs := []string{".config", ".local"}
				isImportant := false
				for _, dir := range importantDirs {
					if entry.Name() == dir {
						isImportant = true
						break
					}
				}
				if !isImportant {
					continue
				}
			}

			// Skip common large/irrelevant directories
			skipDirs := []string{"node_modules", "venv", ".venv", "build", "dist", "target", ".cache", "Library", "Applications"}
			shouldSkip := false
			for _, dir := range skipDirs {
				if entry.Name() == dir {
					shouldSkip = true
					break
				}
			}
			if shouldSkip {
				continue
			}

			scan(fullPath, currentDepth+1)
		}
	}

	scan(startPath, 0)
	return repos
}

// getGitStatusSortValue returns a sort priority for git status
// Lower values = higher priority (sorted first)
func getGitStatusSortValue(item fileItem) int {
	if !item.isGitRepo {
		return 999 // No git status, sort last
	}

	// Priority: dirty repos first, then ahead/behind, then clean
	if item.gitDirty {
		return 0 // Dirty repos (uncommitted changes) - highest priority
	}

	if item.gitAhead > 0 || item.gitBehind > 0 {
		return 1 // Repos with ahead/behind status
	}

	return 2 // Clean repos with no changes
}

// getFileDiff returns the git diff output for a specific file.
// It tries multiple strategies based on the file's git status:
//   - For untracked files (status "??"), returns the full file content with a NEW FILE header
//   - For deleted files (status " D" or "D "), shows content from HEAD via git show
//   - For staged files, uses "git diff --cached -- <path>"
//   - For unstaged modifications, uses "git diff -- <path>"
//   - Falls back to "git diff HEAD -- <path>" to catch both staged and unstaged
func (m *model) getFileDiff(path string, gitStatusCode string) (string, error) {
	gitRoot := m.findGitRoot(m.currentPath)
	if gitRoot == "" {
		return "", fmt.Errorf("not inside a git repository")
	}

	// Get the relative path from git root
	relPath, err := filepath.Rel(gitRoot, path)
	if err != nil {
		relPath = path
	}

	statusCode := strings.TrimSpace(gitStatusCode)

	// Untracked files: show full content with NEW FILE header
	if statusCode == "??" {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("cannot read untracked file: %w", err)
		}
		lines := strings.Split(string(content), "\n")
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("new file: %s\n", relPath))
		sb.WriteString("--- /dev/null\n")
		sb.WriteString(fmt.Sprintf("+++ b/%s\n", relPath))
		sb.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(lines)))
		for _, line := range lines {
			sb.WriteString("+" + line + "\n")
		}
		return sb.String(), nil
	}

	// Deleted files: show content from HEAD
	if statusCode == "D" || strings.HasPrefix(gitStatusCode, "D") || strings.HasSuffix(gitStatusCode, "D") {
		cmd := exec.Command("git", "-C", gitRoot, "show", "HEAD:"+relPath)
		output, err := cmd.Output()
		if err != nil {
			return "", fmt.Errorf("cannot show deleted file: %w", err)
		}
		lines := strings.Split(string(output), "\n")
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("deleted file: %s\n", relPath))
		sb.WriteString(fmt.Sprintf("--- a/%s\n", relPath))
		sb.WriteString("+++ /dev/null\n")
		sb.WriteString(fmt.Sprintf("@@ -1,%d +0,0 @@\n", len(lines)))
		for _, line := range lines {
			sb.WriteString("-" + line + "\n")
		}
		return sb.String(), nil
	}

	// Try git diff HEAD -- <path> (catches both staged and unstaged changes)
	cmd := exec.Command("git", "-C", gitRoot, "diff", "HEAD", "--", relPath)
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return string(output), nil
	}

	// Fallback: try git diff --cached (staged only)
	cmd = exec.Command("git", "-C", gitRoot, "diff", "--cached", "--", relPath)
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		return string(output), nil
	}

	// Fallback: try git diff (unstaged only)
	cmd = exec.Command("git", "-C", gitRoot, "diff", "--", relPath)
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		return string(output), nil
	}

	return "", fmt.Errorf("no diff available for %s", relPath)
}

// extractGitStatusCode extracts the two-character git status code from a changedFiles item name.
// The name format is "[XX] relative/path" where XX is the git status code.
func extractGitStatusCode(itemName string) string {
	if len(itemName) >= 4 && itemName[0] == '[' && itemName[3] == ']' {
		return itemName[1:3]
	}
	return ""
}

// getChangedFiles runs `git status --porcelain` from the git root and returns
// fileItems for every modified, added, deleted, or untracked file.  Each item's
// name is prefixed with the two-character git status indicator (e.g. " M", "??").
// Returns an error if the current directory is not inside a git repository.
func (m *model) getChangedFiles() ([]fileItem, error) {
	// Find git root from current path
	gitRoot := m.findGitRoot(m.currentPath)
	if gitRoot == "" {
		return nil, fmt.Errorf("not inside a git repository")
	}

	cmd := exec.Command("git", "-C", gitRoot, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git status failed: %w", err)
	}

	lines := strings.Split(strings.TrimRight(string(output), "\n"), "\n")
	items := make([]fileItem, 0, len(lines))

	for _, line := range lines {
		if len(line) < 4 {
			continue // malformed line
		}

		// git status --porcelain format: XY <path>
		// X = index status, Y = working tree status
		statusCode := line[:2]
		relPath := strings.TrimSpace(line[3:])

		// Handle renames: "R  old -> new"
		if strings.Contains(relPath, " -> ") {
			parts := strings.SplitN(relPath, " -> ", 2)
			relPath = parts[1]
		}

		fullPath := filepath.Join(gitRoot, relPath)

		// Stat the file to get info (may fail for deleted files)
		info, statErr := os.Stat(fullPath)

		item := fileItem{
			name: fmt.Sprintf("[%s] %s", statusCode, relPath),
			path: fullPath,
		}

		if statErr == nil {
			item.isDir = info.IsDir()
			item.size = info.Size()
			item.modTime = info.ModTime()
			item.mode = info.Mode()
		} else {
			// File was deleted — mark with zero time, keep path for display
			item.isDir = false
			item.size = 0
			item.modTime = time.Time{}
			item.mode = 0
		}

		items = append(items, item)
	}

	return items, nil
}

// sortGitReposList sorts the git repositories list based on sortBy and sortAsc settings
// This is a specialized version of sortFiles() for the gitReposList array
func (m *model) sortGitReposList() {
	if len(m.gitReposList) <= 1 {
		return
	}

	// Sort the repos based on sortBy criteria
	sort.Slice(m.gitReposList, func(i, j int) bool {
		a, b := m.gitReposList[i], m.gitReposList[j]

		// Determine sort result based on sortBy
		var less bool
		switch m.sortBy {
		case "name":
			less = strings.ToLower(a.name) < strings.ToLower(b.name)

		case "branch":
			// Sort by git branch name
			aBranch := a.gitBranch
			bBranch := b.gitBranch
			if aBranch == bBranch {
				// If same branch, sort by name as secondary
				less = strings.ToLower(a.name) < strings.ToLower(b.name)
			} else {
				less = strings.ToLower(aBranch) < strings.ToLower(bBranch)
			}

		case "status":
			// Sort by git status
			// Priority: dirty repos first, then ahead/behind status, then clean
			aStatus := getGitStatusSortValue(a)
			bStatus := getGitStatusSortValue(b)
			if aStatus == bStatus {
				// If same status, sort by name as secondary
				less = strings.ToLower(a.name) < strings.ToLower(b.name)
			} else {
				less = aStatus < bStatus
			}

		case "modified":
			// Sort by last commit time (stored in gitLastCommit)
			if a.gitLastCommit.Equal(b.gitLastCommit) {
				// If same time, sort by name as secondary
				less = strings.ToLower(a.name) < strings.ToLower(b.name)
			} else {
				less = a.gitLastCommit.Before(b.gitLastCommit)
			}

		default:
			// Fallback to name sorting
			less = strings.ToLower(a.name) < strings.ToLower(b.name)
		}

		// Apply sort direction (ascending vs descending)
		if !m.sortAsc {
			less = !less
		}

		return less
	})
}
