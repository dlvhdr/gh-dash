package common

import (
	"testing"
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
