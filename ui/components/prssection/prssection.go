package prssection

import (
	"sort"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/pr"
	"github.com/dlvhdr/gh-dash/ui/components/search"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

const SectionType = "prs"

type Model struct {
	section.Model
	Prs []data.PullRequestData
}

func NewModel(id int, ctx *context.ProgramContext, config config.SectionConfig) Model {
	section := section.Model{
		Id:          id,
		Config:      config,
		Ctx:         ctx,
		Spinner:     spinner.Model{Spinner: spinner.Dot},
		Search:      search.NewModel(SectionType, ctx, config.Filters),
		IsLoading:   true,
		IsSearching: false,
		Type:        SectionType,
	}
	m := Model{
		section,
		[]data.PullRequestData{},
	}

	m.Table = table.NewModel(
		m.GetDimensions(),
		m.GetSectionColumns(),
		m.BuildRows(),
		"PR",
		utils.StringPtr(emptyStateStyle.Render(
			"No PRs were found that match the given filters",
		)),
	)

	return m
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case section.DelegatedSectionMsg:
		if msg.Id != m.Id || msg.Type != m.Type {
			return &m, nil
		}

		switch iMsg := msg.InternalMsg.(type) {

		case SectionPullRequestsFetchedMsg:
			m.Prs = iMsg.Prs
			m.IsLoading = false
			m.Table.SetRows(m.BuildRows())

		case section.SectionTickMsg:
			if !m.IsLoading {
				return &m, nil
			}

			var internalTickCmd tea.Cmd
			m.Spinner, internalTickCmd = m.Spinner.Update(iMsg.InternalTickMsg)
			cmd = m.CreateNextTickCmd(internalTickCmd)

		case search.SearchSubmitted:
			m.SetIsSearching(false)
			cmd = m.FetchSectionRows()

		case search.SearchCancelled:
			m.SetIsSearching(false)

		}
		sm, searchCmd := m.Search.Update(msg.InternalMsg)
		m.Search = sm
		return &m, tea.Batch(m.MakeSectionCmd(searchCmd), cmd)
	}

	return &m, nil
}

func (m *Model) View() string {
	var spinnerText *string
	if m.IsLoading {
		spinnerText = utils.StringPtr(lipgloss.JoinHorizontal(lipgloss.Top,
			spinnerStyle.Copy().Render(m.Spinner.View()),
			"Fetching Pull Requests...",
		))
	}

	var search string
	search = m.Search.View(*m.Ctx)

	return containerStyle.Copy().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			search,
			m.Table.View(spinnerText),
		),
	)
}

func (m *Model) GetSectionColumns() []table.Column {
	return []table.Column{
		{
			Title: "",
			Width: &updatedAtCellWidth,
		},
		{
			Title: "",
			Width: &prRepoCellWidth,
		},
		{
			Title: "Title",
			Grow:  utils.BoolPtr(true),
		},
		{
			Title: "Author",
			Width: &prAuthorCellWidth,
		},
		{
			Title: "",
			Width: utils.IntPtr(4),
		},
		{
			Title: "",
		},
		{
			Title: "",
			Width: &ciCellWidth,
		},
		{
			Title: "",
			Width: &linesCellWidth,
		},
	}
}

func (m *Model) BuildRows() []table.Row {
	var rows []table.Row
	for _, currPr := range m.Prs {
		prModel := pr.PullRequest{Data: currPr}
		rows = append(rows, prModel.ToTableRow())
	}

	return rows
}

func (m *Model) NumRows() int {
	return len(m.Prs)
}

type SectionPullRequestsFetchedMsg struct {
	Prs []data.PullRequestData
}

func (m *Model) GetCurrRow() data.RowData {
	if len(m.Prs) == 0 {
		return nil
	}
	pr := m.Prs[m.Table.GetCurrItem()]
	return &pr
}

func (m *Model) FetchSectionRows() tea.Cmd {
	if m == nil {
		return nil
	}
	m.Prs = nil
	m.Table.ResetCurrItem()
	m.Table.Rows = nil
	m.IsLoading = true
	var cmds []tea.Cmd
	cmds = append(cmds, m.CreateNextTickCmd(spinner.Tick))

	cmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.PrsLimit
		}
		fetchedPrs, err := data.FetchPullRequests(m.GetFilters(), *limit)
		if err != nil {
			return SectionPullRequestsFetchedMsg{
				Prs: []data.PullRequestData{},
			}
		}

		sort.Slice(fetchedPrs, func(i, j int) bool {
			return fetchedPrs[i].UpdatedAt.After(fetchedPrs[j].UpdatedAt)
		})
		return SectionPullRequestsFetchedMsg{
			Prs: fetchedPrs,
		}
	}
	cmds = append(cmds, m.MakeSectionCmd(cmd))

	return tea.Batch(cmds...)
}

func FetchAllSections(ctx context.ProgramContext) (sections []section.Section, fetchAllCmd tea.Cmd) {
	fetchPRsCmds := make([]tea.Cmd, 0, len(ctx.Config.PRSections))
	sections = make([]section.Section, 0, len(ctx.Config.PRSections))
	for i, sectionConfig := range ctx.Config.PRSections {
		sectionModel := NewModel(i, &ctx, sectionConfig)
		sections = append(sections, &sectionModel)
		fetchPRsCmds = append(fetchPRsCmds, sectionModel.FetchSectionRows())
	}
	return sections, tea.Batch(fetchPRsCmds...)
}
