package notificationssection

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// noopStartTask is a stub that returns nil for testing
func noopStartTask(task context.Task) tea.Cmd {
	return nil
}

func TestCheckoutPR(t *testing.T) {
	tests := []struct {
		name      string
		prNumber  int
		repoName  string
		repoPaths map[string]string
		wantErr   bool
		wantNil   bool
	}{
		{
			name:      "returns error when repo path not configured",
			prNumber:  123,
			repoName:  "owner/repo",
			repoPaths: map[string]string{},
			wantErr:   true,
			wantNil:   true,
		},
		{
			name:     "returns command when repo path is configured",
			prNumber: 123,
			repoName: "owner/repo",
			repoPaths: map[string]string{
				"owner/repo": "/path/to/repo",
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:     "returns command with tilde path",
			prNumber: 456,
			repoName: "my-org/my-repo",
			repoPaths: map[string]string{
				"my-org/my-repo": "~/projects/my-repo",
			},
			wantErr: false,
			wantNil: false,
		},
		{
			name:      "returns error for unconfigured repo even with other repos configured",
			prNumber:  789,
			repoName:  "other/repo",
			repoPaths: map[string]string{"owner/repo": "/path/to/repo"},
			wantErr:   true,
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProgramContext{
				Config: &config.Config{
					RepoPaths: tt.repoPaths,
				},
				StartTask: noopStartTask,
			}

			cmd, err := CheckoutPR(ctx, tt.prNumber, tt.repoName)

			if tt.wantErr && err == nil {
				t.Errorf("CheckoutPR() error = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("CheckoutPR() error = %v, want nil", err)
			}
			if tt.wantNil && cmd != nil {
				t.Errorf("CheckoutPR() returned non-nil cmd, want nil")
			}
			if !tt.wantNil && cmd == nil {
				t.Errorf("CheckoutPR() returned nil cmd, want non-nil")
			}
		})
	}
}

func TestCheckoutPRErrorMessage(t *testing.T) {
	ctx := &context.ProgramContext{
		Config: &config.Config{
			RepoPaths: map[string]string{},
		},
		StartTask: noopStartTask,
	}

	_, err := CheckoutPR(ctx, 123, "owner/repo")

	if err == nil {
		t.Fatal("CheckoutPR() expected error, got nil")
	}

	expectedMsg := "local path to repo not specified, set one in your config.yml under repoPaths"
	if err.Error() != expectedMsg {
		t.Errorf("CheckoutPR() error = %q, want %q", err.Error(), expectedMsg)
	}
}
