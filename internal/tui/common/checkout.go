package common

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

// CheckoutTask mirrors the fields needed from context.Task to avoid
// an import cycle (context already imports common).
type CheckoutTask struct {
	Id           string
	StartText    string
	FinishedText string
}

// CheckoutParams holds the parameters for a PR checkout.
type CheckoutParams struct {
	RepoPaths     map[string]string
	WorktreePaths map[string]string
	StartTask     func(task CheckoutTask) tea.Cmd
	PRNumber      int
	RepoName      string
	BranchName    string
}

// RunCommandFunc is the function used to execute shell commands.
// It is a variable so tests can override it.
var RunCommandFunc = func(name string, args []string, dir string) error {
	c := exec.Command(name, args...)
	c.Dir = dir
	var stderr bytes.Buffer
	c.Stderr = &stderr
	if err := c.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("%w: %s", err, msg)
		}
		return err
	}
	return nil
}

// ExpandTilde replaces a leading ~ with the user's home directory.
func ExpandTilde(path string) string {
	if strings.HasPrefix(path, "~") {
		userHomeDir, _ := os.UserHomeDir()
		return strings.Replace(path, "~", userHomeDir, 1)
	}
	return path
}

// CheckoutPR checks out a PR using either worktree or regular checkout,
// depending on config.
func CheckoutPR(params CheckoutParams) (tea.Cmd, error) {
	// Check worktreePaths first (priority over repoPaths)
	if wtParent, ok := GetRepoLocalPath(params.RepoName, params.WorktreePaths); ok {
		// worktreePaths requires repoPaths to also be configured (for the repo root)
		repoRoot, repoOk := GetRepoLocalPath(params.RepoName, params.RepoPaths)
		if !repoOk {
			return nil, errors.New(
				"worktreePaths is configured but repoPaths is not; both are required for worktree checkout",
			)
		}
		return checkoutWorktree(params, repoRoot, wtParent)
	}

	// Fall back to regular checkout via repoPaths
	if repoPath, ok := GetRepoLocalPath(params.RepoName, params.RepoPaths); ok {
		return checkoutRegular(params, repoPath)
	}

	return nil, errors.New(
		"local path to repo not specified, set one in your config.yml under repoPaths or worktreePaths",
	)
}

func checkoutRegular(params CheckoutParams, repoPath string) (tea.Cmd, error) {
	taskId := fmt.Sprintf("checkout_%d", params.PRNumber)
	task := CheckoutTask{
		Id:           taskId,
		StartText:    fmt.Sprintf("Checking out PR #%d", params.PRNumber),
		FinishedText: fmt.Sprintf("PR #%d has been checked out at %s", params.PRNumber, repoPath),
	}
	startCmd := params.StartTask(task)
	return tea.Batch(startCmd, func() tea.Msg {
		expandedPath := ExpandTilde(repoPath)
		err := RunCommandFunc(
			"gh",
			[]string{"pr", "checkout", fmt.Sprint(params.PRNumber)},
			expandedPath,
		)
		return constants.TaskFinishedMsg{TaskId: taskId, Err: err}
	}), nil
}

func checkoutWorktree(params CheckoutParams, repoRoot string, wtParent string) (tea.Cmd, error) {
	expandedRoot := ExpandTilde(repoRoot)
	expandedParent := ExpandTilde(wtParent)

	// Verify the repo root is a git repo
	gitDir := filepath.Join(expandedRoot, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, fmt.Errorf(
			"%s is not a git repository, clone it first", expandedRoot)
	}

	// Determine worktree directory name
	dirName := params.BranchName
	if dirName == "" {
		dirName = fmt.Sprintf("pr-%d", params.PRNumber)
	}
	wtPath := filepath.Join(expandedParent, dirName)

	taskId := fmt.Sprintf("checkout_worktree_%d", params.PRNumber)
	task := CheckoutTask{
		Id:        taskId,
		StartText: fmt.Sprintf("Checking out PR #%d into worktree", params.PRNumber),
		FinishedText: fmt.Sprintf(
			"PR #%d checked out into worktree (path copied to clipboard)",
			params.PRNumber,
		),
	}
	startCmd := params.StartTask(task)

	return tea.Batch(startCmd, func() tea.Msg {
		// If worktree already exists, skip creation
		if _, err := os.Stat(wtPath); err == nil {
			_ = clipboard.WriteAll(wtPath)
			return constants.TaskFinishedMsg{TaskId: taskId, Err: nil}
		}

		// Ensure the worktree parent directory exists
		if err := os.MkdirAll(expandedParent, 0o755); err != nil {
			return constants.TaskFinishedMsg{
				TaskId: taskId,
				Err:    fmt.Errorf("create worktree directory: %w", err),
			}
		}

		// Create a detached worktree
		err := RunCommandFunc(
			"git",
			[]string{"worktree", "add", "--detach", wtPath, "HEAD"},
			expandedRoot,
		)
		if err != nil {
			return constants.TaskFinishedMsg{
				TaskId: taskId,
				Err:    fmt.Errorf("git worktree add: %w", err),
			}
		}

		// Check out the PR inside the new worktree
		err = RunCommandFunc("gh", []string{"pr", "checkout", fmt.Sprint(params.PRNumber)}, wtPath)
		if err != nil {
			// Clean up the worktree on failure
			_ = RunCommandFunc("git", []string{"worktree", "remove", wtPath}, expandedRoot)
			return constants.TaskFinishedMsg{
				TaskId: taskId,
				Err:    fmt.Errorf("gh pr checkout: %w", err),
			}
		}

		// Copy path to clipboard (ignore errors)
		_ = clipboard.WriteAll(wtPath)

		return constants.TaskFinishedMsg{TaskId: taskId, Err: nil}
	}), nil
}
