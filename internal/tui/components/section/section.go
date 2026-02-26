package section

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"text/template"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/go-sprout/sprout"
	timeregistry "github.com/go-sprout/sprout/registry/time"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/common"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prompt"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/search"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

type BaseModel struct {
	Id                        int
	Config                    config.SectionConfig
	Ctx                       *context.ProgramContext
	Spinner                   spinner.Model
	SearchBar                 search.Model
	IsSearching               bool
	SearchValue               string
	Table                     table.Model
	Type                      string
	SingularForm              string
	PluralForm                string
	Columns                   []table.Column
	TotalCount                int
	PageInfo                  *data.PageInfo
	PromptConfirmationBox     prompt.Model
	IsPromptConfirmationShown bool
	PromptConfirmationAction  string
	LastFetchTaskId           string
	IsSearchSupported         bool
	ShowAuthorIcon            bool
	IsFilteredByCurrentRemote bool
	IsLoading                 bool
}

type NewSectionOptions struct {
	Id          int
	Config      config.SectionConfig
	Ctx         *context.ProgramContext
	Type        string
	Columns     []table.Column
	Singular    string
	Plural      string
	LastUpdated time.Time
	CreatedAt   time.Time
}

func (options NewSectionOptions) GetConfigFiltersWithCurrentRemoteAdded(ctx *context.ProgramContext) string {
	searchValue := options.Config.Filters
	if !ctx.Config.SmartFilteringAtLaunch {
		return searchValue
	}
	repo, err := repository.Current()
	if err != nil {
		return searchValue
	}
	for token := range strings.FieldsSeq(searchValue) {
		if strings.HasPrefix(token, "repo:") {
			return searchValue
		}
	}
	return fmt.Sprintf("repo:%s/%s %s", repo.Owner, repo.Name, searchValue)
}

func NewModel(
	ctx *context.ProgramContext,
	options NewSectionOptions,
) BaseModel {
	filters := options.GetConfigFiltersWithCurrentRemoteAdded(ctx)
	isFilteredByCurrentRemote := false
	repo, err := repository.Current()
	if err == nil {
		currentCloneFilter := fmt.Sprintf("repo:%s/%s", repo.Owner, repo.Name)
		for token := range strings.FieldsSeq(filters) {
			if token == currentCloneFilter {
				isFilteredByCurrentRemote = true
				break
			}
		}
	}
	m := BaseModel{
		Ctx:          ctx,
		Id:           options.Id,
		Type:         options.Type,
		Config:       options.Config,
		Spinner:      spinner.Model{Spinner: spinner.Dot},
		Columns:      options.Columns,
		SingularForm: options.Singular,
		PluralForm:   options.Plural,
		SearchBar: search.NewModel(ctx, search.SearchOptions{
			Prefix:       fmt.Sprintf("is:%s", options.Type),
			InitialValue: filters,
		}),
		SearchValue:               filters,
		IsSearching:               false,
		IsFilteredByCurrentRemote: isFilteredByCurrentRemote,
		TotalCount:                0,
		PageInfo:                  nil,
		PromptConfirmationBox:     prompt.NewModel(ctx),
		ShowAuthorIcon:            ctx.Config.ShowAuthorIcons,
	}
	m.Table = table.NewModel(
		*ctx,
		m.GetDimensions(),
		options.LastUpdated,
		options.CreatedAt,
		m.Columns,
		nil,
		m.SingularForm,
		utils.StringPtr(m.Ctx.Styles.Section.EmptyStateStyle.Render(
			fmt.Sprintf(
				"No %s were found that match the given filters",
				m.PluralForm,
			),
		)),
		"Loading...",
		false,
	)
	return m
}

type Section interface {
	Identifier
	Component
	Table
	Search
	PromptConfirmation
	GetConfig() config.SectionConfig
	UpdateProgramContext(ctx *context.ProgramContext)
	MakeSectionCmd(cmd tea.Cmd) tea.Cmd
	GetPagerContent() string
	GetItemSingularForm() string
	GetItemPluralForm() string
	GetTotalCount() int
}

type Identifier interface {
	GetId() int
	GetType() string
}

type Component interface {
	Update(msg tea.Msg) (Section, tea.Cmd)
	View() string
}

type Table interface {
	NumRows() int
	GetCurrRow() data.RowData
	CurrRow() int
	NextRow() int
	PrevRow() int
	FirstItem() int
	LastItem() int
	FetchNextPageSectionRows() []tea.Cmd
	BuildRows() []table.Row
	ResetRows()
	GetIsLoading() bool
	SetIsLoading(val bool)
}

type Search interface {
	SetIsSearching(val bool) tea.Cmd
	IsSearchFocused() bool
	ResetFilters()
	GetFilters() string
	ResetPageInfo()
}

type PromptConfirmation interface {
	SetIsPromptConfirmationShown(val bool) tea.Cmd
	IsPromptConfirmationFocused() bool
	SetPromptConfirmationAction(action string)
	GetPromptConfirmationAction() string
	GetPromptConfirmation() string
}

func (m *BaseModel) GetDimensions() constants.Dimensions {
	return constants.Dimensions{
		Width:  max(0, m.Ctx.MainContentWidth-m.Ctx.Styles.Section.ContainerStyle.GetHorizontalPadding()),
		Height: max(0, m.Ctx.MainContentHeight-common.SearchHeight),
	}
}

func (m *BaseModel) GetConfig() config.SectionConfig {
	return m.Config
}

func (m *BaseModel) HasRepoNameInConfiguredFilter() bool {
	filters := m.SearchValue
	for token := range strings.FieldsSeq(filters) {
		if strings.HasPrefix(token, "repo:") {
			return true
		}
	}
	return false
}

func (m *BaseModel) HasCurrentRepoNameInConfiguredFilter() bool {
	filters := m.SearchValue
	repo, err := repository.Current()
	if err != nil {
		return false
	}
	currentCloneFilter := fmt.Sprintf("repo:%s/%s", repo.Owner, repo.Name)
	for token := range strings.FieldsSeq(filters) {
		if token == currentCloneFilter {
			return true
		}
	}
	return false
}

func (m *BaseModel) SyncSmartFilterWithSearchValue() {
	m.IsFilteredByCurrentRemote = m.HasCurrentRepoNameInConfiguredFilter()
}

func (m *BaseModel) GetSearchValue() string {
	searchValue := m.enrichSearchWithTemplateVars()
	repo, err := repository.Current()
	if err != nil {
		return searchValue
	}

	currentCloneFilter := fmt.Sprintf("repo:%s/%s", repo.Owner, repo.Name)
	var searchValueWithoutCurrentCloneFilter []string
	for token := range strings.FieldsSeq(searchValue) {
		if token != currentCloneFilter {
			searchValueWithoutCurrentCloneFilter = append(searchValueWithoutCurrentCloneFilter, token)
		}
	}
	if m.IsFilteredByCurrentRemote {
		return fmt.Sprintf("%s %s", currentCloneFilter,
			strings.Join(searchValueWithoutCurrentCloneFilter, " "))
	}
	return strings.Join(searchValueWithoutCurrentCloneFilter, " ")
}

func (m *BaseModel) enrichSearchWithTemplateVars() string {
	searchValue := m.SearchValue
	searchVars := struct{ Now time.Time }{
		Now: time.Now(),
	}
	sl := slog.New(log.Default())
	handler := sprout.New(sprout.WithRegistries(timeregistry.NewRegistry(), utils.NewRegistry()), sprout.WithLogger(sl))
	funcs := handler.Build()

	tmpl, err := template.New("search").Funcs(funcs).Parse(searchValue)
	if err != nil {
		log.Error("bad template", "err", err)
		return searchValue
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, searchVars)
	if err != nil {
		return searchValue
	}

	return buf.String()
}

func (m *BaseModel) UpdateProgramContext(ctx *context.ProgramContext) {
	m.Ctx = ctx
	newDimensions := m.GetDimensions()
	tableDimensions := constants.Dimensions{
		Height: max(0, newDimensions.Height-2),
		Width:  max(0, newDimensions.Width),
	}
	m.Table.SetDimensions(tableDimensions)
	m.Table.UpdateProgramContext(ctx)
	m.Table.SyncViewPortContent()
	m.SearchBar.UpdateProgramContext(ctx)
}

type SectionRowsFetchedMsg struct {
	SectionId int
	Issues    []data.RowData
}

func (msg SectionRowsFetchedMsg) GetSectionId() int {
	return msg.SectionId
}

func (m *BaseModel) GetId() int {
	return m.Id
}

func (m *BaseModel) GetType() string {
	return m.Type
}

func (m *BaseModel) CurrRow() int {
	return m.Table.GetCurrItem()
}

func (m *BaseModel) NextRow() int {
	return m.Table.NextItem()
}

func (m *BaseModel) PrevRow() int {
	return m.Table.PrevItem()
}

func (m *BaseModel) FirstItem() int {
	return m.Table.FirstItem()
}

func (m *BaseModel) LastItem() int {
	return m.Table.LastItem()
}

func (m *BaseModel) IsSearchFocused() bool {
	return m.IsSearching
}

func (m *BaseModel) GetIsLoading() bool {
	return m.IsLoading
}

func (m *BaseModel) SetIsSearching(val bool) tea.Cmd {
	m.IsSearching = val
	if val {
		m.SearchBar.Focus()
		return m.SearchBar.Init()
	} else {
		m.SearchBar.Blur()
		return nil
	}
}

func (m *BaseModel) ResetFilters() {
	m.SearchBar.SetValue(m.GetSearchValue())
}

func (m *BaseModel) ResetPageInfo() {
	m.PageInfo = nil
}

func (m *BaseModel) IsPromptConfirmationFocused() bool {
	return m.IsPromptConfirmationShown
}

func (m *BaseModel) SetIsPromptConfirmationShown(val bool) tea.Cmd {
	m.IsPromptConfirmationShown = val
	if val {
		m.PromptConfirmationBox.Focus()
		return m.PromptConfirmationBox.Init()
	}

	m.PromptConfirmationBox.Blur()
	return nil
}

func (m *BaseModel) SetPromptConfirmationAction(action string) {
	m.PromptConfirmationAction = action
}

func (m *BaseModel) GetPromptConfirmationAction() string {
	return m.PromptConfirmationAction
}

type SectionMsg struct {
	Id          int
	Type        string
	InternalMsg tea.Msg
}

func (m *BaseModel) MakeSectionCmd(cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}

	return func() tea.Msg {
		internalMsg := cmd()
		return SectionMsg{
			Id:          m.Id,
			Type:        m.Type,
			InternalMsg: internalMsg,
		}
	}
}

func (m *BaseModel) GetFilters() string {
	return m.GetSearchValue()
}

func (m *BaseModel) GetMainContent() string {
	if m.Table.Rows == nil {
		d := m.GetDimensions()
		return lipgloss.Place(
			d.Width,
			d.Height,
			lipgloss.Center,
			lipgloss.Center,

			fmt.Sprintf(
				"%s you can change the search query by pressing %s and submitting it with %s",
				lipgloss.NewStyle().Bold(true).Render(" Tip:"),
				m.Ctx.Styles.Section.KeyStyle.Render("/"),
				m.Ctx.Styles.Section.KeyStyle.Render("Enter"),
			),
		)
	} else {
		return m.Table.View()
	}
}

func (m *BaseModel) View() string {
	search := m.SearchBar.View(m.Ctx)
	return m.Ctx.Styles.Section.ContainerStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			search,
			m.GetMainContent(),
		),
	)
}

func (m *BaseModel) ResetRows() {
	m.Table.Rows = nil
	m.ResetPageInfo()
	m.Table.ResetCurrItem()
}

func (m *BaseModel) LastUpdated() time.Time {
	return m.Table.LastUpdated()
}

func (m *BaseModel) CreatedAt() time.Time {
	return m.Table.CreatedAt()
}

func (m *BaseModel) UpdateTotalItemsCount(count int) {
	m.Table.UpdateTotalItemsCount(count)
}

func (m *BaseModel) GetPromptConfirmation() string {
	if m.IsPromptConfirmationShown {
		var prompt string
		switch {
		case m.PromptConfirmationAction == "close" && m.Ctx.View == config.PRsView:
			prompt = "Are you sure you want to close this PR? (Y/n) "

		case m.PromptConfirmationAction == "reopen" && m.Ctx.View == config.PRsView:
			prompt = "Are you sure you want to reopen this PR? (Y/n) "

		case m.PromptConfirmationAction == "ready" && m.Ctx.View == config.PRsView:
			prompt = "Are you sure you want to mark this PR as ready? (Y/n) "

		case m.PromptConfirmationAction == "merge" && m.Ctx.View == config.PRsView:
			prompt = "Are you sure you want to merge this PR? (Y/n) "

		case m.PromptConfirmationAction == "update" && m.Ctx.View == config.PRsView:
			prompt = "Are you sure you want to update this PR? (Y/n) "

		case m.PromptConfirmationAction == "approveWorkflows" && m.Ctx.View == config.PRsView:
			prompt = "Are you sure you want to approve all workflows? (Y/n) "

		case m.PromptConfirmationAction == "close" && m.Ctx.View == config.IssuesView:
			prompt = "Are you sure you want to close this issue? (Y/n) "

		case m.PromptConfirmationAction == "reopen" && m.Ctx.View == config.IssuesView:
			prompt = "Are you sure you want to reopen this issue? (Y/n) "
		case m.PromptConfirmationAction == "delete" && m.Ctx.View == config.RepoView:
			prompt = "Are you sure you want to delete this branch? (Y/n) "
		case m.PromptConfirmationAction == "new" && m.Ctx.View == config.RepoView:
			prompt = "Enter branch name: "
		case m.PromptConfirmationAction == "create_pr" && m.Ctx.View == config.RepoView:
			prompt = "Enter PR title: "
		case m.PromptConfirmationAction == "done_all" && m.Ctx.View == config.NotificationsView:
			prompt = "Are you sure you want to mark all as done? (Y/n) "
		}

		m.PromptConfirmationBox.SetPrompt(prompt)

		return m.Ctx.Styles.ListViewPort.PagerStyle.Render(m.PromptConfirmationBox.View())
	}

	return ""
}
