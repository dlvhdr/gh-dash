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
	"time"

	tea "github.com/charmbracelet/bubbletea"
	log "github.com/charmbracelet/log"
	"github.com/charmbracelet/x/exp/teatest"
	gh "github.com/cli/go-gh/v2/pkg/api"
	zone "github.com/lrstanley/bubblezone"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestFullOutput(t *testing.T) {
	if _, debug := os.LookupEnv("DEBUG"); debug {
		f, _ := os.CreateTemp("", "gh-dash-debug.log")
		fmt.Printf("[DEBU] writing debug logs to %s\n", f.Name())
		defer f.Close()
		log.SetOutput(f)
		log.SetLevel(log.DebugLevel)
	}
	setMockClient(t)

	zone.NewGlobal()
	zone.SetEnabled(false)
	m := NewModel(config.Location{RepoPath: "", ConfigFlag: "../config/testdata/test-config.yml"})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(160, 60))

	waitForText(t, tm, "Reading config...")
	waitForText(t, tm, "style: make assignment brief", teatest.WithDuration(6*time.Second))

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("q"),
	})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
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
		body := mustRead(t, req.Body)
		switch {
		case strings.Contains(body, "query SearchPullRequests"):
			d, err := os.ReadFile("./testdata/searchPullRequests.json")
			if err != nil {
				t.Errorf("failed reading mock data file %v", err)
			}
			mustWrite(t, w, string(d))
		default:
			t.Errorf("unexpected request with body %s", req.Body)
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

func bytesContains(t *testing.T, bts []byte, str string) bool {
	t.Helper()
	return bytes.Contains(bts, []byte(str))
}

func waitForText(t *testing.T, tm *teatest.TestModel, text string, options ...teatest.WaitForOption) {
	teatest.WaitFor(t,
		tm.Output(),
		func(bts []byte) bool {
			contains := bytesContains(t, bts, text)
			if _, debug := os.LookupEnv("DEBUG"); debug {
				if contains {
					f, _ := os.CreateTemp("", "gh-dash-debug")
					defer f.Close()
					fmt.Fprintf(f, "%s", string(bts))
					log.Debug("✅ wrote to file while looking for text", "file", f.Name(), "text", text)
				} else {
					f, _ := os.CreateTemp("", "not-found-gh-dash-debug")
					defer f.Close()
					fmt.Fprintf(f, "%s", string(bts))
					log.Debug("❌ text not found", "file", f.Name(), "text", text)
				}
			}
			return contains
		},
		options...,
	)
}
