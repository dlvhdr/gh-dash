package inputbox

import (
	"fmt"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/cmp"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx *context.ProgramContext
	// text area is for multiline inputs
	textArea *textarea.Model
	// text area is for single line inputs
	textInput *textinput.Model
	inputHelp help.Model
	prompt    string
	cmp       *cmp.Model
	cmpSource cmp.Source
}

var inputKeys = []key.Binding{
	key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("Ctrl+d", "submit")),
	key.NewBinding(key.WithKeys("ctrl+c", "esc"), key.WithHelp("Ctrl+c/esc", "cancel")),
	cmp.ToggleSuggestions,
}

const DefaultInputHeight = 5

func DefaultTextArea(ctx *context.ProgramContext) textarea.Model {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.SetHeight(DefaultInputHeight)
	ta.CharLimit = 65536
	base := lipgloss.NewStyle()
	ta.SetStyles(textarea.Styles{
		Focused: textarea.StyleState{
			Base:       base,
			Text:       base.Foreground(ctx.Theme.PrimaryText),
			LineNumber: base.Foreground(ctx.Theme.FaintText),
			CursorLine: base.Background(ctx.Theme.FaintBorder).
				Foreground(ctx.Theme.PrimaryText),
			CursorLineNumber: base.Foreground(ctx.Theme.SecondaryText),
			Placeholder:      base.Foreground(ctx.Theme.FaintText),
			EndOfBuffer:      base.Foreground(ctx.Theme.FaintText),
		},
	})
	return ta
}

func DefaultTextInput(ctx *context.ProgramContext) textinput.Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 65536
	base := lipgloss.NewStyle()
	ti.SetStyles(textinput.Styles{
		Focused: textinput.StyleState{
			Text:        base.Foreground(ctx.Theme.PrimaryText),
			Placeholder: base.Foreground(ctx.Theme.FaintText),
		},
	})
	return ti
}

type ModelOpts struct {
	TextArea  *textarea.Model
	TextInput *textinput.Model
}

func NewModel(ctx *context.ProgramContext, opts ModelOpts) Model {
	if opts.TextArea != nil {
		opts.TextArea.Focus()
	}
	if opts.TextInput != nil {
		opts.TextInput.Focus()
	}

	h := help.New()
	h.Styles = ctx.Styles.Help.BubbleStyles
	return Model{
		ctx:       ctx,
		textArea:  opts.TextArea,
		textInput: opts.TextInput,
		inputHelp: h,
		prompt:    "",
	}
}

func (m *Model) SetAutocomplete(cmp *cmp.Model) {
	m.cmp = cmp
}

func (m *Model) SetAutocompleteSource(src cmp.Source) {
	m.cmpSource = src
}

func (m Model) CurrentAutocompleteContext() cmp.Context {
	if m.cmpSource == nil {
		return cmp.Context{}
	}

	return m.cmpSource.ExtractContext(m.Value(), m.GetAbsoluteCursorPosition())
}

func (m Model) AutocompleteItemsToExclude() []string {
	if m.cmpSource == nil {
		return nil
	}

	return m.cmpSource.ItemsToExclude(m.Value(), m.GetAbsoluteCursorPosition())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Allow toggling suggestions at any time
		if m.cmp != nil && key.Matches(msg, cmp.ToggleSuggestions) {
			if m.cmp.IsVisible() {
				m.cmp.Suppress()
				return m, nil
			}

			m.cmp.Unsuppress()
			currentContext := m.CurrentAutocompleteContext()
			m.cmp.Show(currentContext.Content, m.AutocompleteItemsToExclude())
			return m, nil
		}

		// Allow navigation/selection even if the popup is hidden (as long as there are filtered results)
		if m.cmp != nil &&
			(m.cmp.IsVisible() || m.cmp.HasSuggestions()) {
			switch {
			case key.Matches(msg, cmp.PrevKey):
				m.cmp.Prev()
				return m, nil
			case key.Matches(msg, cmp.NextKey):
				m.cmp.Next()
				return m, nil
			case m.cmp.Selected() != "" && key.Matches(msg, cmp.SelectKey):
				selected := m.cmp.Selected()
				if selected != "" && m.cmpSource != nil {
					currentValue := m.Value()
					currentContext := m.CurrentAutocompleteContext()
					newValue, newCursorPos := m.cmpSource.InsertSuggestion(
						currentValue,
						selected,
						currentContext.Start,
						currentContext.End,
					)
					m.SetValue(newValue)
					m.SetCursorColumn(newCursorPos.X)
					// Refresh cmp to exclude the newly-added item
					if m.AutocompleteItemsToExclude() != nil {
						newContext := m.CurrentAutocompleteContext()
						m.cmp.Show(newContext.Content, m.AutocompleteItemsToExclude())
					}
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	if m.textArea != nil {
		ta, taCmd := m.textArea.Update(msg)
		m.textArea = &ta
		cmd = taCmd
	}
	if m.textInput != nil {
		ti, tiCmd := m.textInput.Update(msg)
		m.textInput = &ti
		cmd = tiCmd
	}
	return m, cmd
}

func (m Model) View() string {
	content := ""
	if m.textInput != nil {
		content = m.textInput.View()
	} else {
		content = m.textArea.View()
	}
	if m.prompt != "" {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			fmt.Sprintf("%s\n", m.prompt),
			content,
			lipgloss.NewStyle().
				MarginTop(1).
				Render(m.inputHelp.ShortHelpView(inputKeys)),
		)
	}

	return content
}

func (m Model) ViewCompletions() string {
	if m.cmp == nil || !m.cmp.IsVisible() {
		return ""
	}

	return m.cmp.View()
}

func (m *Model) Value() string {
	if m.textInput != nil {
		return m.textInput.Value()
	}
	return m.textArea.Value()
}

func (m *Model) SetValue(s string) {
	if m.textInput != nil {
		m.textInput.SetValue(s)
		return
	}
	m.textArea.SetValue(s)
}

func (m *Model) SetCursorColumn(col int) {
	if m.textInput != nil {
		m.textInput.SetCursor(col)
		return
	}
	m.textArea.SetCursorColumn(col)
}

func (m *Model) Blur() {
	if m.textArea != nil {
		m.textArea.Blur()
		return
	}

	m.textInput.Blur()
}

func (m *Model) Focus() tea.Cmd {
	if m.textArea != nil {
		return m.textArea.Focus()
	}

	return m.textInput.Focus()
}

func (m *Model) Focused() bool {
	if m.textArea != nil {
		return m.textArea.Focused()
	}

	return m.textInput.Focused()
}

func (m *Model) Width() int {
	if m.textArea != nil {
		return m.textArea.Width()
	}

	return m.textInput.Width()
}

func (m *Model) SetWidth(width int) {
	if m.textArea != nil {
		m.textArea.SetWidth(width)
		return
	}

	w := max(1, width-lipgloss.Width(m.textInput.Prompt)-1)
	m.textInput.SetWidth(w)
}

func (m *Model) SetHeight(height int) {
	if m.textArea != nil {
		m.textArea.SetHeight(height)
		return
	}
}

func (m *Model) SetPrompt(prompt string) {
	m.prompt = prompt
}

func (m *Model) Reset() {
	if m.textArea != nil {
		m.textArea.Reset()
		return
	}

	m.textInput.Reset()
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.inputHelp.Styles = ctx.Styles.Help.BubbleStyles
}

func (m *Model) GetAbsoluteCursorPosition() tea.Position {
	if m.textInput != nil {
		return tea.Position{X: m.textInput.Position(), Y: 0}
	}
	line := m.textArea.Line()
	col := m.textArea.Column()
	return tea.Position{X: col, Y: line}
}

func (m *Model) CursorStart() {
	if m.textArea != nil {
		m.textArea.CursorStart()
		return
	}

	m.textInput.CursorStart()
}

func (m *Model) CursorEnd() {
	if m.textArea != nil {
		m.textArea.CursorEnd()
		return
	}

	m.textInput.CursorEnd()
}

func (m *Model) Column() int {
	if m.textArea != nil {
		return m.textArea.Column()
	}

	return m.textInput.Position()
}

func (m *Model) LineFromBottom() int {
	if m.textInput != nil {
		return 0
	}
	if m.textArea.LineCount() < m.textArea.Height() {
		return m.textArea.Height() - m.textArea.LineCount()
	}

	return m.textArea.LineCount() - m.textArea.Line() - 1
}
