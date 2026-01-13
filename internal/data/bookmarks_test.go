package data

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNotificationIDStore(t *testing.T) {
	// Create a temp directory for test files
	tempDir, err := os.MkdirTemp("", "gh-dash-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test-store.json")

	t.Run("new store is empty", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: testFile,
			name:     "test",
		}

		if store.Has("id1") {
			t.Error("New store should not have any IDs")
		}
		if len(store.GetAll()) != 0 {
			t.Error("New store should return empty list")
		}
	})

	t.Run("Add and Has", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: testFile,
			name:     "test",
		}

		store.Add("id1")
		if !store.Has("id1") {
			t.Error("Store should have id1 after Add")
		}
		if store.Has("id2") {
			t.Error("Store should not have id2")
		}
	})

	t.Run("Remove", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: testFile,
			name:     "test",
		}

		store.Add("id1")
		store.Remove("id1")
		if store.Has("id1") {
			t.Error("Store should not have id1 after Remove")
		}
	})

	t.Run("Toggle", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: testFile,
			name:     "test",
		}

		// Toggle on
		if !store.Toggle("id1") {
			t.Error("First toggle should return true")
		}
		if !store.Has("id1") {
			t.Error("Store should have id1 after toggle on")
		}

		// Toggle off
		if store.Toggle("id1") {
			t.Error("Second toggle should return false")
		}
		if store.Has("id1") {
			t.Error("Store should not have id1 after toggle off")
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: testFile,
			name:     "test",
		}

		store.Add("id1")
		store.Add("id2")
		store.Add("id3")

		all := store.GetAll()
		if len(all) != 3 {
			t.Errorf("GetAll should return 3 IDs, got %d", len(all))
		}

		// Check all IDs are present (order not guaranteed)
		found := make(map[string]bool)
		for _, id := range all {
			found[id] = true
		}
		for _, id := range []string{"id1", "id2", "id3"} {
			if !found[id] {
				t.Errorf("GetAll should include %s", id)
			}
		}
	})

	t.Run("persistence", func(t *testing.T) {
		persistFile := filepath.Join(tempDir, "persist-test.json")

		// Create store, add IDs, let it save
		store1 := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: persistFile,
			name:     "test",
		}
		store1.Add("persist1")
		store1.Add("persist2")

		// Create new store, load from same file
		store2 := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: persistFile,
			name:     "test",
		}
		if err := store2.load(); err != nil {
			t.Fatalf("Failed to load: %v", err)
		}

		if !store2.Has("persist1") {
			t.Error("Loaded store should have persist1")
		}
		if !store2.Has("persist2") {
			t.Error("Loaded store should have persist2")
		}
	})

	t.Run("convenience methods", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: testFile,
			name:     "test",
		}

		// Bookmark convenience methods
		store.Add("bookmark1")
		if !store.IsBookmarked("bookmark1") {
			t.Error("IsBookmarked should return true")
		}
		if store.ToggleBookmark("bookmark1") {
			t.Error("ToggleBookmark should return false (toggled off)")
		}
		if len(store.GetBookmarkedIds()) != 0 {
			t.Error("GetBookmarkedIds should be empty after toggle off")
		}

		// Done convenience methods
		store.MarkDone("done1")
		if !store.IsDone("done1") {
			t.Error("IsDone should return true after MarkDone")
		}
	})
}
