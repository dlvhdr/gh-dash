package section

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/common"
	"github.com/dlvhdr/gh-dash/ui/components/search"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/utils"
)

type Model struct {
	Id           int
	Config       config.SectionConfig
	Ctx          *context.ProgramContext
	Spinner      spinner.Model
	SearchBar    search.Model
	IsSearching  bool
	SearchValue  string
	Table        table.Model
	Type         string
	SingularForm string
	PluralForm   string
	Columns      []table.Column
	TotalCount   int
	PageInfo     *data.PageInfo
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.SectionConfig,
	sType string,
	columns []table.Column,
	singular, plural string,
	lastUpdated time.Time,
) Model {
	m := Model{
		Id:           id,
		Type:         sType,
		Config:       cfg,
		Ctx:          ctx,
		Spinner:      spinner.Model{Spinner: spinner.Dot},
		Columns:      columns,
		SingularForm: singular,
		PluralForm:   plural,
		SearchBar:    search.NewModel(sType, ctx, cfg.Filters),
		SearchValue:  cfg.Filters,
		IsSearching:  false,
		TotalCount:   0,
		PageInfo:     nil,
	}
	m.Table = table.NewModel(
		*ctx,
		m.GetDimensions(),
		lastUpdated,
		m.Columns,
		nil,
		m.SingularForm,
		utils.StringPtr(m.Ctx.Styles.Section.EmptyStateStyle.Render(
			fmt.Sprintf(
				"No %s were found that match the given filters",
				m.PluralForm,
			),
		)),
	)
	return m
}

type Section interface {
	Identifier
	Component
	Table
	Search
	UpdateProgramContext(ctx *context.ProgramContext)
	MakeSectionCmd(cmd tea.Cmd) tea.Cmd
	LastUpdated() time.Time
	UpdateLastUpdated(time.Time)
	GetPagerContent() string
	GetItemSingularForm() string
	GetItemPluralForm() string
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
}

type Search interface {
	SetIsSearching(val bool) tea.Cmd
	IsSearchFocused() bool
	ResetFilters()
	GetFilters() string
	ResetPageInfo()
}

func (m *Model) CreateNextTickCmd(nextTickCmd tea.Cmd) tea.Cmd {
	if m == nil || nextTickCmd == nil {
		return nil
	}
	return m.MakeSectionCmd(func() tea.Msg {
		return SectionTickMsg{
			InternalTickMsg: nextTickCmd(),
		}
	})
}

func (m *Model) GetDimensions() constants.Dimensions {
	return constants.Dimensions{
		Width:  m.Ctx.MainContentWidth - m.Ctx.Styles.Section.ContainerStyle.GetHorizontalPadding(),
		Height: m.Ctx.MainContentHeight - common.SearchHeight,
	}
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	oldDimensions := m.GetDimensions()
	m.Ctx = ctx
	newDimensions := m.GetDimensions()
	tableDimensions := constants.Dimensions{
		Height: newDimensions.Height - 2,
		Width:  newDimensions.Width,
	}
	m.Table.SetDimensions(tableDimensions)
	m.Table.UpdateProgramContext(ctx)

	if oldDimensions.Height != newDimensions.Height ||
		oldDimensions.Width != newDimensions.Width {
		m.Table.SyncViewPortContent()
		m.SearchBar.UpdateProgramContext(ctx)
	}
}

type SectionRowsFetchedMsg struct {
	SectionId int
	Issues    []data.RowData
}

func (msg SectionRowsFetchedMsg) GetSectionId() int {
	return msg.SectionId
}

type SectionTickMsg struct {
	InternalTickMsg tea.Msg
}

func (m *Model) GetId() int {
	return m.Id
}

func (m *Model) GetType() string {
	return m.Type
}

func (m *Model) CurrRow() int {
	return m.Table.GetCurrItem()
}

func (m *Model) NextRow() int {
	return m.Table.NextItem()
}

func (m *Model) PrevRow() int {
	return m.Table.PrevItem()
}

func (m *Model) FirstItem() int {
	return m.Table.FirstItem()
}

func (m *Model) LastItem() int {
	return m.Table.LastItem()
}

func (m *Model) IsSearchFocused() bool {
	return m.IsSearching
}

func (m *Model) SetIsSearching(val bool) tea.Cmd {
	m.IsSearching = val
	if val {
		m.SearchBar.Focus()
		return m.SearchBar.Init()
	} else {
		m.SearchBar.Blur()
		return nil
	}
}

func (m *Model) ResetFilters() {
	m.SearchBar.SetValue(m.Config.Filters)
}

func (m *Model) ResetPageInfo() {
	m.PageInfo = nil
}

type SectionMsg struct {
	Id          int
	Type        string
	InternalMsg tea.Msg
}

func (m *Model) MakeSectionCmd(cmd tea.Cmd) tea.Cmd {
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

func (m *Model) GetFilters() string {
	return m.SearchBar.Value()
}

func (m *Model) GetMainContent() string {
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

func (m *Model) View() string {
	var search string
	search = m.SearchBar.View(*m.Ctx)

	return m.Ctx.Styles.Section.ContainerStyle.Copy().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			search,
			m.GetMainContent(),
		),
	)
}

func (m *Model) LastUpdated() time.Time {
	return m.Table.LastUpdated()
}

func (m *Model) UpdateLastUpdated(t time.Time) {
	m.Table.UpdateLastUpdated(t)
}

func (m *Model) UpdateTotalItemsCount(count int) {
	m.Table.UpdateTotalItemsCount(count)
}

func (m *Model) GetPagerContent() string {
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
	pager := m.Ctx.Styles.ListViewPort.PagerStyle.Copy().Render(pagerContent)
	return pager
}
