package branchsidebar

import (
	"fmt"
	"strings"

	gitm "github.com/aymanbagabas/git-module"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/git"
	"github.com/dlvhdr/gh-dash/v4/ui/components/branch"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

type Model struct {
	ctx    *context.ProgramContext
	branch *branch.BranchData
	status *gitm.NameStatus
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		branch: nil,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case updateBranchStatusMsg:
		m.status = &msg.status
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}

	s.WriteString(lipgloss.NewStyle().Bold(true).Render("STATUS\n"))
	if m.status == nil {
		s.WriteString("\nLoading...")
	} else {

		if len(m.status.Added) == 0 && len(m.status.Removed) == 0 && len(m.status.Modified) == 0 {
			s.WriteString("\nNo changes")
		}

		for _, file := range m.status.Added {
			s.WriteString(fmt.Sprintf("\nA %s", file))
		}
		for _, file := range m.status.Removed {
			s.WriteString(fmt.Sprintf("\nD %s", file))
		}
		for _, file := range m.status.Modified {
			s.WriteString(fmt.Sprintf("\nM %s", file))
		}
	}

	s.WriteString("\n\n")
	s.WriteString(lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintBorder).Render(
		strings.Repeat(lipgloss.NormalBorder().Bottom, m.ctx.Config.Defaults.Preview.Width-5)),
	)
	s.WriteString("\n\n")

	if m.branch == nil {
		return "No branch selected"
	}

	s.WriteString(m.branch.Data.Name)
	if m.branch.PR != nil {
		s.WriteString("\n")
		s.WriteString(fmt.Sprintf("#%d %s", m.branch.PR.GetNumber(), m.branch.PR.Title))
	}

	return s.String()
}

type updateBranchStatusMsg struct {
	status gitm.NameStatus
}

func (m *Model) SetRow(b *branch.BranchData) tea.Cmd {
	m.branch = b
	return m.refreshBranchStatusCmd
}

func (m *Model) refreshBranchStatusCmd() tea.Msg {
	status, err := git.GetStatus(m.ctx.RepoPath)
	if err != nil {
		return nil
	}
	return updateBranchStatusMsg{
		status: status,
	}
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}
