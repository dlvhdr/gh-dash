package tabs

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/context"
)

type Model struct {
	Context       context.ProgramContext
	CurrSectionId int
}

func NewModel() Model {
	return Model{
		CurrSectionId: 0,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View(ctx context.ProgramContext) string {
	sectionTitles := make([]string, 0, len(ctx.Config.PRSections))
	for _, section := range ctx.Config.PRSections {
		sectionTitles = append(sectionTitles, section.Title)
	}

	var tabs []string
	for i, sectionTitle := range sectionTitles {
		if m.CurrSectionId == i {
			tabs = append(tabs, activeTab.Render(sectionTitle))
		} else {
			tabs = append(tabs, tab.Render(sectionTitle))
		}
	}

	return tabsRow.
		Width(ctx.ScreenWidth).
		MaxWidth(ctx.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, tabs...))
}

func (m *Model) SetCurrSectionId(id int) {
	m.CurrSectionId = id
}
