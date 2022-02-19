package table

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dlvhdr/gh-prs/ui/styles"
)

type Model struct {
	Spinner    spinner.Model
	IsLoading  bool
	Columns    []Column
	Rows       []Row
	Width      int
	EmptyState *string
}

type Column struct {
	Title string
	Width *int
	Grow  *bool
}

type Row []string

func NewModel(width int, columns []Column, rows []Row, emptyState *string) Model {
	return Model{
		Spinner:    spinner.Model{},
		IsLoading:  true,
		Columns:    columns,
		Rows:       rows,
		Width:      width,
		EmptyState: emptyState,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen)
}

func (m Model) View() string {
	headerColumns := m.renderHeader()
	header := headerStyle.Copy().
		Width(m.Width).
		MaxWidth(m.Width).
		Render(lipgloss.JoinHorizontal(lipgloss.Left, headerColumns...))

	rows := m.renderRows(headerColumns)

	return lipgloss.JoinVertical(lipgloss.Left, header, rows)
}

func getColumnWidths(renderedColumns []string) []int {
	widths := make([]int, len(renderedColumns))

	for _, column := range renderedColumns {
		widths = append(widths, lipgloss.Width(column))
	}

	return widths
}

func (m *Model) renderHeader() []string {
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

	leftoverWidth := m.Width - takenWidth
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

func (m *Model) renderRows(headerColumns []string) string {
	if len(m.Rows) == 0 && m.EmptyState != nil {
		return *m.EmptyState
	}

	renderedRows := make([]string, 0, len(m.Rows))
	for i := range m.Rows {
		renderedRows = append(renderedRows, m.renderRow(i, headerColumns))
	}

	return lipgloss.JoinVertical(lipgloss.Left, renderedRows...)

}

func (m *Model) renderRow(rowId int, headerColumns []string) string {
	renderedColumns := make([]string, len(m.Columns))
	for i, column := range m.Rows[rowId] {
		renderedColumns = append(
			renderedColumns,
			cellStyle.Width(lipgloss.Width(headerColumns[i])).Render(column),
		)
	}
	return lipgloss.JoinHorizontal(lipgloss.Left, renderedColumns...)
}

func intPtr(val int) *int {
	return &val
}

func boolPtr(val bool) *bool {
	return &val
}
