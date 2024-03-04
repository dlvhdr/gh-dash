package common_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/ui/common"
)

var configPaths = map[string]string{
	"user/repo": "/path/to/user/repo",
	"user_2/*":  "/path/to/user_2/*",
}

var configPathsWithDefaultPath = map[string]string{
	"user/repo":    "/path/to/user/repo",
	"user_2/*":     "/path/to/user_2/*",
	"default_path": "/path/to/user/dev",
}

func TestGetRepoLocalPath(t *testing.T) {
	testCases := map[string]struct {
		repo        string
		want        string
		found       bool
		configPaths map[string]string
	}{
		"exact match": {
			repo:        "user/repo",
			want:        "/path/to/user/repo",
			found:       true,
			configPaths: configPaths,
		},
		"exact no match": {
			repo:        "user/other_repo",
			want:        "",
			found:       false,
			configPaths: configPaths,
		},
		"wildcard match": {
			repo:        "user_2/repo123",
			want:        "/path/to/user_2/repo123",
			found:       true,
			configPaths: configPaths,
		},
		"bad path": {
			repo:        "invalid-lookup",
			want:        "",
			found:       false,
			configPaths: configPaths,
		},
		"default path": {
			repo:        "user3/repo",
			want:        "/path/to/user/dev/repo",
			found:       true,
			configPaths: configPathsWithDefaultPath,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, found := common.GetRepoLocalPath(tc.repo, tc.configPaths)
			require.Equal(t, tc.want, got)
			require.Equal(t, tc.found, found)
		})
	}
}
