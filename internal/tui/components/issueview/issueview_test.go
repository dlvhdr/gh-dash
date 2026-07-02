package issueview

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newTestContext(t *testing.T) *context.ProgramContext {
	t.Helper()

	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	thm := theme.ParseTheme(&cfg)
	return &context.ProgramContext{
		Config:            &cfg,
		Theme:             thm,
		Styles:            context.InitStyles(thm),
		HasDarkBackground: true,
		BackgroundSource:  "default",
	}
}

func TestNewModelSetsProgramContext(t *testing.T) {
	ctx := newTestContext(t)
	m := NewModel(ctx)

	require.NotNil(t, m.ctx)
	require.Same(t, ctx, m.ctx)
}

func TestRenderBodyDoesNotPanicBeforeContextSync(t *testing.T) {
	ctx := newTestContext(t)
	m := NewModel(ctx)
	m.SetWidth(80)
	m.SetRow(&data.IssueData{
		Title: "Example issue",
		Body:  "Hello **world**",
	})

	require.NotPanics(t, func() {
		_ = m.renderBody()
	})
}
