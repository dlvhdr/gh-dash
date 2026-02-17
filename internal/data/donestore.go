package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charmbracelet/log"
)

// DoneStore persists notification IDs along with the timestamp at which they
// were marked done. When checking whether a notification is still "done" we
// compare the stored timestamp against the notification's current updated_at:
// if the notification has been updated since it was marked done, it resurfaces.
type DoneStore struct {
	mu       sync.RWMutex
	entries  map[string]time.Time // id -> updatedAt when marked done
	filePath string
}

func newDoneStore(filename string) *DoneStore {
	store := &DoneStore{
		entries: make(map[string]time.Time),
	}
	filePath, err := getStateFilePath(filename)
	if err != nil {
		log.Error("Failed to get state file path for done notifications", "err", err)
	}
	store.filePath = filePath
	if err := store.load(); err != nil {
		log.Error("Failed to load done notifications", "err", err)
	}
	return store
}

// load reads the done store from disk. It supports two on-disk formats:
//   - New: {"id": "2024-01-15T10:30:00Z", ...}  (map of ID → RFC 3339 timestamp)
//   - Legacy: ["id1", "id2", ...]                (plain array of IDs)
//
// Legacy entries are assigned the zero time, so they resurface on the
// first load after upgrade. Once re-marked as done they get proper timestamps.
func (s *DoneStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Try new format: map[string]string with RFC 3339 values.
	var tsMap map[string]string
	if err := json.Unmarshal(data, &tsMap); err == nil {
		for id, raw := range tsMap {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				log.Warn("Skipping done entry with invalid timestamp", "id", id, "raw", raw, "err", err)
				continue
			}
			s.entries[id] = t
		}
		log.Debug("Loaded done notifications (new format)", "count", len(s.entries))
		return nil
	}

	// Fall back to legacy format: []string.
	var idList []string
	if err := json.Unmarshal(data, &idList); err != nil {
		return err
	}
	for _, id := range idList {
		s.entries[id] = time.Time{}
	}
	log.Debug("Loaded done notifications (legacy format)", "count", len(s.entries))
	return nil
}

func (s *DoneStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.filePath == "" {
		return nil
	}

	tsMap := make(map[string]string, len(s.entries))
	for id, t := range s.entries {
		tsMap[id] = t.Format(time.RFC3339)
	}

	data, err := json.MarshalIndent(tsMap, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return err
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	if err := os.Rename(tmpPath, s.filePath); err != nil {
		os.Remove(tmpPath)
		return err
	}

	log.Debug("Saved done notifications", "count", len(tsMap))
	return nil
}

// MarkDone records the notification's current updated_at as the "done at"
// timestamp. If the notification later receives new activity (a newer
// updated_at), IsDone will return false.
func (s *DoneStore) MarkDone(id string, updatedAt time.Time) {
	s.mu.Lock()
	s.entries[id] = updatedAt
	s.mu.Unlock()
	go s.save()
}

// IsDone returns true only if the notification has not been updated since it
// was marked done: !updatedAt.After(doneAt).
func (s *DoneStore) IsDone(id string, updatedAt time.Time) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	doneAt, ok := s.entries[id]
	if !ok {
		return false
	}
	return !updatedAt.After(doneAt)
}

// Remove removes a notification from the done store.
func (s *DoneStore) Remove(id string) {
	s.mu.Lock()
	delete(s.entries, id)
	s.mu.Unlock()
	go s.save()
}

// Flush forces an immediate synchronous save.
func (s *DoneStore) Flush() error {
	return s.save()
}

// Singleton

var (
	doneStore     *DoneStore
	doneStoreOnce sync.Once
)

// GetDoneStore returns the singleton done store.
func GetDoneStore() *DoneStore {
	doneStoreOnce.Do(func() {
		doneStore = newDoneStore("done.json")
	})
	return doneStore
}
