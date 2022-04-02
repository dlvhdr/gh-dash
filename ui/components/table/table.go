package table

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-dash/ui/components/listviewport"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/styles"
)

type Model struct {
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
		Columns:      columns,
		Rows:         rows,
		EmptyState:   emptyState,
		dimensions:   dimensions,
		rowsViewport: listviewport.NewModel(dimensions, itemTypeLabel, len(rows), 2),
	}
}

func (m Model) View(spinnerText *string) string {
	header := m.renderHeader()
	body := m.renderBody(spinnerText)

	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (m *Model) SetDimensions(dimensions constants.Dimensions) {
	m.dimensions = dimensions
	m.rowsViewport.SetDimensions(constants.Dimensions{
		Width:  m.dimensions.Width,
		Height: m.dimensions.Height,
	})
}

func (m *Model) ResetCurrItem() {
	m.rowsViewport.ResetCurrItem()
}

func (m *Model) GetCurrItem() int {
	return m.rowsViewport.GetCurrItem()
}

func (m *Model) PrevItem() int {
	currItem := m.rowsViewport.PrevItem()
	m.SyncViewPortContent()

	return currItem
}

func (m *Model) NextItem() int {
	currItem := m.rowsViewport.NextItem()
	m.SyncViewPortContent()

	return currItem
}

func (m *Model) FirstItem() int {
	currItem := m.rowsViewport.FirstItem()
	m.SyncViewPortContent()

	return currItem
}

func (m *Model) LastItem() int {
	currItem := m.rowsViewport.LastItem()
	m.SyncViewPortContent()

	return currItem
}

func (m *Model) SyncViewPortContent() {
	headerColumns := m.renderHeaderColumns()
	renderedRows := make([]string, 0, len(m.Rows))
	for i := range m.Rows {
		renderedRows = append(renderedRows, m.renderRow(i, headerColumns))
	}

	m.rowsViewport.SyncViewPort(
		lipgloss.JoinVertical(lipgloss.Left, renderedRows...),
	)
}

func (m *Model) SetRows(rows []Row) {
	m.Rows = rows

	m.rowsViewport.SetNumItems(len(rows))
	m.SyncViewPortContent()
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
			renderedColumns[i] = titleCellStyle.Copy().
				Width(*column.Width).
				MaxWidth(*column.Width).
				Render(column.Title)
			takenWidth += *column.Width
			continue
		}

		if len(column.Title) == 1 {
			takenWidth += styles.SingleRuneWidth
			renderedColumns[i] = singleRuneTitleCellStyle.Copy().
				Width(styles.SingleRuneWidth).
				MaxWidth(styles.SingleRuneWidth).
				Render(column.Title)
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

		renderedColumns[i] = titleCellStyle.Copy().Width(growCellWidth).MaxWidth(growCellWidth).Render(column.Title)
	}

	return renderedColumns
}

func (m *Model) renderHeader() string {
	headerColumns := m.renderHeaderColumns()
	header := lipgloss.JoinHorizontal(lipgloss.Top, headerColumns...)
	return headerStyle.Copy().
		Width(m.dimensions.Width).
		MaxWidth(m.dimensions.Width).
		Render(header)
}

func (m *Model) renderBody(spinnerText *string) string {
	bodyStyle := lipgloss.NewStyle().
		Height(m.dimensions.Height)
	if spinnerText != nil {
		return bodyStyle.Render(*spinnerText)
	} else if len(m.Rows) == 0 && m.EmptyState != nil {
		return bodyStyle.Render(*m.EmptyState)
	}

	return m.rowsViewport.View()
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
		colWidth := lipgloss.Width(headerColumns[i])
		col := style.Copy().Width(colWidth).MaxWidth(colWidth).Render(column)
		renderedColumns = append(renderedColumns, col)
	}

	return rowStyle.Copy().Render(
		lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...),
	)
}
