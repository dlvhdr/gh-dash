package data

import (
	"net/http"
	"testing"

	gh "github.com/cli/go-gh/v2/pkg/api"
	graphqltest "github.com/dlvhdr/gh-dash/v4/internal/testhelpers/graphql"
)

// newMockGraphQLClient creates a GraphQL client backed by the given handler.
// It is a package-local wrapper around graphqltest.NewMockGraphQLClient so
// that internal/data tests can call it without importing the test-helper
// package name explicitly.
func newMockGraphQLClient(t *testing.T, handler http.Handler) *gh.GraphQLClient {
	t.Helper()
	return graphqltest.NewMockGraphQLClient(t, handler)
}
