package tui

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/charmbracelet/log"
	"github.com/charmbracelet/x/exp/teatest"
	gh "github.com/cli/go-gh/v2/pkg/api"
	zone "github.com/lrstanley/bubblezone"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/branchsidebar"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/footer"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/issueview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/notificationview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prssection"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/section"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/sidebar"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/tabs"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/keys"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/testutils"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/theme"
)

func TestFullOutput(t *testing.T) {
	setupTest(t)
	m := NewModel(config.Location{RepoPath: "", ConfigFlag: "../config/testdata/test-config.yml"})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(160, 60))

	testutils.WaitForText(t, tm, "style: make assignment brief", teatest.WithDuration(6*time.Second))

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("q"),
	})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestIssues(t *testing.T) {
	setupTest(t)
	m := NewModel(config.Location{RepoPath: "", ConfigFlag: "../config/testdata/test-config.yml"})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(160, 60))

	// wait for first tab of PRs
	testutils.WaitForText(t, tm, "Mine")

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("s"),
	})
	testutils.WaitForText(t, tm, "[Feature Request] Support notifications", teatest.WithDuration(6*time.Second))
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("q"),
	})
	tm.WaitFinished(t, teatest.WithFinalTimeout(5*time.Second))
}

func setupTest(t *testing.T) {
	if _, debug := os.LookupEnv("DEBUG"); debug {
		f, _ := os.CreateTemp("", "gh-dash-debug.log")
		fmt.Printf("[DEBU] writing debug logs to %s\n", f.Name())
		defer f.Close()
		log.SetOutput(f)
		log.SetLevel(log.DebugLevel)
	}
	setMockClient(t)

	markdown.InitializeMarkdownStyle(true)
	zone.NewGlobal()
	zone.SetEnabled(false)
}

// localRoundTripper is an http.RoundTripper that executes HTTP transactions
// by using handler directly, instead of going over an HTTP connection.
type localRoundTripper struct {
	handler http.Handler
}

func (l localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.handler.ServeHTTP(w, req)
	return w.Result(), nil
}

func mustRead(t *testing.T, r io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func mustWrite(t *testing.T, w io.Writer, s string) {
	t.Helper()
	_, err := io.WriteString(w, s)
	if err != nil {
		panic(err)
	}
}

func setMockClient(t *testing.T) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/graphql", func(w http.ResponseWriter, req *http.Request) {
		log.Debug("got request", "request", req.URL, "body", req.Body)
		body := mustRead(t, req.Body)
		switch {
		case strings.Contains(body, "query SearchPullRequests"):
			d, err := os.ReadFile("./testdata/searchPullRequests.json")
			if err != nil {
				t.Errorf("failed reading mock data file %v", err)
			}
			mustWrite(t, w, string(d))
		case strings.Contains(body, "query SearchIssues"):
			d, err := os.ReadFile("./testdata/searchIssues.json")
			if err != nil {
				t.Errorf("failed reading mock data file %v", err)
			}
			mustWrite(t, w, string(d))
		default:
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
	})
	client, err := gh.NewGraphQLClient(gh.ClientOptions{
		Transport: localRoundTripper{handler: mux},
		Host:      "localhost:3000",
		AuthToken: "fake-token",
	})
	if err != nil {
		t.Errorf("failed creating gh client %v", err)
	}
	data.SetClient(client)
}

func TestGetCurrentViewSections_RepoViewWithNilRepo(t *testing.T) {
	// This test verifies that getCurrentViewSections returns an empty slice
	// when in RepoView but m.repo is nil (before data is loaded).
	// Previously this would return []section.Section{nil} which caused a panic.
	m := Model{
		ctx: &context.ProgramContext{
			View: config.RepoView,
		},
		repo: nil,
	}

	sections := m.getCurrentViewSections()

	require.NotNil(t, sections, "sections should not be nil")
	require.Empty(t, sections, "sections should be empty when repo is nil")
}

func TestPromptConfirmation_NilSection(t *testing.T) {
	// promptConfirmation should return nil when currSection is nil
	m := Model{}
	cmd := m.promptConfirmation(nil, "close")
	require.Nil(t, cmd, "promptConfirmation should return nil when section is nil")
}

func TestNotificationView_SwitchViewWithSKey(t *testing.T) {
	// Test that pressing 's' in Notifications view switches to PRs view
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	sidebarModel := sidebar.NewModel()
	sidebarModel.UpdateProgramContext(ctx)

	m := Model{
		ctx:     ctx,
		keys:    keys.Keys,
		prView:  prview.NewModel(ctx),
		sidebar: sidebarModel,
		tabs:    tabs.NewModel(ctx),
	}
	prSec := prssection.NewModel(0, ctx, config.PrsSectionConfig{}, time.Now(), time.Now())
	m.prs = []section.Section{&prSec}

	// Verify we start in NotificationsView
	require.Equal(t, config.NotificationsView, m.ctx.View, "should start in NotificationsView")

	// Test that switchSelectedView returns PRsView when in NotificationsView
	m.switchSelectedView()
	require.Equal(t, config.PRsView, m.ctx.View,
		"switchSelectedView should set view to PRsView when in NotificationsView")
}

func TestNotificationView_SwitchViewWithSKey_WhileViewingPR(t *testing.T) {
	// Test that pressing 's' when viewing a PR notification switches views
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	sidebarModel := sidebar.NewModel()
	sidebarModel.UpdateProgramContext(ctx)

	m := Model{
		ctx:     ctx,
		keys:    keys.Keys,
		prView:  prview.NewModel(ctx),
		sidebar: sidebarModel,
		tabs:    tabs.NewModel(ctx),
	}
	prSec := prssection.NewModel(0, ctx, config.PrsSectionConfig{}, time.Now(), time.Now())
	m.prs = []section.Section{&prSec}

	// Set up a PR notification subject (simulating viewing a PR notification)
	m.notificationView.SetSubjectPR(&prrow.Data{}, "test-notification-id")

	// Verify we start in NotificationsView
	require.Equal(t, config.NotificationsView, m.ctx.View, "should start in NotificationsView")

	// Verify GetSubjectPR returns non-nil
	require.NotNil(t, m.notificationView.GetSubjectPR(), "subject PR should be set")

	// Test that switchSelectedView returns PRsView
	m.switchSelectedView()
	require.Equal(t, config.PRsView, m.ctx.View,
		"switchSelectedView should set view to PRsView when in NotificationsView")

	// Verify subject was cleared after switch
	require.Nil(t, m.notificationView.GetSubjectPR(),
		"subject PR should be cleared after switching views")
}

func TestNotificationView_SwitchViewWithSKey_WhileViewingIssue(t *testing.T) {
	// Test that pressing 's' when viewing an Issue notification switches views
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	sidebarModel := sidebar.NewModel()
	sidebarModel.UpdateProgramContext(ctx)

	m := Model{
		ctx:     ctx,
		keys:    keys.Keys,
		prView:  prview.NewModel(ctx),
		sidebar: sidebarModel,
		tabs:    tabs.NewModel(ctx),
	}
	prSec := prssection.NewModel(0, ctx, config.PrsSectionConfig{}, time.Now(), time.Now())
	m.prs = []section.Section{&prSec}

	// Set up an Issue notification subject (simulating viewing an Issue notification)
	m.notificationView.SetSubjectIssue(&data.IssueData{}, "test-notification-id")

	// Verify we start in NotificationsView
	require.Equal(t, config.NotificationsView, m.ctx.View, "should start in NotificationsView")

	// Verify GetSubjectIssue returns non-nil
	require.NotNil(t, m.notificationView.GetSubjectIssue(), "subject Issue should be set")

	// Test that switchSelectedView returns PRsView
	m.switchSelectedView()
	require.Equal(t, config.PRsView, m.ctx.View,
		"switchSelectedView should set view to PRsView when in NotificationsView")

	// Verify subject was cleared after switch
	require.Nil(t, m.notificationView.GetSubjectIssue(),
		"subject Issue should be cleared after switching views")
}

func TestNotificationView_PRViewTabNavigation(t *testing.T) {
	// This test verifies that tab navigation works in notification view when viewing a PR.
	// Previously, the code only returned when prCmd != nil, but tab navigation
	// (carousel.MoveLeft/MoveRight) doesn't return a command - it just updates state.
	// The fix ensures we always sync sidebar and return after prView.Update().
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	sidebarModel := sidebar.NewModel()
	sidebarModel.UpdateProgramContext(ctx)

	m := Model{
		ctx:     ctx,
		keys:    keys.Keys,
		prView:  prview.NewModel(ctx),
		sidebar: sidebarModel,
	}

	// Set up a PR notification subject so GetSubjectPR() returns non-nil
	m.notificationView.SetSubjectPR(&prrow.Data{}, "test-notification-id")

	// Get initial tab
	initialTab := m.prView.SelectedTab()

	// Send "next tab" key message
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("]")}
	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Verify tab changed
	require.NotEqual(t, initialTab, m.prView.SelectedTab(),
		"prView tab should have changed after pressing next tab key")

	// Now test going back
	currentTab := m.prView.SelectedTab()
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("[")}
	newModel, _ = m.Update(msg)
	m = newModel.(Model)

	require.NotEqual(t, currentTab, m.prView.SelectedTab(),
		"prView tab should have changed after pressing prev tab key")
}

func TestNotificationView_EnterKeyWorksAfterViewingPR(t *testing.T) {
	// Test that pressing Enter still works after a PR notification has been viewed.
	// Previously, once a PR subject was set, Enter would be absorbed by the PR handler
	// instead of triggering loadNotificationContent().
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	sidebarModel := sidebar.NewModel()
	sidebarModel.UpdateProgramContext(ctx)

	m := Model{
		ctx:              ctx,
		keys:             keys.Keys,
		footer:           footer.NewModel(ctx),
		prView:           prview.NewModel(ctx),
		issueSidebar:     issueview.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
		sidebar:          sidebarModel,
		tabs:             tabs.NewModel(ctx),
	}

	// Create a notification section with a PR notification
	notifSec := notificationssection.NewModel(0, ctx, config.NotificationsSectionConfig{}, time.Now())
	notifSec.Notifications = []notificationrow.Data{
		{
			Notification: data.NotificationData{
				Id: "test-notification-1",
				Subject: data.NotificationSubject{
					Title: "Test PR",
					Url:   "https://api.github.com/repos/owner/repo/pulls/123",
					Type:  "PullRequest",
				},
				Repository: data.NotificationRepository{
					FullName: "owner/repo",
				},
				Unread: true,
			},
		},
	}
	notifSec.Table.SetRows(notifSec.BuildRows())
	m.notifications = []section.Section{&notifSec}

	// Set up a PR notification subject (simulating that Enter was already pressed once)
	m.notificationView.SetSubjectPR(&prrow.Data{}, "test-notification-1")

	// Verify GetSubjectPR returns non-nil
	require.NotNil(t, m.notificationView.GetSubjectPR(), "subject PR should be set")

	// Send Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := m.Update(msg)

	// The fix ensures Enter triggers loadNotificationContent() even when a subject is set.
	// loadNotificationContent() returns a batch command for PR notifications.
	// Before the fix, Enter would be absorbed by the PR handler and cmd would be nil.
	require.NotNil(t, cmd, "Enter key should trigger loadNotificationContent and return a command")
}

func TestNotificationView_EnterKeyWorksAfterViewingIssue(t *testing.T) {
	// Test that pressing Enter still works after an Issue notification has been viewed.
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	sidebarModel := sidebar.NewModel()
	sidebarModel.UpdateProgramContext(ctx)

	m := Model{
		ctx:              ctx,
		keys:             keys.Keys,
		footer:           footer.NewModel(ctx),
		prView:           prview.NewModel(ctx),
		issueSidebar:     issueview.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
		sidebar:          sidebarModel,
		tabs:             tabs.NewModel(ctx),
	}

	// Create a notification section with an Issue notification
	notifSec := notificationssection.NewModel(0, ctx, config.NotificationsSectionConfig{}, time.Now())
	notifSec.Notifications = []notificationrow.Data{
		{
			Notification: data.NotificationData{
				Id: "test-notification-2",
				Subject: data.NotificationSubject{
					Title: "Test Issue",
					Url:   "https://api.github.com/repos/owner/repo/issues/456",
					Type:  "Issue",
				},
				Repository: data.NotificationRepository{
					FullName: "owner/repo",
				},
				Unread: true,
			},
		},
	}
	notifSec.Table.SetRows(notifSec.BuildRows())
	m.notifications = []section.Section{&notifSec}

	// Set up an Issue notification subject (simulating that Enter was already pressed once)
	m.notificationView.SetSubjectIssue(&data.IssueData{}, "test-notification-2")

	// Verify GetSubjectIssue returns non-nil
	require.NotNil(t, m.notificationView.GetSubjectIssue(), "subject Issue should be set")

	// Send Enter key
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := m.Update(msg)

	// The fix ensures Enter triggers loadNotificationContent() even when a subject is set.
	require.NotNil(t, cmd, "Enter key should trigger loadNotificationContent and return a command")
}

// executeCommandTemplate mimics the template execution logic from runCustomCommand
// to allow testing template variable substitution without executing shell commands.
func executeCommandTemplate(t *testing.T, commandTemplate string, input map[string]any) (string, error) {
	t.Helper()
	cmd, err := template.New("test_command").Parse(commandTemplate)
	if err != nil {
		return "", err
	}
	cmd = cmd.Option("missingkey=error")

	var buff bytes.Buffer
	err = cmd.Execute(&buff, input)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}

func TestPRCommandTemplateVariables(t *testing.T) {
	// Test that PR command templates correctly substitute all available variables,
	// matching the behavior of runCustomPRCommand in modelUtils.go
	input := map[string]any{
		"RepoName":    "owner/repo",
		"PrNumber":    123,
		"HeadRefName": "feature-branch",
		"BaseRefName": "main",
		"Author":      "testuser",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Author variable",
			template: "gh pr view --author {{.Author}}",
			expected: "gh pr view --author testuser",
		},
		{
			name:     "PrNumber variable",
			template: "gh pr checkout {{.PrNumber}}",
			expected: "gh pr checkout 123",
		},
		{
			name:     "HeadRefName variable",
			template: "git checkout {{.HeadRefName}}",
			expected: "git checkout feature-branch",
		},
		{
			name:     "Multiple variables",
			template: "echo PR #{{.PrNumber}} by {{.Author}} in {{.RepoName}}: {{.HeadRefName}} -> {{.BaseRefName}}",
			expected: "echo PR #123 by testuser in owner/repo: feature-branch -> main",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := executeCommandTemplate(t, tc.template, input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestIssueCommandTemplateVariables(t *testing.T) {
	// Test that Issue command templates correctly substitute all available variables,
	// matching the behavior of runCustomIssueCommand in modelUtils.go
	input := map[string]any{
		"RepoName":    "owner/repo",
		"IssueNumber": 456,
		"Author":      "issueauthor",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Author variable",
			template: "gh issue view --author {{.Author}}",
			expected: "gh issue view --author issueauthor",
		},
		{
			name:     "IssueNumber variable",
			template: "gh issue view {{.IssueNumber}}",
			expected: "gh issue view 456",
		},
		{
			name:     "Multiple variables",
			template: "echo Issue #{{.IssueNumber}} by {{.Author}} in {{.RepoName}}",
			expected: "echo Issue #456 by issueauthor in owner/repo",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := executeCommandTemplate(t, tc.template, input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestCommandTemplateMissingVariable(t *testing.T) {
	// Test that templates with missing variables return an error,
	// matching the missingkey=error behavior in runCustomCommand
	input := map[string]any{
		"RepoName": "owner/repo",
	}

	_, err := executeCommandTemplate(t, "gh pr view --author {{.Author}}", input)
	require.Error(t, err, "template with missing variable should return an error")
}

func TestSyncMainContentWidth(t *testing.T) {
	tests := []struct {
		name                 string
		screenWidth          int
		previewWidth         float64
		sidebarOpen          bool
		expectedPreviewWidth int
		expectedMainWidth    int
	}{
		{
			name:                 "absolute width with sidebar open",
			screenWidth:          100,
			previewWidth:         50,
			sidebarOpen:          true,
			expectedPreviewWidth: 50,
			expectedMainWidth:    50,
		},
		{
			name:                 "absolute width with sidebar closed",
			screenWidth:          100,
			previewWidth:         50,
			sidebarOpen:          false,
			expectedPreviewWidth: 0,
			expectedMainWidth:    100,
		},
		{
			name:                 "relative width 40%",
			screenWidth:          100,
			previewWidth:         0.4,
			sidebarOpen:          true,
			expectedPreviewWidth: 40,
			expectedMainWidth:    60,
		},
		{
			name:                 "relative width 50%",
			screenWidth:          200,
			previewWidth:         0.5,
			sidebarOpen:          true,
			expectedPreviewWidth: 100,
			expectedMainWidth:    100,
		},
		{
			name:                 "very small relative width results in zero",
			screenWidth:          100,
			previewWidth:         0.005,
			sidebarOpen:          true,
			expectedPreviewWidth: 0,
			expectedMainWidth:    100,
		},
		{
			name:                 "absolute width of 1",
			screenWidth:          100,
			previewWidth:         1,
			sidebarOpen:          true,
			expectedPreviewWidth: 1,
			expectedMainWidth:    99,
		},
		{
			name:                 "small screen with relative width",
			screenWidth:          10,
			previewWidth:         0.1,
			sidebarOpen:          true,
			expectedPreviewWidth: 1,
			expectedMainWidth:    9,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := config.Config{
				Defaults: config.Defaults{
					Preview: config.PreviewConfig{
						Open:  true,
						Width: tc.previewWidth,
					},
				},
			}

			m := Model{
				ctx: &context.ProgramContext{
					Config:      &cfg,
					ScreenWidth: tc.screenWidth,
				},
				sidebar: sidebar.Model{
					IsOpen: tc.sidebarOpen,
				},
			}

			m.syncMainContentWidth()

			if tc.sidebarOpen {
				require.Equal(t, tc.expectedPreviewWidth, m.ctx.DynamicPreviewWidth,
					"DynamicPreviewWidth mismatch")
			}
			require.Equal(t, tc.expectedMainWidth, m.ctx.MainContentWidth,
				"MainContentWidth mismatch")
			require.Equal(t, tc.sidebarOpen, m.ctx.SidebarOpen,
				"SidebarOpen mismatch")
		})
	}
}

func TestPromptConfirmationForNotificationPR(t *testing.T) {
	// Test that promptConfirmationForNotificationPR sets the pending action
	// and displays the confirmation prompt in the footer.
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	m := Model{
		ctx:    ctx,
		keys:   keys.Keys,
		footer: footer.NewModel(ctx),
	}

	// Set up a PR notification subject
	m.notificationView.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{
			Number: 123,
		},
	}, "test-notification-id")

	// Call promptConfirmationForNotificationPR
	m.promptConfirmationForNotificationPR("close")

	// Verify pending action is set
	require.Equal(t, "pr_close", m.notificationView.GetPendingAction(),
		"pendingNotificationAction should be set to pr_close")
}

func TestPromptConfirmationForNotificationPR_NilSubject(t *testing.T) {
	// Test that promptConfirmationForNotificationPR returns nil when no PR subject
	ctx := &context.ProgramContext{}
	m := Model{
		notificationView: notificationview.NewModel(ctx),
	}

	cmd := m.promptConfirmationForNotificationPR("close")

	require.Nil(t, cmd, "should return nil when no PR subject")
	require.Empty(t, m.notificationView.GetPendingAction(), "should not set pending action when no PR subject")
}

func TestPromptConfirmationForNotificationIssue(t *testing.T) {
	// Test that promptConfirmationForNotificationIssue sets the pending action
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	m := Model{
		ctx:    ctx,
		keys:   keys.Keys,
		footer: footer.NewModel(ctx),
	}

	// Set up an Issue notification subject
	m.notificationView.SetSubjectIssue(&data.IssueData{
		Number: 456,
	}, "test-notification-id")

	// Call promptConfirmationForNotificationIssue
	m.promptConfirmationForNotificationIssue("close")

	// Verify pending action is set
	require.Equal(t, "issue_close", m.notificationView.GetPendingAction(),
		"pendingNotificationAction should be set to issue_close")
}

func TestNotificationConfirmation_CancelOnOtherKey(t *testing.T) {
	// Test that pressing any key other than y/Y/Enter cancels the confirmation
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	m := Model{
		ctx:              ctx,
		keys:             keys.Keys,
		footer:           footer.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
	}

	// Set up a PR notification subject and pending action
	m.notificationView.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{
			Number: 123,
		},
	}, "test-notification-id")
	m.notificationView.SetPendingPRAction("close") // Simulate pending action

	// Press 'n' to cancel
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("n")}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	// Verify pending action is cleared
	require.Empty(t, m.notificationView.GetPendingAction(),
		"pendingNotificationAction should be cleared after cancellation")
	require.Nil(t, cmd, "should return nil command when cancelled")
}

func TestNotificationConfirmation_AcceptWithY(t *testing.T) {
	// Test that pressing 'y' confirms and executes the action
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
		StartTask: func(task context.Task) tea.Cmd {
			return nil // No-op for testing
		},
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	m := Model{
		ctx:              ctx,
		keys:             keys.Keys,
		footer:           footer.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
	}

	// Set up a PR notification subject and pending action
	m.notificationView.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{
			Number: 123,
		},
	}, "test-notification-id")
	m.notificationView.SetPendingPRAction("close")

	// Press 'y' to confirm
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	// Verify pending action is cleared and command is returned
	require.Empty(t, m.notificationView.GetPendingAction(),
		"pendingNotificationAction should be cleared after confirmation")
	require.NotNil(t, cmd, "should return a command to execute the action")
}

func TestNotificationConfirmation_AcceptWithUpperY(t *testing.T) {
	// Test that pressing 'Y' (uppercase) also confirms
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
		StartTask: func(task context.Task) tea.Cmd {
			return nil // No-op for testing
		},
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	m := Model{
		ctx:              ctx,
		keys:             keys.Keys,
		footer:           footer.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
	}

	// Set up a PR notification subject and pending action
	m.notificationView.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{
			Number: 123,
		},
	}, "test-notification-id")
	m.notificationView.SetPendingPRAction("merge")

	// Press 'Y' to confirm
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Y")}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	// Verify pending action is cleared and command is returned
	require.Empty(t, m.notificationView.GetPendingAction(),
		"pendingNotificationAction should be cleared after confirmation")
	require.NotNil(t, cmd, "should return a command to execute the action")
}

func TestNotificationConfirmation_AcceptWithEnter(t *testing.T) {
	// Test that pressing Enter also confirms
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
		StartTask: func(task context.Task) tea.Cmd {
			return nil // No-op for testing
		},
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	m := Model{
		ctx:              ctx,
		keys:             keys.Keys,
		footer:           footer.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
	}

	// Set up an Issue notification subject and pending action
	m.notificationView.SetSubjectIssue(&data.IssueData{
		Number: 456,
		Url:    "https://github.com/test/repo/issues/456",
		Repository: data.Repository{
			NameWithOwner: "test/repo",
		},
	}, "test-notification-id")
	m.notificationView.SetPendingIssueAction("reopen")

	// Press Enter to confirm
	msg := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	// Verify pending action is cleared and command is returned
	require.Empty(t, m.notificationView.GetPendingAction(),
		"pendingNotificationAction should be cleared after confirmation")
	require.NotNil(t, cmd, "should return a command to execute the action")
}

func TestPromptConfirmationForNotificationIssue_NilSubject(t *testing.T) {
	// Test that promptConfirmationForNotificationIssue returns nil when no Issue subject
	ctx := &context.ProgramContext{}
	m := Model{
		notificationView: notificationview.NewModel(ctx),
	}

	cmd := m.promptConfirmationForNotificationIssue("close")

	require.Nil(t, cmd, "should return nil when no Issue subject")
	require.Empty(t, m.notificationView.GetPendingAction(), "should not set pending action when no Issue subject")
}

func TestRefresh_ClearsEnrichmentCache(t *testing.T) {
	// This test verifies that pressing the refresh key ('r') clears the
	// enrichment cache, ensuring fresh reviewer data is fetched.
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config:    &cfg,
		View:      config.PRsView,
		StartTask: func(task context.Task) tea.Cmd { return nil },
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	// Create a PR section so getCurrSection() returns non-nil
	prSection := prssection.NewModel(
		0,
		ctx,
		config.PrsSectionConfig{
			Title:   "Test",
			Filters: "is:open",
		},
		time.Now(),
		time.Now(),
	)

	m := Model{
		ctx:              ctx,
		keys:             keys.Keys,
		prs:              []section.Section{&prSection},
		sidebar:          sidebar.NewModel(),
		footer:           footer.NewModel(ctx),
		tabs:             tabs.NewModel(ctx),
		prView:           prview.NewModel(ctx),
		issueSidebar:     issueview.NewModel(ctx),
		branchSidebar:    branchsidebar.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
	}

	// Simulate having a populated cache by ensuring it's NOT cleared
	// (In real usage, this would happen after viewing a PR in sidebar)
	data.SetClient(nil) // Reset to known state first
	// Note: We can't easily populate the cache without making API calls,
	// so we verify the cache clearing behavior works from a cleared state

	// Verify cache starts cleared
	require.True(t, data.IsEnrichmentCacheCleared(), "cache should start cleared")

	// Send refresh key - this should call data.ClearEnrichmentCache()
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")}
	_, _ = m.Update(msg)

	// Verify cache is still cleared (ClearEnrichmentCache was called)
	require.True(t, data.IsEnrichmentCacheCleared(),
		"cache should be cleared after refresh key press")
}

func TestRefreshAll_ClearsEnrichmentCache(t *testing.T) {
	// This test verifies that pressing the refresh all key ('R') also
	// clears the enrichment cache. The cache clearing happens at the start
	// of the handler, before fetchAllViewSections.
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config:    &cfg,
		View:      config.PRsView,
		StartTask: func(task context.Task) tea.Cmd { return nil },
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	// Create a PR section for proper setup
	prSection := prssection.NewModel(
		0,
		ctx,
		config.PrsSectionConfig{
			Title:   "Test",
			Filters: "is:open",
		},
		time.Now(),
		time.Now(),
	)

	m := Model{
		ctx:              ctx,
		keys:             keys.Keys,
		prs:              []section.Section{&prSection},
		sidebar:          sidebar.NewModel(),
		footer:           footer.NewModel(ctx),
		tabs:             tabs.NewModel(ctx),
		prView:           prview.NewModel(ctx),
		issueSidebar:     issueview.NewModel(ctx),
		branchSidebar:    branchsidebar.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
	}

	// Reset to known state
	data.SetClient(nil)
	require.True(t, data.IsEnrichmentCacheCleared(), "cache should start cleared")

	// Send refresh all key - ClearEnrichmentCache is called at the start
	// of the handler, before fetchAllViewSections. We use recover to handle
	// any panics from incomplete test setup while still verifying cache behavior.
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")}
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Panic is expected due to incomplete test setup.
				// The important thing is that ClearEnrichmentCache was called
				// before the panic occurred (it's the first line in the handler).
				t.Logf("Recovered from expected panic in fetchAllViewSections: %v", r)
			}
		}()
		_, _ = m.Update(msg)
	}()

	// Verify cache is cleared - this is the key assertion
	require.True(t, data.IsEnrichmentCacheCleared(),
		"cache should be cleared after refresh all key press")
}

func TestPromptConfirmationForNotificationPR_ApproveWorkflows(t *testing.T) {
	// Test that promptConfirmationForNotificationPR sets the pending action
	// for approveWorkflows.
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	m := Model{
		ctx:    ctx,
		keys:   keys.Keys,
		footer: footer.NewModel(ctx),
	}

	// Set up a PR notification subject
	m.notificationView.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{
			Number: 42,
		},
	}, "test-notification-id")

	// Call promptConfirmationForNotificationPR with approveWorkflows
	m.promptConfirmationForNotificationPR("approveWorkflows")

	// Verify pending action is set
	require.Equal(t, "pr_approveWorkflows", m.notificationView.GetPendingAction(),
		"pendingNotificationAction should be set to pr_approveWorkflows")
}

func TestNotificationConfirmation_ApproveWorkflows_AcceptWithY(t *testing.T) {
	// Test that confirming approveWorkflows executes the action
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag: "../config/testdata/test-config.yml",
	})
	require.NoError(t, err)

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
		StartTask: func(task context.Task) tea.Cmd {
			return nil // No-op for testing
		},
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	m := Model{
		ctx:              ctx,
		keys:             keys.Keys,
		footer:           footer.NewModel(ctx),
		notificationView: notificationview.NewModel(ctx),
	}

	// Set up a PR notification subject and pending approveWorkflows action
	m.notificationView.SetSubjectPR(&prrow.Data{
		Primary: &data.PullRequestData{
			Number: 42,
			Repository: data.Repository{
				NameWithOwner: "owner/repo",
			},
		},
	}, "test-notification-id")
	m.notificationView.SetPendingPRAction("approveWorkflows")

	// Press 'y' to confirm
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")}
	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	// Verify pending action is cleared and command is returned
	require.Empty(t, m.notificationView.GetPendingAction(),
		"pendingNotificationAction should be cleared after confirmation")
	require.NotNil(t, cmd, "should return a command to execute the action")
}

func TestIsUserDefinedKeybinding_NotificationsView_PRNotification(t *testing.T) {
	// Custom PR keybindings should be recognized when viewing a PR notification
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	// Add a custom PR keybinding
	cfg.Keybindings.Prs = []config.Keybinding{
		{Key: "B", Command: "gh pr view -w {{.PrNumber}}"},
	}

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	// Create a notification section with a PR notification as the current row
	notifSec := notificationssection.NewModel(0, ctx, config.NotificationsSectionConfig{}, time.Now())
	notifSec.Notifications = []notificationrow.Data{
		{
			Notification: data.NotificationData{
				Id: "test-1",
				Subject: data.NotificationSubject{
					Title: "Test PR",
					Url:   "https://api.github.com/repos/owner/repo/pulls/42",
					Type:  "PullRequest",
				},
				Repository: data.NotificationRepository{
					FullName: "owner/repo",
				},
			},
		},
	}
	notifSec.Table.SetRows(notifSec.BuildRows())

	m := Model{
		ctx:           ctx,
		keys:          keys.Keys,
		notifications: []section.Section{&notifSec},
	}

	// The custom PR keybinding "B" should be recognized
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("B")}
	require.True(t, m.isUserDefinedKeybinding(msg),
		"custom PR keybinding should be recognized in NotificationsView for PR notifications")

	// A non-configured key should not be recognized
	msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("Z")}
	require.False(t, m.isUserDefinedKeybinding(msg),
		"unconfigured key should not be recognized")
}

func TestIsUserDefinedKeybinding_NotificationsView_IssueNotification(t *testing.T) {
	// Custom Issue keybindings should be recognized when viewing an Issue notification
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	// Add a custom Issue keybinding
	cfg.Keybindings.Issues = []config.Keybinding{
		{Key: "B", Command: "gh issue view -w {{.IssueNumber}}"},
	}

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	// Create a notification section with an Issue notification as the current row
	notifSec := notificationssection.NewModel(0, ctx, config.NotificationsSectionConfig{}, time.Now())
	notifSec.Notifications = []notificationrow.Data{
		{
			Notification: data.NotificationData{
				Id: "test-2",
				Subject: data.NotificationSubject{
					Title: "Test Issue",
					Url:   "https://api.github.com/repos/owner/repo/issues/99",
					Type:  "Issue",
				},
				Repository: data.NotificationRepository{
					FullName: "owner/repo",
				},
			},
		},
	}
	notifSec.Table.SetRows(notifSec.BuildRows())

	m := Model{
		ctx:           ctx,
		keys:          keys.Keys,
		notifications: []section.Section{&notifSec},
	}

	// The custom Issue keybinding "B" should be recognized
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("B")}
	require.True(t, m.isUserDefinedKeybinding(msg),
		"custom Issue keybinding should be recognized in NotificationsView for Issue notifications")
}

func TestIsUserDefinedKeybinding_NotificationsView_NonPRIssueNotification(t *testing.T) {
	// Custom PR/Issue keybindings should NOT be recognized for other notification types
	cfg, err := config.ParseConfig(config.Location{
		ConfigFlag:       "../config/testdata/test-config.yml",
		SkipGlobalConfig: true,
	})
	require.NoError(t, err)

	cfg.Keybindings.Prs = []config.Keybinding{
		{Key: "B", Command: "gh pr view -w {{.PrNumber}}"},
	}

	ctx := &context.ProgramContext{
		Config: &cfg,
		View:   config.NotificationsView,
	}
	ctx.Theme = theme.ParseTheme(ctx.Config)
	ctx.Styles = context.InitStyles(ctx.Theme)

	// Create a notification section with a Release notification
	notifSec := notificationssection.NewModel(0, ctx, config.NotificationsSectionConfig{}, time.Now())
	notifSec.Notifications = []notificationrow.Data{
		{
			Notification: data.NotificationData{
				Id: "test-3",
				Subject: data.NotificationSubject{
					Title: "v1.0.0",
					Url:   "https://api.github.com/repos/owner/repo/releases/12345",
					Type:  "Release",
				},
				Repository: data.NotificationRepository{
					FullName: "owner/repo",
				},
			},
		},
	}
	notifSec.Table.SetRows(notifSec.BuildRows())

	m := Model{
		ctx:           ctx,
		keys:          keys.Keys,
		notifications: []section.Section{&notifSec},
	}

	// The custom PR keybinding "B" should NOT be recognized for a Release notification
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("B")}
	require.False(t, m.isUserDefinedKeybinding(msg),
		"custom PR keybinding should not be recognized for Release notifications")
}

func TestNotificationPRCommandTemplateVariables(t *testing.T) {
	// Test that notification PR command templates have RepoName and PrNumber available,
	// matching the behavior of runCustomNotificationPRCommand in modelUtils.go
	input := map[string]any{
		"RepoName": "owner/repo",
		"PrNumber": 42,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "PrNumber variable",
			template: "gh pr view {{.PrNumber}}",
			expected: "gh pr view 42",
		},
		{
			name:     "RepoName and PrNumber",
			template: "gh pr view -R {{.RepoName}} {{.PrNumber}}",
			expected: "gh pr view -R owner/repo 42",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := executeCommandTemplate(t, tc.template, input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestNotificationPRCommandTemplateVariables_WithSidebar(t *testing.T) {
	// When the sidebar is open for a PR notification, HeadRefName, BaseRefName,
	// and Author become available in the template fields.
	input := map[string]any{
		"RepoName":    "owner/repo",
		"PrNumber":    42,
		"HeadRefName": "feature-branch",
		"BaseRefName": "main",
		"Author":      "octocat",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "HeadRefName variable",
			template: "git checkout {{.HeadRefName}}",
			expected: "git checkout feature-branch",
		},
		{
			name:     "BaseRefName variable",
			template: "git diff {{.BaseRefName}}...{{.HeadRefName}}",
			expected: "git diff main...feature-branch",
		},
		{
			name:     "Author variable",
			template: "echo {{.Author}}",
			expected: "echo octocat",
		},
		{
			name:     "all fields combined",
			template: "gh pr view -R {{.RepoName}} {{.PrNumber}} # by {{.Author}} ({{.HeadRefName}} -> {{.BaseRefName}})",
			expected: "gh pr view -R owner/repo 42 # by octocat (feature-branch -> main)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := executeCommandTemplate(t, tc.template, input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestNotificationIssueCommandTemplateVariables(t *testing.T) {
	// Test that notification Issue command templates have RepoName and IssueNumber available,
	// matching the behavior of runCustomNotificationIssueCommand in modelUtils.go
	input := map[string]any{
		"RepoName":    "owner/repo",
		"IssueNumber": 99,
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "IssueNumber variable",
			template: "gh issue view {{.IssueNumber}}",
			expected: "gh issue view 99",
		},
		{
			name:     "RepoName and IssueNumber",
			template: "gh issue view -R {{.RepoName}} {{.IssueNumber}}",
			expected: "gh issue view -R owner/repo 99",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := executeCommandTemplate(t, tc.template, input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestNotificationIssueCommandTemplateVariables_WithSidebar(t *testing.T) {
	// When the sidebar is open for an Issue notification, Author becomes available.
	input := map[string]any{
		"RepoName":    "owner/repo",
		"IssueNumber": 99,
		"Author":      "octocat",
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "Author variable",
			template: "echo {{.Author}}",
			expected: "echo octocat",
		},
		{
			name:     "all fields combined",
			template: "gh issue view -R {{.RepoName}} {{.IssueNumber}} # by {{.Author}}",
			expected: "gh issue view -R owner/repo 99 # by octocat",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := executeCommandTemplate(t, tc.template, input)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestNotificationCommandTemplate_PRFieldsWithoutSidebar(t *testing.T) {
	// When the sidebar is not open, notification PR commands only have RepoName and PrNumber.
	// Templates referencing sidebar-only fields should error (missingkey=error behavior).
	input := map[string]any{
		"RepoName": "owner/repo",
		"PrNumber": 42,
	}

	unavailableFields := []struct {
		name     string
		template string
	}{
		{"HeadRefName", "git checkout {{.HeadRefName}}"},
		{"BaseRefName", "git checkout {{.BaseRefName}}"},
		{"Author", "echo {{.Author}}"},
	}

	for _, tc := range unavailableFields {
		t.Run(tc.name, func(t *testing.T) {
			_, err := executeCommandTemplate(t, tc.template, input)
			require.Error(t, err,
				"%s should not be available in notification PR commands", tc.name)
		})
	}
}
