package tabs

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/ui/components/carousel"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

type Model struct {
	sectionsConfigs []config.SectionConfig
	sectionCounts   []*int
	carousel        carousel.Model
	ctx             *context.ProgramContext
	version         string
}

func NewModel(ctx *context.ProgramContext) Model {
	c := carousel.New(carousel.WithHeight(1))

	return Model{
		carousel: c,
		ctx:      ctx,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return m.ctx.Styles.Tabs.TabsRow.
		Width(m.ctx.ScreenWidth).
		MaxWidth(m.ctx.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, m.carousel.View(), m.version))
}

func (m *Model) SetCurrSectionId(id int) {
	log.Debug("SetCurrSectionId", "id", id)
	m.carousel.SetCursor(id)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.carousel.SetStyles(carousel.Styles{
		Item:     ctx.Styles.Tabs.Tab,
		Selected: ctx.Styles.Tabs.ActiveTab,
	})

	m.version = lipgloss.NewStyle().Foreground(ctx.Theme.SecondaryText).Render(ctx.Version)
	m.carousel.SetWidth(ctx.ScreenWidth - lipgloss.Width(m.version))
	log.Debug("dolev", "carousel width", m.carousel.Width())
}

func (m *Model) UpdateSectionsConfigs(ctx *context.ProgramContext) {
	m.sectionsConfigs = ctx.GetViewSectionsConfig()
	sectionTitles := make([]string, 0, len(m.sectionsConfigs))
	for i, section := range m.sectionsConfigs {
		title := section.Title
		// handle search section
		if i > 0 && len(m.sectionCounts) >= i && m.sectionCounts[i] != nil && ctx.Config.Theme.Ui.SectionsShowCount {
			title = fmt.Sprintf("%s (%s)", title, utils.ShortNumber(*m.sectionCounts[i]))
		}
		sectionTitles = append(sectionTitles, title)
	}
	oldCursor := m.carousel.Cursor()
	m.carousel.SetItems(sectionTitles)
	m.carousel.SetCursor(oldCursor)
}

func (m *Model) UpdateSectionCounts(sections []section.Section) {
	m.sectionCounts = make([]*int, len(sections))
	for i, s := range sections {
		m.sectionCounts[i] = s.GetTotalCount()
	}
	m.UpdateSectionsConfigs(m.ctx)
}
