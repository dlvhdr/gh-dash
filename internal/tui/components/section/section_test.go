package section

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/config"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/components/prompt"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

func currentRepoFilter(t *testing.T) string {
	t.Helper()
	t.Setenv("GH_REPO", "https://github.com/dlvhdr/gh-dash")
	repo, err := repository.Current()
	if err != nil {
		t.Fatal("failed to resolve current repository:", err)
	}
	return fmt.Sprintf("repo:%s/%s", repo.Owner, repo.Name)
}

func TestHasRepoNameInConfiguredFilter(t *testing.T) {
	repoFilter := currentRepoFilter(t)

	tests := []struct {
		name        string
		searchValue string
		want        bool
	}{
		{
			name:        "no repo filter",
			searchValue: "is:open author:@me",
			want:        false,
		},
		{
			name:        "has current repo filter",
			searchValue: repoFilter + " is:open",
			want:        true,
		},
		{
			name:        "has different repo filter",
			searchValue: "repo:other/repo is:open",
			want:        true,
		},
		{
			name:        "empty search value",
			searchValue: "",
			want:        false,
		},
		{
			name:        "repo filter with similar prefix",
			searchValue: repoFilter + "-extra is:open",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := BaseModel{SearchValue: tt.searchValue}
			require.Equal(t, tt.want, m.HasRepoNameInConfiguredFilter())
		})
	}
}

func TestHasCurrentRepoNameInConfiguredFilter(t *testing.T) {
	repoFilter := currentRepoFilter(t)

	tests := []struct {
		name        string
		searchValue string
		want        bool
	}{
		{
			name:        "no repo filter",
			searchValue: "is:open author:@me",
			want:        false,
		},
		{
			name:        "has current repo filter",
			searchValue: repoFilter + " is:open",
			want:        true,
		},
		{
			name:        "has different repo filter",
			searchValue: "repo:other/repo is:open",
			want:        false,
		},
		{
			name:        "current repo filter with extra suffix does not match",
			searchValue: repoFilter + "-extra is:open",
			want:        false,
		},
		{
			name:        "current repo filter alone",
			searchValue: repoFilter,
			want:        true,
		},
		{
			name:        "empty search value",
			searchValue: "",
			want:        false,
		},
		{
			name:        "multiple repo filters including current",
			searchValue: "repo:other/repo " + repoFilter + " is:open",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := BaseModel{SearchValue: tt.searchValue}
			require.Equal(t, tt.want, m.HasCurrentRepoNameInConfiguredFilter())
		})
	}
}

func TestSyncSmartFilterWithSearchValue(t *testing.T) {
	repoFilter := currentRepoFilter(t)

	tests := []struct {
		name        string
		searchValue string
		wantFlag    bool
	}{
		{
			name:        "search contains current repo filter",
			searchValue: repoFilter + " is:open",
			wantFlag:    true,
		},
		{
			name:        "search does not contain current repo filter",
			searchValue: "is:open author:@me",
			wantFlag:    false,
		},
		{
			name:        "search contains different repo filter",
			searchValue: "repo:other/repo is:open",
			wantFlag:    false,
		},
		{
			name:        "similar repo name does not set flag",
			searchValue: repoFilter + "-extra is:open",
			wantFlag:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := BaseModel{
				SearchValue:               tt.searchValue,
				IsFilteredByCurrentRemote: !tt.wantFlag,
			}
			m.SyncSmartFilterWithSearchValue()
			require.Equal(t, tt.wantFlag, m.IsFilteredByCurrentRemote)
		})
	}
}

func TestGetSearchValue(t *testing.T) {
	repoFilter := currentRepoFilter(t)

	tests := []struct {
		name                      string
		searchValue               string
		isFilteredByCurrentRemote bool
		wantContainsRepoFilter    bool
		wantContainsOtherFilters  bool
	}{
		{
			name:                      "smart filter on adds repo filter",
			searchValue:               "is:open author:@me",
			isFilteredByCurrentRemote: true,
			wantContainsRepoFilter:    true,
			wantContainsOtherFilters:  true,
		},
		{
			name:                      "smart filter off does not add repo filter",
			searchValue:               "is:open author:@me",
			isFilteredByCurrentRemote: false,
			wantContainsRepoFilter:    false,
			wantContainsOtherFilters:  true,
		},
		{
			name:                      "smart filter on with repo already present does not duplicate",
			searchValue:               repoFilter + " is:open",
			isFilteredByCurrentRemote: true,
			wantContainsRepoFilter:    true,
			wantContainsOtherFilters:  true,
		},
		{
			name:                      "similar repo name is preserved when smart filter is on",
			searchValue:               repoFilter + "-extra is:open",
			isFilteredByCurrentRemote: true,
			wantContainsRepoFilter:    true,
		},
		{
			name:                      "similar repo name is preserved when smart filter is off",
			searchValue:               repoFilter + "-extra is:open",
			isFilteredByCurrentRemote: false,
			wantContainsRepoFilter:    false,
		},
		{
			name:                      "empty search value with smart filter on",
			searchValue:               "",
			isFilteredByCurrentRemote: true,
			wantContainsRepoFilter:    true,
		},
		{
			name:                      "empty search value with smart filter off",
			searchValue:               "",
			isFilteredByCurrentRemote: false,
			wantContainsRepoFilter:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := BaseModel{
				SearchValue:               tt.searchValue,
				IsFilteredByCurrentRemote: tt.isFilteredByCurrentRemote,
			}

			got := m.GetSearchValue()

			hasExactRepoFilter := false
			for token := range strings.FieldsSeq(got) {
				if token == repoFilter {
					hasExactRepoFilter = true
					break
				}
			}
			require.Equal(t, tt.wantContainsRepoFilter, hasExactRepoFilter,
				"GetSearchValue() = %q, expected repo filter present = %v", got, tt.wantContainsRepoFilter)

			if tt.wantContainsOtherFilters {
				require.Contains(t, got, "is:open")
			}
		})
	}
}

func TestGetSearchValue_SimilarRepoNameNotStripped(t *testing.T) {
	repoFilter := currentRepoFilter(t)
	similarRepo := repoFilter + "-extra"

	m := BaseModel{
		SearchValue:               similarRepo + " is:open",
		IsFilteredByCurrentRemote: false,
	}

	got := m.GetSearchValue()

	require.Contains(t, got, similarRepo,
		"similar repo name should not be stripped from search value")
}

func TestGetSearchValue_ManualRepoFilterRemoval(t *testing.T) {
	repoFilter := currentRepoFilter(t)

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

			require.Equal(t, tt.wantContainsRepoFilter, containsRepoFilter,
				"GetSearchValue() = %q, contains %q = %v, want %v",
				got, repoFilter, containsRepoFilter, tt.wantContainsRepoFilter)
		})
	}
}

func TestGetConfigFiltersWithCurrentRemoteAdded(t *testing.T) {
	repoFilter := currentRepoFilter(t)

	tests := []struct {
		name                   string
		filters                string
		smartFilteringAtLaunch bool
		wantContainsRepoFilter bool
	}{
		{
			name:                   "smart filtering enabled, no repo in config",
			filters:                "is:open author:@me",
			smartFilteringAtLaunch: true,
			wantContainsRepoFilter: true,
		},
		{
			name:                   "smart filtering disabled, no repo in config",
			filters:                "is:open author:@me",
			smartFilteringAtLaunch: false,
			wantContainsRepoFilter: false,
		},
		{
			name:                   "smart filtering enabled, repo already in config",
			filters:                repoFilter + " is:open",
			smartFilteringAtLaunch: true,
			wantContainsRepoFilter: true,
		},
		{
			name:                   "smart filtering enabled, different repo in config",
			filters:                "repo:other/repo is:open",
			smartFilteringAtLaunch: true,
			wantContainsRepoFilter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := NewSectionOptions{
				Config: config.SectionConfig{Filters: tt.filters},
			}
			ctx := &context.ProgramContext{
				Config: &config.Config{
					SmartFilteringAtLaunch: tt.smartFilteringAtLaunch,
				},
			}

			got := options.GetConfigFiltersWithCurrentRemoteAdded(ctx)

			hasRepoFilter := false
			for token := range strings.FieldsSeq(got) {
				if token == repoFilter {
					hasRepoFilter = true
					break
				}
			}

			require.Equal(t, tt.wantContainsRepoFilter, hasRepoFilter,
				"GetConfigFiltersWithCurrentRemoteAdded() = %q, expected repo filter = %v",
				got, tt.wantContainsRepoFilter)

			require.Contains(t, got, "is:open",
				"original filters should be preserved")
		})
	}
}

func TestGetPromptConfirmation(t *testing.T) {
	tests := []struct {
		name         string
		action       string
		view         config.ViewType
		wantNonEmpty bool
	}{
		{
			name:         "done_all in notifications view shows confirmation",
			action:       "done_all",
			view:         config.NotificationsView,
			wantNonEmpty: true,
		},
		{
			name:         "close in PRs view shows confirmation",
			action:       "close",
			view:         config.PRsView,
			wantNonEmpty: true,
		},
		{
			name:         "merge in PRs view shows confirmation",
			action:       "merge",
			view:         config.PRsView,
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &context.ProgramContext{
				View: tt.view,
			}
			m := BaseModel{
				IsPromptConfirmationShown: true,
				PromptConfirmationAction:  tt.action,
				PromptConfirmationBox:     prompt.NewModel(ctx),
			}
			m.Ctx = ctx

			result := m.GetPromptConfirmation()
			if tt.wantNonEmpty {
				require.NotEmpty(t, result, "GetPromptConfirmation() should return non-empty for action %q in view %v", tt.action, tt.view)
			}
		})
	}
}
