package section

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/search"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/styles"
	"github.com/dlvhdr/gh-dash/utils"
)

type Model struct {
	Id           int
	Config       config.SectionConfig
	Ctx          *context.ProgramContext
	Spinner      spinner.Model
	IsLoading    bool
	SearchBar    search.Model
	IsSearching  bool
	SearchValue  string
	Table        table.Model
	Type         string
	SingularForm string
	PluralForm   string
	Columns      []table.Column
}

func NewModel(
	id int,
	ctx *context.ProgramContext,
	cfg config.SectionConfig,
	sType string,
	columns []table.Column,
	singular, plural string,
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
		IsLoading:    false,
		SearchBar:    search.NewModel(sType, ctx, cfg.Filters),
		SearchValue:  cfg.Filters,
		IsSearching:  false,
	}
	m.Table = table.NewModel(
		m.GetDimensions(),
		m.Columns,
		nil,
		m.SingularForm,
		utils.StringPtr(emptyStateStyle.Render(
			fmt.Sprintf("No %s were found that match the given filters", m.PluralForm),
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
	NextRow() int
	PrevRow() int
	FirstItem() int
	LastItem() int
	FetchSectionRows() tea.Cmd
	BuildRows() []table.Row
}

type Search interface {
	SetIsSearching(val bool) tea.Cmd
	IsSearchFocused() bool
	ResetFilters()
	GetFilters() string
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
		Width:  m.Ctx.MainContentWidth - containerStyle.GetHorizontalPadding(),
		Height: m.Ctx.MainContentHeight - styles.SearchHeight,
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

	if oldDimensions.Height != newDimensions.Height || oldDimensions.Width != newDimensions.Width {
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
	if m.Table.Rows == nil && m.IsLoading == false {
		d := m.GetDimensions()
		return lipgloss.Place(
			d.Width,
			d.Height,
			lipgloss.Center,
			lipgloss.Center,
			fmt.Sprintf(
				"Enter a query to the search bar above by pressing %s and submit it with %s.",
				keyStyle.Render("/"),
				keyStyle.Render("Enter"),
			),
		)
	} else {
		return m.Table.View(m.GetSpinnerText())
	}
}

func (m *Model) GetSpinnerText() *string {
	var spinnerText *string
	if m.IsLoading {
		spinnerText = utils.StringPtr(lipgloss.JoinHorizontal(lipgloss.Top,
			spinnerStyle.Copy().Render(m.Spinner.View()),
			fmt.Sprintf("Fetching %s...", m.PluralForm),
		))
	}
	return spinnerText
}

func (m *Model) View() string {
	var search string
	search = m.SearchBar.View(*m.Ctx)

	return containerStyle.Copy().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			search,
			m.GetMainContent(),
		),
	)
}
