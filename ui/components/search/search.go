package search

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

type Model struct {
	sectionId    int
	sectionType  string
	InitialValue string
	textInput    textinput.Model
}

func NewModel(sectionId int, sectionType string, ctx context.ProgramContext, initialValue string) Model {
	prompt := " is:pr "
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.Width = ctx.MainContentWidth - lipgloss.Width(prompt) - 6
	ti.PromptStyle = ti.PromptStyle.Copy().Foreground(styles.DefaultTheme.SecondaryText)
	ti.Prompt = prompt
	ti.TextStyle = ti.TextStyle.Copy().Faint(true)
	ti.Blur()
	ti.SetValue(initialValue)
	ti.CursorStart()

	return Model{
		textInput:    ti,
		InitialValue: initialValue,
		sectionId:    sectionId,
		sectionType:  sectionType,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, m.submitSearch
		case tea.KeyCtrlC, tea.KeyEsc:
			m.textInput.SetValue(m.InitialValue)
			return m, m.cancelSearch
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View(ctx context.ProgramContext) string {
	return lipgloss.NewStyle().
		Width(ctx.MainContentWidth - 4).
		MaxHeight(3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.DefaultTheme.PrimaryBorder).
		Render(m.textInput.View())
}

func (m *Model) Focus() {
	m.textInput.TextStyle = m.textInput.TextStyle.Copy().Faint(false)
	m.textInput.CursorEnd()
	m.textInput.Focus()
}

func (m *Model) Blur() {
	m.textInput.TextStyle = m.textInput.TextStyle.Copy().Faint(true)
	m.textInput.CursorStart()
	m.textInput.Blur()
}

func (m *Model) ResetValue() {
	m.textInput.SetValue(m.InitialValue)
}

type SearchCancelled struct {
	sectionId   int
	sectionType string
}

func (sc SearchCancelled) GetSectionId() int {
	return sc.sectionId
}
func (sc SearchCancelled) GetSectionType() string {
	return sc.sectionType
}

func (m Model) cancelSearch() tea.Msg {
	return SearchCancelled{
		sectionId:   m.sectionId,
		sectionType: m.sectionType,
	}
}

type SearchSubmitted struct {
	sectionId   int
	sectionType string
	Term        string
}

func (sc SearchSubmitted) GetSectionId() int {
	return sc.sectionId
}
func (sc SearchSubmitted) GetSectionType() string {
	return sc.sectionType
}

func (m Model) submitSearch() tea.Msg {
	return SearchSubmitted{
		sectionId:   m.sectionId,
		sectionType: m.sectionType,
		Term:        m.textInput.Value(),
	}
}
