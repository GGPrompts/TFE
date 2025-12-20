package main

// Module: git_operations.go
// Purpose: Git repository operations (pull, push, sync, fetch)
// Responsibilities:
// - Execute git commands in repository directories
// - Handle git errors and conflicts
// - Provide user feedback for git operations

import (
	"fmt"
	"os/exec"

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
