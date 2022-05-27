package prssection

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/pr"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

const SectionType = "prs"

type Model struct {
	Prs     []data.PullRequestData
	section section.Model
}

func NewModel(id int, ctx *context.ProgramContext, config config.SectionConfig) Model {
	m := Model{
		Prs: []data.PullRequestData{},
		section: section.Model{
			Id:        id,
			Config:    config,
			Ctx:       ctx,
			Spinner:   spinner.Model{Spinner: spinner.Dot},
			IsLoading: true,
			Type:      SectionType,
		},
	}

	m.section.Table = table.NewModel(
		m.section.GetDimensions(),
		m.GetSectionColumns(),
		m.BuildRows(),
		"PR",
		utils.StringPtr(emptyStateStyle.Render(fmt.Sprintf(
			"No PRs were found that match the given filters: %s",
			lipgloss.NewStyle().Italic(true).Render(m.section.Config.Filters),
		))),
	)

	return m
}

func (m *Model) Id() int {
	return m.section.Id
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case SectionPullRequestsFetchedMsg:
		m.Prs = msg.Prs
		m.section.IsLoading = false
		m.section.Table.SetRows(m.BuildRows())

	case section.SectionTickMsg:
		if !m.section.IsLoading {
			return &m, nil
		}

		var internalTickCmd tea.Cmd
		m.section.Spinner, internalTickCmd = m.section.Spinner.Update(msg.InternalTickMsg)
		cmd = m.section.CreateNextTickCmd(internalTickCmd)
	}

	return &m, cmd
}

func (m *Model) View() string {
	var spinnerText *string
	if m.section.IsLoading {
		spinnerText = utils.StringPtr(lipgloss.JoinHorizontal(lipgloss.Top,
			spinnerStyle.Copy().Render(m.section.Spinner.View()),
			"Fetching Pull Requests...",
		))
	}

	return containerStyle.Copy().Render(
		m.section.Table.View(spinnerText),
	)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.section.UpdateProgramContext(ctx)
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
	SectionId int
	Prs       []data.PullRequestData
}

func (msg SectionPullRequestsFetchedMsg) GetSectionId() int {
	return msg.SectionId
}

func (msg SectionPullRequestsFetchedMsg) GetSectionType() string {
	return SectionType
}

func (m *Model) GetCurrRow() data.RowData {
	if len(m.Prs) == 0 {
		return nil
	}
	pr := m.Prs[m.section.Table.GetCurrItem()]
	return &pr
}

func (m *Model) NextRow() int {
	return m.section.NextRow()
}

func (m *Model) PrevRow() int {
	return m.section.PrevRow()
}

func (m *Model) FetchSectionRows() tea.Cmd {
	if m == nil {
		return nil
	}
	m.Prs = nil
	m.section.Table.ResetCurrItem()
	m.section.Table.Rows = nil
	m.section.IsLoading = true
	var cmds []tea.Cmd
	cmds = append(cmds, m.section.CreateNextTickCmd(spinner.Tick))

	cmds = append(cmds, func() tea.Msg {
		limit := m.section.Config.Limit
		if limit == nil {
			limit = &m.section.Ctx.Config.Defaults.PrsLimit
		}
		fetchedPrs, err := data.FetchPullRequests(m.section.Config.Filters, *limit)
		if err != nil {
			return SectionPullRequestsFetchedMsg{
				SectionId: m.section.Id,
				Prs:       []data.PullRequestData{},
			}
		}

		sort.Slice(fetchedPrs, func(i, j int) bool {
			return fetchedPrs[i].UpdatedAt.After(fetchedPrs[j].UpdatedAt)
		})
		return SectionPullRequestsFetchedMsg{
			SectionId: m.section.Id,
			Prs:       fetchedPrs,
		}
	})

	return tea.Batch(cmds...)
}

func (m *Model) GetIsLoading() bool {
	return m.section.IsLoading
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
