package prssection

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-prs/config"
	"github.com/dlvhdr/gh-prs/data"
	"github.com/dlvhdr/gh-prs/ui/components/pr"
	"github.com/dlvhdr/gh-prs/ui/components/table"
	"github.com/dlvhdr/gh-prs/ui/context"
	"github.com/dlvhdr/gh-prs/utils"
)

type Model struct {
	Id      int
	Context *context.ProgramContext
	Config  config.PRSectionConfig
	Prs     []data.PullRequestData
}

func NewModel(id int, ctx *context.ProgramContext, config config.PRSectionConfig, prs []data.PullRequestData) Model {
	return Model{
		Id:      id,
		Context: ctx,
		Config:  config,
		Prs:     prs,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SectionPullRequestsFetchedMsg:
		if msg.SectionId != m.Id {
			break
		}

		m.Prs = msg.Prs
		return m, nil
	}

	return m, nil
}

func (m *Model) View() string {
	return table.NewModel(
		m.Context.MainViewportWidth,
		m.getSectionColumns(),
		m.renderPullRequestRows(),
		nil,
	).View()
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

func (m *Model) renderPullRequestRows() []table.Row {
	var renderedPRs []table.Row
	for _, currPr := range m.Prs {
		// isSelected := m.cursor.currSectionId == section.Id && m.cursor.currPrId == prId
		prModel := pr.PullRequest{Data: currPr}
		renderedPRs = append(renderedPRs, prModel.Render(false, m.Context.ScreenWidth))
	}

	return renderedPRs
}

func (m *Model) NumPrs() int {
	return len(m.Prs)
}

type FetchPullRequestsMsg struct{}

type SectionPullRequestsFetchedMsg struct {
	SectionId int
	Prs       []data.PullRequestData
}

func (m *Model) FetchSectionPullRequests() tea.Cmd {
	limit := m.Config.Limit
	if limit == nil {
		limit = &m.Context.Config.Defaults.PrsLimit
	}
	return func() tea.Msg {
		fetchedPrs, err := data.FetchRepoPullRequests(m.Config.Filters, *limit)
		if err != nil {
			return SectionPullRequestsFetchedMsg{
				SectionId: m.Id,
				Prs:       []data.PullRequestData{},
			}
		}

		return SectionPullRequestsFetchedMsg{
			SectionId: m.Id,
			Prs:       fetchedPrs,
		}
	}
}
