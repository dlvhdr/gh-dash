package issuesidebar

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/issue"
	"github.com/dlvhdr/gh-dash/ui/markdown"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

type Model struct {
	issue *issue.Issue
	width int
}

func NewModel(data *data.IssueData, width int) Model {
	var s *issue.Issue
	if data == nil {
		s = nil
	} else {
		s = &issue.Issue{Data: *data}
	}
	return Model{
		issue: s,
		width: width,
	}
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
	s.WriteString("\n\n")

	return s.String()
}

func (m *Model) renderTitle() string {
	return styles.MainTextStyle.Copy().Width(m.getIndentedContentWidth()).
		Render(m.issue.Data.Title)
}

func (m *Model) renderStatusPill() string {
	bgColor := ""
	content := ""
	switch m.issue.Data.State {
	case "OPEN":
		bgColor = issue.OpenIssue.Dark
		content = " Open"
	case "CLOSED":
		bgColor = issue.ClosedIssue.Dark
		content = " Closed"
	}

	return pillStyle.
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
			pillStyle.
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
