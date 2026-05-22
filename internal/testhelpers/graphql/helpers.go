// Package graphqltest provides shared test helpers for creating mock GraphQL
// clients backed by in-process HTTP handlers.
package graphqltest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gh "github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/require"
)

// LocalRoundTripper is an http.RoundTripper that routes requests directly to
// an in-process handler, avoiding a real network connection.
type LocalRoundTripper struct {
	Handler http.Handler
}

// RoundTrip implements http.RoundTripper.
func (l LocalRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.Handler.ServeHTTP(w, req)
	return w.Result(), nil
}

// NewMockGraphQLClient creates a *gh.GraphQLClient backed by the given handler.
// It is intended for use in tests that need to intercept GraphQL mutations
// without a real network connection.
func NewMockGraphQLClient(t *testing.T, handler http.Handler) *gh.GraphQLClient {
	t.Helper()
	c, err := gh.NewGraphQLClient(gh.ClientOptions{
		Transport: LocalRoundTripper{Handler: handler},
		Host:      "localhost:3000",
		AuthToken: "fake-token",
	})
	require.NoError(t, err)
	return c
}
