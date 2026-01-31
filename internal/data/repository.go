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
		RequiredStatusCheckContexts  []graphql.String
	}
}

type Repository struct {
	Name                  string
	NameWithOwner         string
	IsArchived            bool
	AllowMergeCommit      bool                  `graphql:"mergeCommitAllowed"`
	AllowSquashMerge      bool                  `graphql:"squashMergeAllowed"`
	AllowRebaseMerge      bool                  `graphql:"rebaseMergeAllowed"`
	BranchProtectionRules BranchProtectionRules `graphql:"branchProtectionRules(first: 1)"`
}
