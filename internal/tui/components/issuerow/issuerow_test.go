package issuerow

import (
	"strings"
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

func TestIssue_RenderTitle(t *testing.T) {
	ctx := newTestContext(t)

	t.Run("renders standard title for top-level issue", func(t *testing.T) {
		issue := Issue{
			Ctx: ctx,
			Data: data.IssueData{
				Number: 1,
				Title:  "Standard issue title",
				State:  "OPEN",
			},
		}
		rendered := issue.renderTitle()
		require.True(t, strings.Contains(rendered, "Standard issue title"))
		require.False(t, strings.Contains(rendered, "↳"))
	})

	t.Run("renders title with child indicator for sub-issue", func(t *testing.T) {
		issue := Issue{
			Ctx: ctx,
			Data: data.IssueData{
				Number: 2,
				Title:  "Sub issue title",
				State:  "OPEN",
				Parent: &data.ParentIssue{
					Number: 1,
					Title:  "Parent issue",
				},
			},
		}
		rendered := issue.renderTitle()
		require.True(t, strings.Contains(rendered, "↳"))
		require.True(t, strings.Contains(rendered, "Sub issue title"))
	})

	t.Run("renders title with sub-issues progress indicator for parent issue", func(t *testing.T) {
		issue := Issue{
			Ctx: ctx,
			Data: data.IssueData{
				Number: 3,
				Title:  "Parent issue title",
				State:  "OPEN",
				SubIssuesSummary: data.SubIssuesSummary{
					Total:     4,
					Completed: 2,
				},
			},
		}
		rendered := issue.renderTitle()
		require.True(t, strings.Contains(rendered, "Parent issue title"))
		require.True(t, strings.Contains(rendered, "(2/4)"))
	})

	t.Run(
		"renders both child indicator and progress indicator for intermediate issue",
		func(t *testing.T) {
			issue := Issue{
				Ctx: ctx,
				Data: data.IssueData{
					Number: 4,
					Title:  "Intermediate issue",
					State:  "OPEN",
					Parent: &data.ParentIssue{
						Number: 1,
					},
					SubIssuesSummary: data.SubIssuesSummary{
						Total:     2,
						Completed: 1,
					},
				},
			}
			rendered := issue.renderTitle()
			require.True(t, strings.Contains(rendered, "↳"))
			require.True(t, strings.Contains(rendered, "Intermediate issue"))
			require.True(t, strings.Contains(rendered, "(1/2)"))
		},
	)
}
