package data

import (
	"time"

	"github.com/maypok86/otter/v2"
)

type Cache struct {
	*otter.Cache[string, string]
}

var cache = Cache{}

func New() Cache {
	cache := otter.Must(&otter.Options[string, string]{
		MaximumSize:       10_000,
		ExpiryCalculator:  otter.ExpiryAccessing[string, string](time.Second),           // Reset timer on reads/writes
		RefreshCalculator: otter.RefreshWriting[string, string](500 * time.Millisecond), // Refresh after writes
	})

	return Cache{
		cache,
	}
}

func Get() Cache {
	return cache
}
