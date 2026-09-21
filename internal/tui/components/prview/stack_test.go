package prview

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func newTestModelForStack(t *testing.T) Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config: &cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
	}

	m := NewModel(ctx)
	m.ctx = ctx
	m.SetWidth(80)
	return m
}

func stackedRow(position, size int) *prrow.Data {
	nodes := make([]data.StackEntryNode, 0, size)
	for i := 1; i <= size; i++ {
		nodes = append(nodes, data.StackEntryNode{
			Position: i,
			PullRequest: data.StackPullRequest{
				Number:      100 + i,
				Title:       "stacked pr",
				State:       "OPEN",
				HeadRefName: "feat",
			},
		})
	}

	enriched := data.EnrichedPullRequestData{Number: 100 + position}
	enriched.StackEntry.Position = position
	enriched.StackEntry.Stack.Size = size
	enriched.StackEntry.Stack.BaseRefName = "main"
	enriched.StackEntry.Stack.Entries.Nodes = nodes

	primary := enriched.ToPullRequestData()

	return &prrow.Data{Primary: &primary, Enriched: enriched, IsEnriched: true}
}

func unstackedRow() *prrow.Data {
	return &prrow.Data{Primary: &data.PullRequestData{Number: 1}, IsEnriched: true}
}

func TestStackTab(t *testing.T) {
	t.Run("is offered only for stacked prs", func(t *testing.T) {
		m := newTestModelForStack(t)

		m.SetRow(unstackedRow())
		assert.NotContains(t, m.carousel.Items(), stackTab)

		m.SetRow(stackedRow(2, 3))
		assert.Contains(t, m.carousel.Items(), stackTab)

		m.SetRow(unstackedRow())
		assert.NotContains(t, m.carousel.Items(), stackTab)
	})

	t.Run("is offered before the enriched data arrives", func(t *testing.T) {
		m := newTestModelForStack(t)
		row := stackedRow(2, 3)
		row.IsEnriched = false
		row.Enriched = data.EnrichedPullRequestData{}

		m.SetRow(row)

		assert.Contains(t, m.carousel.Items(), stackTab)
	})

	t.Run("leaves the cursor valid when it disappears", func(t *testing.T) {
		m := newTestModelForStack(t)
		m.SetRow(stackedRow(1, 2))
		m.carousel.SetCursor(len(m.carousel.Items()) - 1)
		require.Equal(t, stackTab, m.carousel.SelectedItem())

		m.SetRow(unstackedRow())

		assert.Less(t, m.carousel.Cursor(), len(m.carousel.Items()))
		assert.NotEqual(t, stackTab, m.carousel.SelectedItem())
	})
}

func TestStackTabVisibility(t *testing.T) {
	tests := []struct {
		name        string
		width       int
		wantVisible bool
	}{
		{name: "narrow sidebar clips the tab", width: 60, wantVisible: false},
		{name: "wide sidebar shows the tab", width: 100, wantVisible: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestModelForStack(t)
			m.SetWidth(tt.width)
			m.SetRow(stackedRow(2, 3))

			require.Contains(t, m.carousel.Items(), stackTab)
			assert.Equal(t, tt.wantVisible,
				strings.Contains(ansi.Strip(m.carousel.View()), ansi.Strip(stackTab)))
		})
	}
}

func TestStackTabReachableWhenClipped(t *testing.T) {
	m := newTestModelForStack(t)
	m.SetWidth(60)
	m.SetRow(stackedRow(2, 3))

	for range len(m.carousel.Items()) - 1 {
		m.carousel.MoveRight()
	}

	assert.Equal(t, stackTab, m.carousel.SelectedItem())
	assert.Contains(t, ansi.Strip(m.renderStack()), "#101")
}

func TestClippedTabsShowAnOverflowIndicator(t *testing.T) {
	m := newTestModelForStack(t)
	m.SetWidth(60)
	m.SetRow(stackedRow(2, 3))

	assert.Contains(t, ansi.Strip(m.carousel.View()), "→")

	for range len(m.carousel.Items()) - 1 {
		m.carousel.MoveRight()
	}

	assert.NotContains(t, ansi.Strip(m.carousel.View()), "→")
}

func TestWideSidebarNeedsNoOverflowIndicator(t *testing.T) {
	m := newTestModelForStack(t)
	m.SetWidth(100)
	m.SetRow(stackedRow(2, 3))

	assert.NotContains(t, ansi.Strip(m.carousel.View()), "→")
}

func TestRenderStack(t *testing.T) {
	t.Run("lists every entry and the base branch", func(t *testing.T) {
		m := newTestModelForStack(t)
		m.SetRow(stackedRow(2, 3))

		got := ansi.Strip(m.renderStack())

		for _, want := range []string{"#101", "#102", "#103", "main", "Stack of 3"} {
			assert.Contains(t, got, want)
		}
	})

	t.Run("orders entries top first", func(t *testing.T) {
		m := newTestModelForStack(t)
		m.SetRow(stackedRow(2, 3))

		got := ansi.Strip(m.renderStack())

		assert.Less(t, strings.Index(got, "#103"), strings.Index(got, "#101"))
	})

	t.Run("flags a stack truncated by the entries page size", func(t *testing.T) {
		m := newTestModelForStack(t)
		row := stackedRow(2, 3)
		row.Enriched.StackEntry.Stack.Size = 60

		m.SetRow(row)

		got := ansi.Strip(m.renderStack())
		assert.Contains(t, got, "Stack of 60")
		assert.Contains(t, got, "showing 3")
	})

	t.Run("does not flag a complete stack", func(t *testing.T) {
		m := newTestModelForStack(t)
		m.SetRow(stackedRow(2, 3))

		assert.NotContains(t, ansi.Strip(m.renderStack()), "showing")
	})

	t.Run("renders when the row carries no primary data", func(t *testing.T) {
		m := newTestModelForStack(t)
		row := stackedRow(2, 3)
		row.Primary = nil

		m.SetRow(row)

		assert.NotPanics(t, func() { m.renderStack() })
	})

	t.Run("says so when the pr is not stacked", func(t *testing.T) {
		m := newTestModelForStack(t)
		m.SetRow(unstackedRow())

		assert.Contains(t, ansi.Strip(m.renderStack()), "isn't part of a stack")
	})

	t.Run("waits for the enriched data", func(t *testing.T) {
		m := newTestModelForStack(t)
		row := stackedRow(2, 3)
		row.IsEnriched = false

		m.SetRow(row)

		assert.Contains(t, ansi.Strip(m.renderStack()), "Loading")
	})
}
