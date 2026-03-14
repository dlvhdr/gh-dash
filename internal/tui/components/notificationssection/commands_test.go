package notificationssection

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationrow"
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

// TestMarkAsDoneStoresCorrectTimestamp is a regression test for a
// pointer-aliasing bug that occurred in markAsDone().
//
// GetCurrNotification() returns &m.Notifications[idx], a pointer into the
// Notifications slice. When the closure later dereferences this pointer to
// read UpdatedAt, the slice may have been modified (element removed via
// append), causing the pointer to reference a different notification's data.
//
// The fix captures UpdatedAt by value before the closure. This test verifies
// that the correct timestamp reaches the DoneStore even when the slice is
// modified between command creation and execution.
func TestMarkAsDoneStoresCorrectTimestamp(t *testing.T) {
	// Mock the API call to succeed without network access.
	origFunc := markNotificationDoneFunc
	markNotificationDoneFunc = func(string) error { return nil }
	defer func() { markNotificationDoneFunc = origFunc }()

	// Set up a DoneStore backed by a temp file so we don't touch real state.
	tempDir, err := os.MkdirTemp("", "gh-dash-markdone-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store := data.NewDoneStoreForTesting(filepath.Join(tempDir, "done.json"))
	restoreStore := data.OverrideDoneStoreForTesting(store)
	defer restoreStore()

	t1 := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 2, 20, 15, 30, 0, 0, time.UTC)
	t3 := time.Date(2026, 3, 1, 8, 0, 0, 0, time.UTC)

	// Build a minimal Model. A zero-value Table has cursor at index 0,
	// so GetCurrNotification() returns &m.Notifications[0].
	m := Model{
		Notifications: []notificationrow.Data{
			{Notification: data.NotificationData{Id: "notif-A", UpdatedAt: t1}},
			{Notification: data.NotificationData{Id: "notif-B", UpdatedAt: t2}},
			{Notification: data.NotificationData{Id: "notif-C", UpdatedAt: t3}},
		},
		sessionMarkedDone: make(map[string]bool),
		sessionMarkedRead: make(map[string]bool),
	}
	m.Ctx = &context.ProgramContext{
		StartTask: noopStartTask,
	}

	// Step 1: Call markAsDone(). This captures notif-A's ID and UpdatedAt
	// by value, before the closure.
	cmd := m.markAsDone()
	if cmd == nil {
		t.Fatal("markAsDone() returned nil cmd")
	}

	// Step 2: Simulate the race — remove notif-A from the slice.
	// This shifts notif-B into position 0 and notif-C into position 1.
	// If the closure had captured a pointer instead of a value, it would
	// now read notif-B's UpdatedAt (t2) instead of notif-A's (t1).
	m.Notifications = append(m.Notifications[:0], m.Notifications[1:]...)

	// Step 3: Execute the command. tea.Batch returns a BatchMsg containing
	// the inner cmds; execute each one.
	batchMsg := cmd()
	if cmds, ok := batchMsg.(tea.BatchMsg); ok {
		for _, c := range cmds {
			if c != nil {
				c()
			}
		}
	}

	// Step 4: Verify the DoneStore received notif-A's original timestamp (t1),
	// not notif-B's (t2), which is what the shifted pointer would have read.
	if !store.IsDone("notif-A", t1) {
		t.Error("DoneStore should have notif-A marked done at t1")
	}
	// If the bug were present, t2 would have been stored instead of t1.
	// In that case, IsDone("notif-A", t1) would still return true (t1 <= t2),
	// but IsDone with a time between t1 and t2 would incorrectly return true.
	// Use a more precise check: mark done at t1 means t1+1s should resurface
	// only if the stored timestamp is exactly t1.
	justAfterT1 := t1.Add(1 * time.Second)
	if store.IsDone("notif-A", justAfterT1) {
		t.Error(
			"notif-A should resurface for activity after t1 (stored timestamp should be exactly t1)",
		)
	}
	// That is the critical assertion: if the pointer-aliasing bug were present,
	// t2 would be stored, and activity at justAfterT1 would _not_ resurface
	// (because justAfterT1 < t2). The test would fail here.
}
