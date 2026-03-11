package issueview

import (
	"testing"

	tea "charm.land/bubbletea/v2"
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
		{"assign key", "a", IssueActionAssign},
		{"unassign key", "A", IssueActionUnassign},
		{"comment key", "c", IssueActionComment},
		{"checkout key", "C", IssueActionCheckout},
		{"close key", "x", IssueActionClose},
		{"reopen key", "X", IssueActionReopen},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := newTestModelForAction(t)
			msg := tea.KeyPressMsg{Text: tc.keyBinding}

			_, _, action := m.Update(msg)

			require.NotNil(t, action, "expected action for key %q", tc.keyBinding)
			require.Equal(
				t,
				tc.expectedAction,
				action.Type,
				"expected action type %v for key %q, got %v",
				tc.expectedAction,
				tc.keyBinding,
				action.Type,
			)
		})
	}
}

func TestUpdateReturnsNilActionForUnknownKeys(t *testing.T) {
	m := newTestModelForAction(t)
	msg := tea.KeyPressMsg{Text: "z"}

	_, _, action := m.Update(msg)

	require.Nil(t, action, "expected nil action for unknown key")
}

func TestUpdateReturnsNilActionWhenCommenting(t *testing.T) {
	m := newTestModelForAction(t)
	cmd := m.SetIsCommenting(true)

	require.NotNil(t, cmd)
	msg := tea.KeyPressMsg{Text: "L"} // label key

	_, _, action := m.Update(msg)

	require.Nil(t, action, "expected nil action when in commenting mode")
}

func TestUpdateReturnsNilActionWhenLabeling(t *testing.T) {
	m := newTestModelForAction(t)
	cmd := m.SetIsLabeling(true)

	require.NotNil(t, cmd)
	msg := tea.KeyPressMsg{Text: "c"} // comment key

	_, _, action := m.Update(msg)

	require.Nil(t, action, "expected nil action when in labeling mode")
}

func TestUpdateReturnsNilActionWhenAssigning(t *testing.T) {
	m := newTestModelForAction(t)
	cmd := m.SetIsAssigning(true)

	require.NotNil(t, cmd)
	msg := tea.KeyPressMsg{Text: "x"} // close key

	_, _, action := m.Update(msg)

	require.Nil(t, action, "expected nil action when in assigning mode")
}

func TestUpdateReturnsNilActionWhenUnassigning(t *testing.T) {
	m := newTestModelForAction(t)
	cmd := m.SetIsUnassigning(true)

	require.NotNil(t, cmd)
	msg := tea.KeyPressMsg{Text: "X"} // reopen key

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
		IssueActionCheckout,
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
	msg := tea.KeyPressMsg{Text: "l"}

	_, _, action := m.Update(msg)

	require.NotNil(t, action, "expected action for rebound key")
	require.Equal(t, IssueActionLabel, action.Type, "expected label action for rebound key")
}

func TestSetIsCommentingActivatesFocusedInput(t *testing.T) {
	m := newTestModelForAction(t)
	cmd := m.SetIsCommenting(true)

	require.NotNil(t, cmd)
	require.True(t, m.IsTextInputBoxFocused())
	require.True(t, m.GetIsCommenting())
}

func TestSetIsAssigningReturnsCommandOnCacheMiss(t *testing.T) {
	m := newTestModelForAction(t)
	cmd := m.SetIsAssigning(true)

	require.NotNil(t, cmd)
	require.True(t, m.IsTextInputBoxFocused())
	require.True(t, m.GetIsAssigning())
}

func TestSetIsUnassigningActivatesFocusedInput(t *testing.T) {
	m := newTestModelForAction(t)
	cmd := m.SetIsUnassigning(true)

	require.NotNil(t, cmd)
	require.True(t, m.IsTextInputBoxFocused())
	require.True(t, m.GetIsUnassigning())
}

func TestSetIsLabelingActivatesFocusedInput(t *testing.T) {
	m := newTestModelForAction(t)
	cmd := m.SetIsLabeling(true)

	require.NotNil(t, cmd)
	require.True(t, m.IsTextInputBoxFocused())
}
