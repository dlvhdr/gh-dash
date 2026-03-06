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
		{"assign key", "a", IssueActionAssign},
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

// TestAutocompleteStateResetWhenSwitchingModes verifies that autocomplete
// suggestions from labeling mode don't leak into commenting mode.
// This is a regression test for https://github.com/dlvhdr/gh-dash/issues/751
func TestAutocompleteStateResetWhenSwitchingModes(t *testing.T) {
	m := newTestModelForAction(t)

	// Simulate what happens when entering labeling mode:
	// 1. Autocomplete gets populated with label suggestions
	m.ac.SetSuggestions([]string{"bug", "feature", "documentation", "enhancement"})

	// 2. User types something and autocomplete filters/shows suggestions
	// This populates the internal 'filtered' slice
	m.ac.Show("fea", nil) // Would match "feature"

	// Verify autocomplete has suggestions (the bug condition)
	require.True(t, m.ac.HasSuggestions(),
		"autocomplete should have suggestions after Show()")

	// 3. User exits labeling mode (Escape) - in the bug, this only hid but didn't reset
	m.ac.Hide()

	// In the buggy code, HasSuggestions() would still return true here
	// because Hide() only sets visible=false but doesn't clear filtered

	// 4. User enters commenting mode
	m.SetIsCommenting(true)

	// 5. After the fix, autocomplete state should be fully reset
	require.False(t, m.ac.HasSuggestions(),
		"autocomplete should have no suggestions after entering comment mode")
	require.False(t, m.ac.IsVisible(),
		"autocomplete should not be visible after entering comment mode")
}

// TestAutocompleteResetOnAssignMode verifies autocomplete is reset when entering assign mode
func TestAutocompleteResetOnAssignMode(t *testing.T) {
	m := newTestModelForAction(t)

	// Set up autocomplete with label suggestions
	m.ac.SetSuggestions([]string{"bug", "feature"})
	m.ac.Show("bug", nil)
	require.True(t, m.ac.HasSuggestions(), "precondition: autocomplete should have suggestions")

	// Enter assign mode
	m.SetIsAssigning(true)

	// Autocomplete should be reset
	require.False(t, m.ac.HasSuggestions(),
		"autocomplete should have no suggestions after entering assign mode")
}

// TestAutocompleteResetOnUnassignMode verifies autocomplete is reset when entering unassign mode
func TestAutocompleteResetOnUnassignMode(t *testing.T) {
	m := newTestModelForAction(t)

	// Set up autocomplete with label suggestions
	m.ac.SetSuggestions([]string{"bug", "feature"})
	m.ac.Show("bug", nil)
	require.True(t, m.ac.HasSuggestions(), "precondition: autocomplete should have suggestions")

	// Enter unassign mode
	m.SetIsUnassigning(true)

	// Autocomplete should be reset
	require.False(t, m.ac.HasSuggestions(),
		"autocomplete should have no suggestions after entering unassign mode")
}

// TestInputBoxTextNotReplacedByStaleAutocomplete is an end-to-end test that
// verifies the full bug scenario: typing in comment box after exiting label mode
// should not autocomplete with label names.
func TestInputBoxTextNotReplacedByStaleAutocomplete(t *testing.T) {
	m := newTestModelForAction(t)

	// Step 1: Simulate labeling mode with suggestions
	m.ac.SetSuggestions([]string{"bug", "feature", "documentation"})
	m.isLabeling = true
	m.ac.Show("fea", nil) // User typed "fea", matches "feature"

	// Step 2: Exit labeling mode (simulates pressing Escape)
	m.isLabeling = false
	m.ac.Hide()

	// Step 3: Enter commenting mode
	m.SetIsCommenting(true)

	// Step 4: Set some text in the input box (simulates user typing)
	m.inputBox.SetValue("This is my comment about fea")

	// Step 5: Verify that Tab key does NOT trigger autocomplete selection
	// because autocomplete was reset when entering comment mode
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	m.inputBox.Update(tabMsg)

	// The text should remain unchanged (not replaced with "feature")
	require.Equal(t, "This is my comment about fea", m.inputBox.Value(),
		"input text should not be modified by Tab when autocomplete is reset")
}
