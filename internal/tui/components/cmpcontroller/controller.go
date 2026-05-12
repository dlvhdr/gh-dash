// Package cmpcontroller combines an input with a cmp.Source
// and allows switching sources on the fly
package cmpcontroller

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"

	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/fuzzyselect"
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
	Repo                             RepoRef
	EnterFetch                       FetchPolicy
	ConfirmDiscardOnCancel           bool
	HideAutocompleteWhenContextEmpty bool
}

type Submit struct {
	Mode  Mode
	Value string
}

type SourceFetchFailedMsg struct {
	Err error
}

type SourceDataFetchedMsg struct{}

type Controller struct {
	ctx               *context.ProgramContext
	inputBox          inputbox.Model
	fzfSelect         *fuzzyselect.Model
	mode              Mode
	prompt            string
	repo              RepoRef
	confirmDiscard    bool
	showConfirmCancel bool
	hideOnEmpty       bool
}

func New(ctx *context.ProgramContext, opts inputbox.ModelOpts) Controller {
	inputBox := inputbox.NewModel(ctx, opts)
	fzfSelect := fuzzyselect.NewModel(ctx, nil)
	inputBox.SetAutocomplete(&fzfSelect)

	ctl := Controller{
		ctx:       ctx,
		inputBox:  inputBox,
		fzfSelect: &fzfSelect,
	}

	return ctl
}

func (c *Controller) Value() string {
	return c.inputBox.Value()
}

func (c *Controller) SetValue(value string) {
	c.inputBox.SetValue(value)
}

func (c *Controller) SelectStyles() context.SelectStyles {
	return c.fzfSelect.Styles()
}

func (c *Controller) SetSelectStyles(styles context.SelectStyles) {
	c.fzfSelect.SetStyles(styles)
}

func (c *Controller) InputStyles() inputbox.Styles {
	return c.inputBox.Styles()
}

func (c *Controller) SetInputStyles(styles inputbox.Styles) {
	c.inputBox.SetStyles(styles)
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

func (c *Controller) Focused() bool {
	return c.inputBox.Focused()
}

func (c *Controller) ViewCompletions() string {
	return c.inputBox.ViewCompletions()
}

func (c *Controller) Width() int {
	return c.fzfSelect.Width()
}

func (c *Controller) SetWidth(width int) {
	c.inputBox.SetWidth(width)
	c.fzfSelect.SetWidth(width)
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
	c.fzfSelect.UpdateProgramContext(ctx)
}

func (c *Controller) Exit() {
	c.inputBox.Blur()
	c.inputBox.CursorStart()
	c.resetAutocompleteState()
	c.mode = ModeNone
	c.prompt = ""
	c.repo = RepoRef{}
	c.confirmDiscard = false
	c.showConfirmCancel = false
	c.hideOnEmpty = false
}

func (c *Controller) SetAutocompleteSource(src fuzzyselect.Source) {
	c.fzfSelect.Source = src
}

func (c *Controller) Enter(opts EnterOptions) tea.Cmd {
	c.inputBox.Reset()
	c.inputBox.SetValue(opts.InitialValue)
	c.CursorEnd()
	c.resetAutocompleteState()
	c.mode = opts.Mode
	c.prompt = opts.Prompt
	c.repo = opts.Repo
	c.confirmDiscard = opts.ConfirmDiscardOnCancel
	c.showConfirmCancel = false
	c.hideOnEmpty = opts.HideAutocompleteWhenContextEmpty

	c.inputBox.SetPrompt(opts.Prompt)

	cmds := []tea.Cmd{
		textarea.Blink,
		c.inputBox.Focus(),
		c.loadSuggestions(opts.EnterFetch == FetchWithLoading),
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

	case tea.KeyMsg:
		if !c.Active() {
			return nil, false
		}

		switch {
		case key.Matches(msg, fuzzyselect.RefreshSuggestionsKey):
			c.clearRelevantCache()
			cmds = append(cmds, c.loadSuggestions(true))
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

		var previousContext fuzzyselect.Context
		if c.usesAutocomplete() {
			previousContext = c.inputBox.CurrentAutocompleteContext()
		}

		c.inputBox, taCmd = c.inputBox.Update(msg)
		cmds = append(cmds, taCmd)

		if c.usesAutocomplete() {
			currentContext := c.inputBox.CurrentAutocompleteContext()
			if currentContext != previousContext {
				if c.hideOnEmpty && currentContext == (fuzzyselect.Context{}) {
					c.fzfSelect.Hide()
				} else {
					c.ShowCompletions()
				}
			}
		}

		return tea.Batch(cmds...), true
	}

	var acCmd tea.Cmd
	switch msg := msg.(type) {
	case spinner.TickMsg, fuzzyselect.ClearFetchStatusMsg:
		*c.fzfSelect, acCmd = c.fzfSelect.Update(msg)
		return acCmd, c.Active()
	case SourceDataFetchedMsg:
		c.Filter()
		acCmd = c.fzfSelect.SetFetchSuccess()
		return acCmd, c.Active()
	case SourceFetchFailedMsg:
		acCmd = c.fzfSelect.SetFetchError(msg.Err)
		return acCmd, c.Active()
	}

	c.inputBox, taCmd = c.inputBox.Update(msg)

	return taCmd, false
}

func (c *Controller) LineFromBottom() int {
	return c.inputBox.LineFromBottom()
}

func (c *Controller) clearRelevantCache() {
	switch c.fzfSelect.Source.(type) {
	case *fuzzyselect.UserMentionSource:
		if c.repo.NameWithOwner != "" {
			data.ClearRepoUserCache(c.repo.NameWithOwner)
		}
	case *fuzzyselect.LabelSource:
		if c.repo.NameWithOwner != "" {
			data.ClearRepoLabelCache(c.repo.NameWithOwner)
		}
	case *fuzzyselect.SearchQuerySource:
		if c.repo.NameWithOwner != "" {
			data.ClearRepoLabelCache(c.repo.NameWithOwner)
			data.ClearRepoUserCache(c.repo.NameWithOwner)
		}
	}
}

func (c *Controller) Filter() {
	currentContext := c.inputBox.CurrentAutocompleteContext()
	exclude := c.inputBox.AutocompleteItemsToExclude()
	log.Debug("filtering suggestions", "mode", c.mode, "ctx", currentContext)
	c.fzfSelect.Filter(c.inputBox.Value(), currentContext, exclude)
}

func (c *Controller) ShowCompletions() {
	inputValue := c.inputBox.Value()
	ctx := c.fzfSelect.Source.ExtractContext(
		inputValue,
		c.inputBox.GetAbsoluteCursorPosition(),
	)

	if c.hideOnEmpty && !c.fzfSelect.HasSuggestions() {
		return
	}

	lines := c.inputBox.Lines()
	switch src := c.fzfSelect.Source.(type) {
	case *fuzzyselect.UserMentionSource:
		if !src.WithAtSymbol {
			c.fzfSelect.Show()
		} else if x := ctx.Start.X - 1; x >= 0 && len(lines) > ctx.Start.Y && len(lines[ctx.Start.Y]) > x &&
			lines[ctx.Start.Y][x] == '@' {
			c.fzfSelect.Show()
		}
	default:
		c.fzfSelect.Show()
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
	c.fzfSelect.Reset()
	c.fzfSelect.Hide()
}

func (c Controller) usesAutocomplete() bool {
	switch c.mode {
	case ModeComment, ModeApprove, ModeAssign, ModeLabel, ModeSearch:
		return true
	default:
		return false
	}
}

func (c Controller) loadSuggestions(showLoading bool) tea.Cmd {
	var spinnerTickCmd tea.Cmd
	if c.fzfSelect.Source == nil {
		log.Error("cannot load completion suggestion without a source")
		return nil
	}

	if showLoading {
		spinnerTickCmd = c.fzfSelect.SetFetchLoading()
	}

	fetchCmd := func() tea.Msg {
		err := c.fzfSelect.Source.LoadSuggestions(
			fuzzyselect.LoaderContext{RepoOwner: c.repo.Owner, RepoName: c.repo.Name},
		)
		if err != nil {
			return SourceFetchFailedMsg{Err: err}
		}
		return SourceDataFetchedMsg{}
	}

	if spinnerTickCmd != nil {
		return tea.Batch(spinnerTickCmd, fetchCmd)
	}
	return fetchCmd
}
