package table

import (
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/dlvhdr/gh-dash/ui/common"
	"github.com/dlvhdr/gh-dash/ui/components/listviewport"
	"github.com/dlvhdr/gh-dash/ui/constants"
	"github.com/dlvhdr/gh-dash/ui/context"
)

type Model struct {
	ctx          context.ProgramContext
	Columns      []Column
	Rows         []Row
	EmptyState   *string
	dimensions   constants.Dimensions
	rowsViewport listviewport.Model
}

type Column struct {
	Title  string
	Hidden *bool
	Width  *int
	Grow   *bool
}

type Row []string

func NewModel(
	ctx context.ProgramContext,
	dimensions constants.Dimensions,
	lastUpdated time.Time,
	columns []Column,
	rows []Row,
	itemTypeLabel string,
	emptyState *string,
) Model {
	itemHeight := 1
	if ctx.Config.Theme.Ui.Table.ShowSeparator {
		itemHeight = 2
	}
	return Model{
		ctx:        ctx,
		Columns:    columns,
		Rows:       rows,
		EmptyState: emptyState,
		dimensions: dimensions,
		rowsViewport: listviewport.NewModel(
			ctx,
			dimensions,
			lastUpdated,
			itemTypeLabel,
			len(rows),
			itemHeight,
		),
	}
}

func (m Model) View() string {
	header := m.renderHeader()
	body := m.renderBody()

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
	m.rowsViewport.SetNumItems(len(m.Rows))
	m.SyncViewPortContent()
}

func (m *Model) OnLineDown() {
	m.rowsViewport.NextItem()
}

func (m *Model) OnLineUp() {
	m.rowsViewport.PrevItem()
}

func (m *Model) getShownColumns() []Column {
	shownColumns := make([]Column, 0, len(m.Columns))
	for _, col := range m.Columns {
		if col.Hidden != nil && *col.Hidden {
			continue
		}

		shownColumns = append(shownColumns, col)
	}
	return shownColumns
}

func (m *Model) renderHeaderColumns() []string {
	shownColumns := m.getShownColumns()
	renderedColumns := make([]string, len(shownColumns))
	takenWidth := 0
	numGrowingColumns := 0
	for i, column := range shownColumns {
		if column.Grow != nil && *column.Grow {
			numGrowingColumns += 1
			continue
		}

		if column.Width != nil {
			renderedColumns[i] = m.ctx.Styles.Table.TitleCellStyle.Copy().
				Width(*column.Width).
				MaxWidth(*column.Width).
				Render(column.Title)
			takenWidth += *column.Width
			continue
		}

		cell := m.ctx.Styles.Table.TitleCellStyle.Copy().Render(column.Title)
		renderedColumns[i] = cell
		takenWidth += lipgloss.Width(cell)
	}

	if numGrowingColumns == 0 {
		return renderedColumns
	}

	leftoverWidth := m.dimensions.Width - takenWidth
	growCellWidth := leftoverWidth / numGrowingColumns
	for i, column := range shownColumns {
		if column.Grow == nil || !*column.Grow {
			continue
		}

		renderedColumns[i] = m.ctx.Styles.Table.TitleCellStyle.Copy().
			Width(growCellWidth).
			MaxWidth(growCellWidth).
			Render(column.Title)
	}

	return renderedColumns
}

func (m *Model) renderHeader() string {
	headerColumns := m.renderHeaderColumns()
	header := lipgloss.JoinHorizontal(lipgloss.Top, headerColumns...)
	return m.ctx.Styles.Table.HeaderStyle.Copy().
		Width(m.dimensions.Width).
		MaxWidth(m.dimensions.Width).
		Height(common.TableHeaderHeight).
		MaxHeight(common.TableHeaderHeight).
		Render(header)
}

func (m *Model) renderBody() string {
	bodyStyle := lipgloss.NewStyle().
		Height(m.dimensions.Height).
		MaxWidth(m.dimensions.Width)

	if len(m.Rows) == 0 && m.EmptyState != nil {
		return bodyStyle.Render(*m.EmptyState)
	}

	return m.rowsViewport.View()
}

func (m *Model) renderRow(rowId int, headerColumns []string) string {
	var style lipgloss.Style

	if m.rowsViewport.GetCurrItem() == rowId {
		style = m.ctx.Styles.Table.SelectedCellStyle
	} else {
		style = m.ctx.Styles.Table.CellStyle
	}

	renderedColumns := make([]string, 0, len(m.Columns))

	headerColId := 0

	for i, column := range m.Columns {
		if column.Hidden != nil && *column.Hidden {
			continue
		}

		colWidth := lipgloss.Width(headerColumns[headerColId])
		renderedCol := style.Copy().
			Width(colWidth).
			MaxWidth(colWidth).
			Height(1).
			MaxHeight(1).
			Render(m.Rows[rowId][i])
		renderedColumns = append(renderedColumns, renderedCol)
		headerColId++
	}

	return m.ctx.Styles.Table.RowStyle.Copy().
		BorderBottom(m.ctx.Config.Theme.Ui.Table.ShowSeparator).
		MaxWidth(m.dimensions.Width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedColumns...))

}

func (m *Model) UpdateProgramContext(ctx *context.ProgramContext) {
	m.ctx = *ctx
	m.rowsViewport.UpdateProgramContext(ctx)
}

func (m *Model) LastUpdated() time.Time {
	return m.rowsViewport.LastUpdated
}

func (m *Model) UpdateLastUpdated(t time.Time) {
	m.rowsViewport.LastUpdated = t
}

func (m *Model) UpdateTotalItemsCount(count int) {
	m.rowsViewport.SetTotalItems(count)
}
