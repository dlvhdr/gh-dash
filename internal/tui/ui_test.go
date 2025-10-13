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

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/markdown"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/testutils"
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
