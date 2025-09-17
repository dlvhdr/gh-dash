package tabs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type SectionState struct {
	Count     int
	IsLoading bool
	spinner   spinner.Model
}

type Model struct {
	sectionsConfigs []config.SectionConfig
	sectionCounts   []SectionState
	CurrSectionId   int
}

func NewModel(ctx *context.ProgramContext) Model {
	return Model{
		CurrSectionId: 1,
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

func (m Model) View(ctx *context.ProgramContext) string {
	sectionTitles := make([]string, 0, len(m.sectionsConfigs))
	for i, section := range m.sectionsConfigs {
		title := section.Title
		// handle search section
		if i > 0 {
			if m.sectionCounts[i].IsLoading {
				title = fmt.Sprintf("%s %s", title, m.sectionCounts[i].spinner.View())
			} else {
				title = fmt.Sprintf("%s (%s)", title, utils.ShortNumber(m.sectionCounts[i].Count))
			}
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

	version := lipgloss.NewStyle().Foreground(ctx.Theme.SecondaryText).Render(ctx.Version)

	renderedTabs := lipgloss.NewStyle().
		Width(ctx.ScreenWidth - lipgloss.Width(version)).
		MaxWidth(ctx.ScreenWidth - lipgloss.Width(version)).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, strings.Join(tabs, ctx.Styles.Tabs.TabSeparator.Render("|"))))

	return ctx.Styles.Tabs.TabsRow.
		Width(ctx.ScreenWidth).
		MaxWidth(ctx.ScreenWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, renderedTabs, version))
}

func (m *Model) SetCurrSectionId(id int) {
	m.CurrSectionId = id
}

func (m *Model) UpdateSectionsConfigs(ctx *context.ProgramContext) {
	m.sectionsConfigs = ctx.GetViewSectionsConfig()
	m.sectionCounts = make([]SectionState, len(m.sectionsConfigs))
	for i := range m.sectionsConfigs {
		m.sectionCounts[i] = SectionState{
			Count:     0,
			IsLoading: false,
			spinner:   spinner.New(spinner.WithSpinner(spinner.Dot), spinner.WithStyle(lipgloss.NewStyle().Foreground(ctx.Theme.FaintText).PaddingLeft(2))),
		}
	}
}

func (m *Model) UpdateSectionCounts(sections []section.Section) {
	for i, s := range sections {
		m.sectionCounts[i].Count = s.GetTotalCount()
		m.sectionCounts[i].IsLoading = s.GetIsLoading()
	}
}

func (m *Model) SetAllLoading() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i := range m.sectionCounts {
		m.sectionCounts[i].IsLoading = true
		cmds = append(cmds, m.sectionCounts[i].spinner.Tick)
	}

	return cmds
}
