package issuessection

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/issue"
	"github.com/dlvhdr/gh-dash/ui/components/section"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

const SectionType = "issues"

type Model struct {
	Issues  []data.IssueData
	section section.Model
	error   error
}

func NewModel(id int, ctx *context.ProgramContext, config config.SectionConfig) Model {
	m := Model{
		Issues: []data.IssueData{},
		section: section.Model{
			Id:        id,
			Config:    config,
			Ctx:       ctx,
			Spinner:   spinner.Model{Spinner: spinner.Dot},
			IsLoading: true,
			Type:      SectionType,
		},
		error: nil,
	}

	m.section.Table = table.NewModel(
		m.getDimensions(),
		m.GetSectionColumns(),
		m.BuildRows(),
		"Issue",
		utils.StringPtr(emptyStateStyle.Render(fmt.Sprintf(
			"No issues were found that match the given filters: %s",
			lipgloss.NewStyle().Italic(true).Render(m.section.Config.Filters),
		))),
	)

	return m
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case SectionIssuesFetchedMsg:
		m.Issues = msg.Issues
		m.section.IsLoading = false
		m.section.Table.SetRows(m.BuildRows())
		m.error = msg.Err

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

func (m *Model) getDimensions() constants.Dimensions {
	return constants.Dimensions{
		Width:  m.section.Ctx.MainContentWidth - containerStyle.GetHorizontalPadding(),
		Height: m.section.Ctx.MainContentHeight - 2,
	}
}

func (m *Model) View() string {
	var spinnerText *string
	if m.section.IsLoading {
		spinnerText = utils.StringPtr(lipgloss.JoinHorizontal(lipgloss.Top,
			spinnerStyle.Copy().Render(m.section.Spinner.View()),
			"Fetching Issues...",
		))
	}

	if m.error != nil {
		spinnerText = utils.StringPtr(fmt.Sprintf("Error while fetching issues: %v", m.error))
	}

	return containerStyle.Copy().Render(
		m.section.Table.View(spinnerText),
	)
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	oldDimensions := m.getDimensions()
	m.section.Ctx = ctx
	newDimensions := m.getDimensions()
	m.section.Table.SetDimensions(newDimensions)

	if oldDimensions.Height != newDimensions.Height || oldDimensions.Width != newDimensions.Width {
		m.section.Table.SyncViewPortContent()
	}
}

func (m *Model) GetSectionColumns() []table.Column {
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
		issueModel := issue.Issue{Data: currIssue, Width: m.getDimensions().Width}
		rows = append(rows, issueModel.ToTableRow())
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
	issue := m.Issues[m.section.Table.GetCurrItem()]
	return &issue
}

func (m *Model) NextRow() int {
	return m.section.Table.NextItem()
}

func (m *Model) PrevRow() int {
	return m.section.Table.PrevItem()
}

func (m *Model) FirstItem() int {
	return m.section.Table.FirstItem()
}

func (m *Model) LastItem() int {
	return m.section.Table.LastItem()
}

func (m *Model) FetchSectionRows() tea.Cmd {
	m.error = nil
	if m == nil {
		return nil
	}
	m.Issues = nil
	m.section.Table.ResetCurrItem()
	m.section.Table.Rows = nil
	m.section.IsLoading = true
	var cmds []tea.Cmd
	cmds = append(cmds, m.section.CreateNextTickCmd(spinner.Tick))

	cmds = append(cmds, func() tea.Msg {
		limit := m.section.Config.Limit
		if limit == nil {
			limit = &m.section.Ctx.Config.Defaults.IssuesLimit
		}
		fetchedIssues, err := data.FetchIssues(m.section.Config.Filters, *limit)
		if err != nil {
			return SectionIssuesFetchedMsg{
				SectionId: m.section.Id,
				Issues:    []data.IssueData{},
				Err:       err,
			}
		}

		sort.Slice(fetchedIssues, func(i, j int) bool {
			return fetchedIssues[i].UpdatedAt.After(fetchedIssues[j].UpdatedAt)
		})
		return SectionIssuesFetchedMsg{
			SectionId: m.section.Id,
			Issues:    fetchedIssues,
		}
	})

	return tea.Batch(cmds...)
}

func (m *Model) Id() int {
	return m.section.Id
}

func (m *Model) GetIsLoading() bool {
	return m.section.IsLoading
}

func FetchAllSections(ctx context.ProgramContext) (sections []section.Section, fetchAllCmd tea.Cmd) {
	sectionConfigs := ctx.Config.IssuesSections
	fetchIssuesCmds := make([]tea.Cmd, 0, len(sectionConfigs))
	sections = make([]section.Section, 0, len(sectionConfigs))
	for i, sectionConfig := range sectionConfigs {
		sectionModel := NewModel(i, &ctx, sectionConfig)
		sections = append(sections, &sectionModel)
		fetchIssuesCmds = append(fetchIssuesCmds, sectionModel.FetchSectionRows())
	}
	return sections, tea.Batch(fetchIssuesCmds...)
}
