package table

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/components/listviewport"
	"github.com/dlvhdr/gh-prs/ui/constants"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

type Model struct {
	Spinner      spinner.Model
	IsLoading    bool
	Columns      []Column
	Rows         []Row
	EmptyState   *string
	dimensions   constants.Dimensions
	rowsViewport listviewport.Model
}

type Column struct {
	Title string
	Width *int
	Grow  *bool
}

type Row []string

func NewModel(dimensions constants.Dimensions, columns []Column, rows []Row, itemTypeLabel string, emptyState *string) Model {
	return Model{
		Spinner:      spinner.Model{},
		IsLoading:    true,
		Columns:      columns,
		Rows:         rows,
		EmptyState:   emptyState,
		dimensions:   dimensions,
		rowsViewport: listviewport.NewModel(dimensions, itemTypeLabel, len(rows), 2),
	}
}

func (m Model) View() string {
	header := m.renderHeader()
	rows := m.renderRows()

	return lipgloss.JoinVertical(lipgloss.Top, header, rows)
}

func (m *Model) SetDimensions(dimensions constants.Dimensions) {
	m.dimensions = dimensions
	m.rowsViewport.SetDimensions(constants.Dimensions{
		Width:  dimensions.Width,
		Height: dimensions.Height - lipgloss.Height(headerStyle.Render("Header")),
	})
}

func (m *Model) GetCurrItem() int {
	return m.rowsViewport.GetCurrItem()
}

func (m *Model) PrevItem() int {
	currItem := m.rowsViewport.PrevItem()
	m.syncViewPortContent()

	return currItem
}

func (m *Model) NextItem() int {
	currItem := m.rowsViewport.NextItem()
	m.syncViewPortContent()

	return currItem
}

func (m *Model) syncViewPortContent() {
	headerColumns := m.renderHeaderColumns()
	renderedRows := make([]string, 0, len(m.Rows))
	for i := range m.Rows {
		renderedRows = append(renderedRows, m.renderRow(i, headerColumns))
	}

	m.rowsViewport.SyncViewPort(
		lipgloss.JoinVertical(lipgloss.Top, renderedRows...),
	)
}

func (m *Model) SetRows(rows []Row) {
	m.Rows = rows
	m.rowsViewport.NumItems = len(m.Rows)
	m.syncViewPortContent()
}

func (m *Model) OnLineDown() {
	m.rowsViewport.NextItem()
}

func (m *Model) OnLineUp() {
	m.rowsViewport.PrevItem()
}

func (m *Model) renderHeaderColumns() []string {
	renderedColumns := make([]string, len(m.Columns))
	takenWidth := 0
	numGrowingColumns := 0
	for i, column := range m.Columns {
		if column.Grow != nil && *column.Grow {
			numGrowingColumns += 1
			continue
		}

		if column.Width != nil {
			renderedColumns[i] = titleCellStyle.Copy().Width(*column.Width).Render(column.Title)
			takenWidth += *column.Width
			continue
		}

		if len(column.Title) == 1 {
			takenWidth += styles.SingleRuneWidth
			renderedColumns[i] = singleRuneTitleCellStyle.Copy().Width(styles.SingleRuneWidth).Render(column.Title)
			continue
		}

		cell := titleCellStyle.Copy().Render(column.Title)
		renderedColumns[i] = cell
		takenWidth += lipgloss.Width(cell)
	}

	leftoverWidth := m.dimensions.Width - takenWidth
	if numGrowingColumns == 0 {
		return renderedColumns
	}

	growCellWidth := leftoverWidth / numGrowingColumns
	for i, column := range m.Columns {
		if column.Grow == nil || !*column.Grow {
			continue
		}

		renderedColumns[i] = titleCellStyle.Copy().Width(growCellWidth).Render(column.Title)
	}

	return renderedColumns
}

func (m *Model) renderHeader() string {
	headerColumns := m.renderHeaderColumns()
	return headerStyle.Copy().
		Width(m.dimensions.Width).
		MaxWidth(m.dimensions.Width).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, headerColumns...))
}

func (m *Model) renderRows() string {
	if len(m.Rows) == 0 && m.EmptyState != nil {
		return *m.EmptyState
	}

	return lipgloss.JoinVertical(lipgloss.Top, m.rowsViewport.View())

}

func (m *Model) renderRow(rowId int, headerColumns []string) string {
	var style lipgloss.Style
	if m.rowsViewport.GetCurrItem() == rowId {
		style = selectedCellStyle
	} else {
		style = cellStyle
	}

	renderedColumns := make([]string, len(m.Columns))
	for i, column := range m.Rows[rowId] {
		renderedColumns = append(
			renderedColumns,
			style.Copy().Width(lipgloss.Width(headerColumns[i])).Render(column),
		)
	}

	return rowStyle.Copy().
		Render(lipgloss.JoinHorizontal(lipgloss.Left, renderedColumns...))
}
