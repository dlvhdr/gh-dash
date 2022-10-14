package tabs

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/config"
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
	sectionsConfigs := ctx.GetViewSectionsConfig()

	sectionTitles := make([]string, 0, len(sectionsConfigs))
	for _, section := range sectionsConfigs {
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

	viewSwitcher := m.renderViewSwitcher(ctx)
	tabsWidth := ctx.ScreenWidth - lipgloss.Width(viewSwitcher)
	renderedTabs := lipgloss.NewStyle().
		Width(tabsWidth).
		MaxWidth(tabsWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(tabs, tabSeparator.Render("|"))))

	return tabsRow.Copy().
		Width(ctx.ScreenWidth).
		MaxWidth(ctx.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs, viewSwitcher))
}

func (m *Model) SetCurrSectionId(id int) {
	m.CurrSectionId = id
}

func (m *Model) renderViewSwitcher(ctx context.ProgramContext) string {
	var views []string

	for _, tab := range config.GetViewTypes(ctx.Config.Defaults) {

		var tabStyle lipgloss.Style
		if ctx.View == config.ViewType(tab.Name) {
			tabStyle = activeView
		} else {
			tabStyle = inactiveView
		}

		views = append(views, tabStyle.Render(tab.Label))
	}

	return viewSwitcher.Copy().
		Render(lipgloss.JoinHorizontal(lipgloss.Top, views...))
}
