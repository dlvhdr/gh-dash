package cmpcontroller

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/fuzzyselect"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/inputbox"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newTestController(t *testing.T) Controller {
	t.Helper()

	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config: &cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
	}

	ta := inputbox.DefaultTextArea(ctx)
	return New(ctx, inputbox.ModelOpts{TextArea: &ta})
}

func testRepo() RepoRef {
	return RepoRef{
		NameWithOwner: "owner/repo",
		Owner:         "owner",
		Name:          "repo",
	}
}

func suggestions(values ...string) []fuzzyselect.Suggestion {
	items := make([]fuzzyselect.Suggestion, 0, len(values))
	for _, value := range values {
		items = append(items, fuzzyselect.Suggestion{Value: value})
	}
	return items
}

var emptyCtx = fuzzyselect.Context{}

func TestEnterCommentModeResetsAutocompleteState(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.ListSource{Options: suggestions("bug", "feature")})
	c.fzfSelect.Filter("bug", emptyCtx, nil)
	require.True(t, c.fzfSelect.HasSuggestions())

	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})

	require.False(t, c.fzfSelect.HasSuggestions())
	require.False(t, c.fzfSelect.IsVisible())
}

func TestEnterAssignModeResetsAutocompleteState(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.ListSource{Options: suggestions("bug", "feature")})
	c.fzfSelect.Filter("bug", emptyCtx, nil)
	c.fzfSelect.Show()
	require.True(t, c.fzfSelect.HasSuggestions())

	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	require.False(t, c.fzfSelect.HasSuggestions())
	require.False(t, c.fzfSelect.IsVisible())
}

func TestEnterLabelModePrepopulatesCurrentLabels(t *testing.T) {
	data.ClearLabelCache()
	c := newTestController(t)

	c.SetAutocompleteSource(&fuzzyselect.ListSource{Options: suggestions("bug", "docs")})
	c.Enter(EnterOptions{
		Mode:                             ModeLabel,
		Prompt:                           "label",
		InitialValue:                     "bug, docs, ",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	require.Equal(t, "bug, docs, ", c.inputBox.Value())
}

func TestRepoUsersFetchedStartsFiltering(t *testing.T) {
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: false})
	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	c.fzfSelect.Source.(*fuzzyselect.UserMentionSource).Users = []data.User{{Login: "alice"}}

	_, handled := c.Update(
		SourceDataFetchedMsg{},
	)
	require.True(t, handled)
	suggestions := c.fzfSelect.Source.Suggestions("", tea.Position{X: 0, Y: 0})
	require.Len(t, suggestions, 1)
	require.Equal(t, "alice", suggestions[0].Value)
}

func TestRepoLabelsFetchedStartsFiltering(t *testing.T) {
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.LabelSource{})
	c.Enter(EnterOptions{
		Mode:                             ModeLabel,
		Prompt:                           "labels",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	c.fzfSelect.Source.(*fuzzyselect.LabelSource).Labels = []data.Label{{Name: "low-pri"}}

	_, handled := c.Update(
		SourceDataFetchedMsg{},
	)
	require.True(t, handled)
	suggestions := c.fzfSelect.Source.Suggestions("", tea.Position{X: 0, Y: 0})
	require.Len(t, suggestions, 1)
	require.Equal(t, "low-pri", suggestions[0].Value)
}

func TestEnterSilentFetchReturnsCommand(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)

	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{})
	cmd := c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Repo:                             testRepo(),
		EnterFetch:                       FetchSilent,
		HideAutocompleteWhenContextEmpty: false,
	})

	require.NotNil(t, cmd)
}

func TestCommentModeHidesPopupWhenMentionContextDisappears(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: true})
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		InitialValue:                     "@ali",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})

	c.fzfSelect.Source.(*fuzzyselect.UserMentionSource).Users = []data.User{{Login: "alice"}}
	c.Update(SourceDataFetchedMsg{})
	c.ShowCompletions()
	require.True(t, c.fzfSelect.IsVisible())

	c.fzfSelect.Hide()
	c.inputBox.SetValue("done")
	c.Filter()

	// should be a no op since we set `HideAutocompleteWhenContextEmpty` to true
	c.ShowCompletions()
	require.False(t, c.fzfSelect.IsVisible())
}

func TestCommentModeShowsPopupAtMention(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: true})
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})
	c.fzfSelect.Source.(*fuzzyselect.UserMentionSource).Users = []data.User{{Login: "alice"}}
	c.Update(SourceDataFetchedMsg{})

	_, handled := c.Update(tea.KeyPressMsg{Text: "@"})
	require.True(t, handled)
	require.True(t, c.fzfSelect.IsVisible())
}

func TestAssignModeShowsPopupForEmptyContext(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: false})
	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		InitialValue:                     "",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})
	c.fzfSelect.Source.(*fuzzyselect.UserMentionSource).Users = []data.User{{Login: "alice"}}
	c.Update(SourceDataFetchedMsg{})
	c.ShowCompletions()

	require.True(t, c.fzfSelect.IsVisible())
}

func TestEscapeInCommentModeShowsDiscardPrompt(t *testing.T) {
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: true})
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})

	_, handled := c.Update(tea.KeyPressMsg{Text: "esc"})
	require.True(t, handled)
	require.True(t, c.showConfirmCancel)
}

func TestConfirmDiscardExitsMode(t *testing.T) {
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: true})
	c.Enter(EnterOptions{
		Mode:                             ModeApprove,
		Prompt:                           "approve",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: false,
	})

	c.showConfirmCancel = true
	_, handled := c.Update(tea.KeyPressMsg{Text: "y"})
	require.True(t, handled)
}

func TestRejectDiscardRestoresPrompt(t *testing.T) {
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: true})
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})

	c.showConfirmCancel = true
	_, handled := c.Update(tea.KeyPressMsg{Text: "n"})
	require.True(t, handled)
	require.False(t, c.showConfirmCancel)
}

func TestCtrlDReturnsSubmit(t *testing.T) {
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: true})
	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		InitialValue:                     "alice bob",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	_, handled := c.Update(tea.KeyPressMsg{Text: "ctrl+d"})
	require.True(t, handled)
	require.Equal(t, ModeAssign, c.Mode())
	require.Equal(t, "alice bob", c.Value())
}

func TestPasteInCommentMode(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: true})
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})

	msg := tea.PasteMsg{Content: "pasted text"}
	_, handled := c.Update(msg)

	require.True(t, handled)
	require.Contains(t, c.inputBox.Value(), "pasted text",
		"pasted text should appear in the inputbox")
}

func TestPasteInAssignMode(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.SetAutocompleteSource(&fuzzyselect.UserMentionSource{WithAtSymbol: false})
	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Repo:                             testRepo(),
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	msg := tea.PasteMsg{Content: "alice bob"}
	_, handled := c.Update(msg)

	require.True(t, handled)
	require.Contains(t, c.inputBox.Value(), "alice bob",
		"pasted text should appear in the inputbox during assignment")
}

func TestPasteIgnoredWhenInactive(t *testing.T) {
	c := newTestController(t)

	msg := tea.PasteMsg{Content: "should not appear"}
	_, handled := c.Update(msg)

	require.False(t, handled,
		"paste should not be handled when controller is inactive")
}
