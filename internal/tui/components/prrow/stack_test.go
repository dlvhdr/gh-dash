package prrow

import (
	"strings"
	"testing"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

var compactModes = []struct {
	name    string
	compact bool
}{
	{name: "extended", compact: false},
	{name: "compact", compact: true},
}

func testCtx() *context.ProgramContext {
	cfg := &config.Config{Theme: &config.ThemeConfig{}}
	return &context.ProgramContext{Config: cfg, Theme: theme.ParseTheme(cfg)}
}

func stackedPr(position, size int) *PullRequest {
	return &PullRequest{
		Ctx: testCtx(),
		Data: &Data{
			Primary: &data.PullRequestData{
				StackEntry: data.PrStackEntry{
					Position: position,
					Stack:    data.StackSummary{Size: size, BaseRefName: "main"},
				},
			},
		},
	}
}

func TestRenderStack(t *testing.T) {
	tests := []struct {
		name         string
		pr           *PullRequest
		wantContains []string
		wantEmpty    bool
	}{
		{
			name:         "shows position and size",
			pr:           stackedPr(2, 3),
			wantContains: []string{"2/3", constants.StackIcon},
		},
		{
			name:         "shows a single entry stack",
			pr:           stackedPr(1, 1),
			wantContains: []string{"1/1"},
		},
		{
			name:      "is blank for an unstacked pr",
			pr:        stackedPr(0, 0),
			wantEmpty: true,
		},
		{
			name:      "is blank when there is no pr data",
			pr:        &PullRequest{Ctx: testCtx(), Data: &Data{Primary: nil}},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pr.renderStack()

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("renderStack() = %q, want empty", got)
				}
				return
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("renderStack() = %q, want it to contain %q", got, want)
				}
			}
		})
	}
}

func TestToTableRowIncludesStackCell(t *testing.T) {
	for _, tt := range compactModes {
		t.Run(tt.name, func(t *testing.T) {
			pr := stackedPr(2, 3)
			pr.Ctx.Config.Theme.Ui.Table.Compact = tt.compact

			found := false
			for _, cell := range pr.ToTableRow(false) {
				if strings.Contains(cell, "2/3") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("table row has no stack cell")
			}
		})
	}
}
