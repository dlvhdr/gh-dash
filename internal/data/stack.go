package data

import (
	"cmp"
	"slices"
)

type StackPullRequest struct {
	Number         int
	Title          string
	State          string
	IsDraft        bool
	IsInMergeQueue bool
	HeadRefName    string
	ReviewDecision string
	Author         struct {
		Login string
	}
	Commits LastCommitStatus `graphql:"commits(last: 1)"`
}

// Position 1 is the entry closest to the base branch. GitHub does not promise
// an order on the entries connection.
type StackEntryNode struct {
	Position    int
	PullRequest StackPullRequest
}

type StackDetails struct {
	Size        int
	BaseRefName string
	Entries     struct {
		Nodes []StackEntryNode
	} `graphql:"entries(first: 50)"`
}

type StackSummary struct {
	Size        int
	BaseRefName string
}

type PrStackEntry struct {
	Position int
	Stack    StackSummary
}

type EnrichedStackEntry struct {
	Position int
	Stack    StackDetails
}

func (e PrStackEntry) IsStacked() bool {
	return e.Position > 0 && e.Stack.Size > 0
}

func (e EnrichedStackEntry) IsStacked() bool {
	return e.Position > 0 && e.Stack.Size > 0
}

func (e EnrichedStackEntry) SortedEntries() []StackEntryNode {
	nodes := slices.Clone(e.Stack.Entries.Nodes)
	slices.SortFunc(nodes, func(a, b StackEntryNode) int {
		return cmp.Compare(a.Position, b.Position)
	})
	return nodes
}

func (e EnrichedStackEntry) Summary() PrStackEntry {
	return PrStackEntry{
		Position: e.Position,
		Stack: StackSummary{
			Size:        e.Stack.Size,
			BaseRefName: e.Stack.BaseRefName,
		},
	}
}
