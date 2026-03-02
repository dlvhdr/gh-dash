package detailedit

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	dataautocomplete "github.com/dlvhdr/gh-dash/v4/internal/data/autocomplete"
	popupautocomplete "github.com/dlvhdr/gh-dash/v4/internal/tui/components/autocomplete"
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

	return New(ctx)
}

func testRepo() RepoRef {
	return RepoRef{
		NameWithOwner: "owner/repo",
		Owner:         "owner",
		Name:          "repo",
	}
}

func suggestions(values ...string) []popupautocomplete.Suggestion {
	items := make([]popupautocomplete.Suggestion, 0, len(values))
	for _, value := range values {
		items = append(items, popupautocomplete.Suggestion{Value: value})
	}
	return items
}

func TestEnterCommentModeResetsAutocompleteState(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.ac.SetSuggestions(suggestions("bug", "feature"))
	c.ac.Show("bug", nil)
	require.True(t, c.ac.HasSuggestions())

	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Source:                           dataautocomplete.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})

	require.False(t, c.ac.HasSuggestions())
	require.False(t, c.ac.IsVisible())
}

func TestEnterAssignModeResetsAutocompleteState(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c.ac.SetSuggestions(suggestions("bug", "feature"))
	c.ac.Show("bug", nil)
	require.True(t, c.ac.HasSuggestions())

	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Source:                           dataautocomplete.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	require.False(t, c.ac.HasSuggestions())
	require.False(t, c.ac.IsVisible())
}

func TestEnterLabelModePrepopulatesCurrentLabels(t *testing.T) {
	data.ClearLabelCache()
	c := newTestController(t)

	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeLabel,
		Prompt:                           "label",
		InitialValue:                     "bug, docs, ",
		Source:                           dataautocomplete.LabelSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionLabels,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	require.Equal(t, "bug, docs, ", c.inputBox.Value())
}

func TestRepoUsersFetchedUpdatesControllerState(t *testing.T) {
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Source:                           dataautocomplete.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	c, _, _, handled := c.Update(RepoUsersFetchedMsg{Users: []data.User{{Login: "alice", Name: "Alice"}}})
	require.True(t, handled)
	require.Len(t, c.repoUsers, 1)
	require.Equal(t, "alice", c.repoUsers[0].Login)
	require.True(t, c.ac.IsVisible())
}

func TestRepoLabelsFetchedUpdatesControllerState(t *testing.T) {
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeLabel,
		Prompt:                           "label",
		InitialValue:                     "bu",
		Source:                           dataautocomplete.LabelSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionLabels,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	c, _, _, handled := c.Update(RepoLabelsFetchedMsg{Labels: []data.Label{{Name: "bug", Description: "Bug"}}})
	require.True(t, handled)
	require.Len(t, c.repoLabels, 1)
	require.Equal(t, "bug", c.repoLabels[0].Name)
	require.True(t, c.ac.IsVisible())
}

func TestEnterSilentFetchReturnsCommand(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)

	_, cmd := c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		Source:                           dataautocomplete.WhitespaceSource{},
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
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeLabel,
		Prompt:                           "label",
		Source:                           dataautocomplete.LabelSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionLabels,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	_, cmd, _, handled := c.Update(popupautocomplete.FetchSuggestionsRequestedMsg{Force: true})
	require.True(t, handled)
	require.NotNil(t, cmd)
}

func TestCommentModeHidesPopupWhenMentionContextDisappears(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		InitialValue:                     "@ali",
		Source:                           dataautocomplete.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})
	c.ac.SetSuggestions(suggestions("alice"))
	c.showSuggestionsFromCurrentContext()
	require.True(t, c.ac.IsVisible())

	c.inputBox.SetValue("done")
	c.showSuggestionsFromCurrentContext()
	require.False(t, c.ac.IsVisible())
}

func TestCommentModeShowsPopupForBareAtMention(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Source:                           dataautocomplete.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})
	c.ac.SetSuggestions(suggestions("alice"))

	c, _, _, handled := c.Update(tea.KeyPressMsg{Text: "@"})
	require.True(t, handled)
	require.True(t, c.ac.IsVisible())
}

func TestAssignModeShowsPopupForEmptyContext(t *testing.T) {
	data.ClearUserCache()
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		InitialValue:                     "",
		Source:                           dataautocomplete.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})
	c.ac.SetSuggestions(suggestions("alice"))
	c.showSuggestionsFromCurrentContext()

	require.True(t, c.ac.IsVisible())
}

func TestEscapeInCommentModeShowsDiscardPrompt(t *testing.T) {
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Source:                           dataautocomplete.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})

	c, _, _, handled := c.Update(tea.KeyPressMsg{Text: "esc"})
	require.True(t, handled)
	require.True(t, c.showConfirmCancel)
}

func TestConfirmDiscardExitsMode(t *testing.T) {
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeApprove,
		Prompt:                           "approve",
		Source:                           dataautocomplete.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: false,
	})

	c.showConfirmCancel = true
	c, _, _, handled := c.Update(tea.KeyPressMsg{Text: "y"})
	require.True(t, handled)
	require.Equal(t, ModeNone, c.Mode())
}

func TestRejectDiscardRestoresPrompt(t *testing.T) {
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeComment,
		Prompt:                           "comment",
		Source:                           dataautocomplete.UserMentionSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		ConfirmDiscardOnCancel:           true,
		HideAutocompleteWhenContextEmpty: true,
	})

	c.showConfirmCancel = true
	c, _, _, handled := c.Update(tea.KeyPressMsg{Text: "n"})
	require.True(t, handled)
	require.False(t, c.showConfirmCancel)
}

func TestCtrlDReturnsSubmit(t *testing.T) {
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:                             ModeAssign,
		Prompt:                           "assign",
		InitialValue:                     "alice bob",
		Source:                           dataautocomplete.WhitespaceSource{},
		Repo:                             testRepo(),
		SuggestionKind:                   SuggestionUsers,
		EnterFetch:                       FetchNone,
		HideAutocompleteWhenContextEmpty: false,
	})

	c, _, submit, handled := c.Update(tea.KeyPressMsg{Text: "ctrl+d"})
	require.True(t, handled)
	require.Equal(t, ModeNone, c.Mode())
	require.NotNil(t, submit)
	require.Equal(t, ModeAssign, submit.Mode)
	require.Equal(t, "alice bob", submit.Value)
}

func TestUnassignModeDoesNotUseAutocomplete(t *testing.T) {
	c := newTestController(t)
	c, _ = c.Enter(EnterOptions{
		Mode:         ModeUnassign,
		Prompt:       "unassign",
		InitialValue: "alice\nbob",
		Repo:         testRepo(),
	})

	require.False(t, c.usesAutocomplete())
}
