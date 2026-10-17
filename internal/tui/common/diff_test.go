package common

import (
	"errors"
	"io"
	"testing"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
)

func TestDiffPR(t *testing.T) {
	tests := []struct {
		name     string
		prNumber int
		repoName string
		wantNil  bool
	}{
		{
			name:     "returns command for valid PR",
			prNumber: 123,
			repoName: "owner/repo",
			wantNil:  false,
		},
		{
			name:     "returns command for PR number 0",
			prNumber: 0,
			repoName: "owner/repo",
			wantNil:  false,
		},
		{
			name:     "returns command with hyphenated repo name",
			prNumber: 456,
			repoName: "my-org/my-repo",
			wantNil:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := DiffPR(tt.prNumber, tt.repoName, nil)

			if tt.wantNil && cmd != nil {
				t.Errorf("DiffPR() returned non-nil, want nil")
			}
			if !tt.wantNil && cmd == nil {
				t.Errorf("DiffPR() returned nil, want non-nil")
			}
		})
	}
}

func TestDiffPager(t *testing.T) {
	tests := []struct {
		name string
		env  []string
		want string
	}{
		{
			name: "uses configured pager",
			env:  []string{"GH_PAGER=delta --paging always"},
			want: "delta --paging always",
		},
		{
			name: "last pager wins",
			env:  []string{"GH_PAGER=less", "GH_PAGER=delta --paging always"},
			want: "delta --paging always",
		},
		{
			name: "defaults empty pager",
			env:  []string{"GH_PAGER="},
			want: "less",
		},
		{
			name: "defaults missing pager",
			env:  []string{"LESS=CRX"},
			want: "less",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := diffPager(tt.env); got != tt.want {
				t.Fatalf("diffPager() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDiffPRErrorsWhenDiffFetchFails(t *testing.T) {
	wantErr := errors.New("boom")
	cmd := diffPR(123, "owner/repo", nil, func(int, string) (string, error) {
		return "", wantErr
	})

	msg := cmd()
	errMsg, ok := msg.(constants.ErrMsg)
	if !ok {
		t.Fatalf("diffPR() msg = %T, want constants.ErrMsg", msg)
	}
	if !errors.Is(errMsg.Err, wantErr) {
		t.Fatalf("diffPR() err = %v, want %v", errMsg.Err, wantErr)
	}
}

func TestDiffPRReturnsExecMessageAfterDiffFetch(t *testing.T) {
	cmd := diffPR(123, "owner/repo", nil, func(int, string) (string, error) {
		return "diff --git a/file b/file\n", nil
	})

	msg := cmd()
	if msg == nil {
		t.Fatal("diffPR() msg = nil, want exec message")
	}
}

func TestPagerCommandUsesWindowsShell(t *testing.T) {
	t.Setenv("COMSPEC", "")

	cmd := pagerCommandForGOOS("windows", "delta --paging always", "diff text", []string{"GH_PAGER=delta"})

	if cmd.Path != "cmd" {
		t.Fatalf("pagerCommandForGOOS() path = %q, want cmd", cmd.Path)
	}
	if got, want := cmd.Args, []string{"cmd", "/C", "delta --paging always"}; !equalStrings(got, want) {
		t.Fatalf("pagerCommandForGOOS() args = %#v, want %#v", got, want)
	}
	if got, _ := io.ReadAll(cmd.Stdin); string(got) != "diff text" {
		t.Fatalf("pagerCommandForGOOS() stdin = %q, want diff text", string(got))
	}
}

func TestPagerCommandUsesUnixShell(t *testing.T) {
	t.Setenv("SHELL", "/bin/sh")

	cmd := pagerCommandForGOOS("linux", "less", "diff text", []string{"GH_PAGER=less"})

	if cmd.Path != "/bin/sh" {
		t.Fatalf("pagerCommandForGOOS() path = %q, want /bin/sh", cmd.Path)
	}
	if got, want := cmd.Args, []string{"/bin/sh", "-c", "less"}; !equalStrings(got, want) {
		t.Fatalf("pagerCommandForGOOS() args = %#v, want %#v", got, want)
	}
	if got, _ := io.ReadAll(cmd.Stdin); string(got) != "diff text" {
		t.Fatalf("pagerCommandForGOOS() stdin = %q, want diff text", string(got))
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
