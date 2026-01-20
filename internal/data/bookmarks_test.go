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

		// Create store, add IDs, flush to ensure save completes
		store1 := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: persistFile,
			name:     "test",
		}
		store1.Add("persist1")
		store1.Add("persist2")
		if err := store1.Flush(); err != nil {
			t.Fatalf("Failed to flush: %v", err)
		}

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

	t.Run("load from non-existent file", func(t *testing.T) {
		nonExistentFile := filepath.Join(tempDir, "non-existent.json")
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: nonExistentFile,
			name:     "test",
		}

		// load should not error for non-existent file, just return empty store
		err := store.load()
		if err != nil {
			t.Errorf("load() from non-existent file should not error, got: %v", err)
		}
		if len(store.GetAll()) != 0 {
			t.Error("store should be empty after loading from non-existent file")
		}
	})

	t.Run("load from empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty.json")
		if err := os.WriteFile(emptyFile, []byte(""), 0o644); err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: emptyFile,
			name:     "test",
		}

		err := store.load()
		// Empty file should either be handled gracefully or return an error
		// The important thing is it doesn't panic
		if err != nil {
			// This is acceptable - empty file is not valid JSON
			t.Logf("load() from empty file returned error (expected): %v", err)
		}
	})

	t.Run("load from corrupted JSON", func(t *testing.T) {
		corruptedFile := filepath.Join(tempDir, "corrupted.json")
		if err := os.WriteFile(corruptedFile, []byte("{invalid json"), 0o644); err != nil {
			t.Fatalf("Failed to create corrupted file: %v", err)
		}

		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: corruptedFile,
			name:     "test",
		}

		err := store.load()
		if err == nil {
			t.Error("load() from corrupted JSON should return error")
		}
	})

	t.Run("Add same ID multiple times", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: testFile,
			name:     "test",
		}

		store.Add("id1")
		store.Add("id1")
		store.Add("id1")

		// Should still only have one ID
		all := store.GetAll()
		if len(all) != 1 {
			t.Errorf("Adding same ID multiple times should result in 1 entry, got %d", len(all))
		}
	})

	t.Run("Remove non-existent ID", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: testFile,
			name:     "test",
		}

		// Should not panic or error
		store.Remove("non-existent")
		if store.Has("non-existent") {
			t.Error("Has should return false for non-existent ID")
		}
	})

	t.Run("empty string ID", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: testFile,
			name:     "test",
		}

		store.Add("")
		// Empty string should be handled (though it's an edge case)
		if !store.Has("") {
			t.Error("Store should have empty string ID after Add")
		}
	})

	t.Run("special characters in ID", func(t *testing.T) {
		store := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: filepath.Join(tempDir, "special-chars.json"),
			name:     "test",
		}

		specialIds := []string{
			"id-with-dash",
			"id_with_underscore",
			"id.with.dots",
			"id:with:colons",
			"id/with/slashes",
			"12345678901234567890", // long numeric ID
		}

		for _, id := range specialIds {
			store.Add(id)
		}

		// Verify all were added
		for _, id := range specialIds {
			if !store.Has(id) {
				t.Errorf("Store should have ID %q", id)
			}
		}

		// Flush to ensure async save completes before testing persistence
		if err := store.Flush(); err != nil {
			t.Fatalf("Failed to flush: %v", err)
		}

		// Test persistence with special characters
		store2 := &NotificationIDStore{
			ids:      make(map[string]bool),
			filePath: filepath.Join(tempDir, "special-chars.json"),
			name:     "test",
		}
		if err := store2.load(); err != nil {
			t.Fatalf("Failed to load: %v", err)
		}

		for _, id := range specialIds {
			if !store2.Has(id) {
				t.Errorf("Loaded store should have ID %q", id)
			}
		}
	})
}
