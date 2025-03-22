package tabs

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/ui/components/carousel"
	"github.com/dlvhdr/gh-dash/v4/internal/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/ui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type SectionState struct {
	Count     int
	IsLoading bool
	spinner   spinner.Model
}

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
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case spinner.TickMsg:
		for i, s := range m.sectionCounts {
			if s.IsLoading {
				var cmd tea.Cmd
				m.sectionCounts[i].spinner, cmd = s.spinner.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	return m, tea.Batch(cmds...)
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
	for i, s := range sections {
		m.sectionCounts[i].Count = s.GetTotalCount()
		m.sectionCounts[i].IsLoading = s.GetIsLoading()
	}
	m.UpdateSectionsConfigs(m.ctx)
}

func (m *Model) SetAllLoading() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i := range m.sectionCounts {
		m.sectionCounts[i].IsLoading = true
		cmds = append(cmds, m.sectionCounts[i].spinner.Tick)
	}

	return cmds
}
