package reposection

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/config"
	"github.com/dlvhdr/gh-dash/v4/data"
	"github.com/dlvhdr/gh-dash/v4/git"
	"github.com/dlvhdr/gh-dash/v4/ui/components/branch"
	"github.com/dlvhdr/gh-dash/v4/ui/components/section"
	"github.com/dlvhdr/gh-dash/v4/ui/components/table"
	"github.com/dlvhdr/gh-dash/v4/ui/constants"
	"github.com/dlvhdr/gh-dash/v4/ui/context"
	"github.com/dlvhdr/gh-dash/v4/ui/keys"
	"github.com/dlvhdr/gh-dash/v4/utils"
)

const SectionType = "repo"

type Model struct {
	section.BaseModel
	repo     *git.Repo
	Branches []branch.Branch
	Prs      []data.PullRequestData
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.PrsSectionConfig,
	lastUpdated time.Time,
) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		id,
		ctx,
		cfg.ToSectionConfig(),
		SectionType,
		GetSectionColumns(cfg, ctx),
		m.GetItemSingularForm(),
		m.GetItemPluralForm(),
		lastUpdated,
	)
	m.repo = &git.Repo{Branches: []git.Branch{}}
	m.Branches = []branch.Branch{}
	m.Prs = []data.PullRequestData{}

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
				m.ResetRows()
				return &m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

			break
		}
		switch {
		case key.Matches(msg, keys.PRKeys.Checkout):
			cmd, err = m.checkout()
			if err != nil {
				m.Ctx.Error = err
			}
		}

	case UpdatePRMsg:
		for i, currPr := range m.Prs {
			if currPr.Number == msg.PrNumber {
				if msg.IsClosed != nil {
					if *msg.IsClosed {
						currPr.State = "CLOSED"
					} else {
						currPr.State = "OPEN"
					}
				}
				if msg.NewComment != nil {
					currPr.Comments.Nodes = append(currPr.Comments.Nodes, *msg.NewComment)
				}
				if msg.AddedAssignees != nil {
					currPr.Assignees.Nodes = addAssignees(currPr.Assignees.Nodes, msg.AddedAssignees.Nodes)
				}
				if msg.RemovedAssignees != nil {
					currPr.Assignees.Nodes = removeAssignees(currPr.Assignees.Nodes, msg.RemovedAssignees.Nodes)
				}
				if msg.ReadyForReview != nil && *msg.ReadyForReview {
					currPr.IsDraft = false
				}
				if msg.IsMerged != nil && *msg.IsMerged {
					currPr.State = "MERGED"
					currPr.Mergeable = ""
				}
				m.Prs[i] = currPr
				m.Table.SetIsLoading(false)
				m.Table.SetRows(m.BuildRows())
				break
			}
		}

	case repoMsg:
		m.repo = msg.repo
		m.Table.SetIsLoading(false)
		m.updateBranches()
		m.Table.SetRows(m.BuildRows())

	case SectionPullRequestsFetchedMsg:
		if m.LastFetchTaskId == msg.TaskId {
			m.Prs = msg.Prs
			m.TotalCount = msg.TotalCount
			m.PageInfo = &msg.PageInfo
			m.Table.SetIsLoading(false)
			m.updateBranches()
			m.Table.SetRows(m.BuildRows())
			m.Table.UpdateLastUpdated(time.Now())
			m.UpdateTotalItemsCount(m.TotalCount)
		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.Table.SetRows(m.BuildRows())
	m.SearchBar = search

	prompt, promptCmd := m.PromptConfirmationBox.Update(msg)
	m.PromptConfirmationBox = prompt

	table, tableCmd := m.Table.Update(msg)
	m.Table = table

	return &m, tea.Batch(cmd, searchCmd, promptCmd, tableCmd)
}

func GetSectionColumns(
	cfg config.PrsSectionConfig,
	ctx *context.ProgramContext,
) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Prs
	sLayout := cfg.Layout

	updatedAtLayout := config.MergeColumnConfigs(
		dLayout.UpdatedAt,
		sLayout.UpdatedAt,
	)
	repoLayout := config.MergeColumnConfigs(dLayout.Repo, sLayout.Repo)
	titleLayout := config.MergeColumnConfigs(dLayout.Title, sLayout.Title)
	authorLayout := config.MergeColumnConfigs(dLayout.Author, sLayout.Author)
	assigneesLayout := config.MergeColumnConfigs(
		dLayout.Assignees,
		sLayout.Assignees,
	)
	baseLayout := config.MergeColumnConfigs(dLayout.Base, sLayout.Base)
	reviewStatusLayout := config.MergeColumnConfigs(
		dLayout.ReviewStatus,
		sLayout.ReviewStatus,
	)
	stateLayout := config.MergeColumnConfigs(dLayout.State, sLayout.State)
	ciLayout := config.MergeColumnConfigs(dLayout.Ci, sLayout.Ci)
	linesLayout := config.MergeColumnConfigs(dLayout.Lines, sLayout.Lines)

	if !ctx.Config.Theme.Ui.Table.Compact {
		return []table.Column{
			{
				Title:  "",
				Width:  utils.IntPtr(3),
				Hidden: stateLayout.Hidden,
			},
			{
				Title:  "Title",
				Grow:   utils.BoolPtr(true),
				Hidden: titleLayout.Hidden,
			},
			{
				Title:  "Assignees",
				Width:  assigneesLayout.Width,
				Hidden: assigneesLayout.Hidden,
			},
			{
				Title:  "Base",
				Width:  baseLayout.Width,
				Hidden: baseLayout.Hidden,
			},
			{
				Title:  "󰯢",
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
			{
				Title:  "",
				Width:  updatedAtLayout.Width,
				Hidden: updatedAtLayout.Hidden,
			},
		}
	}

	return []table.Column{
		{
			Title:  "",
			Width:  utils.IntPtr(3),
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
			Title:  "Assignees",
			Width:  assigneesLayout.Width,
			Hidden: assigneesLayout.Hidden,
		},
		{
			Title:  "Base",
			Width:  baseLayout.Width,
			Hidden: baseLayout.Hidden,
		},
		{
			Title:  "󰯢",
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
		{
			Title:  "",
			Width:  updatedAtLayout.Width,
			Hidden: updatedAtLayout.Hidden,
		},
	}
}

func (m *Model) updateBranches() {
	branches := make([]branch.Branch, 0)
	for _, ref := range m.repo.Branches {
		b := branch.Branch{Ctx: m.Ctx, Data: ref, Columns: m.Table.Columns}
		b.PR = findPRForRef(m.Prs, ref.Name)

		branches = append(branches, b)
	}
	m.Branches = branches
}

func (m Model) BuildRows() []table.Row {
	var rows []table.Row
	currItem := m.Table.GetCurrItem()

	for i, ref := range m.repo.Branches {
		b := branch.Branch{Ctx: m.Ctx, Data: ref, Columns: m.Table.Columns}
		b.PR = findPRForRef(m.Prs, ref.Name)

		rows = append(
			rows,
			b.ToTableRow(currItem == i),
		)
	}

	if rows == nil {
		rows = []table.Row{}
	}

	return rows
}

func findPRForRef(prs []data.PullRequestData, branch string) *data.PullRequestData {
	for _, pr := range prs {
		if pr.HeadRefName == branch {
			return &pr
		}
	}
	return nil
}

func (m *Model) NumRows() int {
	return len(m.repo.Branches)
}

type SectionPullRequestsFetchedMsg struct {
	Prs        []data.PullRequestData
	TotalCount int
	PageInfo   data.PageInfo
	TaskId     string
}

func (m *Model) GetCurrBranch() *branch.Branch {
	if len(m.repo.Branches) == 0 {
		return nil
	}
	return &m.Branches[m.Table.GetCurrItem()]
}

func (m *Model) GetCurrRow() data.RowData {
	if len(m.repo.Branches) == 0 {
		return nil
	}
	branch := m.repo.Branches[m.Table.GetCurrItem()]
	pr := findPRForRef(m.Prs, branch.Name)
	return pr
}

func (m *Model) FetchNextPageSectionRows() []tea.Cmd {
	if m == nil {
		return nil
	}

	if m.PageInfo != nil && !m.PageInfo.HasNextPage {
		return nil
	}

	var cmds []tea.Cmd

	startCursor := time.Now().String()
	if m.PageInfo != nil {
		startCursor = m.PageInfo.StartCursor
	}
	taskId := fmt.Sprintf("fetching_prs_%d_%s", m.Id, startCursor)
	m.LastFetchTaskId = taskId

	branchesTaskId := fmt.Sprintf("fetching_branches_%d", time.Now().Unix())
	if m.Ctx.RepoPath != nil {
		branchesTask := context.Task{
			Id:        branchesTaskId,
			StartText: "Reading local branches",
			FinishedText: fmt.Sprintf(
				`Read branches successfully for "%s"`,
				*m.Ctx.RepoPath,
			),
			State: context.TaskStart,
			Error: nil,
		}
		bCmd := m.Ctx.StartTask(branchesTask)
		cmds = append(cmds, bCmd)
	}

	task := context.Task{
		Id:           taskId,
		StartText:    "Fetching PRs for your branches",
		FinishedText: "PRs for your branches have been fetched",
		State:        context.TaskStart,
		Error:        nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	var repoCmd tea.Cmd
	if m.Ctx.RepoPath != nil {
		repoCmd = m.makeRepoCmd(branchesTaskId)
	}
	fetchCmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.PrsLimit
		}
		res, err := data.FetchPullRequests(m.GetFilters(), *limit, m.PageInfo)
		// TODO: enrich with branches only for section with branches
		if err != nil {
			return constants.TaskFinishedMsg{
				SectionId:   m.Id,
				SectionType: m.Type,
				TaskId:      taskId,
				Err:         err,
			}
		}

		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg: SectionPullRequestsFetchedMsg{
				Prs:        res.Prs,
				TotalCount: res.TotalCount,
				PageInfo:   res.PageInfo,
				TaskId:     taskId,
			},
		}
	}
	cmds = append(cmds, fetchCmd, repoCmd)

	if m.PageInfo == nil {
		m.Table.SetIsLoading(true)
		cmds = append(cmds, m.Table.StartLoadingSpinner())

	}

	return cmds
}

func (m *Model) ResetRows() {
	m.Prs = nil
	m.BaseModel.ResetRows()
}

type repoMsg struct {
	repo *git.Repo
	err  error
}

func openRepoCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		repo, err := git.GetRepo(dir)
		return repoMsg{repo: repo, err: err}
	}
}

func FetchAllBranches(
	ctx context.ProgramContext,
) (sections []section.Section, fetchAllCmd tea.Cmd) {

	cmds := make([]tea.Cmd, 0)
	if ctx.RepoPath != nil {
		cmds = append(cmds, openRepoCmd(*ctx.RepoPath))
	}

	t := config.RepoView
	cfg := config.PrsSectionConfig{
		Title:   "Local Branches",
		Filters: "author:@me",
		Limit:   utils.IntPtr(20),
		Type:    &t,
	}
	s := NewModel(
		1,
		&ctx,
		cfg,
		time.Now(),
	)
	cmds = append(cmds, s.FetchNextPageSectionRows()...)

	return []section.Section{&s}, tea.Batch(cmds...)
}

type UpdatePRMsg struct {
	PrNumber         int
	IsClosed         *bool
	NewComment       *data.Comment
	ReadyForReview   *bool
	IsMerged         *bool
	AddedAssignees   *data.Assignees
	RemovedAssignees *data.Assignees
}

func addAssignees(assignees, addedAssignees []data.Assignee) []data.Assignee {
	newAssignees := assignees
	for _, assignee := range addedAssignees {
		if !assigneesContains(newAssignees, assignee) {
			newAssignees = append(newAssignees, assignee)
		}
	}

	return newAssignees
}

func removeAssignees(
	assignees, removedAssignees []data.Assignee,
) []data.Assignee {
	newAssignees := []data.Assignee{}
	for _, assignee := range assignees {
		if !assigneesContains(removedAssignees, assignee) {
			newAssignees = append(newAssignees, assignee)
		}
	}

	return newAssignees
}

func assigneesContains(assignees []data.Assignee, assignee data.Assignee) bool {
	for _, a := range assignees {
		if assignee == a {
			return true
		}
	}
	return false
}

func (m Model) GetItemSingularForm() string {
	return "PR"
}

func (m Model) GetItemPluralForm() string {
	return "PRs"
}

func (m Model) GetTotalCount() *int {
	if m.IsLoading() {
		return nil
	}
	return &m.TotalCount
}

func (m Model) IsLoading() bool {
	return m.Table.IsLoading()
}

func (m *Model) makeRepoCmd(taskId string) tea.Cmd {
	return func() tea.Msg {
		repo, err := git.GetRepo(*m.Ctx.RepoPath)
		return constants.TaskFinishedMsg{
			SectionId:   m.Id,
			SectionType: m.Type,
			TaskId:      taskId,
			Msg:         repoMsg{repo: repo},
			Err:         err,
		}
	}
}
