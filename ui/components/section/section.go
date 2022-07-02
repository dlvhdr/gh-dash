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
	GetIsSearching() bool
	ResetFilters()
}

func (m *Model) CreateNextTickCmd(nextTickCmd tea.Cmd) tea.Cmd {
	if m == nil || nextTickCmd == nil {
		return nil
	}
	return func() tea.Msg {
		return SectionTickMsg{
			SectionId:       m.Id,
			InternalTickMsg: nextTickCmd(),
			Type:            m.Type,
		}
	}
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
	}

	m.Search.UpdateProgramContext(ctx)
}

type SectionMsg interface {
	GetSectionId() int
	GetSectionType() string
}

type SectionRowsFetchedMsg struct {
	SectionId int
	Issues    []data.RowData
}

func (msg SectionRowsFetchedMsg) GetSectionId() int {
	return msg.SectionId
}

type SectionTickMsg struct {
	SectionId       int
	InternalTickMsg tea.Msg
	Type            string
}

func (m *Model) GetId() int {
	return m.Id
}

func (m *Model) GetType() string {
	return m.Type
}

func (msg SectionTickMsg) GetSectionId() int {
	return msg.SectionId
}

func (msg SectionTickMsg) GetSectionType() string {
	return msg.Type
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

func (m *Model) GetIsSearching() bool {
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
	m.Search.ResetValue()
}
