package data

import (
	"io"
	"net/http"
	"strings"
	"testing"

	gh "github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClearEnrichmentCache(t *testing.T) {
	// Save original state
	originalCachedClient := cachedClient
	defer func() {
		cachedClient = originalCachedClient
	}()

	t.Run("clears nil cache without panic", func(t *testing.T) {
		cachedClient = nil
		require.True(t, IsEnrichmentCacheCleared(), "cache should be cleared initially")

		ClearEnrichmentCache()
		require.True(t, IsEnrichmentCacheCleared(), "cache should remain cleared")
	})

	t.Run("clears non-nil cache", func(t *testing.T) {
		// Simulate having a cached client (we use an empty struct pointer
		// since we can't create a real GraphQL client without credentials)
		cachedClient = &gh.GraphQLClient{}
		require.False(
			t,
			IsEnrichmentCacheCleared(),
			"cache should not be cleared when client is set",
		)

		ClearEnrichmentCache()
		require.True(
			t,
			IsEnrichmentCacheCleared(),
			"cache should be cleared after ClearEnrichmentCache",
		)
	})
}

func TestIsEnrichmentCacheCleared(t *testing.T) {
	// Save original state
	originalCachedClient := cachedClient
	defer func() {
		cachedClient = originalCachedClient
	}()

	t.Run("returns true when cache is nil", func(t *testing.T) {
		cachedClient = nil
		require.True(t, IsEnrichmentCacheCleared())
	})

	t.Run("returns false when cache is set", func(t *testing.T) {
		cachedClient = &gh.GraphQLClient{}
		require.False(t, IsEnrichmentCacheCleared())
	})
}

func TestSetClient(t *testing.T) {
	// Save original state
	originalClient := client
	originalCachedClient := cachedClient
	defer func() {
		client = originalClient
		cachedClient = originalCachedClient
	}()

	t.Run("sets both client and cachedClient", func(t *testing.T) {
		client = nil
		cachedClient = nil

		// SetClient with nil should set both to nil
		SetClient(nil)
		require.Nil(t, client)
		require.True(t, IsEnrichmentCacheCleared())
	})
}

// setMockMergeClient installs a GraphQL client backed by the given handler and
// resets the package-level clients afterwards.
func setMockMergeClient(t *testing.T, handler http.Handler) {
	t.Helper()
	SetClient(newMockGraphQLClient(t, handler))
	t.Cleanup(func() {
		client = nil
		cachedClient = nil
	})
}

// mergePullRequestResponse returns a JSON body for a successful mergePullRequest mutation.
func mergePullRequestResponse(state string) string {
	return `{"data":{"mergePullRequest":{"pullRequest":{"state":"` + state + `"}}}}`
}

// enableAutoMergeResponse returns a JSON body for a successful enablePullRequestAutoMerge mutation.
func enableAutoMergeResponse(hasAutoMerge bool) string {
	autoMergeValue := "null"
	if hasAutoMerge {
		autoMergeValue = `{"enabledAt":"2024-01-01T00:00:00Z"}`
	}
	return `{"data":{"enablePullRequestAutoMerge":{"pullRequest":{"state":"OPEN","autoMergeRequest":` + autoMergeValue + `}}}}`
}

// graphqlErrorResponse returns a JSON body representing a GraphQL-level error.
func graphqlErrorResponse(message string) string {
	return `{"errors":[{"message":"` + message + `","locations":[],"path":[]}]}`
}

// mergeMockHandler returns an http.Handler that serves controlled responses for
// each of the merge-related GraphQL mutations. Each response body is
// keyed by the operation name embedded in the request body by the GraphQL client.
func mergeMockHandler(t *testing.T, responses map[string]string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		bodyStr := string(body)

		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(bodyStr, "mutation MergePullRequest"):
			resp, ok := responses["MergePullRequest"]
			if !ok {
				t.Error("unexpected call to MergePullRequest mutation")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = io.WriteString(w, resp)
		case strings.Contains(bodyStr, "mutation EnablePullRequestAutoMerge"):
			resp, ok := responses["EnablePullRequestAutoMerge"]
			if !ok {
				t.Error("unexpected call to EnablePullRequestAutoMerge mutation")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = io.WriteString(w, resp)
		default:
			t.Errorf("unexpected GraphQL request body: %s", bodyStr)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

// --- mergeMethodForRepo tests ---

func TestMergeMethodForRepo(t *testing.T) {
	tests := []struct {
		name           string
		repo           Repository
		expectedMethod string
		expectError    bool
	}{
		{
			name:           "only merge commits allowed",
			repo:           Repository{AllowMergeCommit: true},
			expectedMethod: "MERGE",
		},
		{
			name:           "only squash merge allowed",
			repo:           Repository{AllowSquashMerge: true},
			expectedMethod: "SQUASH",
		},
		{
			name:           "only rebase merge allowed",
			repo:           Repository{AllowRebaseMerge: true},
			expectedMethod: "REBASE",
		},
		{
			name:           "all three allowed - MERGE wins",
			repo:           Repository{AllowMergeCommit: true, AllowSquashMerge: true, AllowRebaseMerge: true},
			expectedMethod: "MERGE",
		},
		{
			name:           "merge and squash allowed - MERGE wins",
			repo:           Repository{AllowMergeCommit: true, AllowSquashMerge: true},
			expectedMethod: "MERGE",
		},
		{
			name:           "squash and rebase allowed - SQUASH wins",
			repo:           Repository{AllowSquashMerge: true, AllowRebaseMerge: true},
			expectedMethod: "SQUASH",
		},
		{
			name:        "no methods allowed returns error",
			repo:        Repository{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, err := mergeMethodForRepo(tt.repo)
			if tt.expectError {
				require.Error(t, err)
				assert.Empty(t, method)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedMethod, string(method))
			}
		})
	}
}

// --- MergePullRequest tests ---

func TestMergePullRequest_DirectMerge_Clean(t *testing.T) {
	setMockMergeClient(t, mergeMockHandler(t, map[string]string{
		"MergePullRequest": mergePullRequestResponse("MERGED"),
	}))

	repo := Repository{AllowMergeCommit: true}
	status, err := MergePullRequest("PR_node123", "CLEAN", repo)

	require.NoError(t, err)
	assert.Equal(t, "MERGED", status.State)
	assert.False(t, status.IsInMergeQueue)
	assert.False(t, status.HasAutoMerge)
}

func TestMergePullRequest_DirectMerge_Unstable(t *testing.T) {
	setMockMergeClient(t, mergeMockHandler(t, map[string]string{
		"MergePullRequest": mergePullRequestResponse("MERGED"),
	}))

	repo := Repository{AllowSquashMerge: true}
	status, err := MergePullRequest("PR_node456", "UNSTABLE", repo)

	require.NoError(t, err)
	assert.Equal(t, "MERGED", status.State)
	assert.False(t, status.IsInMergeQueue)
	assert.False(t, status.HasAutoMerge)
}

func TestMergePullRequest_Blocked_EnablesAutoMerge(t *testing.T) {
	setMockMergeClient(t, mergeMockHandler(t, map[string]string{
		"EnablePullRequestAutoMerge": enableAutoMergeResponse(true),
	}))

	repo := Repository{AllowMergeCommit: true}
	status, err := MergePullRequest("PR_node789", "BLOCKED", repo)

	require.NoError(t, err)
	assert.Equal(t, "OPEN", status.State)
	assert.True(t, status.HasAutoMerge)
	assert.False(t, status.IsInMergeQueue)
}

func TestMergePullRequest_AutoMergeEnabled(t *testing.T) {
	setMockMergeClient(t, mergeMockHandler(t, map[string]string{
		"EnablePullRequestAutoMerge": enableAutoMergeResponse(true),
	}))

	repo := Repository{AllowMergeCommit: true}
	// Any status other than CLEAN, UNSTABLE, or BLOCKED falls through to auto-merge.
	status, err := MergePullRequest("PR_node999", "BEHIND", repo)

	require.NoError(t, err)
	assert.True(t, status.HasAutoMerge)
	assert.False(t, status.IsInMergeQueue)
}

// TestMergePullRequest_MockDataFlag_UsesLocalServer verifies that when the
// FF_MOCK_DATA environment variable is set, MergePullRequest initialises the
// GraphQL client targeting localhost:3000 rather than calling
// gh.DefaultGraphQLClient(). Because the production code hardcodes the host
// and also replaces http.DefaultTransport's TLS config, we cannot intercept
// the connection via a test server. Instead we confirm the behaviour
// indirectly: with FF_MOCK_DATA set and client == nil, MergePullRequest must
// return an error that references localhost:3000 (connection refused), proving
// it attempted the mock-data host rather than the real GitHub API (which would
// produce a credentials error, not a dial error to localhost).
func TestMergePullRequest_MockDataFlag_UsesLocalServer(t *testing.T) {
	// Ensure client starts nil so MergePullRequest performs lazy init.
	originalClient := client
	client = nil
	t.Cleanup(func() { client = originalClient })

	// Activate the feature flag. No server is listening on localhost:3000,
	// so the call will fail — but with a dial error to localhost:3000, which
	// is what we want to assert.
	t.Setenv("FF_MOCK_DATA", "1")

	repo := Repository{AllowMergeCommit: true}
	_, err := MergePullRequest("PR_mockflag", "CLEAN", repo)

	require.Error(t, err, "MergePullRequest should fail when no mock server is running")
	assert.Contains(t, err.Error(), "localhost:3000",
		"error should reference localhost:3000, confirming the mock-data init path was taken")
}

func TestMergePullRequest_DirectMerge_HasHooks(t *testing.T) {
	setMockMergeClient(t, mergeMockHandler(t, map[string]string{
		"MergePullRequest": mergePullRequestResponse("MERGED"),
	}))

	repo := Repository{AllowMergeCommit: true}
	status, err := MergePullRequest("PR_hooks", "HAS_HOOKS", repo)

	require.NoError(t, err)
	assert.Equal(t, "MERGED", status.State)
	assert.False(t, status.IsInMergeQueue)
	assert.False(t, status.HasAutoMerge)
}

func TestMergePullRequest_Behind_EnablesAutoMerge(t *testing.T) {
	setMockMergeClient(t, mergeMockHandler(t, map[string]string{
		"EnablePullRequestAutoMerge": enableAutoMergeResponse(true),
	}))

	repo := Repository{AllowMergeCommit: true}
	status, err := MergePullRequest("PR_behind", "BEHIND", repo)

	require.NoError(t, err)
	assert.True(t, status.HasAutoMerge)
	assert.False(t, status.IsInMergeQueue)
}
