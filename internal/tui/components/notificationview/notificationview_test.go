package notificationview

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func TestSetPendingPRAction(t *testing.T) {
	tests := []struct {
		name           string
		action         string
		prNumber       int
		expectedAction string
		expectedPrompt string
	}{
		{
			name:           "close action",
			action:         "close",
			prNumber:       123,
			expectedAction: "pr_close",
			expectedPrompt: "Are you sure you want to close PR #123? (y/N)",
		},
		{
			name:           "reopen action",
			action:         "reopen",
			prNumber:       456,
			expectedAction: "pr_reopen",
			expectedPrompt: "Are you sure you want to reopen PR #456? (y/N)",
		},
		{
			name:           "ready action displays as mark as ready",
			action:         "ready",
			prNumber:       789,
			expectedAction: "pr_ready",
			expectedPrompt: "Are you sure you want to mark as ready PR #789? (y/N)",
		},
		{
			name:           "merge action",
			action:         "merge",
			prNumber:       100,
			expectedAction: "pr_merge",
			expectedPrompt: "Are you sure you want to merge PR #100? (y/N)",
		},
		{
			name:           "update action",
			action:         "update",
			prNumber:       200,
			expectedAction: "pr_update",
			expectedPrompt: "Are you sure you want to update PR #200? (y/N)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(&context.ProgramContext{})
			m.SetSubjectPR(&prrow.Data{
				Primary: &data.PullRequestData{Number: tt.prNumber},
			}, "notif-id")

			prompt := m.SetPendingPRAction(tt.action)

			require.Equal(t, tt.expectedAction, m.GetPendingAction())
			require.Equal(t, tt.expectedPrompt, prompt)
			require.True(t, m.HasPendingAction())
		})
	}
}

func TestSetPendingPRAction_NilSubject(t *testing.T) {
	m := NewModel(&context.ProgramContext{})

	prompt := m.SetPendingPRAction("close")

	require.Empty(t, prompt, "should return empty prompt when no PR subject")
	require.Empty(t, m.GetPendingAction(), "should not set pending action")
	require.False(t, m.HasPendingAction())
}

func TestSetPendingIssueAction(t *testing.T) {
	tests := []struct {
		name           string
		action         string
		issueNumber    int
		expectedAction string
		expectedPrompt string
	}{
		{
			name:           "close action",
			action:         "close",
			issueNumber:    123,
			expectedAction: "issue_close",
			expectedPrompt: "Are you sure you want to close Issue #123? (y/N)",
		},
		{
			name:           "reopen action",
			action:         "reopen",
			issueNumber:    456,
			expectedAction: "issue_reopen",
			expectedPrompt: "Are you sure you want to reopen Issue #456? (y/N)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewModel(&context.ProgramContext{})
			m.SetSubjectIssue(&data.IssueData{Number: tt.issueNumber}, "notif-id")

			prompt := m.SetPendingIssueAction(tt.action)

			require.Equal(t, tt.expectedAction, m.GetPendingAction())
			require.Equal(t, tt.expectedPrompt, prompt)
			require.True(t, m.HasPendingAction())
		})
	}
}

func TestSetPendingIssueAction_NilSubject(t *testing.T) {
	m := NewModel(&context.ProgramContext{})

	prompt := m.SetPendingIssueAction("close")

	require.Empty(t, prompt, "should return empty prompt when no Issue subject")
	require.Empty(t, m.GetPendingAction(), "should not set pending action")
	require.False(t, m.HasPendingAction())
}

func TestClearPendingAction(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	require.True(t, m.HasPendingAction(), "should have pending action before clear")

	m.ClearPendingAction()

	require.False(t, m.HasPendingAction(), "should not have pending action after clear")
	require.Empty(t, m.GetPendingAction())
}

func TestHasPendingAction(t *testing.T) {
	m := NewModel(&context.ProgramContext{})

	require.False(t, m.HasPendingAction(), "should be false initially")

	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("merge")

	require.True(t, m.HasPendingAction(), "should be true after setting action")

	m.ClearPendingAction()

	require.False(t, m.HasPendingAction(), "should be false after clearing")
}

// Update method tests

func TestUpdate_NoPendingAction(t *testing.T) {
	// When there's no pending action, Update should return early with no command
	m := NewModel(&context.ProgramContext{})
	callbackCalled := false
	m.SetOnConfirmAction(func(action string) tea.Cmd {
		callbackCalled = true
		return nil
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, cmd := m.Update(msg)

	require.False(t, callbackCalled, "callback should not be called when no pending action")
	require.Nil(t, cmd, "should return nil command")
	require.False(t, newModel.HasPendingAction())
}

func TestUpdate_ConfirmWithLowercaseY(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	var receivedAction string
	m.SetOnConfirmAction(func(action string) tea.Cmd {
		receivedAction = action
		return func() tea.Msg { return "test-cmd" }
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, cmd := m.Update(msg)

	require.Equal(t, "pr_close", receivedAction, "callback should receive the action")
	require.NotNil(t, cmd, "should return a command")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_ConfirmWithUppercaseY(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 456},
	}, "notif-id")
	m.SetPendingPRAction("merge")

	var receivedAction string
	m.SetOnConfirmAction(func(action string) tea.Cmd {
		receivedAction = action
		return func() tea.Msg { return "test-cmd" }
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")}
	newModel, cmd := m.Update(msg)

	require.Equal(t, "pr_merge", receivedAction, "callback should receive the action")
	require.NotNil(t, cmd, "should return a command")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_ConfirmWithEnter(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectIssue(&data.IssueData{Number: 789}, "notif-id")
	m.SetPendingIssueAction("reopen")

	var receivedAction string
	m.SetOnConfirmAction(func(action string) tea.Cmd {
		receivedAction = action
		return func() tea.Msg { return "test-cmd" }
	})

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)

	require.Equal(t, "issue_reopen", receivedAction, "callback should receive the action")
	require.NotNil(t, cmd, "should return a command")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_CancelWithN(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	callbackCalled := false
	m.SetOnConfirmAction(func(action string) tea.Cmd {
		callbackCalled = true
		return nil
	})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	newModel, cmd := m.Update(msg)

	require.False(t, callbackCalled, "callback should not be called on cancel")
	require.Nil(t, cmd, "should return nil command on cancel")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_CancelWithEscape(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("ready")

	callbackCalled := false
	m.SetOnConfirmAction(func(action string) tea.Cmd {
		callbackCalled = true
		return nil
	})

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, cmd := m.Update(msg)

	require.False(t, callbackCalled, "callback should not be called on escape")
	require.Nil(t, cmd, "should return nil command on escape")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_CancelWithRandomKey(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("update")

	callbackCalled := false
	m.SetOnConfirmAction(func(action string) tea.Cmd {
		callbackCalled = true
		return nil
	})

	// Press a random key like 'x'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	newModel, cmd := m.Update(msg)

	require.False(t, callbackCalled, "callback should not be called on random key")
	require.Nil(t, cmd, "should return nil command on random key")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_NilCallback(t *testing.T) {
	// When callback is nil, confirming should not panic
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")
	// Don't set callback - leave it nil

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, cmd := m.Update(msg)

	require.Nil(t, cmd, "should return nil command when callback is nil")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_NonKeyMsg(t *testing.T) {
	// Non-KeyMsg messages should be ignored
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	callbackCalled := false
	m.SetOnConfirmAction(func(action string) tea.Cmd {
		callbackCalled = true
		return nil
	})

	// Send a non-key message (e.g., WindowSizeMsg)
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, cmd := m.Update(msg)

	require.False(t, callbackCalled, "callback should not be called for non-key messages")
	require.Nil(t, cmd, "should return nil command for non-key messages")
	require.True(t, newModel.HasPendingAction(), "pending action should remain for non-key messages")
}

func TestUpdate_AllPRActions(t *testing.T) {
	// Test that all PR action types work correctly
	actions := []string{"close", "reopen", "ready", "merge", "update"}

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			m := NewModel(&context.ProgramContext{})
			m.SetSubjectPR(&prrow.Data{
				Primary: &data.PullRequestData{Number: 123},
			}, "notif-id")
			m.SetPendingPRAction(action)

			var receivedAction string
			m.SetOnConfirmAction(func(a string) tea.Cmd {
				receivedAction = a
				return func() tea.Msg { return "test" }
			})

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
			newModel, cmd := m.Update(msg)

			require.Equal(t, "pr_"+action, receivedAction)
			require.NotNil(t, cmd)
			require.False(t, newModel.HasPendingAction())
		})
	}
}

func TestUpdate_AllIssueActions(t *testing.T) {
	// Test that all Issue action types work correctly
	actions := []string{"close", "reopen"}

	for _, action := range actions {
		t.Run(action, func(t *testing.T) {
			m := NewModel(&context.ProgramContext{})
			m.SetSubjectIssue(&data.IssueData{Number: 456}, "notif-id")
			m.SetPendingIssueAction(action)

			var receivedAction string
			m.SetOnConfirmAction(func(a string) tea.Cmd {
				receivedAction = a
				return func() tea.Msg { return "test" }
			})

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			newModel, cmd := m.Update(msg)

			require.Equal(t, "issue_"+action, receivedAction)
			require.NotNil(t, cmd)
			require.False(t, newModel.HasPendingAction())
		})
	}
}

func TestSetOnConfirmAction(t *testing.T) {
	m := NewModel(&context.ProgramContext{})

	called := false
	callback := func(action string) tea.Cmd {
		called = true
		return nil
	}

	m.SetOnConfirmAction(callback)
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	m.Update(msg)

	require.True(t, called, "callback should be invoked after setting and confirming")
}
