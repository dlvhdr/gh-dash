package issuessection

import (
	"github.com/dlvhdr/gh-dash/v4/internal/data"
)

type issueKey struct {
	repo   string
	number int
}

// NestSubIssues orders issues such that child issues appear immediately after their parent issue
// if the parent issue is present in the list. Non-child issues preserve their relative order.
func NestSubIssues(issues []data.IssueData) []data.IssueData {
	if len(issues) <= 1 {
		return issues
	}

	inList := make(map[issueKey]bool, len(issues))
	childrenMap := make(map[issueKey][]data.IssueData)
	var topLevel []data.IssueData

	for _, issue := range issues {
		key := issueKey{
			repo:   issue.GetRepoNameWithOwner(),
			number: issue.Number,
		}
		inList[key] = true
	}

	for _, issue := range issues {
		if issue.HasParent() {
			pKey := issueKey{
				repo:   issue.Parent.Repository.NameWithOwner,
				number: issue.Parent.Number,
			}
			if pKey.repo == "" {
				pKey.repo = issue.GetRepoNameWithOwner()
			}
			if inList[pKey] {
				childrenMap[pKey] = append(childrenMap[pKey], issue)
				continue
			}
		}
		topLevel = append(topLevel, issue)
	}

	result := make([]data.IssueData, 0, len(issues))
	visited := make(map[issueKey]bool, len(issues))

	var appendWithChildren func(item data.IssueData)
	appendWithChildren = func(item data.IssueData) {
		k := issueKey{
			repo:   item.GetRepoNameWithOwner(),
			number: item.Number,
		}
		if visited[k] {
			return
		}
		visited[k] = true
		result = append(result, item)
		for _, child := range childrenMap[k] {
			appendWithChildren(child)
		}
	}

	for _, item := range topLevel {
		appendWithChildren(item)
	}

	// Safety check to handle cycles or unvisited issues
	for _, item := range issues {
		k := issueKey{
			repo:   item.GetRepoNameWithOwner(),
			number: item.Number,
		}
		if !visited[k] {
			result = append(result, item)
		}
	}

	return result
}
