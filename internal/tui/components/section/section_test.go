package section

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
)

func TestGetSearchValue_ManualRepoFilterRemoval(t *testing.T) {
	repo, err := repository.Current()
	if err != nil {
		t.Fatal("failed to resolve current repository:", err)
	}
	repoFilter := fmt.Sprintf("repo:%s/%s", repo.Owner, repo.Name)

	tests := []struct {
		name                      string
		configFilters             string
		searchValue               string
		isFilteredByCurrentRemote bool
		wantContainsRepoFilter    bool
	}{
		{
			name:                      "smart filtering on, repo filter in search value",
			configFilters:             "is:open author:@me",
			searchValue:               repoFilter + " is:open author:@me",
			isFilteredByCurrentRemote: true,
			wantContainsRepoFilter:    true,
		},
		{
			name:                      "smart filtering off via toggle, repo filter not in search value",
			configFilters:             "is:open author:@me",
			searchValue:               "is:open author:@me",
			isFilteredByCurrentRemote: false,
			wantContainsRepoFilter:    false,
		},
		{
			name:                      "user manually removed repo filter from search bar",
			configFilters:             "is:open author:@me",
			searchValue:               "is:open author:@me",
			isFilteredByCurrentRemote: true,
			wantContainsRepoFilter:    false,
		},
		{
			name:                      "user replaced repo filter with a different repo",
			configFilters:             "is:open author:@me",
			searchValue:               "repo:other/repo is:open author:@me",
			isFilteredByCurrentRemote: true,
			wantContainsRepoFilter:    false,
		},
		{
			name:                      "config already has repo filter, search value unchanged",
			configFilters:             repoFilter + " is:open",
			searchValue:               repoFilter + " is:open",
			isFilteredByCurrentRemote: false,
			wantContainsRepoFilter:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := BaseModel{
				Config:                    config.SectionConfig{Filters: tt.configFilters},
				SearchValue:               tt.searchValue,
				IsFilteredByCurrentRemote: tt.isFilteredByCurrentRemote,
			}

			m.SyncSmartFilterWithSearchValue()
			got := m.GetSearchValue()

			containsRepoFilter := false
			for token := range strings.FieldsSeq(got) {
				if token == repoFilter {
					containsRepoFilter = true
					break
				}
			}

			if containsRepoFilter != tt.wantContainsRepoFilter {
				t.Errorf("GetSearchValue() = %q, contains %q = %v, want %v",
					got, repoFilter, containsRepoFilter, tt.wantContainsRepoFilter)
			}
		})
	}
}
