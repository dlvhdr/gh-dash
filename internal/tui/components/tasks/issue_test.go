package tasks

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

// mockIssue implements data.RowData for testing
type mockIssue struct {
	number    int
	repoName  string
	title     string
	url       string
	updatedAt time.Time
}

func (m mockIssue) GetNumber() int               { return m.number }
func (m mockIssue) GetRepoNameWithOwner() string { return m.repoName }
func (m mockIssue) GetTitle() string             { return m.title }
func (m mockIssue) GetUrl() string               { return m.url }
func (m mockIssue) GetUpdatedAt() time.Time      { return m.updatedAt }

// noopStartTask is a stub that returns nil for testing
func noopStartTask(task context.Task) tea.Cmd {
	return nil
}

func TestUpdateIssueMsg_Fields(t *testing.T) {
	t.Run("all fields can be set", func(t *testing.T) {
		isClosed := true
		labels := &data.IssueLabels{Nodes: []data.Label{{Name: "bug", Color: "ff0000"}}}
		comment := &data.IssueComment{Body: "test comment"}
		addedAssignees := &data.Assignees{Nodes: []data.Assignee{{Login: "user1"}}}
		removedAssignees := &data.Assignees{Nodes: []data.Assignee{{Login: "user2"}}}

		msg := UpdateIssueMsg{
			IssueNumber:      123,
			Labels:           labels,
			NewComment:       comment,
			IsClosed:         &isClosed,
			AddedAssignees:   addedAssignees,
			RemovedAssignees: removedAssignees,
		}

		require.Equal(t, 123, msg.IssueNumber)
		require.Equal(t, labels, msg.Labels)
		require.Equal(t, comment, msg.NewComment)
		require.NotNil(t, msg.IsClosed)
		require.True(t, *msg.IsClosed)
		require.Equal(t, addedAssignees, msg.AddedAssignees)
		require.Equal(t, removedAssignees, msg.RemovedAssignees)
	})

	t.Run("nil pointer fields are valid", func(t *testing.T) {
		msg := UpdateIssueMsg{
			IssueNumber: 456,
		}

		require.Equal(t, 456, msg.IssueNumber)
		require.Nil(t, msg.Labels)
		require.Nil(t, msg.NewComment)
		require.Nil(t, msg.IsClosed)
		require.Nil(t, msg.AddedAssignees)
		require.Nil(t, msg.RemovedAssignees)
	})
}

func TestCloseIssue(t *testing.T) {
	tests := []struct {
		name        string
		issueNumber int
		repoName    string
	}{
		{
			name:        "closes issue with standard number",
			issueNumber: 123,
			repoName:    "owner/repo",
		},
		{
			name:        "closes issue with large number",
			issueNumber: 99999,
			repoName:    "my-org/my-project",
		},
		{
			name:        "closes issue with hyphenated repo name",
			issueNumber: 1,
			repoName:    "some-owner/some-repo-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProgramContext{
				StartTask: noopStartTask,
			}
			section := SectionIdentifier{Id: 1, Type: "issue"}
			issue := mockIssue{
				number:   tt.issueNumber,
				repoName: tt.repoName,
			}

			cmd := CloseIssue(ctx, section, issue)

			require.NotNil(t, cmd, "CloseIssue should return a non-nil command")
		})
	}
}

func TestCloseIssue_TaskConfiguration(t *testing.T) {
	var capturedTask context.Task

	ctx := &context.ProgramContext{
		StartTask: func(task context.Task) tea.Cmd {
			capturedTask = task
			return nil
		},
	}
	section := SectionIdentifier{Id: 5, Type: "issue"}
	issue := mockIssue{
		number:   42,
		repoName: "test/repo",
	}

	_ = CloseIssue(ctx, section, issue)

	require.Equal(t, "issue_close_42", capturedTask.Id)
	require.Equal(t, "Closing issue #42", capturedTask.StartText)
	require.Equal(t, "Issue #42 has been closed", capturedTask.FinishedText)
	require.Equal(t, context.TaskStart, capturedTask.State)
	require.Nil(t, capturedTask.Error)
}

func TestReopenIssue(t *testing.T) {
	tests := []struct {
		name        string
		issueNumber int
		repoName    string
	}{
		{
			name:        "reopens issue with standard number",
			issueNumber: 123,
			repoName:    "owner/repo",
		},
		{
			name:        "reopens issue with large number",
			issueNumber: 99999,
			repoName:    "my-org/my-project",
		},
		{
			name:        "reopens issue with hyphenated repo name",
			issueNumber: 1,
			repoName:    "some-owner/some-repo-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProgramContext{
				StartTask: noopStartTask,
			}
			section := SectionIdentifier{Id: 1, Type: "issue"}
			issue := mockIssue{
				number:   tt.issueNumber,
				repoName: tt.repoName,
			}

			cmd := ReopenIssue(ctx, section, issue)

			require.NotNil(t, cmd, "ReopenIssue should return a non-nil command")
		})
	}
}

func TestReopenIssue_TaskConfiguration(t *testing.T) {
	var capturedTask context.Task

	ctx := &context.ProgramContext{
		StartTask: func(task context.Task) tea.Cmd {
			capturedTask = task
			return nil
		},
	}
	section := SectionIdentifier{Id: 3, Type: "issue"}
	issue := mockIssue{
		number:   99,
		repoName: "example/project",
	}

	_ = ReopenIssue(ctx, section, issue)

	require.Equal(t, "issue_reopen_99", capturedTask.Id)
	require.Equal(t, "Reopening issue #99", capturedTask.StartText)
	require.Equal(t, "Issue #99 has been reopened", capturedTask.FinishedText)
	require.Equal(t, context.TaskStart, capturedTask.State)
	require.Nil(t, capturedTask.Error)
}

func TestCloseIssue_SectionIdentifierPropagation(t *testing.T) {
	tests := []struct {
		name        string
		sectionId   int
		sectionType string
	}{
		{
			name:        "issue section type",
			sectionId:   1,
			sectionType: "issue",
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
			issue := mockIssue{number: 1, repoName: "o/r"}

			cmd := CloseIssue(ctx, section, issue)

			require.NotNil(t, cmd)
		})
	}
}

func TestReopenIssue_SectionIdentifierPropagation(t *testing.T) {
	tests := []struct {
		name        string
		sectionId   int
		sectionType string
	}{
		{
			name:        "issue section type",
			sectionId:   1,
			sectionType: "issue",
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
			issue := mockIssue{number: 1, repoName: "o/r"}

			cmd := ReopenIssue(ctx, section, issue)

			require.NotNil(t, cmd)
		})
	}
}

func TestUpdateIssueMsg_ImplementsTeaMsg(t *testing.T) {
	// Verify UpdateIssueMsg can be used as a tea.Msg
	var msg tea.Msg = UpdateIssueMsg{IssueNumber: 1}

	_, ok := msg.(UpdateIssueMsg)
	require.True(t, ok, "UpdateIssueMsg should be usable as tea.Msg")
}

func TestCloseIssue_UsesCorrectIssueNumber(t *testing.T) {
	issueNumbers := []int{1, 100, 12345, 999999}

	for _, num := range issueNumbers {
		t.Run(fmt.Sprintf("issue_%d", num), func(t *testing.T) {
			var capturedTask context.Task
			ctx := &context.ProgramContext{
				StartTask: func(task context.Task) tea.Cmd {
					capturedTask = task
					return nil
				},
			}
			issue := mockIssue{number: num, repoName: "o/r"}

			CloseIssue(ctx, SectionIdentifier{}, issue)

			expectedId := fmt.Sprintf("issue_close_%d", num)
			require.Equal(t, expectedId, capturedTask.Id)
			require.Contains(t, capturedTask.StartText, fmt.Sprintf("#%d", num))
		})
	}
}

func TestReopenIssue_UsesCorrectIssueNumber(t *testing.T) {
	issueNumbers := []int{1, 100, 12345, 999999}

	for _, num := range issueNumbers {
		t.Run(fmt.Sprintf("issue_%d", num), func(t *testing.T) {
			var capturedTask context.Task
			ctx := &context.ProgramContext{
				StartTask: func(task context.Task) tea.Cmd {
					capturedTask = task
					return nil
				},
			}
			issue := mockIssue{number: num, repoName: "o/r"}

			ReopenIssue(ctx, SectionIdentifier{}, issue)

			expectedId := fmt.Sprintf("issue_reopen_%d", num)
			require.Equal(t, expectedId, capturedTask.Id)
			require.Contains(t, capturedTask.StartText, fmt.Sprintf("#%d", num))
		})
	}
}

func TestCloseIssue_MsgCallbackReturnsCorrectUpdateIssueMsg(t *testing.T) {
	issueNumber := 42

	// Create a GitHubTask directly to test the Msg callback
	task := GitHubTask{
		Id: fmt.Sprintf("issue_close_%d", issueNumber),
		Args: []string{
			"issue",
			"close",
			fmt.Sprint(issueNumber),
			"-R",
			"owner/repo",
		},
		Section:      SectionIdentifier{Id: 1, Type: "issue"},
		StartText:    fmt.Sprintf("Closing issue #%d", issueNumber),
		FinishedText: fmt.Sprintf("Issue #%d has been closed", issueNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdateIssueMsg{
				IssueNumber: issueNumber,
				IsClosed:    boolPtr(true),
			}
		},
	}

	msg := task.Msg(nil, nil)
	updateMsg, ok := msg.(UpdateIssueMsg)

	require.True(t, ok, "Msg should return UpdateIssueMsg")
	require.Equal(t, issueNumber, updateMsg.IssueNumber)
	require.NotNil(t, updateMsg.IsClosed)
	require.True(t, *updateMsg.IsClosed, "IsClosed should be true for close action")
}

func TestReopenIssue_MsgCallbackReturnsCorrectUpdateIssueMsg(t *testing.T) {
	issueNumber := 42

	// Create a GitHubTask directly to test the Msg callback
	task := GitHubTask{
		Id: fmt.Sprintf("issue_reopen_%d", issueNumber),
		Args: []string{
			"issue",
			"reopen",
			fmt.Sprint(issueNumber),
			"-R",
			"owner/repo",
		},
		Section:      SectionIdentifier{Id: 1, Type: "issue"},
		StartText:    fmt.Sprintf("Reopening issue #%d", issueNumber),
		FinishedText: fmt.Sprintf("Issue #%d has been reopened", issueNumber),
		Msg: func(c *exec.Cmd, err error) tea.Msg {
			return UpdateIssueMsg{
				IssueNumber: issueNumber,
				IsClosed:    boolPtr(false),
			}
		},
	}

	msg := task.Msg(nil, nil)
	updateMsg, ok := msg.(UpdateIssueMsg)

	require.True(t, ok, "Msg should return UpdateIssueMsg")
	require.Equal(t, issueNumber, updateMsg.IssueNumber)
	require.NotNil(t, updateMsg.IsClosed)
	require.False(t, *updateMsg.IsClosed, "IsClosed should be false for reopen action")
}

func TestCloseIssue_CommandArgs(t *testing.T) {
	tests := []struct {
		name         string
		issueNumber  int
		repoName     string
		expectedArgs []string
	}{
		{
			name:        "standard repo",
			issueNumber: 123,
			repoName:    "owner/repo",
			expectedArgs: []string{
				"issue", "close", "123", "-R", "owner/repo",
			},
		},
		{
			name:        "org repo with hyphens",
			issueNumber: 456,
			repoName:    "my-org/my-repo-name",
			expectedArgs: []string{
				"issue", "close", "456", "-R", "my-org/my-repo-name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := GitHubTask{
				Args: []string{
					"issue",
					"close",
					fmt.Sprint(tt.issueNumber),
					"-R",
					tt.repoName,
				},
			}

			require.Equal(t, tt.expectedArgs, task.Args)
		})
	}
}

func TestReopenIssue_CommandArgs(t *testing.T) {
	tests := []struct {
		name         string
		issueNumber  int
		repoName     string
		expectedArgs []string
	}{
		{
			name:        "standard repo",
			issueNumber: 123,
			repoName:    "owner/repo",
			expectedArgs: []string{
				"issue", "reopen", "123", "-R", "owner/repo",
			},
		},
		{
			name:        "org repo with hyphens",
			issueNumber: 456,
			repoName:    "my-org/my-repo-name",
			expectedArgs: []string{
				"issue", "reopen", "456", "-R", "my-org/my-repo-name",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := GitHubTask{
				Args: []string{
					"issue",
					"reopen",
					fmt.Sprint(tt.issueNumber),
					"-R",
					tt.repoName,
				},
			}

			require.Equal(t, tt.expectedArgs, task.Args)
		})
	}
}

// boolPtr is a helper to create a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}
