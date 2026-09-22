package prssection

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/table"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
	"github.com/dlvhdr/gh-dash/v4/internal/utils"
)

var compactModes = []struct {
	name    string
	compact bool
}{
	{name: "extended", compact: false},
	{name: "compact", compact: true},
}

func hideDeveloperConfigFromDiscovery(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("GH_DASH_CONFIG", "")
}

func columnsTestCtx(t *testing.T, compact bool) *context.ProgramContext {
	t.Helper()
	hideDeveloperConfigFromDiscovery(t)

	cfg, err := config.ParseConfig(config.Location{SkipGlobalConfig: true})
	require.NoError(t, err)
	cfg.Theme.Ui.Table.Compact = compact

	return &context.ProgramContext{Config: &cfg, Theme: theme.ParseTheme(&cfg)}
}

func findStackColumn(
	t *testing.T,
	cfg config.PrsSectionConfig,
	ctx *context.ProgramContext,
) table.Column {
	t.Helper()
	for _, col := range GetSectionColumns(cfg, ctx) {
		if col.Title == constants.StackIcon {
			return col
		}
	}
	t.Fatal("no stack column declared")
	return table.Column{}
}

func TestSectionColumnsMatchRowCells(t *testing.T) {
	for _, tt := range compactModes {
		t.Run(tt.name, func(t *testing.T) {
			ctx := columnsTestCtx(t, tt.compact)
			columns := GetSectionColumns(config.PrsSectionConfig{}, ctx)

			pr := prrow.PullRequest{
				Ctx:     ctx,
				Columns: columns,
				Data: &prrow.Data{
					Primary: &data.PullRequestData{
						StackEntry: data.PrStackEntry{
							Position: 2,
							Stack:    data.StackSummary{Size: 3, BaseRefName: "main"},
						},
					},
				},
			}

			assert.Len(t, pr.ToTableRow(false), len(columns))
		})
	}
}

func TestStackColumn(t *testing.T) {
	for _, tt := range compactModes {
		t.Run(tt.name+" is hidden by default", func(t *testing.T) {
			col := findStackColumn(t, config.PrsSectionConfig{}, columnsTestCtx(t, tt.compact))

			require.NotNil(t, col.Hidden)
			assert.True(t, *col.Hidden)
		})
	}

	t.Run("honors a section override", func(t *testing.T) {
		cfg := config.PrsSectionConfig{
			Layout: config.PrsLayoutConfig{
				Stack: config.ColumnConfig{Hidden: utils.BoolPtr(false), Width: utils.IntPtr(9)},
			},
		}

		col := findStackColumn(t, cfg, columnsTestCtx(t, true))

		require.NotNil(t, col.Hidden)
		assert.False(t, *col.Hidden)
		require.NotNil(t, col.Width)
		assert.Equal(t, 9, *col.Width)
	})
}
