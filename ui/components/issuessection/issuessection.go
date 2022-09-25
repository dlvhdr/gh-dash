package issuessection

import (
	"log"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/issue"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/keys"
	"github.com/dlvhdr/gh-dash/utils"
)

const SectionType = "issue"

type Model struct {
	section.Model
	Issues []data.IssueData
}

func NewModel(id int, ctx *context.ProgramContext, cfg config.IssuesSectionConfig) Model {
	m := Model{
		section.NewModel(
			id,
			ctx,
			cfg.ToSectionConfig(),
			SectionType,
			GetSectionColumns(ctx),
			"Issue",
			"Issues",
		),
		[]data.IssueData{},
	}

	return m
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch {
		case key.Matches(msg, keys.IssueKeys.Close):
			cmd = m.close()

		case key.Matches(msg, keys.IssueKeys.Reopen):
			cmd = m.reopen()

		case msg.Type == tea.KeyEnter:
			m.SearchValue = m.SearchBar.Value()
			m.SetIsSearching(false)
			return &m, m.FetchSectionRows()

		case msg.Type == tea.KeyCtrlC, msg.Type == tea.KeyEsc:
			m.SearchBar.SetValue(m.SearchValue)
			blinkCmd := m.SetIsSearching(false)
			return &m, blinkCmd

		}

	case UpdateIssueMsg:
		for i, currIssue := range m.Issues {
			if currIssue.Number == msg.IssueNumber {
				if msg.IsClosed != nil {
					if *msg.IsClosed == true {
						currIssue.State = "CLOSED"
					} else {
						currIssue.State = "OPEN"
					}
				}
				if msg.NewComment != nil {
					currIssue.Comments.Nodes = append(currIssue.Comments.Nodes, *msg.NewComment)
				}
				m.Issues[i] = currIssue
				m.Table.SetRows(m.BuildRows())
				break
			}
		}

	case section.SectionMsg:
		if msg.Id != m.Id || msg.Type != m.Type {
			return &m, nil
		}

		switch iMsg := msg.InternalMsg.(type) {

		case SectionIssuesFetchedMsg:
			m.Issues = iMsg.Issues
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

func GetSectionColumns(ctx *context.ProgramContext) []table.Column {
	layout := ctx.Config.Defaults.Layout.Issues
	log.Printf("updated at: %v\n", *layout.UpdatedAt.Width)
	return []table.Column{
		{
			Title:  "",
			Width:  layout.UpdatedAt.Width,
			Hidden: layout.UpdatedAt.Hidden,
		},
		{
			Title:  "",
			Width:  layout.State.Width,
			Hidden: layout.State.Hidden,
		},
		{
			Title:  "",
			Width:  layout.Repo.Width,
			Hidden: layout.Repo.Hidden,
		},
		{
			Title:  "Title",
			Grow:   utils.BoolPtr(true),
			Hidden: layout.Title.Hidden,
		},
		{
			Title:  "Creator",
			Width:  layout.Creator.Width,
			Hidden: layout.Creator.Hidden,
		},
		{
			Title:  "Assignees",
			Width:  layout.Assignees.Width,
			Hidden: layout.Assignees.Hidden,
		},
		{
			Title:  "",
			Width:  &issueNumCommentsCellWidth,
			Hidden: layout.Comments.Hidden,
		},
		{
			Title:  "",
			Width:  &issueNumCommentsCellWidth,
			Hidden: layout.Reactions.Hidden,
		},
	}
}

func (m *Model) BuildRows() []table.Row {
	var rows []table.Row
	for _, currIssue := range m.Issues {
		issueModel := issue.Issue{Data: currIssue}
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

type SectionIssuesFetchedMsg struct {
	SectionId int
	Issues    []data.IssueData
	Err       error
}

func (msg SectionIssuesFetchedMsg) GetSectionId() int {
	return msg.SectionId
}

func (msg SectionIssuesFetchedMsg) GetSectionType() string {
	return SectionType
}

func (m *Model) GetCurrRow() data.RowData {
	if len(m.Issues) == 0 {
		return nil
	}
	issue := m.Issues[m.Table.GetCurrItem()]
	return &issue
}

func (m *Model) FetchSectionRows() tea.Cmd {
	if m == nil {
		return nil
	}
	m.Issues = nil
	m.Table.ResetCurrItem()
	m.Table.Rows = nil
	m.IsLoading = true
	var cmds []tea.Cmd
	cmds = append(cmds, m.CreateNextTickCmd(spinner.Tick))

	cmd := func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.IssuesLimit
		}
		fetchedIssues, err := data.FetchIssues(m.GetFilters(), *limit)
		if err != nil {
			return SectionIssuesFetchedMsg{
				Issues: []data.IssueData{},
			}
		}

		sort.Slice(fetchedIssues, func(i, j int) bool {
			return fetchedIssues[i].UpdatedAt.After(fetchedIssues[j].UpdatedAt)
		})
		return SectionIssuesFetchedMsg{
			Issues: fetchedIssues,
		}
	}
	cmds = append(cmds, m.MakeSectionCmd(cmd))

	return tea.Batch(cmds...)
}

func FetchAllSections(ctx context.ProgramContext) (sections []section.Section, fetchAllCmd tea.Cmd) {
	sectionConfigs := ctx.Config.IssuesSections
	fetchIssuesCmds := make([]tea.Cmd, 0, len(sectionConfigs))
	sections = make([]section.Section, 0, len(sectionConfigs))
	for i, sectionConfig := range sectionConfigs {
		sectionModel := NewModel(i+1, &ctx, sectionConfig) // 0 is the search section
		sections = append(sections, &sectionModel)
		fetchIssuesCmds = append(fetchIssuesCmds, sectionModel.FetchSectionRows())
	}
	return sections, tea.Batch(fetchIssuesCmds...)
}

type UpdateIssueMsg struct {
	IssueNumber int
	NewComment  *data.Comment
	IsClosed    *bool
}
