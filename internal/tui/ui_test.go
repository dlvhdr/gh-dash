package tui

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/charmbracelet/log"
	"github.com/charmbracelet/x/exp/teatest"
	gh "github.com/cli/go-gh/v2/pkg/api"
	zone "github.com/lrstanley/bubblezone"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prrow"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prview"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/sidebar"
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
	}

	// Verify we start in NotificationsView
	require.Equal(t, config.NotificationsView, m.ctx.View, "should start in NotificationsView")

	// Test that switchSelectedView returns PRsView when in NotificationsView
	newView := m.switchSelectedView()
	require.Equal(t, config.PRsView, newView,
		"switchSelectedView should return PRsView when in NotificationsView")
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
	}

	// Set up a PR notification subject (simulating viewing a PR notification)
	m.notificationView.SetSubjectPR(&prrow.Data{}, "test-notification-id")

	// Verify we start in NotificationsView
	require.Equal(t, config.NotificationsView, m.ctx.View, "should start in NotificationsView")

	// Verify GetSubjectPR returns non-nil
	require.NotNil(t, m.notificationView.GetSubjectPR(), "subject PR should be set")

	// Test that switchSelectedView returns PRsView
	newView := m.switchSelectedView()
	require.Equal(t, config.PRsView, newView,
		"switchSelectedView should return PRsView when in NotificationsView")

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
	}

	// Set up an Issue notification subject (simulating viewing an Issue notification)
	m.notificationView.SetSubjectIssue(&data.IssueData{}, "test-notification-id")

	// Verify we start in NotificationsView
	require.Equal(t, config.NotificationsView, m.ctx.View, "should start in NotificationsView")

	// Verify GetSubjectIssue returns non-nil
	require.NotNil(t, m.notificationView.GetSubjectIssue(), "subject Issue should be set")

	// Test that switchSelectedView returns PRsView
	newView := m.switchSelectedView()
	require.Equal(t, config.PRsView, newView,
		"switchSelectedView should return PRsView when in NotificationsView")

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
