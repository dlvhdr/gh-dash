package tasks

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
	graphqltest "github.com/dlvhdr/gh-dash/v4/internal/testhelpers/graphql"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
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

// --- extractPullRequestData tests ---

// unknownRowData is a RowData implementation that is neither *data.PullRequestData
// nor a primaryPRDataProvider — used to exercise the nil fallback branch.
type unknownRowData struct{}

func (u unknownRowData) GetRepoNameWithOwner() string { return "owner/repo" }
func (u unknownRowData) GetTitle() string             { return "title" }
func (u unknownRowData) GetNumber() int               { return 1 }
func (u unknownRowData) GetUrl() string               { return "https://github.com/owner/repo/pull/1" }
func (u unknownRowData) GetUpdatedAt() time.Time      { return time.Time{} }

func TestExtractPullRequestData_DirectPRData(t *testing.T) {
	prData := &data.PullRequestData{ID: "PR_xyz"}
	result := extractPullRequestData(prData)
	assert.Same(t, prData, result, "should return the same *data.PullRequestData pointer")
}

func TestExtractPullRequestData_PrrowData(t *testing.T) {
	prData := &data.PullRequestData{ID: "PR_abc"}
	row := &prrow.Data{Primary: prData}
	result := extractPullRequestData(row)
	assert.Same(t, prData, result, "should return the Primary field from prrow.Data")
}

func TestExtractPullRequestData_PrrowData_NilPrimary(t *testing.T) {
	row := &prrow.Data{Primary: nil}
	result := extractPullRequestData(row)
	assert.Nil(t, result)
}

func TestExtractPullRequestData_NilInput(t *testing.T) {
	// A bare nil becomes a nil interface value; the type switch has no
	// matching case and falls through to return nil.
	result := extractPullRequestData(nil)
	assert.Nil(t, result)
}

func TestExtractPullRequestData_UnknownRowDataType(t *testing.T) {
	result := extractPullRequestData(unknownRowData{})
	assert.Nil(t, result, "unknown RowData type with no GetPrimaryPRData() method should return nil")
}

// --- MergePR tests ---

// setMockClient installs a mock GraphQL client backed by handler and resets it after the test.
func setMockClient(t *testing.T, handler http.Handler) {
	t.Helper()
	data.SetClient(graphqltest.NewMockGraphQLClient(t, handler))
	t.Cleanup(func() { data.SetClient(nil) })
}

// mergeMockHandler serves controlled responses keyed by GraphQL operation name.
func mergeMockHandler(t *testing.T, responses map[string]string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		bodyStr := string(body)

		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(bodyStr, "mutation MergePullRequest"):
			_, _ = io.WriteString(w, responses["MergePullRequest"])
		case strings.Contains(bodyStr, "mutation EnablePullRequestAutoMerge"):
			_, _ = io.WriteString(w, responses["EnablePullRequestAutoMerge"])
		default:
			t.Errorf("unexpected GraphQL request body: %s", bodyStr)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

// noopCtx returns a minimal ProgramContext whose StartTask is a no-op.
// tea.Batch(nil, workFn) returns workFn directly, so calling the result of
// MergePR() will invoke the work function synchronously in tests.
func noopCtx() *context.ProgramContext {
	return &context.ProgramContext{
		StartTask: func(task context.Task) tea.Cmd { return nil },
	}
}

// prDataFor builds a *data.PullRequestData with the fields MergePR needs.
func prDataFor(id string, status data.MergeStateStatus, repo data.Repository) *data.PullRequestData {
	return &data.PullRequestData{
		ID:               id,
		Number:           42,
		MergeStateStatus: status,
		Repository:       repo,
	}
}

// runMergePR calls MergePR and immediately invokes the returned tea.Cmd to
// obtain the TaskFinishedMsg synchronously (no bubbletea runtime required).
func runMergePR(t *testing.T, pr data.RowData) constants.TaskFinishedMsg {
	t.Helper()
	ctx := noopCtx()
	section := SectionIdentifier{Id: 1, Type: "prs"}
	cmd := MergePR(ctx, section, pr)
	require.NotNil(t, cmd, "MergePR should return a non-nil tea.Cmd")
	msg := cmd()
	finished, ok := msg.(constants.TaskFinishedMsg)
	require.True(t, ok, "expected TaskFinishedMsg, got %T", msg)
	return finished
}

func TestMergePR_DirectMerge(t *testing.T) {
	setMockClient(t, mergeMockHandler(t, map[string]string{
		"MergePullRequest": `{"data":{"mergePullRequest":{"pullRequest":{"state":"MERGED"}}}}`,
	}))

	pr := prDataFor("PR_clean", "CLEAN", data.Repository{AllowMergeCommit: true})
	msg := runMergePR(t, pr)

	require.NoError(t, msg.Err)
	assert.Equal(t, "PR #42 has been merged", msg.FinishedText)
	update, ok := msg.Msg.(UpdatePRMsg)
	require.True(t, ok)
	require.NotNil(t, update.IsMerged)
	assert.True(t, *update.IsMerged)
	assert.Nil(t, update.AutoMergeEnabled)
}

func TestMergePR_Blocked_EnablesAutoMerge(t *testing.T) {
	setMockClient(t, mergeMockHandler(t, map[string]string{
		"EnablePullRequestAutoMerge": `{"data":{"enablePullRequestAutoMerge":{"pullRequest":{"state":"OPEN","autoMergeRequest":{"enabledAt":"2024-01-01T00:00:00Z"}}}}}`,
	}))

	pr := prDataFor("PR_blocked", "BLOCKED", data.Repository{AllowMergeCommit: true})
	msg := runMergePR(t, pr)

	require.NoError(t, msg.Err)
	assert.Equal(t, "Auto-merge enabled for PR #42", msg.FinishedText)
	update, ok := msg.Msg.(UpdatePRMsg)
	require.True(t, ok)
	require.NotNil(t, update.AutoMergeEnabled)
	assert.True(t, *update.AutoMergeEnabled)
	assert.Nil(t, update.IsMerged)
}

func TestMergePR_AutoMergeEnabled(t *testing.T) {
	setMockClient(t, mergeMockHandler(t, map[string]string{
		"EnablePullRequestAutoMerge": `{"data":{"enablePullRequestAutoMerge":{"pullRequest":{"state":"OPEN","autoMergeRequest":{"enabledAt":"2024-01-01T00:00:00Z"}}}}}`,
	}))

	pr := prDataFor("PR_behind", "BEHIND", data.Repository{AllowMergeCommit: true})
	msg := runMergePR(t, pr)

	require.NoError(t, msg.Err)
	assert.Equal(t, "Auto-merge enabled for PR #42", msg.FinishedText)
	update, ok := msg.Msg.(UpdatePRMsg)
	require.True(t, ok)
	require.NotNil(t, update.AutoMergeEnabled)
	assert.True(t, *update.AutoMergeEnabled)
	assert.Nil(t, update.IsMerged)
}

func TestMergePR_NilPRData_ReturnsError(t *testing.T) {
	// unknownRowData has no *data.PullRequestData and no GetPrimaryPRData(),
	// so extractPullRequestData returns nil, triggering the nil-guard error path.
	msg := runMergePR(t, unknownRowData{})

	require.Error(t, msg.Err)
	assert.Contains(t, msg.Err.Error(), "could not resolve PR data")
}

func TestMergePR_DraftPR_ReturnsError(t *testing.T) {
	// No mock client needed: the IsDraft guard fires before any network call.
	pr := &data.PullRequestData{
		ID:               "PR_draft",
		Number:           42,
		IsDraft:          true,
		MergeStateStatus: "CLEAN",
		Repository:       data.Repository{AllowMergeCommit: true},
	}
	msg := runMergePR(t, pr)

	require.Error(t, msg.Err)
	assert.Contains(t, msg.Err.Error(), "draft")
}

func TestMergePR_GraphQLError_ReturnsError(t *testing.T) {
	setMockClient(t, mergeMockHandler(t, map[string]string{
		"MergePullRequest": `{"errors":[{"message":"Pull request is not mergeable","locations":[],"path":[]}]}`,
	}))

	pr := prDataFor("PR_clean2", "CLEAN", data.Repository{AllowMergeCommit: true})
	msg := runMergePR(t, pr)

	require.Error(t, msg.Err)
	assert.Contains(t, msg.Err.Error(), "Pull request is not mergeable")
}

func TestMergePR_RepoWithNoAllowedMethods_ReturnsError(t *testing.T) {
	// Empty repository - no allowed merge methods
	pr := prDataFor("PR_nomethods", "CLEAN", data.Repository{})
	msg := runMergePR(t, pr)

	require.Error(t, msg.Err)
	assert.Contains(t, msg.Err.Error(), "no allowed merge methods")
}
