package issuesidebar

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/ui/common"
	"github.com/dlvhdr/gh-dash/v4/ui/components/inputbox"
	"github.com/dlvhdr/gh-dash/v4/ui/components/issue"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/ui/markdown"
)

type Model struct {
	ctx       *context.ProgramContext
	issue     *issue.Issue
	sectionId int
	width     int

	isCommenting  bool
	isAssigning   bool
	isUnassigning bool

	inputBox inputbox.Model
}

func NewModel(ctx *context.ProgramContext) Model {
	inputBox := inputbox.NewModel(ctx)
	inputBox.SetHeight(common.InputBoxHeight)

	return Model{
		issue: nil,

		isCommenting:  false,
		isAssigning:   false,
		isUnassigning: false,

		inputBox: inputBox,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var (
		cmds  []tea.Cmd
		cmd   tea.Cmd
		taCmd tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.isCommenting {
			switch msg.Type {

			case tea.KeyCtrlD:
				if len(strings.Trim(m.inputBox.Value(), " ")) != 0 {
					cmd = m.comment(m.inputBox.Value())
				}
				m.inputBox.Blur()
				m.isCommenting = false
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isCommenting = false
				return m, nil
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
		} else if m.isAssigning {
			switch msg.Type {

			case tea.KeyCtrlD:
				usernames := strings.Fields(m.inputBox.Value())
				if len(usernames) > 0 {
					cmd = m.assign(usernames)
				}
				m.inputBox.Blur()
				m.isAssigning = false
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isAssigning = false
				return m, nil
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
		} else if m.isUnassigning {
			switch msg.Type {

			case tea.KeyCtrlD:
				usernames := strings.Fields(m.inputBox.Value())
				if len(usernames) > 0 {
					cmd = m.unassign(usernames)
				}
				m.inputBox.Blur()
				m.isUnassigning = false
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				m.inputBox.Blur()
				m.isUnassigning = false
				return m, nil
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
		} else {
			return m, nil
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	s := strings.Builder{}

	s.WriteString(m.renderFullNameAndNumber())
	s.WriteString("\n")

	s.WriteString(m.renderTitle())
	s.WriteString("\n\n")
	s.WriteString(m.renderStatusPill())
	s.WriteString("\n\n")

	labels := m.renderLabels()
	if labels != "" {
		s.WriteString(labels)
		s.WriteString("\n\n")
	}

	s.WriteString(m.renderBody())
	s.WriteString("\n\n")
	s.WriteString(m.renderActivity())

	if m.isCommenting || m.isAssigning || m.isUnassigning {
		s.WriteString(m.inputBox.View())
	}

	return s.String()
}

func (m *Model) renderFullNameAndNumber() string {
	return lipgloss.NewStyle().
		Foreground(m.ctx.Theme.SecondaryText).
		Render(fmt.Sprintf("#%d · %s", m.issue.Data.GetNumber(), m.issue.Data.GetRepoNameWithOwner()))
}

func (m *Model) renderTitle() string {
	return m.ctx.Styles.Common.MainTextStyle.Width(m.getIndentedContentWidth()).
		Render(m.issue.Data.Title)
}

func (m *Model) renderStatusPill() string {
	bgColor := ""
	content := ""
	switch m.issue.Data.State {
	case "OPEN":
		bgColor = m.ctx.Styles.Colors.OpenIssue.Dark
		content = " Open"
	case "CLOSED":
		bgColor = m.ctx.Styles.Colors.ClosedIssue.Dark
		content = " Closed"
	}

	return m.ctx.Styles.PrSidebar.PillStyle.
		Background(lipgloss.Color(bgColor)).
		Render(content)
}

func (m *Model) renderBody() string {
	width := m.getIndentedContentWidth()
	regex := regexp.MustCompile("(?U)<!--(.|[[:space:]])*-->")
	body := regex.ReplaceAllString(m.issue.Data.Body, "")

	regex = regexp.MustCompile(`((\n)+|^)([^\r\n]*\|[^\r\n]*(\n)?)+`)
	body = regex.ReplaceAllString(body, "")

	body = strings.TrimSpace(body)
	if body == "" {
		return lipgloss.NewStyle().Italic(true).Render("No description provided.")
	}

	markdownRenderer := markdown.GetMarkdownRenderer(width)
	rendered, err := markdownRenderer.Render(body)
	if err != nil {
		return ""
	}

	return lipgloss.NewStyle().
		Width(width).
		MaxWidth(width).
		Align(lipgloss.Left).
		Render(rendered)
}

func (m *Model) renderLabels() string {
	width := m.getIndentedContentWidth()
	labels := m.issue.Data.Labels.Nodes
	style := m.ctx.Styles.PrSidebar.PillStyle

	return common.RenderLabels(width, labels, style)
}

func (m *Model) getIndentedContentWidth() int {
	return m.width - 6
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.inputBox.SetWidth(width)
}

func (m *Model) SetSectionId(id int) {
	m.sectionId = id
}

func (m *Model) SetRow(data *data.IssueData) {
	if data == nil {
		m.issue = nil
	} else {
		m.issue = &issue.Issue{Ctx: m.ctx, Data: *data}
	}
}

func (m *Model) IsTextInputBoxFocused() bool {
	return m.isCommenting || m.isAssigning || m.isUnassigning
}

func (m *Model) GetIsCommenting() bool {
	return m.isCommenting
}

func (m *Model) SetIsCommenting(isCommenting bool) tea.Cmd {
	if !m.isCommenting && isCommenting {
		m.inputBox.Reset()
	}
	m.isCommenting = isCommenting
	m.inputBox.SetPrompt("Leave a comment...")

	if isCommenting {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) GetIsAssigning() bool {
	return m.isAssigning
}

func (m *Model) SetIsAssigning(isAssigning bool) tea.Cmd {
	if !m.isAssigning && isAssigning {
		m.inputBox.Reset()
	}
	m.isAssigning = isAssigning
	m.inputBox.SetPrompt("Assign users (whitespace-separated)...")
	if !m.userAssignedToIssue(m.ctx.User) {
		m.inputBox.SetValue(m.ctx.User)
	}

	if isAssigning {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) userAssignedToIssue(login string) bool {
	for _, a := range m.issue.Data.Assignees.Nodes {
		if login == a.Login {
			return true
		}
	}
	return false
}

func (m *Model) GetIsUnassigning() bool {
	return m.isUnassigning
}

func (m *Model) SetIsUnassigning(isUnassigning bool) tea.Cmd {
	if !m.isUnassigning && isUnassigning {
		m.inputBox.Reset()
	}
	m.isUnassigning = isUnassigning
	m.inputBox.SetPrompt("Unassign users (whitespace-separated)...")
	m.inputBox.SetValue(strings.Join(m.issueAssignees(), "\n"))

	if isUnassigning {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) issueAssignees() []string {
	var assignees []string
	for _, n := range m.issue.Data.Assignees.Nodes {
		assignees = append(assignees, n.Login)
	}
	return assignees
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.inputBox.UpdateProgramContext(ctx)
}
