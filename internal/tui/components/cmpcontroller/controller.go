// Package cmpcontroller is used to load completions (e.g. from the network)
// for the various modes and using the cmp package to display them
package cmpcontroller

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"

	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmp"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/inputbox"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Mode int

const (
	ModeNone Mode = iota
	ModeComment
	ModeApprove
	ModeAssign
	ModeUnassign
	ModeLabel
	ModeSearch
)

type SuggestionKind int

const (
	SuggestionNone SuggestionKind = iota
	SuggestionUsers
	SuggestionLabels
)

type FetchPolicy int

const (
	FetchNone FetchPolicy = iota
	FetchSilent
	FetchWithLoading
)

type RepoRef struct {
	NameWithOwner string
	Owner         string
	Name          string
}

type EnterOptions struct {
	Mode                             Mode
	Prompt                           string
	InitialValue                     string
	Source                           cmp.Source
	Repo                             RepoRef
	SuggestionKind                   SuggestionKind
	EnterFetch                       FetchPolicy
	ConfirmDiscardOnCancel           bool
	HideAutocompleteWhenContextEmpty bool
}

type Submit struct {
	Mode  Mode
	Value string
}

type RepoLabelsFetchedMsg struct {
	Labels []data.Label
}

type RepoLabelsFetchFailedMsg struct {
	Err error
}

type RepoUsersFetchedMsg struct {
	Users []data.User
}

type RepoUsersFetchFailedMsg struct {
	Err error
}

type Controller struct {
	ctx               *context.ProgramContext
	inputBox          inputbox.Model
	cmp               *cmp.Model
	mode              Mode
	prompt            string
	repo              RepoRef
	suggestionKind    SuggestionKind
	confirmDiscard    bool
	showConfirmCancel bool
	hideOnEmpty       bool
	repoLabels        []data.Label
	repoUsers         []data.User
}

func New(ctx *context.ProgramContext, ta textarea.Model) Controller {
	inputBox := inputbox.NewModel(ctx, ta)
	cmp := cmp.NewModel(ctx)
	inputBox.SetAutocomplete(&cmp)

	return Controller{
		ctx:      ctx,
		inputBox: inputBox,
		cmp:      &cmp,
	}
}

func (c *Controller) Value() string {
	return c.inputBox.Value()
}

func (c *Controller) SetValue(value string) {
	c.inputBox.SetValue(value)
}

func (c *Controller) Mode() Mode {
	return c.mode
}

func (c *Controller) Active() bool {
	return c.mode != ModeNone
}

func (c *Controller) View() string {
	return c.inputBox.View()
}

func (c *Controller) ViewCompletions() string {
	return c.inputBox.ViewCompletions()
}

func (c *Controller) Width() int {
	return c.cmp.Width()
}

func (c *Controller) SetWidth(width int) {
	c.inputBox.SetWidth(width)
	c.cmp.SetWidth(width)
}

func (c *Controller) Column() int {
	return c.inputBox.Column()
}

func (c *Controller) CursorEnd() {
	c.inputBox.CursorEnd()
}

func (c *Controller) UpdateProgramContext(ctx *context.ProgramContext) {
	c.ctx = ctx
	c.inputBox.UpdateProgramContext(ctx)
	c.cmp.UpdateProgramContext(ctx)
}

func (c *Controller) Exit() {
	c.inputBox.Blur()
	c.inputBox.CursorStart()
	c.resetAutocompleteState()
	c.mode = ModeNone
	c.prompt = ""
	c.repo = RepoRef{}
	c.suggestionKind = SuggestionNone
	c.confirmDiscard = false
	c.showConfirmCancel = false
	c.hideOnEmpty = false
}

// todo make pinter
func (c *Controller) Enter(opts EnterOptions) tea.Cmd {
	c.inputBox.Reset()
	c.inputBox.SetValue(opts.InitialValue)
	c.CursorEnd()
	c.resetAutocompleteState()
	c.mode = opts.Mode
	c.prompt = opts.Prompt
	c.repo = opts.Repo
	c.suggestionKind = opts.SuggestionKind
	c.confirmDiscard = opts.ConfirmDiscardOnCancel
	c.showConfirmCancel = false
	c.hideOnEmpty = opts.HideAutocompleteWhenContextEmpty

	c.inputBox.SetPrompt(opts.Prompt)
	c.inputBox.SetAutocompleteSource(opts.Source)

	cmds := []tea.Cmd{textarea.Blink, c.inputBox.Focus()}

	switch opts.SuggestionKind {
	case SuggestionUsers:
		if users, ok := data.CachedRepoUsers(opts.Repo.NameWithOwner); ok {
			c.repoUsers = users
			c.cmp.SetSuggestions(userSuggestions(users))
			c.showSuggestionsFromCurrentContext()
		} else if opts.EnterFetch != FetchNone {
			cmds = append([]tea.Cmd{c.fetchUsers(opts.EnterFetch == FetchWithLoading)}, cmds...)
		}
	case SuggestionLabels:
		if labels, ok := data.CachedRepoLabels(opts.Repo.NameWithOwner); ok {
			c.repoLabels = labels
			c.cmp.SetSuggestions(labelSuggestions(labels))
			c.showSuggestionsFromCurrentContext()
		} else if opts.EnterFetch != FetchNone {
			cmds = append([]tea.Cmd{c.fetchLabels(opts.EnterFetch == FetchWithLoading)}, cmds...)
		}
	}

	return tea.Sequence(cmds...)
}

func (c *Controller) Update(msg tea.Msg) (tea.Cmd, bool) {
	var (
		cmds  []tea.Cmd
		taCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.PasteMsg:
		if c.Active() {
			c.inputBox, taCmd = c.inputBox.Update(msg)
			cmds = append(cmds, taCmd)
			return tea.Batch(cmds...), true
		}
		return nil, false

	case RepoLabelsFetchedMsg:
		c.repoLabels = msg.Labels
		c.cmp.SetSuggestions(labelSuggestions(msg.Labels))
		cmds = append(cmds, c.cmp.SetFetchSuccess())
		if c.mode == ModeLabel {
			c.showSuggestionsFromCurrentContext()
		}
		return tea.Batch(cmds...), true

	case RepoLabelsFetchFailedMsg:
		return c.cmp.SetFetchError(msg.Err), true

	case RepoUsersFetchedMsg:
		c.repoUsers = msg.Users
		c.cmp.SetSuggestions(userSuggestions(msg.Users))
		cmds = append(cmds, c.cmp.SetFetchSuccess())
		if c.mode == ModeComment || c.mode == ModeApprove || c.mode == ModeAssign {
			c.showSuggestionsFromCurrentContext()
		}
		return tea.Batch(cmds...), true

	case RepoUsersFetchFailedMsg:
		return c.cmp.SetFetchError(msg.Err), true

	case cmp.FetchSuggestionsRequestedMsg:
		if !c.Active() || c.suggestionKind == SuggestionNone {
			return nil, false
		}
		if msg.Force {
			c.clearRelevantCache()
		}
		switch c.suggestionKind {
		case SuggestionUsers:
			return c.fetchUsers(true), true
		case SuggestionLabels:
			return c.fetchLabels(true), true
		default:
			return nil, false
		}

	case tea.KeyMsg:
		if !c.Active() {
			return nil, false
		}

		switch {
		case key.Matches(msg, cmp.RefreshSuggestionsKey):
			if c.suggestionKind == SuggestionNone {
				return nil, true
			}
			c.clearRelevantCache()
			switch c.suggestionKind {
			case SuggestionUsers:
				return c.fetchUsers(true), true
			case SuggestionLabels:
				return c.fetchLabels(true), true
			}
		}

		switch msg.String() {
		case "esc", "ctrl+c":
			if c.confirmDiscard {
				if !c.showConfirmCancel {
					c.setDiscardPrompt()
					return nil, true
				}
				c.Exit()
				return nil, true
			}
			c.Exit()
			return nil, true

		default:
			if c.confirmDiscard {
				if msg.String() == "Y" || msg.String() == "y" {
					if c.showConfirmCancel {
						c.Exit()
						return nil, true
					}
				}
				if c.showConfirmCancel && (msg.String() == "N" || msg.String() == "n") {
					c.restorePrompt()
					return nil, true
				}
				if c.showConfirmCancel {
					c.restorePrompt()
				}
			}
		}

		var previousContext cmp.Context
		if c.usesAutocomplete() {
			previousContext = c.inputBox.CurrentAutocompleteContext()
		}

		c.inputBox, taCmd = c.inputBox.Update(msg)
		cmds = append(cmds, taCmd)

		if c.usesAutocomplete() {
			currentContext := c.inputBox.CurrentAutocompleteContext()
			if currentContext != previousContext {
				if c.hideOnEmpty && currentContext == (cmp.Context{}) {
					c.cmp.Hide()
				} else {
					c.cmp.Show(currentContext.Content, c.inputBox.AutocompleteItemsToExclude())
				}
			}
		}

		return tea.Batch(cmds...), true
	}

	switch msg.(type) {
	case spinner.TickMsg, cmp.ClearFetchStatusMsg:
		var acCmd tea.Cmd
		*c.cmp, acCmd = c.cmp.Update(msg)
		return acCmd, c.Active() || c.suggestionKind != SuggestionNone
	}

	c.inputBox, taCmd = c.inputBox.Update(msg)

	return taCmd, false
}

func (c *Controller) LineFromBottom() int {
	return c.inputBox.LineFromBottom()
}

func (c *Controller) clearRelevantCache() {
	switch c.suggestionKind {
	case SuggestionUsers:
		if c.repo.NameWithOwner != "" {
			data.ClearRepoUserCache(c.repo.NameWithOwner)
		}
	case SuggestionLabels:
		if c.repo.NameWithOwner != "" {
			data.ClearRepoLabelCache(c.repo.NameWithOwner)
		}
	}
}

func (c *Controller) restorePrompt() {
	c.inputBox.SetPrompt(c.prompt)
	c.showConfirmCancel = false
}

func (c *Controller) setDiscardPrompt() {
	c.inputBox.SetPrompt(
		lipgloss.NewStyle().Foreground(c.ctx.Theme.ErrorText).Render("Discard comment? (y/N)"),
	)
	c.showConfirmCancel = true
}

func (c *Controller) resetAutocompleteState() {
	c.cmp.Reset()
	c.cmp.Hide()
	c.cmp.SetSuggestions(nil)
}

func (c Controller) showSuggestionsFromCurrentContext() {
	if !c.usesAutocomplete() {
		return
	}
	currentContext := c.inputBox.CurrentAutocompleteContext()
	if c.hideOnEmpty && currentContext == (cmp.Context{}) {
		c.cmp.Hide()
		return
	}
	c.cmp.Show(currentContext.Content, c.inputBox.AutocompleteItemsToExclude())
}

func (c Controller) usesAutocomplete() bool {
	switch c.mode {
	case ModeComment, ModeApprove, ModeAssign, ModeLabel:
		return true
	default:
		return false
	}
}

func (c Controller) fetchLabels(showLoading bool) tea.Cmd {
	var spinnerTickCmd tea.Cmd
	if showLoading {
		spinnerTickCmd = c.cmp.SetFetchLoading()
	}

	fetchCmd := func() tea.Msg {
		labels, err := data.FetchRepoLabels(c.repo.NameWithOwner)
		if err != nil {
			return RepoLabelsFetchFailedMsg{Err: err}
		}
		return RepoLabelsFetchedMsg{Labels: labels}
	}

	if spinnerTickCmd != nil {
		return tea.Batch(spinnerTickCmd, fetchCmd)
	}
	return fetchCmd
}

func (c Controller) fetchUsers(showLoading bool) tea.Cmd {
	var spinnerTickCmd tea.Cmd
	if showLoading {
		spinnerTickCmd = c.cmp.SetFetchLoading()
	}

	fetchCmd := func() tea.Msg {
		users, err := data.FetchRepoUsers(c.repo.Owner, c.repo.Name)
		if err != nil {
			return RepoUsersFetchFailedMsg{Err: err}
		}
		return RepoUsersFetchedMsg{Users: users}
	}

	if spinnerTickCmd != nil {
		return tea.Batch(spinnerTickCmd, fetchCmd)
	}
	return fetchCmd
}

func userSuggestions(users []data.User) []cmp.Suggestion {
	suggestions := make([]cmp.Suggestion, 0, len(users))
	for _, user := range users {
		suggestions = append(suggestions, cmp.Suggestion{
			Value:  user.Login,
			Detail: strings.TrimSpace(user.Name),
		})
	}
	return suggestions
}

func labelSuggestions(labels []data.Label) []cmp.Suggestion {
	suggestions := make([]cmp.Suggestion, 0, len(labels))
	for _, label := range labels {
		suggestions = append(suggestions, cmp.Suggestion{
			Value:  label.Name,
			Detail: strings.TrimSpace(label.Description),
		})
	}
	return suggestions
}
