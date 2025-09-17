package tabs

import (
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/golden"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func TestTabs(t *testing.T) {
	t.Parallel()
	t.Run("Should display empty tabs", func(t *testing.T) {
		t.Parallel()
		ctx := &context.ProgramContext{
			Config: &config.Config{
				PRSections: []config.PrsSectionConfig{
					{
						Title: "My PRs",
					},
					{
						Title: "Involved",
					},
					{
						Title: "Commented",
					},
				},
			},
			ScreenWidth:  40,
			ScreenHeight: 30,
			View:         config.PRsView,
		}
		m := NewModel(ctx)
		m.UpdateSectionsConfigs(ctx)

		golden.RequireEqual(t, []byte(m.View(ctx)))
	})

	t.Run("Should display loading tabs", func(t *testing.T) {
		t.Parallel()
		ctx := &context.ProgramContext{
			Config: &config.Config{
				PRSections: []config.PrsSectionConfig{
					{
						Title: "My PRs",
					},
					{
						Title: "Involved",
					},
					{
						Title: "Commented",
					},
				},
			},
			ScreenWidth:  80,
			ScreenHeight: 30,
			View:         config.PRsView,
		}
		m := NewModel(ctx)
		m.UpdateSectionsConfigs(ctx)
		execCmd(m, tea.Batch(m.SetAllLoading()...))

		golden.RequireEqual(t, []byte(m.View(ctx)))
	})

	t.Run("Should display tab counts", func(t *testing.T) {
		t.Parallel()
		ctx := &context.ProgramContext{
			Config: &config.Config{
				PRSections: []config.PrsSectionConfig{
					{
						Title: "My PRs",
					},
					{
						Title: "Involved",
					},
					{
						Title: "Commented",
					},
				},
			},
			ScreenWidth:  80,
			ScreenHeight: 30,
			View:         config.PRsView,
		}
		m := NewModel(ctx)
		m.UpdateSectionsConfigs(ctx)
		s := prssection.Model{
			BaseModel: section.BaseModel{},
			Prs:       []data.PullRequestData{},
		}
		s.TotalCount = 10

		// TODO: remove search section logic.
		// First section for search section
		m.UpdateSectionCounts([]section.Section{&s, &s, &s, &s})

		golden.RequireEqual(t, []byte(m.View(ctx)))
	})
}

func execCmd(m Model, cmd tea.Cmd) {
	for cmd != nil {
		msg := cmd()
		m, cmd = m.Update(msg)
	}
}

func WriteGoldenFile(str string) {
	f, err := os.OpenFile("/tmp/golden.txt", os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}
	_, err = f.Write([]byte(str))
	if err != nil {
		panic(err)
	}
}
