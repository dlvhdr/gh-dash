package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDoneStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "gh-dash-donestore-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	baseTime := time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

	t.Run("MarkDone and IsDone with matching timestamp", func(t *testing.T) {
		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test1.json"),
		}

		store.MarkDone("id1", baseTime)
		if !store.IsDone("id1", baseTime) {
			t.Error("Should be done when updatedAt equals doneAt")
		}
	})

	t.Run("IsDone with newer updatedAt returns false", func(t *testing.T) {
		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test2.json"),
		}

		store.MarkDone("id1", baseTime)
		newerTime := baseTime.Add(1 * time.Hour)
		if store.IsDone("id1", newerTime) {
			t.Error("Should NOT be done when notification has new activity (updatedAt > doneAt)")
		}
	})

	t.Run("IsDone with older updatedAt returns true", func(t *testing.T) {
		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test3.json"),
		}

		store.MarkDone("id1", baseTime)
		olderTime := baseTime.Add(-1 * time.Hour)
		if !store.IsDone("id1", olderTime) {
			t.Error("Should be done when updatedAt is older than doneAt")
		}
	})

	t.Run("IsDone for unknown ID returns false", func(t *testing.T) {
		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test4.json"),
		}

		if store.IsDone("unknown", baseTime) {
			t.Error("Should NOT be done for an ID not in the store")
		}
	})

	t.Run("Remove", func(t *testing.T) {
		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "test5.json"),
		}

		store.MarkDone("id1", baseTime)
		store.Remove("id1")
		if store.IsDone("id1", baseTime) {
			t.Error("Should NOT be done after Remove")
		}
	})

	t.Run("persistence round-trip", func(t *testing.T) {
		persistFile := filepath.Join(tempDir, "persist.json")

		store1 := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: persistFile,
		}
		store1.MarkDone("id1", baseTime)
		store1.MarkDone("id2", baseTime.Add(30*time.Minute))
		if err := store1.Flush(); err != nil {
			t.Fatalf("Flush failed: %v", err)
		}

		store2 := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: persistFile,
		}
		if err := store2.load(); err != nil {
			t.Fatalf("load failed: %v", err)
		}

		if !store2.IsDone("id1", baseTime) {
			t.Error("Loaded store should have id1 as done")
		}
		if !store2.IsDone("id2", baseTime.Add(30*time.Minute)) {
			t.Error("Loaded store should have id2 as done")
		}
		// New activity should still resurface
		if store2.IsDone("id1", baseTime.Add(1*time.Hour)) {
			t.Error("Loaded store should detect new activity for id1")
		}
	})

	t.Run("backward compatibility: load legacy format", func(t *testing.T) {
		legacyFile := filepath.Join(tempDir, "legacy.json")
		legacyData, _ := json.Marshal([]string{"id1", "id2", "id3"})
		if err := os.WriteFile(legacyFile, legacyData, 0o644); err != nil {
			t.Fatalf("Failed to write legacy file: %v", err)
		}

		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: legacyFile,
		}
		if err := store.load(); err != nil {
			t.Fatalf("load failed on legacy format: %v", err)
		}

		// Legacy entries get zero time. Any real updatedAt is after zero,
		// so all legacy entries resurface. This is a one-time cost on
		// upgrade; once re-marked as done they get proper timestamps.
		pastTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		if store.IsDone("id1", pastTime) {
			t.Error("Legacy ID should resurface (zero time is before any real updatedAt)")
		}
		if store.IsDone("id2", pastTime) {
			t.Error("Legacy ID should resurface")
		}
		if store.IsDone("id3", pastTime) {
			t.Error("Legacy ID should resurface")
		}

		// Unknown IDs still return false
		if store.IsDone("id4", pastTime) {
			t.Error("Unknown ID should not be done")
		}
	})

	t.Run("load from non-existent file", func(t *testing.T) {
		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "nonexistent.json"),
		}
		if err := store.load(); err != nil {
			t.Errorf("load from non-existent file should not error, got: %v", err)
		}
	})

	t.Run("save then load preserves RFC3339 format", func(t *testing.T) {
		file := filepath.Join(tempDir, "format.json")
		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: file,
		}
		store.MarkDone("id1", baseTime)
		if err := store.Flush(); err != nil {
			t.Fatalf("Flush failed: %v", err)
		}

		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}

		var tsMap map[string]string
		if err := json.Unmarshal(raw, &tsMap); err != nil {
			t.Fatalf("Should be a JSON object: %v", err)
		}
		if _, ok := tsMap["id1"]; !ok {
			t.Error("Should have id1 key")
		}
		if _, err := time.Parse(time.RFC3339, tsMap["id1"]); err != nil {
			t.Errorf("Timestamp should be RFC3339: %v", err)
		}
	})

	t.Run("legacy entries upgrade to new format on save", func(t *testing.T) {
		legacyFile := filepath.Join(tempDir, "legacy-upgrade.json")
		legacyData, _ := json.Marshal([]string{"id1", "id2"})
		if err := os.WriteFile(legacyFile, legacyData, 0o644); err != nil {
			t.Fatalf("Failed to write legacy file: %v", err)
		}

		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: legacyFile,
		}
		if err := store.load(); err != nil {
			t.Fatalf("load failed: %v", err)
		}
		if err := store.Flush(); err != nil {
			t.Fatalf("Flush failed: %v", err)
		}

		// Re-read file: should now be a JSON object, not an array
		raw, err := os.ReadFile(legacyFile)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		var tsMap map[string]string
		if err := json.Unmarshal(raw, &tsMap); err != nil {
			t.Fatalf("Saved legacy file should be new format (JSON object): %v", err)
		}
		if len(tsMap) != 2 {
			t.Errorf("Expected 2 entries, got %d", len(tsMap))
		}
	})

	t.Run("re-mark legacy entry as done with real timestamp", func(t *testing.T) {
		legacyFile := filepath.Join(tempDir, "legacy-remark.json")
		legacyData, _ := json.Marshal([]string{"id1"})
		if err := os.WriteFile(legacyFile, legacyData, 0o644); err != nil {
			t.Fatalf("Failed to write legacy file: %v", err)
		}

		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: legacyFile,
		}
		if err := store.load(); err != nil {
			t.Fatalf("load failed: %v", err)
		}

		// Legacy entry resurfaces (zero time)
		if store.IsDone("id1", baseTime) {
			t.Error("Legacy entry should resurface initially")
		}

		// User re-marks it as done with a real timestamp
		store.MarkDone("id1", baseTime)
		if !store.IsDone("id1", baseTime) {
			t.Error("Should be done after re-marking with real timestamp")
		}
		// New activity after re-mark should still resurface
		if store.IsDone("id1", baseTime.Add(1*time.Hour)) {
			t.Error("Should resurface on new activity after re-mark")
		}
	})

	t.Run("MarkDone overwrites timestamp on same ID", func(t *testing.T) {
		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: filepath.Join(tempDir, "overwrite.json"),
		}

		t1 := baseTime
		t2 := baseTime.Add(2 * time.Hour)

		store.MarkDone("id1", t1)
		// Activity at t1+1h would resurface
		if store.IsDone("id1", t1.Add(1*time.Hour)) {
			t.Error("Should not be done for activity after first mark")
		}

		// Re-mark done at t2 (user saw the new activity and dismissed it)
		store.MarkDone("id1", t2)
		// Now activity at t1+1h is before t2, so it should be done
		if !store.IsDone("id1", t1.Add(1*time.Hour)) {
			t.Error("Should be done after re-marking at later timestamp")
		}
		// Activity after t2 still resurfaces
		if store.IsDone("id1", t2.Add(1*time.Minute)) {
			t.Error("Should resurface for activity after second mark")
		}
	})

	t.Run("load from empty file", func(t *testing.T) {
		emptyFile := filepath.Join(tempDir, "empty.json")
		if err := os.WriteFile(emptyFile, []byte(""), 0o644); err != nil {
			t.Fatalf("Failed to create empty file: %v", err)
		}

		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: emptyFile,
		}
		err := store.load()
		if err == nil {
			t.Error("load from empty file should return an error (invalid JSON)")
		}
	})

	t.Run("load from corrupted JSON", func(t *testing.T) {
		corruptedFile := filepath.Join(tempDir, "corrupted.json")
		if err := os.WriteFile(corruptedFile, []byte("{invalid json"), 0o644); err != nil {
			t.Fatalf("Failed to create corrupted file: %v", err)
		}

		store := &DoneStore{
			entries:  make(map[string]time.Time),
			filePath: corruptedFile,
		}
		err := store.load()
		if err == nil {
			t.Error("load from corrupted JSON should return an error")
		}
	})
}
