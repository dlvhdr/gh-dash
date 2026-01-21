package prview

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
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
	m.pr = &prrow.PullRequest{
		Ctx: ctx,
		Data: &prrow.Data{
			Primary:    &data.PullRequestData{},
			IsEnriched: true,
		},
	}
	return m
}

func TestMsgToActionReturnsCorrectActions(t *testing.T) {
	testCases := []struct {
		name           string
		keyBinding     string
		expectedAction PRActionType
	}{
		{"approve key", "v", PRActionApprove},
		{"assign key", "a", PRActionAssign},
		{"unassign key", "A", PRActionUnassign},
		{"comment key", "c", PRActionComment},
		{"diff key", "d", PRActionDiff},
		{"checkout key C", "C", PRActionCheckout},
		{"checkout key space", " ", PRActionCheckout},
		{"close key", "x", PRActionClose},
		{"ready key", "W", PRActionReady},
		{"reopen key", "X", PRActionReopen},
		{"merge key", "m", PRActionMerge},
		{"update key", "u", PRActionUpdate},
		{"summary view more key", "e", PRActionSummaryViewMore},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tc.keyBinding)}

			action := MsgToAction(msg)

			require.NotNil(t, action, "expected action for key %q", tc.keyBinding)
			require.Equal(t, tc.expectedAction, action.Type,
				"expected action type %v for key %q, got %v", tc.expectedAction, tc.keyBinding, action.Type)
		})
	}
}

func TestMsgToActionReturnsNilForUnknownKeys(t *testing.T) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")}

	action := MsgToAction(msg)

	require.Nil(t, action, "expected nil action for unknown key")
}

func TestIsTextInputBoxFocusedWhenCommenting(t *testing.T) {
	m := newTestModelForAction(t)
	m.isCommenting = true

	require.True(t, m.IsTextInputBoxFocused(), "expected text input box focused when in commenting mode")
}

func TestIsTextInputBoxFocusedWhenApproving(t *testing.T) {
	m := newTestModelForAction(t)
	m.isApproving = true

	require.True(t, m.IsTextInputBoxFocused(), "expected text input box focused when in approving mode")
}

func TestIsTextInputBoxFocusedWhenAssigning(t *testing.T) {
	m := newTestModelForAction(t)
	m.isAssigning = true

	require.True(t, m.IsTextInputBoxFocused(), "expected text input box focused when in assigning mode")
}

func TestIsTextInputBoxFocusedWhenUnassigning(t *testing.T) {
	m := newTestModelForAction(t)
	m.isUnassigning = true

	require.True(t, m.IsTextInputBoxFocused(), "expected text input box focused when in unassigning mode")
}

func TestUpdateHandlesSidebarTabNavigation(t *testing.T) {
	t.Run("prev sidebar tab", func(t *testing.T) {
		m := newTestModelForAction(t)
		// Move to a non-first tab first
		m.carousel.MoveRight()
		initialTab := m.carousel.SelectedItem()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")}
		m, _ = m.Update(msg)

		require.NotEqual(t, initialTab, m.carousel.SelectedItem(),
			"carousel should have moved to previous tab")
	})

	t.Run("next sidebar tab", func(t *testing.T) {
		m := newTestModelForAction(t)
		initialTab := m.carousel.SelectedItem()

		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")}
		m, _ = m.Update(msg)

		require.NotEqual(t, initialTab, m.carousel.SelectedItem(),
			"carousel should have moved to next tab")
	})
}

func TestPRActionTypes(t *testing.T) {
	// Verify all action types are distinct
	actionTypes := []PRActionType{
		PRActionNone,
		PRActionApprove,
		PRActionAssign,
		PRActionUnassign,
		PRActionComment,
		PRActionDiff,
		PRActionCheckout,
		PRActionClose,
		PRActionReady,
		PRActionReopen,
		PRActionMerge,
		PRActionUpdate,
		PRActionSummaryViewMore,
	}

	seen := make(map[PRActionType]bool)
	for _, at := range actionTypes {
		require.False(t, seen[at], "duplicate action type value: %v", at)
		seen[at] = true
	}

	// Verify PRActionNone is zero value
	require.Equal(t, PRActionType(0), PRActionNone, "PRActionNone should be zero value")
}

func TestMsgToActionWithReboundKeys(t *testing.T) {
	// Save original key bindings
	originalApproveKeys := keys.PRKeys.Approve.Keys()

	// Rebind approve key to "V" (uppercase)
	keys.PRKeys.Approve.SetKeys("V")
	defer func() {
		// Restore original bindings
		keys.PRKeys.Approve.SetKeys(originalApproveKeys...)
	}()

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("V")}

	action := MsgToAction(msg)

	require.NotNil(t, action, "expected action for rebound key")
	require.Equal(t, PRActionApprove, action.Type, "expected approve action for rebound key")
}
