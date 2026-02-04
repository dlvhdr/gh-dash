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
	// When there's no pending action, Update should return early with no action
	m := NewModel(&context.ProgramContext{})

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, action := m.Update(msg)

	require.Empty(t, action, "should return empty action when no pending action")
	require.False(t, newModel.HasPendingAction())
}

func TestUpdate_ConfirmWithLowercaseY(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, action := m.Update(msg)

	require.Equal(t, "pr_close", action, "should return the confirmed action")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_ConfirmWithUppercaseY(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 456},
	}, "notif-id")
	m.SetPendingPRAction("merge")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")}
	newModel, action := m.Update(msg)

	require.Equal(t, "pr_merge", action, "should return the confirmed action")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_ConfirmWithEnter(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectIssue(&data.IssueData{Number: 789}, "notif-id")
	m.SetPendingIssueAction("reopen")

	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, action := m.Update(msg)

	require.Equal(t, "issue_reopen", action, "should return the confirmed action")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_CancelWithN(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	newModel, action := m.Update(msg)

	require.Empty(t, action, "should return empty action on cancel")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_CancelWithEscape(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("ready")

	msg := tea.KeyMsg{Type: tea.KeyEsc}
	newModel, action := m.Update(msg)

	require.Empty(t, action, "should return empty action on escape")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_CancelWithRandomKey(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("update")

	// Press a random key like 'x'
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}
	newModel, action := m.Update(msg)

	require.Empty(t, action, "should return empty action on random key")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_ConfirmReturnsAction(t *testing.T) {
	// Confirming should return the action string
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, action := m.Update(msg)

	require.Equal(t, "pr_close", action, "should return the confirmed action")
	require.False(t, newModel.HasPendingAction(), "pending action should be cleared")
}

func TestUpdate_NonKeyMsg(t *testing.T) {
	// Non-KeyMsg messages should be ignored
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	// Send a non-key message (e.g., WindowSizeMsg)
	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	newModel, action := m.Update(msg)

	require.Empty(t, action, "should return empty action for non-key messages")
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

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
			newModel, confirmedAction := m.Update(msg)

			require.Equal(t, "pr_"+action, confirmedAction)
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

			msg := tea.KeyMsg{Type: tea.KeyEnter}
			newModel, confirmedAction := m.Update(msg)

			require.Equal(t, "issue_"+action, confirmedAction)
			require.False(t, newModel.HasPendingAction())
		})
	}
}

func TestUpdate_ReturnsActionOnConfirm(t *testing.T) {
	m := NewModel(&context.ProgramContext{})
	m.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{Number: 123},
	}, "notif-id")
	m.SetPendingPRAction("close")

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	_, action := m.Update(msg)

	require.Equal(t, "pr_close", action, "should return the action on confirm")
}
