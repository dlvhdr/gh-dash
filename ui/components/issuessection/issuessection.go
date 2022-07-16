package issuessection

import (
	"sort"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/issue"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

const SectionType = "issue"

type Model struct {
	section.Model
	Issues []data.IssueData
}

func NewModel(id int, ctx *context.ProgramContext, cfg config.SectionConfig) Model {
	m := Model{
		section.NewModel(
			id,
			ctx,
			cfg,
			SectionType,
			GetSectionColumns(),
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

func GetSectionColumns() []table.Column {
	return []table.Column{
		{
			Title: "",
			Width: &updatedAtCellWidth,
		},
		{
			Title: "",
		},
		{
			Title: "",
			Width: &issueRepoCellWidth,
		},
		{
			Title: "Title",
			Grow:  utils.BoolPtr(true),
		},
		{
			Title: "Creator",
		},
		{
			Title: "Assignees",
			Width: &issueAssigneesCellWidth,
		},
		{
			Title: "",
			Width: &issueNumCommentsCellWidth,
		},
		{
			Title: "",
			Width: &issueNumCommentsCellWidth,
		},
	}
}

func (m *Model) BuildRows() []table.Row {
	var rows []table.Row
	for _, currIssue := range m.Issues {
		issueModel := issue.Issue{Data: currIssue, Width: m.GetDimensions().Width}
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
