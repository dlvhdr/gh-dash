package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/log"
)

const (
	bookmarksFileName  = "bookmarks.json"
	defaultXDGStateDir = ".local/state"
	dashDir            = "gh-dash"
)

// BookmarkStore manages locally stored notification bookmarks
type BookmarkStore struct {
	mu        sync.RWMutex
	bookmarks map[string]bool // notification ID -> bookmarked
	filePath  string
}

var (
	bookmarkStore     *BookmarkStore
	bookmarkStoreOnce sync.Once
)

// GetBookmarkStore returns the singleton bookmark store instance
func GetBookmarkStore() *BookmarkStore {
	bookmarkStoreOnce.Do(func() {
		store := &BookmarkStore{
			bookmarks: make(map[string]bool),
		}
		store.filePath = store.getBookmarksFilePath()
		if err := store.load(); err != nil {
			log.Error("Failed to load bookmarks", "err", err)
		}
		bookmarkStore = store
	})
	return bookmarkStore
}

func (s *BookmarkStore) getBookmarksFilePath() string {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Error("Failed to get home directory", "err", err)
			return ""
		}
		stateDir = filepath.Join(homeDir, defaultXDGStateDir)
	}
	return filepath.Join(stateDir, dashDir, bookmarksFileName)
}

// load reads bookmarks from the JSON file
func (s *BookmarkStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No bookmarks file yet, not an error
		}
		return err
	}

	var bookmarkList []string
	if err := json.Unmarshal(data, &bookmarkList); err != nil {
		return err
	}

	for _, id := range bookmarkList {
		s.bookmarks[id] = true
	}
	log.Debug("Loaded bookmarks", "count", len(s.bookmarks))
	return nil
}

// save writes bookmarks to the JSON file
func (s *BookmarkStore) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.filePath == "" {
		return nil
	}

	bookmarkList := make([]string, 0, len(s.bookmarks))
	for id := range s.bookmarks {
		bookmarkList = append(bookmarkList, id)
	}

	data, err := json.MarshalIndent(bookmarkList, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(s.filePath, data, 0o644); err != nil {
		return err
	}

	log.Debug("Saved bookmarks", "count", len(bookmarkList))
	return nil
}

// IsBookmarked checks if a notification is bookmarked
func (s *BookmarkStore) IsBookmarked(notificationId string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.bookmarks[notificationId]
}

// ToggleBookmark toggles the bookmark state for a notification
// Returns the new bookmark state
func (s *BookmarkStore) ToggleBookmark(notificationId string) bool {
	s.mu.Lock()
	newState := !s.bookmarks[notificationId]
	if newState {
		s.bookmarks[notificationId] = true
	} else {
		delete(s.bookmarks, notificationId)
	}
	s.mu.Unlock()
	s.save()
	return newState
}

// AddBookmark adds a bookmark for a notification
func (s *BookmarkStore) AddBookmark(notificationId string) {
	s.mu.Lock()
	s.bookmarks[notificationId] = true
	s.mu.Unlock()
	s.save()
}

// RemoveBookmark removes a bookmark for a notification
func (s *BookmarkStore) RemoveBookmark(notificationId string) {
	s.mu.Lock()
	delete(s.bookmarks, notificationId)
	s.mu.Unlock()
	s.save()
}

// GetBookmarkedIds returns all bookmarked notification IDs
func (s *BookmarkStore) GetBookmarkedIds() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.bookmarks))
	for id := range s.bookmarks {
		ids = append(ids, id)
	}
	return ids
}
