package common_test

import (
	"testing"

	"github.com/dlvhdr/gh-dash/ui/common"
)

var (
	configPaths = map[string]string{
		"user/repo": "/path/to/user/repo",
		"user_2/*":  "/path/to/user_2/*",
	}
)

func TestGetRepoLocalPathExactMatch(t *testing.T) {
	want := "/path/to/user/repo"
	got := common.GetRepoLocalPath("user/repo", configPaths)
	if want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestGetRepoLocalPathExactNoMatch(t *testing.T) {
	want := ""
	got := common.GetRepoLocalPath("user/other_repo", configPaths)
	if want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestGetRepoLocalPathWildcardMatch(t *testing.T) {
	want := "/path/to/user_2/repo123"
	got := common.GetRepoLocalPath("user_2/repo123", configPaths)
	if want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}

func TestGetRepoLocalPathBadPath(t *testing.T) {
	want := ""
	got := common.GetRepoLocalPath("invalid-lookup", configPaths)
	if want != got {
		t.Errorf("want '%s', got '%s'", want, got)
	}
}
