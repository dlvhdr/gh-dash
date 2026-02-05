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
