package data

import (
	"encoding/json"
	"testing"
)

const enrichedStackJSON = `{
  "position": 2,
  "stack": {
    "number": 7,
    "size": 3,
    "baseRefName": "main",
    "entries": {
      "nodes": [
        {"position": 3, "pullRequest": {"number": 103, "title": "third", "state": "OPEN", "isDraft": true, "headRefName": "feat-3", "baseRefName": "feat-2"}},
        {"position": 1, "pullRequest": {"number": 101, "title": "first", "state": "MERGED", "isDraft": false, "headRefName": "feat-1", "baseRefName": "main"}},
        {"position": 2, "pullRequest": {"number": 102, "title": "second", "state": "OPEN", "isDraft": false, "headRefName": "feat-2", "baseRefName": "feat-1"}}
      ]
    }
  }
}`

func mustEnriched(t *testing.T) EnrichedStackEntry {
	t.Helper()
	var e EnrichedStackEntry
	if err := json.Unmarshal([]byte(enrichedStackJSON), &e); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return e
}

func TestEnrichedStackEntryDecodes(t *testing.T) {
	e := mustEnriched(t)

	if !e.IsStacked() {
		t.Fatal("expected entry to be stacked")
	}
	if e.Position != 2 {
		t.Errorf("Position = %d, want 2", e.Position)
	}
	if e.Stack.Size != 3 {
		t.Errorf("Stack.Size = %d, want 3", e.Stack.Size)
	}
	if e.Stack.BaseRefName != "main" {
		t.Errorf("Stack.BaseRefName = %q, want main", e.Stack.BaseRefName)
	}
	if got := len(e.Stack.Entries.Nodes); got != 3 {
		t.Fatalf("entries = %d, want 3", got)
	}
}

func TestSortedEntriesOrdersBottomFirst(t *testing.T) {
	e := mustEnriched(t)

	want := []int{101, 102, 103}
	sorted := e.SortedEntries()
	if len(sorted) != len(want) {
		t.Fatalf("len = %d, want %d", len(sorted), len(want))
	}
	for i, node := range sorted {
		if node.Position != i+1 {
			t.Errorf("entry %d has position %d, want %d", i, node.Position, i+1)
		}
		if node.PullRequest.Number != want[i] {
			t.Errorf("entry %d is #%d, want #%d", i, node.PullRequest.Number, want[i])
		}
	}
}

func TestSortedEntriesDoesNotMutateResponse(t *testing.T) {
	e := mustEnriched(t)
	first := e.Stack.Entries.Nodes[0].PullRequest.Number

	e.SortedEntries()

	if got := e.Stack.Entries.Nodes[0].PullRequest.Number; got != first {
		t.Errorf(
			"SortedEntries mutated the underlying nodes: first is now #%d, was #%d",
			got,
			first,
		)
	}
}

func TestUnstackedPullRequestDecodesToZero(t *testing.T) {
	var e EnrichedStackEntry
	if err := json.Unmarshal([]byte(`null`), &e); err != nil {
		t.Fatalf("unmarshal null: %v", err)
	}
	if e.IsStacked() {
		t.Error("null stackEntry should not report as stacked")
	}
	if got := len(e.SortedEntries()); got != 0 {
		t.Errorf("SortedEntries on null entry = %d, want 0", got)
	}
}

func TestSummaryNarrowsEnrichedEntry(t *testing.T) {
	s := mustEnriched(t).Summary()

	if !s.IsStacked() {
		t.Error("summary of a stacked entry should report as stacked")
	}
	if s.Position != 2 || s.Stack.Size != 3 || s.Stack.BaseRefName != "main" {
		t.Errorf("summary = %+v, want position 2 of a size-3 stack based on main", s)
	}
}
