package tabs

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

type Model struct {
	sectionsConfigs []config.SectionConfig
	sectionCounts   []*int
	CurrSectionId   int
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		CurrSectionId: 1,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View(ctx *context.ProgramContext) string {
	sectionTitles := make([]string, 0, len(m.sectionsConfigs))
	for i, section := range m.sectionsConfigs {
		title := section.Title
		// handle search section
		if i > 0 && m.sectionCounts[i] != nil && ctx.Config.Theme.Ui.SectionsShowCount {
			title = fmt.Sprintf("%s (%s)", title, utils.ShortNumber(*m.sectionCounts[i]))
		}
		sectionTitles = append(sectionTitles, title)
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

	return ctx.Styles.Tabs.TabsRow.
		Width(ctx.ScreenWidth).
		MaxWidth(ctx.ScreenWidth).
		Render(renderedTabs)
}

func (m *Model) SetCurrSectionId(id int) {
	m.CurrSectionId = id
}

func (m *Model) UpdateSectionsConfigs(ctx *context.ProgramContext) {
	m.sectionsConfigs = ctx.GetViewSectionsConfig()
}

func (m *Model) UpdateSectionCounts(sections []section.Section) {
	m.sectionCounts = make([]*int, len(sections))
	for i, s := range sections {
		m.sectionCounts[i] = s.GetTotalCount()
	}
}
