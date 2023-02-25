package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/ui/context"
)

type Model struct {
	CurrSectionId int
}

func NewModel() Model {
	return Model{
		CurrSectionId: 1,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View(ctx context.ProgramContext) string {
	log.Debug("tabs view")
	sectionsConfigs := ctx.GetViewSectionsConfig()
	log.Debug("tabs.View", "configs", sectionsConfigs)
	sectionTitles := make([]string, 0, len(sectionsConfigs))
	for _, section := range sectionsConfigs {
		sectionTitles = append(sectionTitles, section.Title)
	}

	var tabs []string
	for i, sectionTitle := range sectionTitles {
		if m.CurrSectionId == i {
			tabs = append(tabs, ctx.Styles.Tabs.ActiveTab.Render(sectionTitle))
		} else {
			tabs = append(tabs, ctx.Styles.Tabs.Tab.Render(sectionTitle))
		}
	}

	renderedTabs := lipgloss.NewStyle().
		Width(ctx.ScreenWidth).
		MaxWidth(ctx.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(tabs, ctx.Styles.Tabs.TabSeparator.Render("|"))))

	return ctx.Styles.Tabs.TabsRow.Copy().
		Width(ctx.ScreenWidth).
		MaxWidth(ctx.ScreenWidth).
		Render(renderedTabs)
}

func (m *Model) SetCurrSectionId(id int) {
	m.CurrSectionId = id
}

func (m *Model) GoToFirstSection() {
	m.CurrSectionId = 1
}
