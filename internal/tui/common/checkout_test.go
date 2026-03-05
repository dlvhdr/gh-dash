package common_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
)

func noopStartTask(_ common.CheckoutTask) tea.Cmd {
	return nil
}

func TestCheckoutPR_RegularFlow(t *testing.T) {
	tests := []struct {
		name      string
		params    common.CheckoutParams
		wantErr   bool
		wantNil   bool
		errSubstr string
	}{
		{
			name: "returns error when no path configured",
			params: common.CheckoutParams{
				RepoPaths:     map[string]string{},
				WorktreePaths: map[string]string{},
				StartTask:     noopStartTask,
				PRNumber:      123,
				RepoName:      "owner/repo",
				BranchName:    "fix/bug",
			},
			wantErr:   true,
			wantNil:   true,
			errSubstr: "repoPaths or worktreePaths",
		},
		{
			name: "returns cmd when repoPaths configured",
			params: common.CheckoutParams{
				RepoPaths:     map[string]string{"owner/repo": "/path/to/repo"},
				WorktreePaths: map[string]string{},
				StartTask:     noopStartTask,
				PRNumber:      123,
				RepoName:      "owner/repo",
				BranchName:    "fix/bug",
			},
			wantErr: false,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := common.CheckoutPR(tt.params)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errSubstr)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil && cmd != nil {
				t.Error("expected nil cmd")
			}
			if !tt.wantNil && cmd == nil {
				t.Error("expected non-nil cmd")
			}
		})
	}
}

// executeBatchCmd calls the cmd and handles both BatchMsg (when startTask returns
// a non-nil cmd) and direct TaskFinishedMsg (when startTask returns nil).
func executeBatchCmd(t *testing.T, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	msg := cmd()
	switch m := msg.(type) {
	case tea.BatchMsg:
		for _, c := range m {
			if c != nil {
				c()
			}
		}
	default:
		// Single cmd case (Batch optimized away nil startCmd)
		_ = m
	}
}

// makeRepoDir creates a temp directory with a .git subdirectory to simulate a git repo.
func makeRepoDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestCheckoutPR_WorktreeFlow(t *testing.T) {
	t.Run("creates worktree and checks out PR", func(t *testing.T) {
		origRunCmd := common.RunCommandFunc
		defer func() { common.RunCommandFunc = origRunCmd }()

		repoDir := makeRepoDir(t)
		wtParent := t.TempDir()

		var commands []string
		common.RunCommandFunc = func(name string, args []string, dir string) error {
			commands = append(commands, name+" "+strings.Join(args, " ")+" ["+dir+"]")
			return nil
		}

		cmd, err := common.CheckoutPR(common.CheckoutParams{
			RepoPaths:     map[string]string{"owner/repo": repoDir},
			WorktreePaths: map[string]string{"owner/repo": wtParent},
			StartTask:     noopStartTask,
			PRNumber:      42,
			RepoName:      "owner/repo",
			BranchName:    "feat/cool",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cmd == nil {
			t.Fatal("expected non-nil cmd")
		}

		executeBatchCmd(t, cmd)

		expectedWtPath := filepath.Join(wtParent, "feat/cool")
		if len(commands) < 2 {
			t.Fatalf("expected at least 2 commands, got %d: %v", len(commands), commands)
		}
		// git worktree add should run from the repo root
		if !strings.Contains(commands[0], "git worktree add --detach "+expectedWtPath+" HEAD") {
			t.Errorf("first command should be git worktree add, got: %s", commands[0])
		}
		if !strings.Contains(commands[0], "["+repoDir+"]") {
			t.Errorf("git worktree add should run in repo root %s, got: %s", repoDir, commands[0])
		}
		// gh pr checkout should run from the new worktree path
		if !strings.Contains(commands[1], "gh pr checkout 42") {
			t.Errorf("second command should be gh pr checkout, got: %s", commands[1])
		}
		if !strings.Contains(commands[1], "["+expectedWtPath+"]") {
			t.Errorf("gh pr checkout should run in worktree %s, got: %s", expectedWtPath, commands[1])
		}
	})

	t.Run("skips creation when worktree already exists", func(t *testing.T) {
		origRunCmd := common.RunCommandFunc
		defer func() { common.RunCommandFunc = origRunCmd }()

		repoDir := makeRepoDir(t)
		wtParent := t.TempDir()
		// Pre-create the worktree directory
		wtPath := filepath.Join(wtParent, "feat/existing")
		if err := os.MkdirAll(wtPath, 0o755); err != nil {
			t.Fatal(err)
		}

		var commands []string
		common.RunCommandFunc = func(name string, args []string, dir string) error {
			commands = append(commands, name+" "+strings.Join(args, " "))
			return nil
		}

		cmd, err := common.CheckoutPR(common.CheckoutParams{
			RepoPaths:     map[string]string{"owner/repo": repoDir},
			WorktreePaths: map[string]string{"owner/repo": wtParent},
			StartTask:     noopStartTask,
			PRNumber:      42,
			RepoName:      "owner/repo",
			BranchName:    "feat/existing",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		executeBatchCmd(t, cmd)

		for _, c := range commands {
			if strings.Contains(c, "worktree add") {
				t.Errorf("should not run git worktree add, but got: %s", c)
			}
		}
	})

	t.Run("errors when repo root is not a git repo", func(t *testing.T) {
		repoDir := t.TempDir() // No .git directory
		wtParent := t.TempDir()

		_, err := common.CheckoutPR(common.CheckoutParams{
			RepoPaths:     map[string]string{"owner/repo": repoDir},
			WorktreePaths: map[string]string{"owner/repo": wtParent},
			StartTask:     noopStartTask,
			PRNumber:      42,
			RepoName:      "owner/repo",
			BranchName:    "feat/cool",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "clone it first") {
			t.Errorf("error %q should contain %q", err.Error(), "clone it first")
		}
	})

	t.Run("errors when worktreePaths set but repoPaths missing", func(t *testing.T) {
		wtParent := t.TempDir()

		_, err := common.CheckoutPR(common.CheckoutParams{
			RepoPaths:     map[string]string{},
			WorktreePaths: map[string]string{"owner/repo": wtParent},
			StartTask:     noopStartTask,
			PRNumber:      42,
			RepoName:      "owner/repo",
			BranchName:    "feat/cool",
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "repoPaths is not") {
			t.Errorf("error %q should mention repoPaths requirement", err.Error())
		}
	})

	t.Run("cleans up worktree on gh pr checkout failure", func(t *testing.T) {
		origRunCmd := common.RunCommandFunc
		defer func() { common.RunCommandFunc = origRunCmd }()

		repoDir := makeRepoDir(t)
		wtParent := t.TempDir()

		var commands []string
		common.RunCommandFunc = func(name string, args []string, dir string) error {
			cmd := name + " " + strings.Join(args, " ")
			commands = append(commands, cmd)
			// Fail the gh pr checkout command
			if name == "gh" && len(args) > 0 && args[0] == "pr" {
				return fmt.Errorf("checkout failed")
			}
			return nil
		}

		cmd, err := common.CheckoutPR(common.CheckoutParams{
			RepoPaths:     map[string]string{"owner/repo": repoDir},
			WorktreePaths: map[string]string{"owner/repo": wtParent},
			StartTask:     noopStartTask,
			PRNumber:      42,
			RepoName:      "owner/repo",
			BranchName:    "feat/broken",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		executeBatchCmd(t, cmd)

		foundRemove := false
		for _, c := range commands {
			if strings.Contains(c, "worktree remove") {
				foundRemove = true
				break
			}
		}
		if !foundRemove {
			t.Errorf("expected git worktree remove to be called, commands: %v", commands)
		}
	})

	t.Run("worktreePaths takes priority over repoPaths", func(t *testing.T) {
		origRunCmd := common.RunCommandFunc
		defer func() { common.RunCommandFunc = origRunCmd }()

		repoDir := makeRepoDir(t)
		wtParent := t.TempDir()

		var commands []string
		common.RunCommandFunc = func(name string, args []string, dir string) error {
			commands = append(commands, name+" "+strings.Join(args, " "))
			return nil
		}

		cmd, err := common.CheckoutPR(common.CheckoutParams{
			RepoPaths:     map[string]string{"owner/repo": repoDir},
			WorktreePaths: map[string]string{"owner/repo": wtParent},
			StartTask:     noopStartTask,
			PRNumber:      42,
			RepoName:      "owner/repo",
			BranchName:    "feat/priority",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cmd == nil {
			t.Fatal("expected non-nil cmd")
		}

		executeBatchCmd(t, cmd)

		if len(commands) == 0 {
			t.Fatal("expected commands to be run")
		}
		// Should use worktree flow (git worktree add), not regular flow (gh pr checkout directly)
		if !strings.Contains(commands[0], "git worktree add") {
			t.Errorf("expected worktree flow, got: %s", commands[0])
		}
	})

	t.Run("falls back to pr-N when branch name empty", func(t *testing.T) {
		origRunCmd := common.RunCommandFunc
		defer func() { common.RunCommandFunc = origRunCmd }()

		repoDir := makeRepoDir(t)
		wtParent := t.TempDir()

		var commands []string
		common.RunCommandFunc = func(name string, args []string, dir string) error {
			commands = append(commands, name+" "+strings.Join(args, " ")+" ["+dir+"]")
			return nil
		}

		cmd, err := common.CheckoutPR(common.CheckoutParams{
			RepoPaths:     map[string]string{"owner/repo": repoDir},
			WorktreePaths: map[string]string{"owner/repo": wtParent},
			StartTask:     noopStartTask,
			PRNumber:      42,
			RepoName:      "owner/repo",
			BranchName:    "",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		executeBatchCmd(t, cmd)

		expectedWtPath := filepath.Join(wtParent, "pr-42")
		if len(commands) < 1 {
			t.Fatal("expected at least 1 command")
		}
		if !strings.Contains(commands[0], expectedWtPath) {
			t.Errorf("expected path containing %s, got: %s", expectedWtPath, commands[0])
		}
	})
}
