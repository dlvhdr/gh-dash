package prsidebar

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/commentbox"
	"github.com/dlvhdr/gh-dash/ui/components/pr"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/markdown"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

type Model struct {
	ctx          *context.ProgramContext
	sectionId    int
	pr           *pr.PullRequest
	width        int
	isCommenting bool
	commentBox   commentbox.Model
	commentHelp  help.Model
}

func NewModel() Model {
	commentBox := commentbox.NewModel()

	return Model{
		pr:           nil,
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
	s.WriteString("\n")
	s.WriteString(m.renderBranches())
	s.WriteString("\n\n")
	s.WriteString(m.renderPills())
	s.WriteString("\n\n")
	s.WriteString(m.renderDescription())
	s.WriteString("\n\n")
	s.WriteString(m.renderChecks())
	s.WriteString("\n\n")
	s.WriteString(m.renderActivity())

	if m.isCommenting {
		s.WriteString(m.commentBox.View())
	}

	return s.String()
}

func (m *Model) renderTitle() string {
	return styles.MainTextStyle.Copy().Width(m.getIndentedContentWidth()).
		Render(m.pr.Data.Title)
}

func (m *Model) renderBranches() string {
	return lipgloss.NewStyle().
		Foreground(styles.DefaultTheme.SecondaryText).
		Render(m.pr.Data.BaseRefName + "  " + m.pr.Data.HeadRefName)
}

func (m *Model) renderStatusPill() string {
	bgColor := ""
	switch m.pr.Data.State {
	case "OPEN":
		if m.pr.Data.IsDraft {
			bgColor = styles.DefaultTheme.FaintText.Dark
		} else {
			bgColor = openPR.Dark
		}
	case "CLOSED":
		bgColor = closedPR.Dark
	case "MERGED":
		bgColor = mergedPR.Dark
	}

	return pillStyle.
		Background(lipgloss.Color(bgColor)).
		Render(m.pr.RenderState())
}

func (m *Model) renderMergeablePill() string {
	status := m.pr.Data.Mergeable
	if status == "CONFLICTING" {
		return pillStyle.Copy().
			Background(styles.DefaultTheme.WarningText).
			Render(" Merge Conflicts")
	} else if status == "MERGEABLE" {
		return pillStyle.Copy().
			Background(styles.DefaultTheme.SuccessText).
			Render(" Mergeable")
	}

	return ""
}

func (m *Model) renderChecksPill() string {
	status := m.pr.GetStatusChecksRollup()
	if status == "FAILURE" {
		return pillStyle.Copy().
			Background(styles.DefaultTheme.WarningText).
			Render(" Checks")
	} else if status == "PENDING" {
		return pillStyle.Copy().
			Background(styles.DefaultTheme.FaintText).
			Foreground(styles.DefaultTheme.PrimaryText).
			Faint(true).
			Render(" Checks")
	}

	return pillStyle.Copy().
		Background(styles.DefaultTheme.SuccessText).
		Foreground(styles.DefaultTheme.InvertedText).
		Render(" Checks")
}

func (m *Model) renderPills() string {
	statusPill := m.renderStatusPill()
	mergeablePill := m.renderMergeablePill()
	checksPill := m.renderChecksPill()
	return lipgloss.JoinHorizontal(lipgloss.Top, statusPill, " ", mergeablePill, " ", checksPill)
}

func (m *Model) renderDescription() string {
	width := m.getIndentedContentWidth()
	regex := regexp.MustCompile("(?U)<!--(.|[[:space:]])*-->")
	body := regex.ReplaceAllString(m.pr.Data.Body, "")

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

func (m *Model) SetSectionId(id int) {
	m.sectionId = id
}

func (m *Model) SetRow(data *data.PullRequestData) {
	if data == nil {
		m.pr = nil
	} else {
		m.pr = &pr.PullRequest{Data: *data}
	}
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.commentBox.SetWidth(width)
}

func (m *Model) GetIsCommenting() bool {
	return m.isCommenting
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
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

func (m *Model) getIndentedContentWidth() int {
	return m.width - 6
}
