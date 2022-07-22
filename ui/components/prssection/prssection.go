package prssection

import (
	"sort"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/pr"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

const SectionType = "pr"

type Model struct {
	section.Model
	Prs []data.PullRequestData
}

func NewModel(id int, ctx *context.ProgramContext, cfg config.SectionConfig) Model {
	m := Model{
		section.NewModel(
			id,
			ctx,
			cfg,
			SectionType,
			GetSectionColumns(),
			"PR",
			"Pull Requests",
		),
		[]data.PullRequestData{},
	}

	return m
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.Type {

		case tea.KeyEnter:
			m.SearchValue = m.SearchBar.Value()
			m.SetIsSearching(false)
			return &m, m.FetchSectionRows()

		case tea.KeyCtrlC, tea.KeyEsc:
			m.SearchBar.SetValue(m.SearchValue)
			blinkCmd := m.SetIsSearching(false)
			return &m, blinkCmd

		}

	case section.SectionMsg:
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

		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.SearchBar = search
	return &m, tea.Batch(cmd, searchCmd)
}

func GetSectionColumns() []table.Column {
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

	if rows == nil {
		rows = []table.Row{}
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
		sectionModel := NewModel(i+1, &ctx, sectionConfig) // 0 is the search section
		sections = append(sections, &sectionModel)
		fetchPRsCmds = append(fetchPRsCmds, sectionModel.FetchSectionRows())
	}
	return sections, tea.Batch(fetchPRsCmds...)
}
