package data

import "testing"

func TestSplitDiffRepoName(t *testing.T) {
	tests := []struct {
		name     string
		repoName string
		wantHost string
		wantRepo string
		wantErr  bool
	}{
		{
			name:     "github repo",
			repoName: "owner/repo",
			wantRepo: "owner/repo",
		},
		{
			name:     "enterprise repo",
			repoName: "ghe.example.com/owner/repo",
			wantHost: "ghe.example.com",
			wantRepo: "owner/repo",
		},
		{
			name:     "invalid repo",
			repoName: "owner",
			wantErr:  true,
		},
		{
			name:     "too many parts",
			repoName: "one/two/three/four",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHost, gotRepo, err := splitDiffRepoName(tt.repoName)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("splitDiffRepoName() err = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("splitDiffRepoName() err = %v", err)
			}
			if gotHost != tt.wantHost {
				t.Fatalf("splitDiffRepoName() host = %q, want %q", gotHost, tt.wantHost)
			}
			if gotRepo != tt.wantRepo {
				t.Fatalf("splitDiffRepoName() repo = %q, want %q", gotRepo, tt.wantRepo)
			}
		})
	}
}
