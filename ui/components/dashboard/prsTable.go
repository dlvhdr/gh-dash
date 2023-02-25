package dashboard

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/pr"
	"github.com/dlvhdr/gh-dash/ui/components/prssection"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
)

type prsTable struct {
	ctx   *context.ProgramContext
	cfg   config.PrsSectionConfig
	prs   []data.PullRequestData
	table table.Model
}

func newPrsTable(ctx context.ProgramContext, cfg config.PrsSectionConfig) prsTable {
	var m prsTable
	m.ctx = &ctx
	m.table = table.NewModel(
		*m.ctx,
		constants.Dimensions{Width: m.ctx.MainContentWidth},
		time.Now(),
		prssection.GetSectionColumns(cfg, m.ctx),
		nil,
		"PR",
		nil,
	)
	m.table.EmptyState = nil
	return prsTable{ctx: &ctx}
}

func (m prsTable) update(msg tea.Msg) (prsTable, tea.Cmd) {
	switch msg := msg.(type) {
	case prssection.SectionPullRequestsFetchedMsg:
		if msg.PageInfo.StartCursor == "" {
			m.prs = msg.Prs
			m.table.SetRows(m.buildRows())
		}
	}
	return m, nil
}

func (m *prsTable) buildRows() []table.Row {
	var rows []table.Row
	for _, currPr := range m.prs {
		prModel := pr.PullRequest{Ctx: m.ctx, Data: currPr}
		rows = append(rows, prModel.ToTableRow())
	}

	if rows == nil {
		rows = []table.Row{}
	}

	return rows
}

func (m prsTable) view() string {
	return m.table.View()
}

func (m *prsTable) updateProgramContext(ctx *context.ProgramContext) {
	m.ctx = ctx
}
