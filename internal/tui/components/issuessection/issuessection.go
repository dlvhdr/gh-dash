package issuessection

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issue"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

const SectionType = "issue"

type Model struct {
	section.BaseModel
	Issues []data.IssueData
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.IssuesSectionConfig,
	lastUpdated time.Time,
	createdAt time.Time,
) Model {
	m := Model{}
	m.BaseModel = section.NewModel(
		ctx,
		section.NewSectionOptions{
			Id:          id,
			Config:      cfg.ToSectionConfig(),
			Type:        SectionType,
			Columns:     GetSectionColumns(cfg, ctx),
			Singular:    m.GetItemSingularForm(),
			Plural:      m.GetItemPluralForm(),
			LastUpdated: lastUpdated,
			CreatedAt:   createdAt,
		},
	)
	m.Issues = []data.IssueData{}

	return m
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:

		if m.IsSearchFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.SearchBar.SetValue(m.SearchValue)
				blinkCmd := m.SetIsSearching(false)
				return &m, blinkCmd

			case tea.KeyEnter:
				m.SearchValue = m.SearchBar.Value()
				m.SetIsSearching(false)
				m.ResetRows()
				return &m, tea.Batch(m.FetchNextPageSectionRows()...)
			}

			break
		}

		if m.IsPromptConfirmationFocused() {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.PromptConfirmationBox.Reset()
				cmd = m.SetIsPromptConfirmationShown(false)
				return &m, cmd

			case tea.KeyEnter:
				input := m.PromptConfirmationBox.Value()
				action := m.GetPromptConfirmationAction()
				if input == "Y" || input == "y" {
					switch action {
					case "close":
						cmd = m.close()
					case "reopen":
						cmd = m.reopen()
					}
				}

				m.PromptConfirmationBox.Reset()
				blinkCmd := m.SetIsPromptConfirmationShown(false)

				return &m, tea.Batch(cmd, blinkCmd)
			}
			break
		}

		switch {
		case key.Matches(msg, keys.IssueKeys.ToggleSmartFiltering):
			if !m.HasRepoNameInConfiguredFilter() {
				m.IsFilteredByCurrentRemote = !m.IsFilteredByCurrentRemote
			}
			searchValue := m.GetSearchValue()
			if m.SearchValue != searchValue {
				m.SearchValue = searchValue
				m.SearchBar.SetValue(searchValue)
				m.SetIsSearching(false)
				m.ResetRows()
				return &m, tea.Batch(m.FetchNextPageSectionRows()...)
			}
		}

	case UpdateIssueMsg:
		for i, currIssue := range m.Issues {
			if currIssue.Number == msg.IssueNumber {
				if msg.IsClosed != nil {
					if *msg.IsClosed {
						currIssue.State = "CLOSED"
					} else {
						currIssue.State = "OPEN"
					}
				}
				if msg.Labels != nil {
					currIssue.Labels.Nodes = msg.Labels.Nodes
				}
				if msg.NewComment != nil {
					currIssue.Comments.Nodes = append(currIssue.Comments.Nodes, *msg.NewComment)
				}
				if msg.AddedAssignees != nil {
					currIssue.Assignees.Nodes = addAssignees(
						currIssue.Assignees.Nodes, msg.AddedAssignees.Nodes)
				}
				if msg.RemovedAssignees != nil {
					currIssue.Assignees.Nodes = removeAssignees(
						currIssue.Assignees.Nodes, msg.RemovedAssignees.Nodes)
				}
				m.Issues[i] = currIssue
				m.SetIsLoading(false)
				m.Table.SetRows(m.BuildRows())
				break
			}
		}

	case SectionIssuesFetchedMsg:
		if m.LastFetchTaskId == msg.TaskId {
			if m.PageInfo != nil {
				m.Issues = append(m.Issues, msg.Issues...)
			} else {
				m.Issues = msg.Issues
			}
			m.TotalCount = msg.TotalCount
			m.SetIsLoading(false)
			m.PageInfo = &msg.PageInfo
			m.Table.SetRows(m.BuildRows())
			m.UpdateLastUpdated(time.Now())
			m.UpdateTotalItemsCount(m.TotalCount)
		}
	}

	search, searchCmd := m.SearchBar.Update(msg)
	m.SearchBar = search

	prompt, promptCmd := m.PromptConfirmationBox.Update(msg)
	m.PromptConfirmationBox = prompt

	table, tableCmd := m.Table.Update(msg)
	m.Table = table

	return &m, tea.Batch(cmd, searchCmd, promptCmd, tableCmd)
}

func GetSectionColumns(
	cfg config.IssuesSectionConfig,
	ctx *context.ProgramContext,
) []table.Column {
	dLayout := ctx.Config.Defaults.Layout.Issues
	sLayout := cfg.Layout

	updatedAtLayout := config.MergeColumnConfigs(
		dLayout.UpdatedAt,
		sLayout.UpdatedAt,
	)
	createdAtLayout := config.MergeColumnConfigs(
		dLayout.CreatedAt,
		sLayout.CreatedAt,
	)
	stateLayout := config.MergeColumnConfigs(dLayout.State, sLayout.State)
	repoLayout := config.MergeColumnConfigs(dLayout.Repo, sLayout.Repo)
	titleLayout := config.MergeColumnConfigs(dLayout.Title, sLayout.Title)
	creatorLayout := config.MergeColumnConfigs(dLayout.Creator, sLayout.Creator)
	assigneesLayout := config.MergeColumnConfigs(
		dLayout.Assignees,
		sLayout.Assignees,
	)
	commentsLayout := config.MergeColumnConfigs(
		dLayout.Comments,
		sLayout.Comments,
	)
	reactionsLayout := config.MergeColumnConfigs(
		dLayout.Reactions,
		sLayout.Reactions,
	)

	return []table.Column{
		{
			Title:  "",
			Width:  stateLayout.Width,
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
			Title:  "Creator",
			Width:  creatorLayout.Width,
			Hidden: creatorLayout.Hidden,
		},
		{
			Title:  "Assignees",
			Width:  assigneesLayout.Width,
			Hidden: assigneesLayout.Hidden,
		},
		{
			Title:  "",
			Width:  &issueNumCommentsCellWidth,
			Hidden: commentsLayout.Hidden,
		},
		{
			Title:  "",
			Width:  &issueNumCommentsCellWidth,
			Hidden: reactionsLayout.Hidden,
		},
		{
			Title:  "󱦻",
			Width:  updatedAtLayout.Width,
			Hidden: updatedAtLayout.Hidden,
		},
		{
			Title:  "󱡢",
			Width:  createdAtLayout.Width,
			Hidden: createdAtLayout.Hidden,
		},
	}
}

func (m Model) BuildRows() []table.Row {
	var rows []table.Row
	for _, currIssue := range m.Issues {
		issueModel := issue.Issue{Ctx: m.Ctx, Data: currIssue, ShowAuthorIcon: m.ShowAuthorIcon}
		rows = append(rows, issueModel.ToTableRow())
	}

	if rows == nil {
		rows = []table.Row{}
	}

	return rows
}

func (m *Model) NumRows() int {
	return len(m.Issues)
}

func (m *Model) GetCurrRow() data.RowData {
	if len(m.Issues) == 0 {
		return nil
	}
	issue := m.Issues[m.Table.GetCurrItem()]
	return &issue
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
	taskId := fmt.Sprintf("fetching_issues_%d_%s", m.Id, startCursor)
	m.LastFetchTaskId = taskId
	task := context.Task{
		Id:        taskId,
		StartText: fmt.Sprintf(`Fetching issues for "%s"`, m.Config.Title),
		FinishedText: fmt.Sprintf(
			`Issues for "%s" have been fetched`,
			m.Config.Title,
		),
		State: context.TaskStart,
		Error: nil,
	}
	startCmd := m.Ctx.StartTask(task)
	cmds = append(cmds, startCmd)

	fetchCmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.IssuesLimit
		}
		res, err := data.FetchIssues(m.GetFilters(), *limit, m.PageInfo)
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
			Msg: SectionIssuesFetchedMsg{
				Issues:     res.Issues,
				TotalCount: res.TotalCount,
				PageInfo:   res.PageInfo,
				TaskId:     taskId,
			},
		}
	}
	cmds = append(cmds, fetchCmd)

	return cmds
}

func (m *Model) UpdateLastUpdated(t time.Time) {
	m.Table.UpdateLastUpdated(t)
}

func (m *Model) ResetRows() {
	m.Issues = nil
	m.BaseModel.ResetRows()
}

func FetchAllSections(
	ctx *context.ProgramContext,
) (sections []section.Section, fetchAllCmd tea.Cmd) {
	sectionConfigs := ctx.Config.IssuesSections
	fetchIssuesCmds := make([]tea.Cmd, 0, len(sectionConfigs))
	sections = make([]section.Section, 0, len(sectionConfigs))
	for i, sectionConfig := range sectionConfigs {
		sectionModel := NewModel(
			i+1,
			ctx,
			sectionConfig,
			time.Now(),
			time.Now(),
		) // 0 is the search section
		if sectionConfig.Layout.CreatorIcon.Hidden != nil {
			sectionModel.ShowAuthorIcon = !*sectionConfig.Layout.CreatorIcon.Hidden
		}
		sections = append(sections, &sectionModel)
		fetchIssuesCmds = append(
			fetchIssuesCmds,
			sectionModel.FetchNextPageSectionRows()...)
	}
	return sections, tea.Batch(fetchIssuesCmds...)
}

type SectionIssuesFetchedMsg struct {
	Issues     []data.IssueData
	TotalCount int
	PageInfo   data.PageInfo
	TaskId     string
}

type UpdateIssueMsg struct {
	IssueNumber      int
	Labels           *data.IssueLabels
	NewComment       *data.IssueComment
	IsClosed         *bool
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
	return "Issue"
}

func (m Model) GetItemPluralForm() string {
	return "Issues"
}

func (m Model) GetTotalCount() int {
	return m.TotalCount
}

func (m *Model) GetIsLoading() bool {
	return m.IsLoading
}

func (m *Model) SetIsLoading(val bool) {
	m.IsLoading = val
	m.Table.SetIsLoading(val)
}

func (m Model) GetPagerContent() string {
	pagerContent := ""
	if m.TotalCount > 0 {
		pagerContent = fmt.Sprintf(
			"%v %v • %v %v/%v • Fetched %v",
			constants.WaitingIcon,
			m.LastUpdated().Format("01/02 15:04:05"),
			m.SingularForm,
			m.Table.GetCurrItem()+1,
			m.TotalCount,
			len(m.Table.Rows),
		)
	}
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Render(pagerContent)
	return pager
}
