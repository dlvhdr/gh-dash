package tabs

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/carousel"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type SectionTab struct {
	section section.Section
	spinner spinner.Model
}

type Model struct {
	sections      []section.Section
	sectionTabs   []SectionTab
	carousel      carousel.Model
	ctx           *context.ProgramContext
	latestVersion string
}

func NewModel(ctx *context.ProgramContext) Model {
	c := carousel.New(carousel.WithHeight(1), carousel.WithOverflowIndicators("←", "→"), carousel.WithSeparators())
	m := Model{
		carousel: c,
	}
	m.UpdateProgramContext(ctx)

	return m
}

func (m Model) Init() tea.Cmd {
	return m.fetchHasNewVersion()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := make([]tea.Cmd, 0)
	switch msg := msg.(type) {
	case latestVersionMsg:
		m.latestVersion = msg.version
	case spinner.TickMsg:
		for i, tab := range m.sectionTabs {
			if tab.section.GetIsLoading() {
				var cmd tea.Cmd
				m.sectionTabs[i].spinner, cmd = tab.spinner.Update(msg)
				cmds = append(cmds, cmd)
			}
		}
	}

	m.UpdateTabTitles()

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	c := m.carousel.View()
	logo := m.viewLogo()
	return m.ctx.Styles.Tabs.TabsRow.
		Width(m.ctx.ScreenWidth).
		Height(common.HeaderHeight).
		Render(lipgloss.JoinHorizontal(lipgloss.Bottom,
			lipgloss.NewStyle().Width(
				m.ctx.ScreenWidth-lipgloss.Width(logo)).Render(c), logo))
}

type latestVersionMsg struct {
	version string
	err     error
}

func (m *Model) fetchHasNewVersion() tea.Cmd {
	return func() tea.Msg {
		r, err := data.FetchLatestVersion()
		return latestVersionMsg{
			version: r.Repository.LatestRelease.TagName,
			err:     err,
		}
	}
}

func (m *Model) CurrSectionId() int {
	return m.carousel.Cursor()
}

func (m *Model) SetCurrSectionId(id int) {
	m.carousel.SetCursor(id)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.carousel.SetStyles(carousel.Styles{
		Item:              ctx.Styles.Tabs.Tab,
		Selected:          ctx.Styles.Tabs.ActiveTab,
		OverflowIndicator: ctx.Styles.Tabs.OverflowIndicator,
		Separator:         ctx.Styles.Tabs.TabSeparator,
	})

	m.carousel.SetWidth(ctx.ScreenWidth - lipgloss.Width(m.viewLogo()))
}

func (m *Model) SetSections(sections []section.Section) {
	sectionTabs := make([]SectionTab, 0)
	for _, s := range sections {
		tab := SectionTab{section: s, spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot), spinner.WithStyle(
				lipgloss.NewStyle().Foreground(m.ctx.Theme.FaintText).PaddingLeft(2)))}
		sectionTabs = append(sectionTabs, tab)
	}
	m.sectionTabs = sectionTabs
	m.UpdateTabTitles()
}

func (m *Model) UpdateTabTitles() {
	titles := make([]string, 0)
	for i, tab := range m.sectionTabs {
		cfg := tab.section.GetConfig()
		title := cfg.Title
		// handle search section
		if i == 0 {
			// noop
		} else if tab.section.GetIsLoading() {
			title = fmt.Sprintf("%s %s", title, m.sectionTabs[i].spinner.View())
		} else if m.ctx.Config.Theme.Ui.SectionsShowCount {
			title = fmt.Sprintf("%s (%s)", title,
				utils.ShortNumber(tab.section.GetTotalCount()))
		}

		titles = append(titles, title)
	}

	oldCursor := m.carousel.Cursor()
	m.carousel.SetItems(titles)
	m.carousel.SetCursor(oldCursor)
}

func (m *Model) viewLogo() string {
	version := lipgloss.NewStyle().Foreground(m.ctx.Theme.SecondaryText).Render(m.ctx.Version)
	if m.latestVersion != "" && m.ctx.Version != "dev" && m.ctx.Version != m.latestVersion {
		version = lipgloss.JoinVertical(lipgloss.Left,
			version,
			lipgloss.NewStyle().Foreground(m.ctx.Styles.Colors.SuccessText).Render(" Update available!"),
		)
	} else {
		version = lipgloss.PlaceVertical(2, lipgloss.Bottom, version)
	}

	return lipgloss.NewStyle().Padding(0, 1, 0, 2).Height(2).Render(lipgloss.JoinHorizontal(lipgloss.Bottom,
		lipgloss.NewStyle().Foreground(context.LogoColor).Render(constants.Logo),
		" ",
		version,
	))
}

func (m *Model) SetAllLoading() []tea.Cmd {
	cmds := make([]tea.Cmd, 0)
	for i := range m.sectionTabs {
		cmds = append(cmds, m.sectionTabs[i].spinner.Tick)
	}

	return cmds
}
