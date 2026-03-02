package detailedit

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	dataautocomplete "github.com/dlvhdr/gh-dash/v4/internal/data/autocomplete"
	popupautocomplete "github.com/dlvhdr/gh-dash/v4/internal/tui/components/autocomplete"
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
	Source                           dataautocomplete.Source
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
	ac                *popupautocomplete.Model
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

func New(ctx *context.ProgramContext) Controller {
	inputBox := inputbox.NewModel(ctx)
	ac := popupautocomplete.NewModel(ctx)
	inputBox.SetAutocomplete(&ac)

	return Controller{
		ctx:      ctx,
		inputBox: inputBox,
		ac:       &ac,
	}
}

func (c Controller) Mode() Mode {
	return c.mode
}

func (c Controller) Active() bool {
	return c.mode != ModeNone
}

func (c Controller) View() string {
	if !c.Active() {
		return ""
	}

	switch c.mode {
	case ModeComment, ModeApprove, ModeAssign, ModeLabel:
		return c.inputBox.ViewWithAutocomplete()
	default:
		return c.inputBox.View()
	}
}

func (c *Controller) SetWidth(width int) {
	c.inputBox.SetWidth(width)
	c.ac.SetWidth(width - 4)
}

func (c *Controller) UpdateProgramContext(ctx *context.ProgramContext) {
	c.ctx = ctx
	c.inputBox.UpdateProgramContext(ctx)
	c.ac.UpdateProgramContext(ctx)
}

func (c Controller) Exit() Controller {
	c.inputBox.Blur()
	c.resetAutocompleteState()
	c.mode = ModeNone
	c.prompt = ""
	c.repo = RepoRef{}
	c.suggestionKind = SuggestionNone
	c.confirmDiscard = false
	c.showConfirmCancel = false
	c.hideOnEmpty = false
	return c
}

func (c Controller) Enter(opts EnterOptions) (Controller, tea.Cmd) {
	c.inputBox.Reset()
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
	c.inputBox.SetValue(opts.InitialValue)

	cmds := []tea.Cmd{textarea.Blink, c.inputBox.Focus()}

	switch opts.SuggestionKind {
	case SuggestionUsers:
		if users, ok := data.CachedRepoUsers(opts.Repo.NameWithOwner); ok {
			c.repoUsers = users
			c.ac.SetSuggestions(userSuggestions(users))
			c.showSuggestionsFromCurrentContext()
		} else if opts.EnterFetch != FetchNone {
			cmds = append([]tea.Cmd{c.fetchUsers(opts.EnterFetch == FetchWithLoading)}, cmds...)
		}
	case SuggestionLabels:
		if labels, ok := data.CachedRepoLabels(opts.Repo.NameWithOwner); ok {
			c.repoLabels = labels
			c.ac.SetSuggestions(labelSuggestions(labels))
			c.showSuggestionsFromCurrentContext()
		} else if opts.EnterFetch != FetchNone {
			cmds = append([]tea.Cmd{c.fetchLabels(opts.EnterFetch == FetchWithLoading)}, cmds...)
		}
	}

	return c, tea.Sequence(cmds...)
}

func (c Controller) Update(msg tea.Msg) (Controller, tea.Cmd, *Submit, bool) {
	var (
		cmds  []tea.Cmd
		taCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case RepoLabelsFetchedMsg:
		c.repoLabels = msg.Labels
		c.ac.SetSuggestions(labelSuggestions(msg.Labels))
		cmds = append(cmds, c.ac.SetFetchSuccess())
		if c.mode == ModeLabel {
			c.showSuggestionsFromCurrentContext()
		}
		return c, tea.Batch(cmds...), nil, true

	case RepoLabelsFetchFailedMsg:
		return c, c.ac.SetFetchError(msg.Err), nil, true

	case RepoUsersFetchedMsg:
		c.repoUsers = msg.Users
		c.ac.SetSuggestions(userSuggestions(msg.Users))
		cmds = append(cmds, c.ac.SetFetchSuccess())
		if c.mode == ModeComment || c.mode == ModeApprove || c.mode == ModeAssign {
			c.showSuggestionsFromCurrentContext()
		}
		return c, tea.Batch(cmds...), nil, true

	case RepoUsersFetchFailedMsg:
		return c, c.ac.SetFetchError(msg.Err), nil, true

	case popupautocomplete.FetchSuggestionsRequestedMsg:
		if !c.Active() || c.suggestionKind == SuggestionNone {
			return c, nil, nil, false
		}
		if msg.Force {
			c.clearRelevantCache()
		}
		switch c.suggestionKind {
		case SuggestionUsers:
			return c, c.fetchUsers(true), nil, true
		case SuggestionLabels:
			return c, c.fetchLabels(true), nil, true
		default:
			return c, nil, nil, false
		}

	case tea.KeyMsg:
		if !c.Active() {
			return c, nil, nil, false
		}

		switch {
		case key.Matches(msg, popupautocomplete.RefreshSuggestionsKey):
			if c.suggestionKind == SuggestionNone {
				return c, nil, nil, true
			}
			c.clearRelevantCache()
			switch c.suggestionKind {
			case SuggestionUsers:
				return c, c.fetchUsers(true), nil, true
			case SuggestionLabels:
				return c, c.fetchLabels(true), nil, true
			}
		}

		switch msg.Type {
		case tea.KeyCtrlD:
			value := c.inputBox.Value()
			mode := c.mode
			c = c.Exit()
			return c, nil, &Submit{Mode: mode, Value: value}, true

		case tea.KeyEsc, tea.KeyCtrlC:
			if c.confirmDiscard {
				if !c.showConfirmCancel {
					c.setDiscardPrompt()
					return c, nil, nil, true
				}
				c = c.Exit()
				return c, nil, nil, true
			}
			c = c.Exit()
			return c, nil, nil, true

		default:
			if c.confirmDiscard {
				if msg.String() == "Y" || msg.String() == "y" {
					if c.showConfirmCancel {
						c = c.Exit()
						return c, nil, nil, true
					}
				}
				if c.showConfirmCancel && (msg.String() == "N" || msg.String() == "n") {
					c.restorePrompt()
					return c, nil, nil, true
				}
				if c.showConfirmCancel {
					c.restorePrompt()
				}
			}
		}

		var previousContext dataautocomplete.Context
		if c.usesAutocomplete() {
			previousContext = c.inputBox.CurrentAutocompleteContext()
		}

		c.inputBox, taCmd = c.inputBox.Update(msg)
		cmds = append(cmds, taCmd)

		if c.usesAutocomplete() {
			currentContext := c.inputBox.CurrentAutocompleteContext()
			if currentContext != previousContext {
				if c.hideOnEmpty && currentContext == (dataautocomplete.Context{}) {
					c.ac.Hide()
				} else {
					c.ac.Show(currentContext.Content, c.inputBox.AutocompleteItemsToExclude())
				}
			}
		}

		return c, tea.Batch(cmds...), nil, true
	}

	switch msg.(type) {
	case spinner.TickMsg, popupautocomplete.ClearFetchStatusMsg:
		var acCmd tea.Cmd
		*c.ac, acCmd = c.ac.Update(msg)
		return c, acCmd, nil, c.Active() || c.suggestionKind != SuggestionNone
	}

	return c, nil, nil, false
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
	c.inputBox.SetPrompt(lipgloss.NewStyle().Foreground(c.ctx.Theme.ErrorText).Render("Discard comment? (y/N)"))
	c.showConfirmCancel = true
}

func (c *Controller) resetAutocompleteState() {
	c.ac.Reset()
	c.ac.Hide()
	c.ac.SetSuggestions(nil)
}

func (c Controller) showSuggestionsFromCurrentContext() {
	if !c.usesAutocomplete() {
		return
	}
	currentContext := c.inputBox.CurrentAutocompleteContext()
	if c.hideOnEmpty && currentContext == (dataautocomplete.Context{}) {
		c.ac.Hide()
		return
	}
	c.ac.Show(currentContext.Content, c.inputBox.AutocompleteItemsToExclude())
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
		spinnerTickCmd = c.ac.SetFetchLoading()
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
		spinnerTickCmd = c.ac.SetFetchLoading()
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

func userSuggestions(users []data.User) []popupautocomplete.Suggestion {
	suggestions := make([]popupautocomplete.Suggestion, 0, len(users))
	for _, user := range users {
		suggestions = append(suggestions, popupautocomplete.Suggestion{
			Value:  user.Login,
			Detail: strings.TrimSpace(user.Name),
		})
	}
	return suggestions
}

func labelSuggestions(labels []data.Label) []popupautocomplete.Suggestion {
	suggestions := make([]popupautocomplete.Suggestion, 0, len(labels))
	for _, label := range labels {
		suggestions = append(suggestions, popupautocomplete.Suggestion{
			Value:  label.Name,
			Detail: strings.TrimSpace(label.Description),
		})
	}
	return suggestions
}
