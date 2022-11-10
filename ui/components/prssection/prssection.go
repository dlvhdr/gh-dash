package prssection

import (
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/pr"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/keys"
	"github.com/dlvhdr/gh-dash/utils"
)

const SectionType = "pr"

type Model struct {
	section.Model
	Prs []data.PullRequestData
}

func NewModel(id int, ctx *context.ProgramContext, cfg config.PrsSectionConfig, lastUpdated time.Time) Model {
	m := Model{
		section.NewModel(
			id,
			ctx,
			cfg.ToSectionConfig(),
			SectionType,
			GetSectionColumns(cfg, ctx),
			"PR",
			"Pull Requests",
			lastUpdated,
		),
		[]data.PullRequestData{},
	}

	return m
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd
	var err error

	switch msg := msg.(type) {

	case tea.KeyMsg:

		if m.IsSearchFocused() {
			switch {

			case msg.Type == tea.KeyCtrlC, msg.Type == tea.KeyEsc:
				m.SearchBar.SetValue(m.SearchValue)
				blinkCmd := m.SetIsSearching(false)
				return &m, blinkCmd

			case msg.Type == tea.KeyEnter:
				m.SearchValue = m.SearchBar.Value()
				m.SetIsSearching(false)
				return &m, m.FetchSectionRows()
			}

			break
		}

		switch {

		case key.Matches(msg, keys.PRKeys.Diff):
			cmd = m.diff()

		case key.Matches(msg, keys.PRKeys.Checkout):
			cmd, err = m.checkout()
			if err != nil {
				m.Ctx.Error = err
			}

		case key.Matches(msg, keys.PRKeys.Close):
			cmd = m.close()

		case key.Matches(msg, keys.PRKeys.Ready):
			cmd = m.ready()

		case key.Matches(msg, keys.PRKeys.Merge):
			cmd = m.merge()

		case key.Matches(msg, keys.PRKeys.Reopen):
			cmd = m.reopen()

		}

	case UpdatePRMsg:
		for i, currPr := range m.Prs {
			if currPr.Number == msg.PrNumber {
				if msg.IsClosed != nil {
					if *msg.IsClosed == true {
						currPr.State = "CLOSED"
					} else {
						currPr.State = "OPEN"
					}
				}
				if msg.NewComment != nil {
					currPr.Comments.Nodes = append(currPr.Comments.Nodes, *msg.NewComment)
				}
				if msg.ReadyForReview != nil && *msg.ReadyForReview {
					currPr.IsDraft = false
				}
				if msg.IsMerged != nil && *msg.IsMerged {
					currPr.State = "MERGED"
					currPr.Mergeable = ""
				}
				m.Prs[i] = currPr
				m.Table.SetRows(m.BuildRows())
				break
			}
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
			m.UpdatedLastUpdated(time.Now())

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

func GetSectionColumns(cfg config.PrsSectionConfig, ctx *context.ProgramContext) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Prs
	sLayout := cfg.Layout

	updatedAtLayout := config.MergeColumnConfigs(dLayout.UpdatedAt, sLayout.UpdatedAt)
	repoLayout := config.MergeColumnConfigs(dLayout.Repo, sLayout.Repo)
	titleLayout := config.MergeColumnConfigs(dLayout.Title, sLayout.Title)
	authorLayout := config.MergeColumnConfigs(dLayout.Author, sLayout.Author)
	reviewStatusLayout := config.MergeColumnConfigs(dLayout.ReviewStatus, sLayout.ReviewStatus)
	stateLayout := config.MergeColumnConfigs(dLayout.State, sLayout.State)
	ciLayout := config.MergeColumnConfigs(dLayout.Ci, sLayout.Ci)
	linesLayout := config.MergeColumnConfigs(dLayout.Lines, sLayout.Lines)

	return []table.Column{
		{
			Title:  "",
			Width:  updatedAtLayout.Width,
			Hidden: updatedAtLayout.Hidden,
		},
		{
			Title:  "",
			Hidden: stateLayout.Hidden,
		},
		{
			Title:  "",
			Width:  repoLayout.Width,
			Hidden: repoLayout.Hidden,
		},
		{
			Title:  "Title",
			Grow:   utils.BoolPtr(true),
			Hidden: titleLayout.Hidden,
		},
		{
			Title:  "Author",
			Width:  authorLayout.Width,
			Hidden: authorLayout.Hidden,
		},
		{
			Title:  "",
			Width:  utils.IntPtr(4),
			Hidden: reviewStatusLayout.Hidden,
		},
		{
			Title:  "",
			Width:  &ctx.Styles.PrSection.CiCellWidth,
			Grow:   new(bool),
			Hidden: ciLayout.Hidden,
		},
		{
			Title:  "",
			Width:  linesLayout.Width,
			Hidden: linesLayout.Hidden,
		},
	}
}

func (m *Model) BuildRows() []table.Row {
	var rows []table.Row
	for _, currPr := range m.Prs {
		prModel := pr.PullRequest{Ctx: m.Ctx, Data: currPr}
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

		filteredPrs := m.excludeArchivedPullRequests(fetchedPrs)

		sort.Slice(filteredPrs, func(i, j int) bool {
			return filteredPrs[i].UpdatedAt.After(filteredPrs[j].UpdatedAt)
		})
		return SectionPullRequestsFetchedMsg{
			Prs: filteredPrs,
		}
	}
	cmds = append(cmds, m.MakeSectionCmd(cmd))

	return tea.Batch(cmds...)
}

func (m *Model) excludeArchivedPullRequests(fetchedPrs []data.PullRequestData) []data.PullRequestData {
	prs := make([]data.PullRequestData, 0)
	for _, v := range fetchedPrs {
		if v.Repository.IsArchived {
			continue
		}
		prs = append(prs, v)
	}
	return prs
}

func FetchAllSections(ctx context.ProgramContext) (sections []section.Section, fetchAllCmd tea.Cmd) {
	fetchPRsCmds := make([]tea.Cmd, 0, len(ctx.Config.PRSections))
	sections = make([]section.Section, 0, len(ctx.Config.PRSections))
	for i, sectionConfig := range ctx.Config.PRSections {
		sectionModel := NewModel(i+1, &ctx, sectionConfig, time.Now()) // 0 is the search section
		sections = append(sections, &sectionModel)
		fetchPRsCmds = append(fetchPRsCmds, sectionModel.FetchSectionRows())
	}
	return sections, tea.Batch(fetchPRsCmds...)
}

type UpdatePRMsg struct {
	PrNumber       int
	IsClosed       *bool
	NewComment     *data.Comment
	ReadyForReview *bool
	IsMerged       *bool
}
