package testdata

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type TestSection struct {
	Config  config.SectionConfig
	loading bool
}

// BuildRows implements section.Section.
func (t *TestSection) BuildRows() []table.Row {
	panic("unimplemented")
}

// CurrRow implements section.Section.
func (t *TestSection) CurrRow() int {
	panic("unimplemented")
}

// FetchNextPageSectionRows implements section.Section.
func (t *TestSection) FetchNextPageSectionRows() []tea.Cmd {
	panic("unimplemented")
}

// FirstItem implements section.Section.
func (t *TestSection) FirstItem() int {
	panic("unimplemented")
}

// GetConfig implements section.Section.
func (t *TestSection) GetConfig() config.SectionConfig {
	return t.Config
}

// GetCurrRow implements section.Section.
func (t *TestSection) GetCurrRow() data.RowData {
	panic("unimplemented")
}

// GetFilters implements section.Section.
func (t *TestSection) GetFilters() string {
	panic("unimplemented")
}

// GetId implements section.Section.
func (t *TestSection) GetId() int {
	panic("unimplemented")
}

// GetIsLoading implements section.Section.
func (t *TestSection) GetIsLoading() bool {
	return t.loading
}

// GetItemPluralForm implements section.Section.
func (t *TestSection) GetItemPluralForm() string {
	panic("unimplemented")
}

// GetItemSingularForm implements section.Section.
func (t *TestSection) GetItemSingularForm() string {
	panic("unimplemented")
}

// GetPagerContent implements section.Section.
func (t *TestSection) GetPagerContent() string {
	panic("unimplemented")
}

// GetPromptConfirmation implements section.Section.
func (t *TestSection) GetPromptConfirmation() string {
	panic("unimplemented")
}

// GetPromptConfirmationAction implements section.Section.
func (t *TestSection) GetPromptConfirmationAction() string {
	panic("unimplemented")
}

// GetTotalCount implements section.Section.
func (t *TestSection) GetTotalCount() int {
	return 10
}

// GetType implements section.Section.
func (t *TestSection) GetType() string {
	panic("unimplemented")
}

// IsPromptConfirmationFocused implements section.Section.
func (t *TestSection) IsPromptConfirmationFocused() bool {
	panic("unimplemented")
}

// IsSearchFocused implements section.Section.
func (t *TestSection) IsSearchFocused() bool {
	panic("unimplemented")
}

// LastItem implements section.Section.
func (t *TestSection) LastItem() int {
	panic("unimplemented")
}

// MakeSectionCmd implements section.Section.
func (t *TestSection) MakeSectionCmd(cmd tea.Cmd) tea.Cmd {
	panic("unimplemented")
}

// NextRow implements section.Section.
func (t *TestSection) NextRow() int {
	panic("unimplemented")
}

// NumRows implements section.Section.
func (t *TestSection) NumRows() int {
	panic("unimplemented")
}

// PrevRow implements section.Section.
func (t *TestSection) PrevRow() int {
	panic("unimplemented")
}

// ResetFilters implements section.Section.
func (t *TestSection) ResetFilters() {
	panic("unimplemented")
}

// ResetPageInfo implements section.Section.
func (t *TestSection) ResetPageInfo() {
	panic("unimplemented")
}

// ResetRows implements section.Section.
func (t *TestSection) ResetRows() {
	panic("unimplemented")
}

// SetIsLoading implements section.Section.
func (t *TestSection) SetIsLoading(val bool) {
	t.loading = val
}

// SetIsPromptConfirmationShown implements section.Section.
func (t *TestSection) SetIsPromptConfirmationShown(val bool) tea.Cmd {
	panic("unimplemented")
}

// SetIsSearching implements section.Section.
func (t *TestSection) SetIsSearching(val bool) tea.Cmd {
	panic("unimplemented")
}

// SetPromptConfirmationAction implements section.Section.
func (t *TestSection) SetPromptConfirmationAction(action string) {
	panic("unimplemented")
}

// Update implements section.Section.
func (t *TestSection) Update(msg tea.Msg) (section.Section, tea.Cmd) {
	panic("unimplemented")
}

// UpdateProgramContext implements section.Section.
func (t *TestSection) UpdateProgramContext(ctx *context.ProgramContext) {
	panic("unimplemented")
}

// View implements section.Section.
func (t *TestSection) View() string {
	panic("unimplemented")
}
