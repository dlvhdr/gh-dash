package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/log"
)

const (
	defaultXDGStateDir = ".local/state"
	dashDir            = "gh-dash"
)

// NotificationIDStore is a generic store for persisting sets of notification IDs.
// Used for bookmarks, done notifications, and similar features.
type NotificationIDStore struct {
	mu       sync.RWMutex
	ids      map[string]bool
	filePath string
	name     string // for logging
}

func newNotificationIDStore(filename, name string) *NotificationIDStore {
	store := &NotificationIDStore{
		ids:  make(map[string]bool),
		name: name,
	}
	filePath, err := getStateFilePath(filename)
	if err != nil {
		log.Error("Failed to get state file path for "+name, "err", err)
	}
	store.filePath = filePath
	if err := store.load(); err != nil {
		log.Error("Failed to load "+name, "err", err)
	}
	return store
}

func getStateFilePath(filename string) (string, error) {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		stateDir = filepath.Join(homeDir, defaultXDGStateDir)
	}
	return filepath.Join(stateDir, dashDir, filename), nil
}

func (s *NotificationIDStore) load() error {
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

	var idList []string
	if err := json.Unmarshal(data, &idList); err != nil {
		return err
	}

	for _, id := range idList {
		s.ids[id] = true
	}
	log.Debug("Loaded "+s.name, "count", len(s.ids))
	return nil
}

func (s *NotificationIDStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.filePath == "" {
		return nil
	}

	idList := make([]string, 0, len(s.ids))
	for id := range s.ids {
		idList = append(idList, id)
	}

	data, err := json.MarshalIndent(idList, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Use atomic write: write to temp file, then rename.
	// This prevents races when multiple async saves run concurrently.
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

	log.Debug("Saved "+s.name, "count", len(idList))
	return nil
}

// Has checks if an ID is in the store
func (s *NotificationIDStore) Has(id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ids[id]
}

// Add adds an ID to the store
func (s *NotificationIDStore) Add(id string) {
	s.mu.Lock()
	s.ids[id] = true
	s.mu.Unlock()
	go s.save() // Async save to avoid blocking UI
}

// Remove removes an ID from the store
func (s *NotificationIDStore) Remove(id string) {
	s.mu.Lock()
	delete(s.ids, id)
	s.mu.Unlock()
	go s.save() // Async save to avoid blocking UI
}

// Toggle toggles an ID in the store, returns the new state
func (s *NotificationIDStore) Toggle(id string) bool {
	s.mu.Lock()
	newState := !s.ids[id]
	if newState {
		s.ids[id] = true
	} else {
		delete(s.ids, id)
	}
	s.mu.Unlock()
	go s.save() // Async save to avoid blocking UI
	return newState
}

// GetAll returns all IDs in the store
func (s *NotificationIDStore) GetAll() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.ids))
	for id := range s.ids {
		ids = append(ids, id)
	}
	return ids
}

// Flush forces an immediate synchronous save. Useful for testing.
func (s *NotificationIDStore) Flush() error {
	return s.save()
}

// Singletons for bookmark and done stores

var (
	bookmarkStore     *NotificationIDStore
	bookmarkStoreOnce sync.Once
	doneStore         *NotificationIDStore
	doneStoreOnce     sync.Once
)

// GetBookmarkStore returns the singleton bookmark store
func GetBookmarkStore() *NotificationIDStore {
	bookmarkStoreOnce.Do(func() {
		bookmarkStore = newNotificationIDStore("bookmarks.json", "bookmarks")
	})
	return bookmarkStore
}

// GetDoneStore returns the singleton done store
func GetDoneStore() *NotificationIDStore {
	doneStoreOnce.Do(func() {
		doneStore = newNotificationIDStore("done.json", "done notifications")
	})
	return doneStore
}

// Convenience methods for BookmarkStore (maintains API compatibility)

// IsBookmarked checks if a notification is bookmarked
func (s *NotificationIDStore) IsBookmarked(id string) bool {
	return s.Has(id)
}

// ToggleBookmark toggles the bookmark state
func (s *NotificationIDStore) ToggleBookmark(id string) bool {
	return s.Toggle(id)
}

// GetBookmarkedIds returns all bookmarked IDs
func (s *NotificationIDStore) GetBookmarkedIds() []string {
	return s.GetAll()
}

// Convenience methods for DoneStore

// IsDone checks if a notification is marked as done
func (s *NotificationIDStore) IsDone(id string) bool {
	return s.Has(id)
}

// MarkDone marks a notification as done
func (s *NotificationIDStore) MarkDone(id string) {
	s.Add(id)
}
