package issuessection

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

func TestNestSubIssues(t *testing.T) {
	makeIssue := func(number int, parentNumber *int, repo string) data.IssueData {
		issue := data.IssueData{
			Number: number,
		}
		issue.Repository.NameWithOwner = repo
		if parentNumber != nil {
			issue.Parent = &data.ParentIssue{
				Number: *parentNumber,
			}
			issue.Parent.Repository.NameWithOwner = repo
		}
		return issue
	}

	t.Run("returns empty slice when given empty slice", func(t *testing.T) {
		res := NestSubIssues([]data.IssueData{})
		require.Empty(t, res)
	})

	t.Run("returns single issue unchanged", func(t *testing.T) {
		issue := makeIssue(1, nil, "owner/repo")
		res := NestSubIssues([]data.IssueData{issue})
		require.Len(t, res, 1)
		require.Equal(t, 1, res[0].Number)
	})

	t.Run("preserves order when no issues have parent-child relations", func(t *testing.T) {
		i1 := makeIssue(1, nil, "owner/repo")
		i2 := makeIssue(2, nil, "owner/repo")
		i3 := makeIssue(3, nil, "owner/repo")

		res := NestSubIssues([]data.IssueData{i1, i2, i3})
		require.Equal(t, []int{1, 2, 3}, []int{res[0].Number, res[1].Number, res[2].Number})
	})

	t.Run("nests child directly below its parent issue", func(t *testing.T) {
		p10 := 10
		parent := makeIssue(10, nil, "owner/repo")
		unrelated := makeIssue(20, nil, "owner/repo")
		child := makeIssue(11, &p10, "owner/repo")

		// Input: parent, unrelated, child
		res := NestSubIssues([]data.IssueData{parent, unrelated, child})
		// Expected: parent, child, unrelated
		require.Equal(t, []int{10, 11, 20}, []int{res[0].Number, res[1].Number, res[2].Number})
	})

	t.Run("nests multiple children under their respective parents", func(t *testing.T) {
		p10 := 10
		p30 := 30
		p1 := makeIssue(10, nil, "owner/repo")
		p2 := makeIssue(30, nil, "owner/repo")
		c1a := makeIssue(11, &p10, "owner/repo")
		c1b := makeIssue(12, &p10, "owner/repo")
		c2a := makeIssue(31, &p30, "owner/repo")

		res := NestSubIssues([]data.IssueData{c1a, p1, p2, c2a, c1b})
		expected := []int{10, 11, 12, 30, 31}
		actual := make([]int, len(res))
		for i, issue := range res {
			actual[i] = issue.Number
		}
		require.Equal(t, expected, actual)
	})

	t.Run("handles multi-level hierarchy", func(t *testing.T) {
		p1 := 1
		p2 := 2
		root := makeIssue(1, nil, "owner/repo")
		child := makeIssue(2, &p1, "owner/repo")
		grandchild := makeIssue(3, &p2, "owner/repo")
		unrelated := makeIssue(4, nil, "owner/repo")

		res := NestSubIssues([]data.IssueData{unrelated, grandchild, root, child})
		expected := []int{4, 1, 2, 3}
		actual := make([]int, len(res))
		for i, issue := range res {
			actual[i] = issue.Number
		}
		require.Equal(t, expected, actual)
	})

	t.Run("keeps child issue when parent is not present in the list", func(t *testing.T) {
		missingParent := 999
		child := makeIssue(5, &missingParent, "owner/repo")
		normal := makeIssue(6, nil, "owner/repo")

		res := NestSubIssues([]data.IssueData{child, normal})
		require.Equal(t, []int{5, 6}, []int{res[0].Number, res[1].Number})
	})

	t.Run("distinguishes same issue numbers in different repositories", func(t *testing.T) {
		p1 := 1
		repoA := "org/repo-a"
		repoB := "org/repo-b"

		parentA := makeIssue(1, nil, repoA)
		childA := makeIssue(2, &p1, repoA)
		parentB := makeIssue(1, nil, repoB)
		childB := makeIssue(2, &p1, repoB)

		res := NestSubIssues([]data.IssueData{parentA, parentB, childA, childB})
		require.Equal(t, repoA, res[0].GetRepoNameWithOwner())
		require.Equal(t, 1, res[0].Number)
		require.Equal(t, repoA, res[1].GetRepoNameWithOwner())
		require.Equal(t, 2, res[1].Number)
		require.Equal(t, repoB, res[2].GetRepoNameWithOwner())
		require.Equal(t, 1, res[2].Number)
		require.Equal(t, repoB, res[3].GetRepoNameWithOwner())
		require.Equal(t, 2, res[3].Number)
	})

	t.Run("handles cyclic references safely without dropping items", func(t *testing.T) {
		p1 := 1
		p2 := 2
		i1 := makeIssue(1, &p2, "owner/repo")
		i2 := makeIssue(2, &p1, "owner/repo")

		res := NestSubIssues([]data.IssueData{i1, i2})
		require.Len(t, res, 2)
	})
}
