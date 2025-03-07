package prsidebar

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
	"github.com/dlvhdr/gh-dash/v4/ui/components/pr"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/ui/markdown"
)

var (
	htmlCommentRegex = regexp.MustCompile("(?U)<!--(.|[[:space:]])*-->")
	lineCleanupRegex = regexp.MustCompile(`((\n)+|^)([^\r\n]*\|[^\r\n]*(\n)?)+`)
	commentPrompt    = "Leave a comment..."
	approvalPrompt   = "Approve with comment..."
)

type Model struct {
	ctx       *context.ProgramContext
	sectionId int
	pr        *pr.PullRequest
	width     int

	ShowConfirmCancel bool
	isCommenting      bool
	isApproving       bool
	isAssigning       bool
	isUnassigning     bool

	inputBox inputbox.Model
}

func NewModel(ctx *context.ProgramContext) Model {
	inputBox := inputbox.NewModel(ctx)
	inputBox.SetHeight(common.InputBoxHeight)

	return Model{
		pr: nil,

		isCommenting:  false,
		isApproving:   false,
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
				if !m.ShowConfirmCancel {
					m.shouldCancelComment()
				}
			default:
				if msg.String() == "Y" || msg.String() == "y" {
					if m.shouldCancelComment() {
						return m, nil
					}
				}
				if m.ShowConfirmCancel && (msg.String() == "N" || msg.String() == "n") {
					m.inputBox.SetPrompt(commentPrompt)
					m.ShowConfirmCancel = false
					return m, nil
				}
				m.inputBox.SetPrompt(commentPrompt)
				m.ShowConfirmCancel = false
			}

			m.inputBox, taCmd = m.inputBox.Update(msg)
			cmds = append(cmds, cmd, taCmd)
		} else if m.isApproving {
			switch msg.Type {

			case tea.KeyCtrlD:
				comment := ""
				if len(strings.Trim(m.inputBox.Value(), " ")) != 0 {
					comment = m.inputBox.Value()
				}
				cmd = m.approve(comment)
				m.inputBox.Blur()
				m.isApproving = false
				return m, cmd

			case tea.KeyEsc, tea.KeyCtrlC:
				if m.shouldCancelComment() {
					return m, nil
				}
			default:
				m.inputBox.SetPrompt(approvalPrompt)
				m.ShowConfirmCancel = false
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
	s.WriteString(m.renderUrl())
	s.WriteString("\n")

	s.WriteString(m.renderTitle())
	s.WriteString("\n\n")
	s.WriteString(m.renderBranches())

	labels := m.renderLabels()
	if labels != "" {
		s.WriteString("\n\n")
		s.WriteString(labels)
	}

	s.WriteString("\n")
	s.WriteString(lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintBorder).Render(strings.Repeat("―", m.width)))
	s.WriteString("\n\n")

	s.WriteString(m.renderDescription())
	s.WriteString("\n\n")
	s.WriteString(m.renderChecks())

	s.WriteString("\n\n")
	s.WriteString(m.renderActivity())

	if m.isCommenting || m.isApproving || m.isAssigning || m.isUnassigning {
		s.WriteString(m.inputBox.View())
	}

	return s.String()
}

func (m *Model) renderFullNameAndNumber() string {
	return lipgloss.NewStyle().Foreground(m.ctx.Theme.SecondaryText).Render(fmt.Sprintf("#%d · %s", m.pr.Data.GetNumber(), m.pr.Data.GetRepoNameWithOwner()))
}

func (m *Model) renderUrl() string {
	return lipgloss.NewStyle().
		Foreground(m.ctx.Theme.SecondaryText).
		Render(m.pr.Data.GetUrl())
}

func (m *Model) renderTitle() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.ctx.Styles.Common.MainTextStyle.Width(m.getIndentedContentWidth()).Render(m.pr.Data.Title),
	)
}

func (m *Model) renderBranches() string {
	return lipgloss.JoinHorizontal(lipgloss.Left,
		m.renderStatusPill(),
		" ",
		lipgloss.NewStyle().
			Foreground(m.ctx.Theme.SecondaryText).
			Render(m.pr.Data.BaseRefName+"  "+m.pr.Data.HeadRefName))
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
		BorderForeground(lipgloss.Color(bgColor)).
		Background(lipgloss.Color(bgColor)).
		Render(m.pr.RenderState())
}

func (m *Model) renderLabels() string {
	width := m.getIndentedContentWidth()
	labels := m.pr.Data.Labels.Nodes
	style := m.ctx.Styles.PrSidebar.PillStyle
	if len(labels) == 0 {
		return ""
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).Underline(true).Render("Labels"),
		"",
		common.RenderLabels(width, labels, style),
	)
}

func (m *Model) renderDescription() string {
	width := m.getIndentedContentWidth()
	// Strip HTML comments from body and cleanup body.
	body := htmlCommentRegex.ReplaceAllString(m.pr.Data.Body, "")
	body = lineCleanupRegex.ReplaceAllString(body, "")

	body = strings.TrimSpace(body)
	if body == "" {
		return lipgloss.NewStyle().Italic(true).Foreground(m.ctx.Theme.FaintText).Render("No description provided.")
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

func (m *Model) SetRow(d *data.PullRequestData) {
	if d == nil {
		m.pr = nil
	} else {
		m.pr = &pr.PullRequest{Ctx: m.ctx, Data: d}
	}
}

func (m *Model) SetWidth(width int) {
	m.width = width
	m.inputBox.SetWidth(width)
}

func (m *Model) IsTextInputBoxFocused() bool {
	return m.isCommenting || m.isAssigning || m.isApproving || m.isUnassigning
}

func (m *Model) GetIsCommenting() bool {
	return m.isCommenting
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.inputBox.UpdateProgramContext(ctx)
}

func (m *Model) shouldCancelComment() bool {
	if !m.ShowConfirmCancel {
		m.inputBox.SetPrompt(lipgloss.NewStyle().Foreground(m.ctx.Theme.ErrorText).Render("Discard comment? (y/N)"))
		m.ShowConfirmCancel = true
		return false
	}
	m.inputBox.Blur()
	m.isCommenting = false
	m.isApproving = false
	m.ShowConfirmCancel = false
	return true
}

func (m *Model) SetIsCommenting(isCommenting bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !m.isCommenting && isCommenting {
		m.inputBox.Reset()
	}
	m.isCommenting = isCommenting
	m.inputBox.SetPrompt(commentPrompt)

	if isCommenting {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) getIndentedContentWidth() int {
	return m.width - 4
}

func (m *Model) GetIsApproving() bool {
	return m.isApproving
}

func (m *Model) SetIsApproving(isApproving bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !m.isApproving && isApproving {
		m.inputBox.Reset()
	}
	m.isApproving = isApproving
	m.inputBox.SetPrompt(approvalPrompt)
	m.inputBox.SetValue("LGTM")

	if isApproving {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) GetIsAssigning() bool {
	return m.isAssigning
}

func (m *Model) SetIsAssigning(isAssigning bool) tea.Cmd {
	if m.pr == nil {
		return nil
	}

	if !m.isAssigning && isAssigning {
		m.inputBox.Reset()
	}
	m.isAssigning = isAssigning
	m.inputBox.SetPrompt("Assign users (whitespace-separated)...")
	if !m.userAssignedToPr(m.ctx.User) {
		m.inputBox.SetValue(m.ctx.User)
	}

	if isAssigning {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) userAssignedToPr(login string) bool {
	for _, a := range m.pr.Data.Assignees.Nodes {
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
	if m.pr == nil {
		return nil
	}

	if !m.isUnassigning && isUnassigning {
		m.inputBox.Reset()
	}
	m.isUnassigning = isUnassigning
	m.inputBox.SetPrompt("Unassign users (whitespace-separated)...")
	m.inputBox.SetValue(strings.Join(m.prAssignees(), "\n"))

	if isUnassigning {
		return tea.Sequence(textarea.Blink, m.inputBox.Focus())
	}
	return nil
}

func (m *Model) prAssignees() []string {
	var assignees []string
	for _, n := range m.pr.Data.Assignees.Nodes {
		assignees = append(assignees, n.Login)
	}
	return assignees
}
