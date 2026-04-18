package cmpcontroller

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmp"
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

func suggestions(values ...string) []cmp.Suggestion {
	items := make([]cmp.Suggestion, 0, len(values))
	for _, value := range values {
		items = append(items, cmp.Suggestion{Value: value})
	}
	return items
}

func TestEnterCommentModeResetsAutocompleteState(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.cmp.SetSuggestions(suggestions("bug", "feature"))
	c.cmp.Show("bug", nil)
	require.True(t, c.cmp.HasSuggestions())

	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Source:                           cmp.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})

	require.False(t, c.cmp.HasSuggestions())
	require.False(t, c.cmp.IsVisible())
}

func TestEnterAssignModeResetsAutocompleteState(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.cmp.SetSuggestions(suggestions("bug", "feature"))
	c.cmp.Show("bug", nil)
	require.True(t, c.cmp.HasSuggestions())

	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Source:                           cmp.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	require.False(t, c.cmp.HasSuggestions())
	require.False(t, c.cmp.IsVisible())
}

func TestEnterLabelModePrepopulatesCurrentLabels(t *testing.T) {
	data.ClearLabelCache()
	c := newTestController(t)

	c.Enter(EnterOptions{
		Mode:                             ModeLabel,
		Prompt:                           "label",
		InitialValue:                     "bug, docs, ",
		Source:                           cmp.LabelSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionLabels,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	require.Equal(t, "bug, docs, ", c.inputBox.Value())
}

func TestRepoUsersFetchedUpdatesControllerState(t *testing.T) {
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Source:                           cmp.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	_, handled := c.Update(
		RepoUsersFetchedMsg{Users: []data.User{{Login: "alice", Name: "Alice"}}},
	)
	require.True(t, handled)
	require.Len(t, c.repoUsers, 1)
	require.Equal(t, "alice", c.repoUsers[0].Login)
	require.True(t, c.cmp.IsVisible())
}

func TestRepoLabelsFetchedUpdatesControllerState(t *testing.T) {
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:                             ModeLabel,
		Prompt:                           "label",
		InitialValue:                     "bu",
		Source:                           cmp.LabelSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionLabels,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	_, handled := c.Update(
		RepoLabelsFetchedMsg{Labels: []data.Label{{Name: "bug", Description: "Bug"}}},
	)
	require.True(t, handled)
	require.Len(t, c.repoLabels, 1)
	require.Equal(t, "bug", c.repoLabels[0].Name)
	require.True(t, c.cmp.IsVisible())
}

func TestEnterSilentFetchReturnsCommand(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)

	cmd := c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Source:                           cmp.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchSilent,
		HideAutocompleteWhenContextEmpty: false,
	})

	require.NotNil(t, cmd)
}

func TestForceRefreshClearsRelevantCache(t *testing.T) {
	data.ClearLabelCache()
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:                             ModeLabel,
		Prompt:                           "label",
		Source:                           cmp.LabelSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionLabels,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	cmd, handled := c.Update(cmp.FetchSuggestionsRequestedMsg{Force: true})
	require.True(t, handled)
	require.NotNil(t, cmd)
}

func TestCommentModeHidesPopupWhenMentionContextDisappears(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		InitialValue:                     "@ali",
		Source:                           cmp.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})
	c.cmp.SetSuggestions(suggestions("alice"))
	c.showSuggestionsFromCurrentContext()
	require.True(t, c.cmp.IsVisible())

	c.inputBox.SetValue("done")
	c.showSuggestionsFromCurrentContext()
	require.False(t, c.cmp.IsVisible())
}

func TestCommentModeHidesPopupWhenMentionContextDisappearsWhitespace(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		InitialValue:                     "@ali ",
		Source:                           cmp.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})
	c.cmp.SetSuggestions(suggestions("alice"))
	c.Update(c.inputBox.Focus())
	c.inputBox.CursorEnd()
	c.showSuggestionsFromCurrentContext()
	require.False(t, c.cmp.IsVisible())
}

func TestCommentModeShowsPopupForBareAtMention(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Source:                           cmp.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})
	c.cmp.SetSuggestions(suggestions("alice"))

	_, handled := c.Update(tea.KeyPressMsg{Text: "@"})
	require.True(t, handled)
	require.True(t, c.cmp.IsVisible())
}

func TestAssignModeShowsPopupForEmptyContext(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		InitialValue:                     "",
		Source:                           cmp.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})
	c.cmp.SetSuggestions(suggestions("alice"))
	c.showSuggestionsFromCurrentContext()

	require.True(t, c.cmp.IsVisible())
}

func TestEscapeInCommentModeShowsDiscardPrompt(t *testing.T) {
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Source:                           cmp.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
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
	c.Enter(EnterOptions{
		Mode:                             ModeApprove,
		Prompt:                           "approve",
		Source:                           cmp.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
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
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Source:                           cmp.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
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
	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		InitialValue:                     "alice bob",
		Source:                           cmp.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	_, handled := c.Update(tea.KeyPressMsg{Text: "ctrl+d"})
	require.True(t, handled)
	require.Equal(t, ModeAssign, c.Mode())
	require.Equal(t, "alice bob", c.Value())
}

func TestUnassignModeDoesNotUseAutocomplete(t *testing.T) {
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:         ModeUnassign,
		Prompt:       "unassign",
		InitialValue: "alice\nbob",
		Repo:         testRepo(),
	})

	require.False(t, c.usesAutocomplete())
}

func TestPasteInCommentMode(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Source:                           cmp.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
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
	c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Source:                           cmp.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
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
