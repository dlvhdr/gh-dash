package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/log"
)

const (
	bookmarksFileName   = "bookmarks.json"
	defaultXDGConfigDir = ".config"
	dashDir             = "gh-dash"
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
		store.load()
		bookmarkStore = store
	})
	return bookmarkStore
}

func (s *BookmarkStore) getBookmarksFilePath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Error("Failed to get home directory", "err", err)
			return ""
		}
		configDir = filepath.Join(homeDir, defaultXDGConfigDir)
	}
	return filepath.Join(configDir, dashDir, bookmarksFileName)
}

// load reads bookmarks from the JSON file
func (s *BookmarkStore) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.filePath == "" {
		return
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Error("Failed to read bookmarks file", "err", err)
		}
		return
	}

	var bookmarkList []string
	if err := json.Unmarshal(data, &bookmarkList); err != nil {
		log.Error("Failed to parse bookmarks file", "err", err)
		return
	}

	for _, id := range bookmarkList {
		s.bookmarks[id] = true
	}
	log.Debug("Loaded bookmarks", "count", len(s.bookmarks))
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
	if s.bookmarks[notificationId] {
		delete(s.bookmarks, notificationId)
		s.mu.Unlock()
		s.save()
		return false
	}
	s.bookmarks[notificationId] = true
	s.mu.Unlock()
	s.save()
	return true
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
