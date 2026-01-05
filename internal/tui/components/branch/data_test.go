package branch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/git"
)

func TestBranchData_GetRepoNameWithOwner_EmptyRemotes(t *testing.T) {
	// This test verifies that GetRepoNameWithOwner returns empty string
	// when Remotes slice is empty, instead of panicking with index out of bounds.
	b := BranchData{
		Data: git.Branch{
			Remotes: []string{}, // Empty remotes
		},
	}

	result := b.GetRepoNameWithOwner()

	require.Equal(t, "", result, "should return empty string for empty remotes")
}

func TestBranchData_GetRepoNameWithOwner_WithRemotes(t *testing.T) {
	b := BranchData{
		Data: git.Branch{
			Remotes: []string{"origin/main", "upstream/main"},
		},
	}

	result := b.GetRepoNameWithOwner()

	require.Equal(t, "origin/main", result, "should return first remote")
}

func TestBranchData_GetUpdatedAt_NilLastUpdatedAt(t *testing.T) {
	// This test verifies that GetUpdatedAt returns zero time
	// when LastUpdatedAt is nil, instead of panicking with nil pointer dereference.
	b := BranchData{
		Data: git.Branch{
			LastUpdatedAt: nil, // Nil pointer
		},
	}

	result := b.GetUpdatedAt()

	require.True(t, result.IsZero(), "should return zero time for nil LastUpdatedAt")
}

func TestBranchData_GetUpdatedAt_WithValue(t *testing.T) {
	now := time.Now()
	b := BranchData{
		Data: git.Branch{
			LastUpdatedAt: &now,
		},
	}

	result := b.GetUpdatedAt()

	require.Equal(t, now, result, "should return the actual time")
}
