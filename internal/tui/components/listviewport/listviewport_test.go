package listviewport

import (
	"testing"
	"time"

	"github.com/dlvhdr/gh-dash/v4/internal/tui/constants"
	"github.com/dlvhdr/gh-dash/v4/internal/tui/context"
)

type testModelOpts struct {
	numItems       int
	viewportHeight int
	itemHeight     int
}

// numItems, viewportHeight, itemHeight int
func newTestModel(opts testModelOpts) Model {
	ctx := context.ProgramContext{}
	dims := constants.Dimensions{Width: 80, Height: opts.viewportHeight}
	now := time.Now()
	return NewModel(ctx, dims, now, now, "items", opts.numItems, opts.itemHeight)
}

func TestPrevItemScrollsAtTopBound(t *testing.T) {
	// 10 items, viewport height 5, item height 1 => 5 items per page
	// Initial bounds: topBoundId=0, bottomBoundId=4
	m := newTestModel(testModelOpts{numItems: 10, viewportHeight: 5, itemHeight: 1})

	// Navigate down past the first page boundary
	// After 5 NextItem calls: currId=5, topBoundId=1, bottomBoundId=5
	for range 5 {
		m.NextItem()
	}

	if m.GetCurrItem() != 5 {
		t.Fatalf("expected currId=5 after 5 NextItem calls, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 1 {
		t.Fatalf("expected topBoundId=1, got %d", m.topBoundId)
	}

	// Navigate back up to the top bound
	// 4 PrevItem calls: currId goes 4, 3, 2, 1
	for range 4 {
		m.PrevItem()
	}

	if m.GetCurrItem() != 1 {
		t.Fatalf("expected currId=1, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 1 {
		t.Fatalf("expected topBoundId=1 (no scroll yet), got %d", m.topBoundId)
	}

	// Now one more PrevItem: currId should go to 0, and the viewport
	// should scroll up so that item 0 is visible.
	// Before fix: topBoundId stays at 1, item 0 is above the viewport.
	// After fix: topBoundId becomes 0, item 0 is at the top of the viewport.
	m.PrevItem()

	if m.GetCurrItem() != 0 {
		t.Fatalf("expected currId=0, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 0 {
		t.Errorf(
			"PrevItem did not scroll up at the top boundary: expected topBoundId=0, got %d (item 0 is above the visible viewport)",
			m.topBoundId,
		)
	}
}

func TestNextItemScrollsAtBottomBound(t *testing.T) {
	// Verify NextItem scrolling works symmetrically (this should pass already)
	m := newTestModel(testModelOpts{numItems: 10, viewportHeight: 5, itemHeight: 1})

	// Navigate down to bottomBoundId (4)
	for range 4 {
		m.NextItem()
	}

	if m.GetCurrItem() != 4 {
		t.Fatalf("expected currId=4, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 0 {
		t.Fatalf("expected topBoundId=0 (no scroll yet), got %d", m.topBoundId)
	}

	// One more NextItem at the bottom boundary should scroll
	m.NextItem()

	if m.GetCurrItem() != 5 {
		t.Fatalf("expected currId=5, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 1 {
		t.Errorf(
			"NextItem did not scroll down at the bottom boundary: expected topBoundId=1, got %d",
			m.topBoundId,
		)
	}
}

func TestPrevItemAtFirstItem(t *testing.T) {
	// Pressing up when already at the first item should stay at 0
	// and should not corrupt viewport bounds.
	m := newTestModel(testModelOpts{numItems: 10, viewportHeight: 5, itemHeight: 1})

	m.PrevItem()

	if m.GetCurrItem() != 0 {
		t.Errorf("expected currId=0, got %d", m.GetCurrItem())
	}
	if m.topBoundId != 0 {
		t.Errorf("expected topBoundId=0 (should not scroll), got %d", m.topBoundId)
	}
	if m.bottomBoundId != 4 {
		t.Errorf("expected bottomBoundId=4 (should not scroll), got %d", m.bottomBoundId)
	}
}

func TestNextItemAtLastItem(t *testing.T) {
	// Pressing down when already at the last item should stay at last
	m := newTestModel(testModelOpts{numItems: 10, viewportHeight: 5, itemHeight: 1})

	for range 20 {
		m.NextItem()
	}

	if m.GetCurrItem() != 9 {
		t.Errorf("expected currId=9, got %d", m.GetCurrItem())
	}
}
