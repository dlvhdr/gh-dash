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
			tabs = append(tabs, ctx.Styles.Tabs.ActiveTab.Render(sectionTitle))
		} else {
			tabs = append(tabs, ctx.Styles.Tabs.Tab.Render(sectionTitle))
		}
	}

	viewSwitcher := m.renderViewSwitcher(ctx)
	tabsWidth := ctx.ScreenWidth - lipgloss.Width(viewSwitcher)
	renderedTabs := lipgloss.NewStyle().
		Width(tabsWidth).
		MaxWidth(tabsWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(tabs, ctx.Styles.Tabs.TabSeparator.Render("|"))))

	return ctx.Styles.Tabs.TabsRow.Copy().
		Width(ctx.ScreenWidth).
		MaxWidth(ctx.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs, viewSwitcher))
}

func (m *Model) SetCurrSectionId(id int) {
	m.CurrSectionId = id
}

func (m *Model) renderViewSwitcher(ctx context.ProgramContext) string {
	var prsStyle, issuesStyle lipgloss.Style
	if ctx.View == config.PRsView {
		prsStyle = ctx.Styles.Tabs.ActiveView
		issuesStyle = ctx.Styles.Tabs.InactiveView
	} else {
		prsStyle = ctx.Styles.Tabs.InactiveView
		issuesStyle = ctx.Styles.Tabs.ActiveView
	}

	prs := prsStyle.Render("[ PRs]")
	issues := issuesStyle.Render("[ Issues]")
	return ctx.Styles.Tabs.ViewSwitcher.Copy().
		Render(lipgloss.JoinHorizontal(lipgloss.Top, prs, issues))
}
