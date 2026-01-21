package issueview

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issuerow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newTestModelForAction(t *testing.T) Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config: &cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
	}

	m := NewModel(ctx)
	m.ctx = ctx
	m.issue = &issuerow.Issue{
		Ctx:  ctx,
		Data: data.IssueData{},
	}
	return m
}

func TestUpdateReturnsCorrectActions(t *testing.T) {
	testCases := []struct {
		name           string
		keyBinding     string
		expectedAction IssueActionType
	}{
		{"label key", "L", IssueActionLabel},
		// Note: IssueKeys.Assign has no default binding in issueKeys.go
		{"unassign key", "A", IssueActionUnassign},
		{"comment key", "c", IssueActionComment},
		{"close key", "x", IssueActionClose},
		{"reopen key", "X", IssueActionReopen},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModelForAction(t)
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.keyBinding)}

			_, _, action := m.Update(msg)

			require.NotNil(t, action, "expected action for key %q", tc.keyBinding)
			require.Equal(t, tc.expectedAction, action.Type,
				"expected action type %v for key %q, got %v", tc.expectedAction, tc.keyBinding, action.Type)
		})
	}
}

func TestUpdateReturnsNilActionForUnknownKeys(t *testing.T) {
	m := newTestModelForAction(t)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")}

	_, _, action := m.Update(msg)

	require.Nil(t, action, "expected nil action for unknown key")
}

func TestUpdateReturnsNilActionWhenCommenting(t *testing.T) {
	m := newTestModelForAction(t)
	m.isCommenting = true
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("L")} // label key

	_, _, action := m.Update(msg)

	require.Nil(t, action, "expected nil action when in commenting mode")
}

func TestUpdateReturnsNilActionWhenLabeling(t *testing.T) {
	m := newTestModelForAction(t)
	m.isLabeling = true
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("c")} // comment key

	_, _, action := m.Update(msg)

	require.Nil(t, action, "expected nil action when in labeling mode")
}

func TestUpdateReturnsNilActionWhenAssigning(t *testing.T) {
	m := newTestModelForAction(t)
	m.isAssigning = true
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")} // close key

	_, _, action := m.Update(msg)

	require.Nil(t, action, "expected nil action when in assigning mode")
}

func TestUpdateReturnsNilActionWhenUnassigning(t *testing.T) {
	m := newTestModelForAction(t)
	m.isUnassigning = true
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("X")} // reopen key

	_, _, action := m.Update(msg)

	require.Nil(t, action, "expected nil action when in unassigning mode")
}

func TestIssueActionTypes(t *testing.T) {
	// Verify all action types are distinct
	actionTypes := []IssueActionType{
		IssueActionNone,
		IssueActionLabel,
		IssueActionAssign,
		IssueActionUnassign,
		IssueActionComment,
		IssueActionClose,
		IssueActionReopen,
	}

	seen := make(map[IssueActionType]bool)
	for _, at := range actionTypes {
		require.False(t, seen[at], "duplicate action type value: %v", at)
		seen[at] = true
	}

	// Verify IssueActionNone is zero value
	require.Equal(t, IssueActionType(0), IssueActionNone, "IssueActionNone should be zero value")
}

func TestUpdateWithReboundKeys(t *testing.T) {
	// Save original key bindings
	originalLabelKeys := keys.IssueKeys.Label.Keys()

	// Rebind label key to "l" (lowercase)
	keys.IssueKeys.Label.SetKeys("l")
	defer func() {
		// Restore original bindings
		keys.IssueKeys.Label.SetKeys(originalLabelKeys...)
	}()

	m := newTestModelForAction(t)
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("l")}

	_, _, action := m.Update(msg)

	require.NotNil(t, action, "expected action for rebound key")
	require.Equal(t, IssueActionLabel, action.Type, "expected label action for rebound key")
}
