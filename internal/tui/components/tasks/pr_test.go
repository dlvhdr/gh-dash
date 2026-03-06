package tasks

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func TestApproveWorkflows_TaskConfiguration(t *testing.T) {
	var capturedTask context.Task

	ctx := &context.ProgramContext{
		StartTask: func(task context.Task) tea.Cmd {
			capturedTask = task
			return nil
		},
	}
	section := SectionIdentifier{Id: 2, Type: "pr"}
	pr := mockIssue{
		number:   42,
		repoName: "owner/repo",
	}

	_ = ApproveWorkflows(ctx, section, pr)

	require.Equal(t, "pr_approve_workflows_42", capturedTask.Id)
	require.Equal(t, "Approving workflows for PR #42", capturedTask.StartText)
	require.Equal(t, "Workflows for PR #42 have been approved", capturedTask.FinishedText)
	require.Equal(t, context.TaskStart, capturedTask.State)
	require.Nil(t, capturedTask.Error)
}

func TestApproveWorkflows_ReturnsNonNilCommand(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		repoName string
	}{
		{
			name:     "standard PR",
			prNumber: 123,
			repoName: "owner/repo",
		},
		{
			name:     "large PR number",
			prNumber: 99999,
			repoName: "my-org/my-project",
		},
		{
			name:     "hyphenated repo name",
			prNumber: 1,
			repoName: "some-owner/some-repo-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProgramContext{
				StartTask: noopStartTask,
			}
			section := SectionIdentifier{Id: 1, Type: "pr"}
			pr := mockIssue{
				number:   tt.prNumber,
				repoName: tt.repoName,
			}

			cmd := ApproveWorkflows(ctx, section, pr)

			require.NotNil(t, cmd, "ApproveWorkflows should return a non-nil command")
		})
	}
}

func TestApproveWorkflows_UsesCorrectPRNumber(t *testing.T) {
	prNumbers := []int{1, 100, 12345, 999999}

	for _, num := range prNumbers {
		t.Run(fmt.Sprintf("pr_%d", num), func(t *testing.T) {
			var capturedTask context.Task
			ctx := &context.ProgramContext{
				StartTask: func(task context.Task) tea.Cmd {
					capturedTask = task
					return nil
				},
			}
			pr := mockIssue{number: num, repoName: "o/r"}

			ApproveWorkflows(ctx, SectionIdentifier{}, pr)

			expectedId := fmt.Sprintf("pr_approve_workflows_%d", num)
			require.Equal(t, expectedId, capturedTask.Id)
			require.Contains(t, capturedTask.StartText, fmt.Sprintf("#%d", num))
			require.Contains(t, capturedTask.FinishedText, fmt.Sprintf("#%d", num))
		})
	}
}

func TestApproveWorkflows_SectionIdentifierPropagation(t *testing.T) {
	tests := []struct {
		name        string
		sectionId   int
		sectionType string
	}{
		{
			name:        "pr section type",
			sectionId:   1,
			sectionType: "pr",
		},
		{
			name:        "notification section type",
			sectionId:   10,
			sectionType: "notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProgramContext{
				StartTask: noopStartTask,
			}
			section := SectionIdentifier{Id: tt.sectionId, Type: tt.sectionType}
			pr := mockIssue{number: 1, repoName: "o/r"}

			cmd := ApproveWorkflows(ctx, section, pr)

			require.NotNil(t, cmd)
		})
	}
}
