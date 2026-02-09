package prview

import (
	"strings"
	"testing"

	graphql "github.com/cli/shurcooL-graphql"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
	checks "github.com/dlvhdr/x/gh-checks"
)

type checksTestOptions struct {
	checkSuites          data.CheckSuites
	checkRuns            []data.CheckRun
	rollupState          string
	requiredStatusChecks []string
}

func newTestModelForChecks(t *testing.T, opts checksTestOptions) Model {
	t.Helper()
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../../../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	thm := theme.ParseTheme(&cfg)
	ctx := &context.ProgramContext{
		Config: &cfg,
		Theme:  thm,
		Styles: context.InitStyles(thm),
	}

	// Build branch protection rules
	branchRules := data.BranchProtectionRules{}
	if len(opts.requiredStatusChecks) > 0 {
		ruleNode := struct {
			RequiredApprovingReviewCount int
			RequiresApprovingReviews     graphql.Boolean
			RequiresCodeOwnerReviews     graphql.Boolean
			RequiresStatusChecks         graphql.Boolean
			RequiredStatusCheckContexts  []graphql.String
		}{
			RequiresStatusChecks: true,
		}
		for _, ctx := range opts.requiredStatusChecks {
			ruleNode.RequiredStatusCheckContexts = append(
				ruleNode.RequiredStatusCheckContexts,
				graphql.String(ctx),
			)
		}
		branchRules.Nodes = append(branchRules.Nodes, ruleNode)
	}

	// Build the enriched data with commits
	enriched := data.EnrichedPullRequestData{}

	// We need to directly construct the Commits field
	// Since it uses anonymous struct types, we build it inline
	enriched.Commits.TotalCount = 1
	enriched.Commits.Nodes = make([]struct {
		Commit struct {
			Deployments struct {
				Nodes []struct {
					Task        graphql.String
					Description graphql.String
				}
			} `graphql:"deployments(last: 10)"`
			CommitUrl         graphql.String
			StatusCheckRollup struct {
				State    graphql.String
				Contexts struct {
					TotalCount                 graphql.Int
					CheckRunCount              graphql.Int
					CheckRunCountsByState      []data.ContextCountByState
					StatusContextCount         graphql.Int
					StatusContextCountsByState []data.ContextCountByState
					Nodes                      []struct {
						Typename      graphql.String     `graphql:"__typename"`
						CheckRun      data.CheckRun      `graphql:"... on CheckRun"`
						StatusContext data.StatusContext `graphql:"... on StatusContext"`
					}
				} `graphql:"contexts(last: 100)"`
			}
			CheckSuites data.CheckSuites `graphql:"checkSuites(last: 20)"`
		}
	}, 1)

	// Set up the commit data
	enriched.Commits.Nodes[0].Commit.CheckSuites = opts.checkSuites
	enriched.Commits.Nodes[0].Commit.StatusCheckRollup.State = graphql.String(opts.rollupState)
	enriched.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.TotalCount = graphql.Int(len(opts.checkRuns))
	enriched.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.CheckRunCount = graphql.Int(len(opts.checkRuns))

	// Build context nodes from check runs and aggregate counts by state
	stateCounts := make(map[string]int)
	for _, cr := range opts.checkRuns {
		contextNode := struct {
			Typename      graphql.String     `graphql:"__typename"`
			CheckRun      data.CheckRun      `graphql:"... on CheckRun"`
			StatusContext data.StatusContext `graphql:"... on StatusContext"`
		}{
			Typename: "CheckRun",
			CheckRun: cr,
		}
		enriched.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.Nodes = append(
			enriched.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.Nodes,
			contextNode,
		)

		// Aggregate by state (status for in-progress, conclusion for completed)
		state := string(cr.Status)
		if state == "COMPLETED" {
			state = string(cr.Conclusion)
		}
		stateCounts[state]++
	}

	// Populate CheckRunCountsByState from aggregated counts
	for state, count := range stateCounts {
		enriched.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.CheckRunCountsByState = append(
			enriched.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.CheckRunCountsByState,
			data.ContextCountByState{
				State: checks.CheckRunState(state),
				Count: graphql.Int(count),
			},
		)
	}

	m := NewModel(ctx)
	m.ctx = ctx
	m.width = 80
	m.pr = &prrow.PullRequest{
		Ctx: ctx,
		Data: &prrow.Data{
			Primary: &data.PullRequestData{
				Repository: data.Repository{
					BranchProtectionRules: branchRules,
				},
			},
			IsEnriched: true,
			Enriched:   enriched,
		},
	}
	return m
}

func makeCheckRun(name string, status string, conclusion checks.CheckRunState) data.CheckRun {
	return data.CheckRun{
		Name:       graphql.String(name),
		Status:     graphql.String(status),
		Conclusion: conclusion,
	}
}

func makeCheckSuite(workflowName string, status string, conclusion string) data.CheckSuiteNode {
	return data.CheckSuiteNode{
		Status:     graphql.String(status),
		Conclusion: graphql.String(conclusion),
		WorkflowRun: struct {
			Workflow struct {
				Name graphql.String
			}
		}{
			Workflow: struct {
				Name graphql.String
			}{
				Name: graphql.String(workflowName),
			},
		},
	}
}

func TestRenderChecks_AwaitingApproval(t *testing.T) {
	// Test that CheckSuites with conclusion: ACTION_REQUIRED are shown
	// under "Awaiting Approval" section
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 3,
			Nodes: []data.CheckSuiteNode{
				makeCheckSuite("Check Redirects", "COMPLETED", "ACTION_REQUIRED"),
				makeCheckSuite("Check URL issues", "COMPLETED", "ACTION_REQUIRED"),
				makeCheckSuite("PR Test", "COMPLETED", "ACTION_REQUIRED"),
			},
		},
		checkRuns:   []data.CheckRun{},
		rollupState: "SUCCESS",
	}

	m := newTestModelForChecks(t, opts)
	got := m.renderChecks()

	// Should show "Awaiting Approval" section header
	require.True(t, strings.Contains(got, "Awaiting Approval"),
		"expected 'Awaiting Approval' section, got: %q", got)

	// Should show all three workflow names
	require.True(t, strings.Contains(got, "Check Redirects"),
		"expected 'Check Redirects' workflow, got: %q", got)
	require.True(t, strings.Contains(got, "Check URL issues"),
		"expected 'Check URL issues' workflow, got: %q", got)
	require.True(t, strings.Contains(got, "PR Test"),
		"expected 'PR Test' workflow, got: %q", got)

	// Should show the action required icon
	require.True(t, strings.Contains(got, constants.ActionRequiredIcon),
		"expected ActionRequiredIcon, got: %q", got)
}

func TestRenderChecks_PendingCheckSuites(t *testing.T) {
	// Test that CheckSuites with status: QUEUED/PENDING/WAITING are shown
	// under "Pending" section
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 2,
			Nodes: []data.CheckSuiteNode{
				makeCheckSuite("Build", "QUEUED", ""),
				makeCheckSuite("Deploy", "PENDING", ""),
			},
		},
		checkRuns:   []data.CheckRun{},
		rollupState: "PENDING",
	}

	m := newTestModelForChecks(t, opts)
	got := m.renderChecks()

	// Should show "Pending" section header
	require.True(t, strings.Contains(got, "Pending"),
		"expected 'Pending' section, got: %q", got)

	// Should show workflow names
	require.True(t, strings.Contains(got, "Build"),
		"expected 'Build' workflow, got: %q", got)
	require.True(t, strings.Contains(got, "Deploy"),
		"expected 'Deploy' workflow, got: %q", got)

	// Should show waiting icon
	require.True(t, strings.Contains(got, constants.WaitingIcon),
		"expected WaitingIcon, got: %q", got)
}

func TestRenderChecks_RequiredButNotReported(t *testing.T) {
	// Test that required status checks from branch protection rules
	// that haven't been reported yet are shown as pending
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 0,
			Nodes:      []data.CheckSuiteNode{},
		},
		checkRuns: []data.CheckRun{
			// Only one check has been reported
			makeCheckRun("lint", "COMPLETED", "SUCCESS"),
		},
		rollupState: "SUCCESS",
		requiredStatusChecks: []string{
			"lint",            // This one is reported
			"check-redirects", // This one is NOT reported
			"tests",           // This one is NOT reported
		},
	}

	m := newTestModelForChecks(t, opts)
	got := m.renderChecks()

	// Should show "Pending" section for unreported required checks
	require.True(t, strings.Contains(got, "Pending"),
		"expected 'Pending' section for unreported required checks, got: %q", got)

	// Should show the unreported required checks
	require.True(t, strings.Contains(got, "check-redirects"),
		"expected 'check-redirects' as pending, got: %q", got)
	require.True(t, strings.Contains(got, "tests"),
		"expected 'tests' as pending, got: %q", got)

	// Should NOT show "lint" in pending since it was reported
	// Count occurrences - lint should only appear once (in the success section)
	lintCount := strings.Count(got, "lint")
	require.Equal(t, 1, lintCount,
		"expected 'lint' to appear exactly once (not in pending), got %d occurrences in: %q", lintCount, got)
}

func TestRenderChecks_MixedStates(t *testing.T) {
	// Test a realistic scenario with:
	// - Awaiting approval workflows
	// - Some successful checks
	// - Required checks not yet reported
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 3,
			Nodes: []data.CheckSuiteNode{
				// Awaiting approval
				makeCheckSuite("Check Redirects", "COMPLETED", "ACTION_REQUIRED"),
				makeCheckSuite("PR Test", "COMPLETED", "ACTION_REQUIRED"),
				// Completed successfully (not ACTION_REQUIRED)
				makeCheckSuite("PR labeler", "COMPLETED", "SUCCESS"),
			},
		},
		checkRuns: []data.CheckRun{
			// Successful checks from PR labeler
			makeCheckRun("Label by path", "COMPLETED", "SUCCESS"),
			makeCheckRun("Label by size", "COMPLETED", "SUCCESS"),
		},
		rollupState: "SUCCESS",
		requiredStatusChecks: []string{
			"Label by path",   // Reported
			"check-redirects", // NOT reported (from awaiting approval workflow)
			"tests",           // NOT reported (from awaiting approval workflow)
		},
	}

	m := newTestModelForChecks(t, opts)
	got := m.renderChecks()

	// Should have "Awaiting Approval" section
	require.True(t, strings.Contains(got, "Awaiting Approval"),
		"expected 'Awaiting Approval' section, got: %q", got)
	require.True(t, strings.Contains(got, "Check Redirects"),
		"expected 'Check Redirects' in awaiting approval, got: %q", got)
	require.True(t, strings.Contains(got, "PR Test"),
		"expected 'PR Test' in awaiting approval, got: %q", got)

	// Should have "Pending" section for required but not reported
	require.True(t, strings.Contains(got, "Pending"),
		"expected 'Pending' section, got: %q", got)
	require.True(t, strings.Contains(got, "check-redirects"),
		"expected 'check-redirects' as pending, got: %q", got)
	require.True(t, strings.Contains(got, "tests"),
		"expected 'tests' as pending, got: %q", got)

	// Should have successful checks
	require.True(t, strings.Contains(got, "Label by path"),
		"expected 'Label by path' check, got: %q", got)
	require.True(t, strings.Contains(got, "Label by size"),
		"expected 'Label by size' check, got: %q", got)
	require.True(t, strings.Contains(got, constants.SuccessIcon),
		"expected SuccessIcon for successful checks, got: %q", got)
}

func TestRenderChecks_NoChecks(t *testing.T) {
	// Test that "No checks to display..." is shown when there are no checks
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 0,
			Nodes:      []data.CheckSuiteNode{},
		},
		checkRuns:            []data.CheckRun{},
		rollupState:          "SUCCESS",
		requiredStatusChecks: []string{},
	}

	m := newTestModelForChecks(t, opts)
	got := m.renderChecks()

	require.True(t, strings.Contains(got, "No checks to display"),
		"expected 'No checks to display...' message, got: %q", got)
}

func TestRenderChecks_FailedChecks(t *testing.T) {
	// Test that failed checks are displayed with failure icon
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 0,
			Nodes:      []data.CheckSuiteNode{},
		},
		checkRuns: []data.CheckRun{
			makeCheckRun("build", "COMPLETED", "FAILURE"),
			makeCheckRun("lint", "COMPLETED", "SUCCESS"),
		},
		rollupState: "FAILURE",
	}

	m := newTestModelForChecks(t, opts)
	got := m.renderChecks()

	// Should show both checks
	require.True(t, strings.Contains(got, "build"),
		"expected 'build' check, got: %q", got)
	require.True(t, strings.Contains(got, "lint"),
		"expected 'lint' check, got: %q", got)

	// Should have failure icon
	require.True(t, strings.Contains(got, constants.FailureIcon),
		"expected FailureIcon for failed check, got: %q", got)

	// Should have success icon
	require.True(t, strings.Contains(got, constants.SuccessIcon),
		"expected SuccessIcon for successful check, got: %q", got)
}

func TestRenderChecks_InProgressChecks(t *testing.T) {
	// Test that in-progress checks are displayed with waiting icon
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 0,
			Nodes:      []data.CheckSuiteNode{},
		},
		checkRuns: []data.CheckRun{
			makeCheckRun("build", "IN_PROGRESS", ""),
			makeCheckRun("lint", "QUEUED", ""),
		},
		rollupState: "PENDING",
	}

	m := newTestModelForChecks(t, opts)
	got := m.renderChecks()

	// Should show both checks
	require.True(t, strings.Contains(got, "build"),
		"expected 'build' check, got: %q", got)
	require.True(t, strings.Contains(got, "lint"),
		"expected 'lint' check, got: %q", got)

	// Should have waiting icon for in-progress checks
	require.True(t, strings.Contains(got, constants.WaitingIcon),
		"expected WaitingIcon for in-progress checks, got: %q", got)
}

func TestGetChecksStats_AwaitingApproval(t *testing.T) {
	// Test that getChecksStats correctly counts awaiting approval
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 3,
			Nodes: []data.CheckSuiteNode{
				makeCheckSuite("Check Redirects", "COMPLETED", "ACTION_REQUIRED"),
				makeCheckSuite("Check URL issues", "COMPLETED", "ACTION_REQUIRED"),
				makeCheckSuite("PR Test", "COMPLETED", "ACTION_REQUIRED"),
			},
		},
		checkRuns:   []data.CheckRun{},
		rollupState: "SUCCESS",
	}

	m := newTestModelForChecks(t, opts)
	stats := m.getChecksStats()

	require.Equal(t, 3, stats.awaitingApproval,
		"expected 3 awaiting approval, got: %d", stats.awaitingApproval)
}

func TestGetChecksStats_PendingCheckSuites(t *testing.T) {
	// Test that getChecksStats correctly counts pending check suites as inProgress
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 2,
			Nodes: []data.CheckSuiteNode{
				makeCheckSuite("Build", "QUEUED", ""),
				makeCheckSuite("Deploy", "WAITING", ""),
			},
		},
		checkRuns:   []data.CheckRun{},
		rollupState: "PENDING",
	}

	m := newTestModelForChecks(t, opts)
	stats := m.getChecksStats()

	require.Equal(t, 2, stats.inProgress,
		"expected 2 in progress (from pending check suites), got: %d", stats.inProgress)
}

func TestGetChecksStats_Mixed(t *testing.T) {
	// Test a mix of check states
	opts := checksTestOptions{
		checkSuites: data.CheckSuites{
			TotalCount: 2,
			Nodes: []data.CheckSuiteNode{
				makeCheckSuite("Check Redirects", "COMPLETED", "ACTION_REQUIRED"),
				makeCheckSuite("Build", "QUEUED", ""),
			},
		},
		checkRuns: []data.CheckRun{
			makeCheckRun("lint", "COMPLETED", "SUCCESS"),
			makeCheckRun("test", "COMPLETED", "FAILURE"),
			makeCheckRun("build", "IN_PROGRESS", ""),
		},
		rollupState: "FAILURE",
	}

	m := newTestModelForChecks(t, opts)
	stats := m.getChecksStats()

	require.Equal(t, 1, stats.awaitingApproval,
		"expected 1 awaiting approval, got: %d", stats.awaitingApproval)
	require.Equal(t, 1, stats.succeeded,
		"expected 1 succeeded, got: %d", stats.succeeded)
	require.Equal(t, 1, stats.failed,
		"expected 1 failed, got: %d", stats.failed)
	// 1 from IN_PROGRESS check run + 1 from QUEUED check suite
	require.Equal(t, 2, stats.inProgress,
		"expected 2 in progress, got: %d", stats.inProgress)
}
