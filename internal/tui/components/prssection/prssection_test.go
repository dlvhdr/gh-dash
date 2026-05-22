package prssection

import (
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prompt"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tasks"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

// newTestModel creates a minimal Model with the prompt confirmation box
// focused and a single PR row so that GetCurrRow returns non-nil.
func newTestModel(action string) Model {
	ctx := &context.ProgramContext{
		StartTask: func(task context.Task) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}
	m := Model{
		BaseModel: section.BaseModel{
			Ctx:                       ctx,
			IsPromptConfirmationShown: true,
			PromptConfirmationAction:  action,
			PromptConfirmationBox:     prompt.NewModel(ctx),
		},
		Prs: []prrow.Data{
			{Primary: &data.PullRequestData{Number: 42}},
		},
	}
	m.PromptConfirmationBox.Focus()
	return m
}

// newFullTestModel creates a Model via the real NewModel constructor (so that
// all internal sub-models — Table, SearchBar, etc. — are properly wired) and
// then injects a single PR row with the given PR number.
//
// ctx.Config, ctx.Theme, and ctx.Styles are all populated so that
// m.Update() can invoke BuildRows() end-to-end without panicking.
func newFullTestModel(prNumber int) Model {
	thm := *theme.DefaultTheme
	s := context.InitStyles(thm)
	cfg := &config.Config{
		// A non-nil ThemeConfig is required because table.NewModel and
		// prrow.ToTableRow both dereference Config.Theme.Ui.Table fields.
		Theme: &config.ThemeConfig{},
	}

	ctx := &context.ProgramContext{
		Config: cfg,
		Theme:  thm,
		Styles: s,
		StartTask: func(task context.Task) tea.Cmd {
			return func() tea.Msg { return nil }
		},
	}

	// Use the real constructor so the embedded Table, SearchBar, and
	// PromptConfirmationBox are fully initialised with the right ctx.
	m := NewModel(0, ctx, config.PrsSectionConfig{}, time.Time{}, time.Time{})
	m.Prs = []prrow.Data{
		{Primary: &data.PullRequestData{Number: prNumber}},
	}
	return m
}

func TestConfirmation_AcceptWithEmptyInput(t *testing.T) {
	// Pressing Enter without typing anything should confirm, since the
	// prompt says (Y/n) indicating Y is the default.
	m := newTestModel("close")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "empty input (default Y) should execute the action")
	require.False(t, m.IsPromptConfirmationShown,
		"confirmation prompt should be dismissed")
}

func TestConfirmation_AcceptWithLowercaseY(t *testing.T) {
	m := newTestModel("merge")
	m.PromptConfirmationBox.SetValue("y")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "lowercase y should execute the action")
}

func TestConfirmation_AcceptWithUppercaseY(t *testing.T) {
	m := newTestModel("reopen")
	m.PromptConfirmationBox.SetValue("Y")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	require.NotNil(t, cmd, "uppercase Y should execute the action")
}

func TestConfirmation_RejectWithN(t *testing.T) {
	m := newTestModel("close")
	m.PromptConfirmationBox.SetValue("n")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := m.Update(msg)

	// cmd is a batch of (nil, blinkCmd) -- the nil means no action was taken.
	// We verify the prompt is dismissed regardless.
	require.False(t, m.IsPromptConfirmationShown,
		"confirmation prompt should be dismissed on rejection")
	_ = cmd
}

func TestConfirmation_CancelWithEsc(t *testing.T) {
	m := newTestModel("merge")

	msg := tea.KeyPressMsg{Code: tea.KeyEsc}
	_, cmd := m.Update(msg)

	require.False(t, m.IsPromptConfirmationShown,
		"Esc should dismiss the confirmation prompt")
	_ = cmd
}

func TestConfirmation_CancelWithCtrlC(t *testing.T) {
	m := newTestModel("update")

	msg := tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}
	_, cmd := m.Update(msg)

	require.False(t, m.IsPromptConfirmationShown,
		"Ctrl+C should dismiss the confirmation prompt")
	_ = cmd
}

func TestUpdatePRMsg_AutoMergeEnabled_SetsFlag(t *testing.T) {
	// Test that when UpdatePRMsg with AutoMergeEnabled=true is processed via
	// m.Update(), the AutoMergeEnabled flag on prrow.Data is set to true.
	//
	// This test uses newFullTestModel() which wires up ctx.Config, ctx.Theme,
	// and ctx.Styles so that Update() can call BuildRows() end-to-end.
	m := newFullTestModel(42)

	require.False(t, m.Prs[0].AutoMergeEnabled, "AutoMergeEnabled should start false")

	autoMerge := true
	msg := tasks.UpdatePRMsg{
		PrNumber:         42,
		AutoMergeEnabled: &autoMerge,
	}

	result, _ := m.Update(msg)
	updated := result.(*Model)

	require.True(t, updated.Prs[0].AutoMergeEnabled,
		"AutoMergeEnabled should be set to true after processing AutoMergeEnabled update")
	require.Nil(t, updated.Prs[0].Primary.AutoMergeRequest,
		"AutoMergeRequest should remain nil (only real API data should populate it)")
}

func TestConfirmation_AllActions(t *testing.T) {
	actions := []string{"close", "reopen", "ready", "merge", "update", "approveWorkflows"}

	for _, action := range actions {
		t.Run(action+"_empty_input", func(t *testing.T) {
			m := newTestModel(action)

			msg := tea.KeyPressMsg{Code: tea.KeyEnter}
			_, cmd := m.Update(msg)

			require.NotNil(t, cmd,
				"empty input should confirm for action %q", action)
		})

		t.Run(action+"_explicit_y", func(t *testing.T) {
			m := newTestModel(action)
			m.PromptConfirmationBox.SetValue("y")

			msg := tea.KeyPressMsg{Code: tea.KeyEnter}
			_, cmd := m.Update(msg)

			require.NotNil(t, cmd,
				"explicit y should confirm for action %q", action)
		})
	}
}
