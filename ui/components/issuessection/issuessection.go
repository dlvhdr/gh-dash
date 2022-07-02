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
	"github.com/dlvhdr/gh-dash/ui/styles"
	"github.com/dlvhdr/gh-dash/utils"
)

const SectionType = "issues"

type Model struct {
	section.Model
	Issues []data.IssueData
	error  error
}

func NewModel(id int, ctx *context.ProgramContext, config config.SectionConfig) Model {
	section := section.Model{
		Id:        id,
		Config:    config,
		Ctx:       ctx,
		Spinner:   spinner.Model{Spinner: spinner.Dot},
		IsLoading: true,
		Type:      SectionType,
	}
	m := Model{
		section,
		[]data.IssueData{},
		nil,
	}

	m.Table = table.NewModel(
		m.getDimensions(),
		m.GetSectionColumns(),
		m.BuildRows(),
		"Issue",
		utils.StringPtr(emptyStateStyle.Render(fmt.Sprintf(
			"No issues were found that match the given filters: %s",
			lipgloss.NewStyle().Italic(true).Render(m.Config.Filters),
		))),
	)

	return m
}

func (m Model) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case SectionIssuesFetchedMsg:
		m.Issues = msg.Issues
		m.IsLoading = false
		m.Table.SetRows(m.BuildRows())
		m.error = msg.Err

	case section.SectionTickMsg:
		if !m.IsLoading {
			return &m, nil
		}

		var internalTickCmd tea.Cmd
		m.Spinner, internalTickCmd = m.Spinner.Update(msg.InternalTickMsg)
		cmd = m.CreateNextTickCmd(internalTickCmd)
	}

	return &m, cmd
}

func (m *Model) getDimensions() constants.Dimensions {
	return constants.Dimensions{
		Width:  m.Ctx.MainContentWidth - containerStyle.GetHorizontalPadding(),
		Height: m.Ctx.MainContentHeight - 2 - styles.SearchHeight,
	}
}

func (m *Model) View() string {
	var spinnerText *string
	if m.IsLoading {
		spinnerText = utils.StringPtr(lipgloss.JoinHorizontal(lipgloss.Top,
			spinnerStyle.Copy().Render(m.Spinner.View()),
			"Fetching Issues...",
		))
	}

	if m.error != nil {
		spinnerText = utils.StringPtr(fmt.Sprintf("Error while fetching issues: %v", m.error))
	}

	return containerStyle.Copy().Render(
		m.Table.View(spinnerText),
	)
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
	issue := m.Issues[m.Table.GetCurrItem()]
	return &issue
}

func (m *Model) FetchSectionRows() tea.Cmd {
	m.error = nil
	if m == nil {
		return nil
	}
	m.Issues = nil
	m.Table.ResetCurrItem()
	m.Table.Rows = nil
	m.IsLoading = true
	var cmds []tea.Cmd
	cmds = append(cmds, m.CreateNextTickCmd(spinner.Tick))

	cmds = append(cmds, func() tea.Msg {
		limit := m.Config.Limit
		if limit == nil {
			limit = &m.Ctx.Config.Defaults.IssuesLimit
		}
		fetchedIssues, err := data.FetchIssues(m.Config.Filters, *limit)
		if err != nil {
			return SectionIssuesFetchedMsg{
				SectionId: m.Id,
				Issues:    []data.IssueData{},
				Err:       err,
			}
		}

		sort.Slice(fetchedIssues, func(i, j int) bool {
			return fetchedIssues[i].UpdatedAt.After(fetchedIssues[j].UpdatedAt)
		})
		return SectionIssuesFetchedMsg{
			SectionId: m.Id,
			Issues:    fetchedIssues,
		}
	})

	return tea.Batch(cmds...)
}

func (m *Model) SetIsSearching(val bool) tea.Cmd {
	m.IsSearching = val
	if val {
		m.Search.Focus()
		return m.Search.Init()
	} else {
		m.Search.Blur()
		return nil
	}
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
