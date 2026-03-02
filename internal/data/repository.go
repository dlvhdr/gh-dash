package data

import (
	graphql "github.com/cli/shurcooL-graphql"
)

type BranchProtectionRules struct {
	Nodes []struct {
		RequiredApprovingReviewCount int
		RequiresApprovingReviews     graphql.Boolean
		RequiresCodeOwnerReviews     graphql.Boolean
		RequiresStatusChecks         graphql.Boolean
	}
}

type GitHubLogin string

type Owner struct {
	Login string
}

type Repository struct {
	Name                  string
	Owner                 Owner
	NameWithOwner         string
	IsArchived            bool
	BranchProtectionRules BranchProtectionRules `graphql:"branchProtectionRules(first: 1)"`
}
