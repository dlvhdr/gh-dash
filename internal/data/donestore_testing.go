package data

import (
	"time"
)

// NewDoneStoreForTesting creates a DoneStore backed by the given file path.
func NewDoneStoreForTesting(filePath string) *DoneStore {
	return &DoneStore{
		entries:  make(map[string]time.Time),
		filePath: filePath,
	}
}

// OverrideDoneStoreForTesting replaces the singleton DoneStore with the given
// store. It returns a function that restores the original store.
func OverrideDoneStoreForTesting(store *DoneStore) func() {
	// Ensure the singleton is initialized so sync.Once has fired.
	GetDoneStore()
	old := doneStore
	doneStore = store
	return func() { doneStore = old }
}
