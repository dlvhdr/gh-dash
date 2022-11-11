package issuesidebar

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/common"
	"github.com/dlvhdr/gh-dash/ui/components/commentbox"
	"github.com/dlvhdr/gh-dash/ui/components/issue"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/markdown"
)

type Model struct {
	ctx          *context.ProgramContext
	issue        *issue.Issue
	sectionId    int
	width        int
	isCommenting bool
	commentBox   commentbox.Model
}

func NewModel(ctx context.ProgramContext) Model {
	commentBox := commentbox.NewModel(ctx)
	commentBox.SetHeight(common.CommentBoxHeight)

	return Model{
		issue:        nil,
		isCommenting: false,
		commentBox:   commentBox,
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

		if !m.isCommenting {
			return m, nil
		}

		switch msg.Type {

		case tea.KeyCtrlD:
			if len(strings.Trim(m.commentBox.Value(), " ")) != 0 {
				cmd = m.comment(m.commentBox.Value())
			}
			m.commentBox.Blur()
			m.isCommenting = false
			return m, cmd

		case tea.KeyEsc, tea.KeyCtrlC:
			m.commentBox.Blur()
			m.isCommenting = false
			return m, nil
		}
	}

	m.commentBox, taCmd = m.commentBox.Update(msg)
	cmds = append(cmds, cmd, taCmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	s := strings.Builder{}
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

	if m.isCommenting {
		s.WriteString(m.commentBox.View())
	}

	return s.String()
}

func (m *Model) renderTitle() string {
	return m.ctx.Styles.Common.MainTextStyle.Copy().Width(m.getIndentedContentWidth()).
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

	return m.ctx.Styles.PrSidebar.PillStyle.Copy().
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
	labels := make([]string, 0, len(m.issue.Data.Labels.Nodes))
	for _, label := range m.issue.Data.Labels.Nodes {
		labels = append(
			labels,
			m.ctx.Styles.PrSidebar.PillStyle.Copy().
				Background(lipgloss.Color("#"+label.Color)).
				Render(label.Name),
		)
		labels = append(labels, " ")
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, labels...)
}

func (m *Model) getIndentedContentWidth() int {
	return m.width - 6
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.commentBox.SetWidth(width)
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

func (m *Model) GetIsCommenting() bool {
	return m.isCommenting
}

func (m *Model) SetIsCommenting(isCommenting bool) tea.Cmd {
	if m.isCommenting == false && isCommenting == true {
		m.commentBox.Reset()
	}
	m.isCommenting = isCommenting

	if isCommenting == true {
		return tea.Sequentially(textarea.Blink, m.commentBox.Focus())
	}
	return nil
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}
