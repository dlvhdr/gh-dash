package prsidebar

import (
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/common"
	"github.com/dlvhdr/gh-dash/ui/components/commentbox"
	"github.com/dlvhdr/gh-dash/ui/components/pr"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/markdown"
)

type Model struct {
	ctx          *context.ProgramContext
	sectionId    int
	pr           *pr.PullRequest
	width        int
	isCommenting bool
	commentBox   commentbox.Model
}

func NewModel(ctx context.ProgramContext) Model {
	commentBox := commentbox.NewModel(ctx)
	commentBox.SetHeight(common.CommentBoxHeight)

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
	return m.ctx.Styles.Common.MainTextStyle.Copy().Width(m.getIndentedContentWidth()).
		Render(m.pr.Data.Title)
}

func (m *Model) renderBranches() string {
	return lipgloss.NewStyle().
		Foreground(m.ctx.Theme.SecondaryText).
		Render(m.pr.Data.BaseRefName + "  " + m.pr.Data.HeadRefName)
}

func (m *Model) renderStatusPill() string {
	bgColor := ""
	switch m.pr.Data.State {
	case "OPEN":
		if m.pr.Data.IsDraft {
			bgColor = m.ctx.Theme.FaintText.Dark
		} else {
			bgColor = m.ctx.Styles.Colors.OpenPR.Dark
		}
	case "CLOSED":
		bgColor = m.ctx.Styles.Colors.ClosedPR.Dark
	case "MERGED":
		bgColor = m.ctx.Styles.Colors.MergedPR.Dark
	}

	return m.ctx.Styles.PrSidebar.PillStyle.
		Background(lipgloss.Color(bgColor)).
		Render(m.pr.RenderState())
}

func (m *Model) renderMergeablePill() string {
	status := m.pr.Data.Mergeable
	if status == "CONFLICTING" {
		return m.ctx.Styles.PrSidebar.PillStyle.Copy().
			Background(m.ctx.Theme.WarningText).
			Render(" Merge Conflicts")
	} else if status == "MERGEABLE" {
		return m.ctx.Styles.PrSidebar.PillStyle.Copy().
			Background(m.ctx.Theme.SuccessText).
			Render(" Mergeable")
	}

	return ""
}

func (m *Model) renderChecksPill() string {
	s := m.ctx.Styles.PrSidebar.PillStyle
	t := m.ctx.Theme

	status := m.pr.GetStatusChecksRollup()
	if status == "FAILURE" {
		return s.Copy().
			Background(t.WarningText).
			Render(" Checks")
	} else if status == "PENDING" {
		return s.Copy().
			Background(t.FaintText).
			Foreground(t.PrimaryText).
			Faint(true).
			Render(" Checks")
	}

	return s.Copy().
		Background(t.SuccessText).
		Foreground(t.InvertedText).
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
		m.pr = &pr.PullRequest{Ctx: m.ctx, Data: *data}
	}
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.commentBox.SetWidth(m.width)
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
	return m.width - 4
}
