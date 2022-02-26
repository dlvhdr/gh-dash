package prssection

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/config"
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/ui/components/pr"
	"github.com/dlvhdr/gh-prs/ui/components/table"
	"github.com/dlvhdr/gh-prs/ui/constants"
	"github.com/dlvhdr/gh-prs/ui/context"
	"github.com/dlvhdr/gh-prs/utils"
)

type Model struct {
	Id        int
	Config    config.PRSectionConfig
	Prs       []data.PullRequestData
	ctx       *context.ProgramContext
	spinner   spinner.Model
	isLoading bool
	table     table.Model
}

func NewModel(id int, ctx *context.ProgramContext, config config.PRSectionConfig, prs []data.PullRequestData) Model {
	m := Model{
		Id:        id,
		Config:    config,
		ctx:       ctx,
		isLoading: true,
		Prs:       prs,
		spinner:   spinner.Model{Spinner: spinner.Dot},
	}

	m.table = table.NewModel(
		m.getDimensions(),
		m.getSectionColumns(),
		m.buildPullRequestRows(),
		"PR",
		utils.StringPtr(emptyStateStyle.Render(fmt.Sprintf(
			"No PRs were found that match the given filters: %s",
			lipgloss.NewStyle().Italic(true).Render(m.Config.Filters),
		))),
	)

	return m
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case SectionPullRequestsFetchedMsg:
		m.Prs = msg.Prs
		m.isLoading = false
		m.table.SetRows(m.buildPullRequestRows())

	case SectionTickMsg:
		if !m.isLoading {
			return m, nil
		}

		var internalTickCmd tea.Cmd
		m.spinner, internalTickCmd = m.spinner.Update(msg.internalTickMsg)
		cmd = m.createNextTickCmd(internalTickCmd)
	}

	return m, cmd
}

func (m *Model) createNextTickCmd(nextTickCmd tea.Cmd) tea.Cmd {
	if m == nil || nextTickCmd == nil {
		return nil
	}
	return func() tea.Msg {
		return SectionTickMsg{
			SectionId:       m.Id,
			internalTickMsg: nextTickCmd(),
		}
	}

}

func (m *Model) getDimensions() constants.Dimensions {
	return constants.Dimensions{
		Width:  m.ctx.MainContentWidth - containerStyle.GetHorizontalPadding(),
		Height: m.ctx.MainContentHeight,
	}
}

func (m *Model) View() string {
	var spinnerText *string
	if m.isLoading {
		spinnerText = utils.StringPtr(lipgloss.JoinHorizontal(lipgloss.Left,
			spinnerStyle.Copy().Render(m.spinner.View()),
			"Fetching Pull Requests...",
		))
	}

	return containerStyle.Copy().Render(
		lipgloss.JoinVertical(lipgloss.Top, m.table.View(spinnerText)),
	)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
	m.table.SetDimensions(m.getDimensions())
}

func (m *Model) getSectionColumns() []table.Column {
	return []table.Column{
		{
			Title: " Updated",
		},
		{
			Title: "",
		},
		{
			Title: "Repo",
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
			Title: "",
		},
		{
			Title: "CI",
			Width: &ciCellWidth,
		},
		{
			Title: "Lines",
			Width: &linesCellWidth,
		},
	}
}

func (m *Model) buildPullRequestRows() []table.Row {
	var rows []table.Row
	for _, currPr := range m.Prs {
		prModel := pr.PullRequest{Data: currPr}
		rows = append(rows, prModel.Render(false, m.getDimensions().Width))
	}

	return rows
}

func (m *Model) NumPrs() int {
	return len(m.Prs)
}

type SectionMsg interface {
	GetSectionId() int
}

type SectionPullRequestsFetchedMsg struct {
	SectionId int
	Prs       []data.PullRequestData
}

func (msg SectionPullRequestsFetchedMsg) GetSectionId() int {
	return msg.SectionId
}

type SectionTickMsg struct {
	SectionId       int
	internalTickMsg tea.Msg
}

func (msg SectionTickMsg) GetSectionId() int {
	return msg.SectionId
}

func (m *Model) GetCurrPr() *data.PullRequestData {
	if len(m.Prs) == 0 {
		return nil
	}
	pr := m.Prs[m.table.GetCurrItem()]
	return &pr
}

func (m *Model) NextPr() int {
	return m.table.NextItem()
}

func (m *Model) PrevPr() int {
	return m.table.PrevItem()
}

func (m *Model) FetchSectionPullRequests() tea.Cmd {
	if m == nil {
		return nil
	}
	m.Prs = nil
	m.table.ResetCurrItem()
	m.table.Rows = nil
	m.isLoading = true
	var cmds []tea.Cmd
	cmds = append(cmds, m.createNextTickCmd(spinner.Tick))

	cmds = append(cmds, func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.ctx.Config.Defaults.PrsLimit
		}
		fetchedPrs, err := data.FetchRepoPullRequests(m.Config.Filters, *limit)
		if err != nil {
			return SectionPullRequestsFetchedMsg{
				SectionId: m.Id,
				Prs:       []data.PullRequestData{},
			}
		}

		sort.Slice(fetchedPrs, func(i, j int) bool {
			return fetchedPrs[i].UpdatedAt.After(fetchedPrs[j].UpdatedAt)
		})
		return SectionPullRequestsFetchedMsg{
			SectionId: m.Id,
			Prs:       fetchedPrs,
		}
	})

	return tea.Batch(cmds...)
}

func (m *Model) IsLoading() bool {
	return m.isLoading
}
