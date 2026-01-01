package inputbox

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/autocomplete"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type Model struct {
	ctx          *context.ProgramContext
	textArea     textarea.Model
	inputHelp    help.Model
	prompt       string
	autocomplete *autocomplete.Model
}

// LabelInfo contains information about a label at a specific cursor position
// in a comma-separated list of labels.
type LabelInfo struct {
	Label    string // The trimmed label text
	StartIdx int    // Start position in original string (inclusive)
	EndIdx   int    // End position in original string (exclusive)
	IsFirst  bool   // First label in list
	IsLast   bool   // Last label in list
}

var inputKeys = []key.Binding{
	key.NewBinding(key.WithKeys(tea.KeyCtrlD.String()), key.WithHelp("Ctrl+d", "submit")),
	key.NewBinding(key.WithKeys(tea.KeyCtrlC.String(), tea.KeyEsc.String()), key.WithHelp("Ctrl+c/esc", "cancel")),
}

func NewModel(ctx *context.ProgramContext) Model {
	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.Prompt = ""
	ta.CharLimit = 65536
	ta.FocusedStyle.Base = lipgloss.NewStyle()
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().
		Background(ctx.Theme.FaintBorder).
		Foreground(ctx.Theme.PrimaryText)
	ta.FocusedStyle.LineNumber = lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
	ta.FocusedStyle.CursorLineNumber = lipgloss.NewStyle().Foreground(ctx.Theme.SecondaryText)
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
	ta.FocusedStyle.Text = lipgloss.NewStyle().Foreground(ctx.Theme.PrimaryText)
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle().Foreground(ctx.Theme.FaintText)
	ta.Focus()

	h := help.New()
	h.Styles = ctx.Styles.Help.BubbleStyles
	return Model{
		ctx:       ctx,
		textArea:  ta,
		inputHelp: h,
		prompt:    "",
	}
}

func (m *Model) SetAutocomplete(ac *autocomplete.Model) {
	m.autocomplete = ac
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.autocomplete != nil && m.autocomplete.IsVisible() {
			switch {
			case key.Matches(msg, autocomplete.PrevKey):
				m.autocomplete.Prev()
				return m, nil
			case key.Matches(msg, autocomplete.NextKey):
				m.autocomplete.Next()
				return m, nil
			case key.Matches(msg, autocomplete.SelectKey):
				selected := m.autocomplete.Selected()
				if selected != "" {
					currentInput := m.textArea.Value()
					cursorPos := m.GetCursorPosition()
					labelInfo := extractLabelAtCursor(currentInput, cursorPos)

					// Build replacement with consistent spacing:
					// - Single space after comma (before label) for non-first labels
					// - Always add ", " after the label for easy continuation
					var replacement string
					if labelInfo.IsFirst {
						replacement = selected + ", "
					} else {
						replacement = " " + selected + ", "
					}

					// Determine what comes after the current label
					// Skip existing comma and spaces if present
					remainingInput := currentInput[labelInfo.EndIdx:]
					if strings.HasPrefix(remainingInput, ",") {
						// Skip the comma
						remainingInput = remainingInput[1:]
						// Skip any spaces after the comma
						remainingInput = strings.TrimLeft(remainingInput, " \t")
					}

					// Build new input by replacing the label at cursor position
					newInput := currentInput[:labelInfo.StartIdx] + replacement + remainingInput
					m.textArea.SetValue(newInput)

					// Position cursor after the ", " we added
					newCursorPos := labelInfo.StartIdx + len(replacement)
					m.textArea.SetCursor(newCursorPos)
				}
				m.autocomplete.Hide()
				return m, nil
			}
		}
	}

	m.textArea, cmd = m.textArea.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return lipgloss.NewStyle().
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(m.ctx.Theme.SecondaryBorder).
		MarginTop(1).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				fmt.Sprintf("%s\n", m.prompt),
				m.textArea.View(),
				lipgloss.NewStyle().
					MarginTop(1).
					Render(m.inputHelp.ShortHelpView(inputKeys)),
			),
		)
}

func (m Model) ViewWithAutocomplete() string {
	baseView := m.View()
	autocompleteView := ""
	if m.autocomplete != nil {
		autocompleteView = m.autocomplete.View()
	}
	return baseView + "\n" + autocompleteView
}

func (m *Model) Value() string {
	return m.textArea.Value()
}

func (m *Model) SetValue(s string) {
	m.textArea.SetValue(s)
}

func (m *Model) Blur() {
	m.textArea.Blur()
}

func (m *Model) Focus() tea.Cmd {
	return m.textArea.Focus()
}

func (m *Model) SetWidth(width int) {
	m.textArea.SetWidth(width)
}

func (m *Model) SetHeight(height int) {
	m.textArea.SetHeight(height)
}

func (m *Model) SetPrompt(prompt string) {
	m.prompt = prompt
}

func (m *Model) Reset() {
	m.textArea.Reset()
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.inputHelp.Styles = ctx.Styles.Help.BubbleStyles
}

func (m *Model) GetCursorPosition() int {
	lineInfo := m.textArea.LineInfo()
	return lineInfo.StartColumn + lineInfo.CharOffset
}

// extractLabelAtCursor extracts information about the label at the given cursor position
// in a comma-separated list. It considers the entire word containing the cursor as the
// current label.
func extractLabelAtCursor(input string, cursorPos int) LabelInfo {
	if input == "" {
		return LabelInfo{
			Label:    "",
			StartIdx: 0,
			EndIdx:   0,
			IsFirst:  true,
			IsLast:   true,
		}
	}

	// Clamp cursor position to valid range
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(input) {
		cursorPos = len(input)
	}

	// Find the comma before the cursor (or start of string)
	startIdx := 0
	for i := cursorPos - 1; i >= 0; i-- {
		if input[i] == ',' {
			startIdx = i + 1
			break
		}
	}

	// Find the comma after the cursor (or end of string)
	endIdx := len(input)
	for i := cursorPos; i < len(input); i++ {
		if input[i] == ',' {
			endIdx = i
			break
		}
	}

	// Extract and trim the label
	label := strings.TrimSpace(input[startIdx:endIdx])

	// Determine if this is the first or last label
	isFirst := startIdx == 0
	isLast := endIdx == len(input)

	return LabelInfo{
		Label:    label,
		StartIdx: startIdx,
		EndIdx:   endIdx,
		IsFirst:  isFirst,
		IsLast:   isLast,
	}
}

func extractCurrentLabel(input string) string {
	lastComma := strings.LastIndex(input, ",")
	if lastComma == -1 {
		return input
	}
	return strings.TrimSpace(input[lastComma+1:])
}

func (m *Model) GetCurrentLabel() string {
	cursorPos := m.GetCursorPosition()
	labelInfo := extractLabelAtCursor(m.textArea.Value(), cursorPos)
	return labelInfo.Label
}

func (m *Model) GetAllLabels() []string {
	value := m.textArea.Value()
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			labels = append(labels, trimmed)
		}
	}
	return labels
}
