package section

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlvhdr/gh-dash/config"
	"github.com/dlvhdr/gh-dash/data"
	"github.com/dlvhdr/gh-dash/ui/components/search"
	"github.com/dlvhdr/gh-dash/ui/components/table"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

type Model struct {
	Id          int
	Config      config.SectionConfig
	Ctx         *context.ProgramContext
	Spinner     spinner.Model
	IsLoading   bool
	Search      search.Model
	IsSearching bool
	Table       table.Model
	Type        string
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
	GetSectionColumns() []table.Column
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
		Height: m.Ctx.MainContentHeight - 2 - styles.SearchHeight,
	}
}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	oldDimensions := m.GetDimensions()
	m.Ctx = ctx
	newDimensions := m.GetDimensions()
	m.Table.SetDimensions(newDimensions)

	if oldDimensions.Height != newDimensions.Height || oldDimensions.Width != newDimensions.Width {
		m.Table.SyncViewPortContent()
		m.Search.UpdateProgramContext(ctx)
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
		m.Search.Focus()
		return m.Search.Init()
	} else {
		m.Search.Blur()
		return nil
	}
}

func (m *Model) ResetFilters() {
	m.Search.SetValue(m.Config.Filters)
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
	return m.Search.Value()
}
