package branchsidebar

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/ui/components/branch"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
)

type Model struct {
	ctx    *context.ProgramContext
	branch *branch.BranchData
}

func NewModel(ctx context.ProgramContext) Model {
	return Model{
		branch: nil,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	s := strings.Builder{}

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

func (m *Model) SetRow(b *branch.BranchData) {
	m.branch = b
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}
