package geocode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/adrg/xdg"
)

// CacheData represents the JSON structure of the cache file
type CacheData struct {
	Version   string            `json:"version"`
	Locations map[string]Result `json:"locations"`
}

// Cache handles persistent storage of geocoding results
type Cache struct {
	mu       sync.RWMutex
	filePath string
	data     CacheData
}

// NewCache creates a new cache instance
func NewCache() (*Cache, error) {
	// Determine cache file path using xdg
	cacheDir := filepath.Join(xdg.CacheHome, "dindoa")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("create cache directory: %w", err)
	}

	filePath := filepath.Join(cacheDir, "geocode.json")

	cache := &Cache{
		filePath: filePath,
		data: CacheData{
			Version:   "1.0",
			Locations: make(map[string]Result),
		},
	}

	// Try to load existing cache
	if err := cache.load(); err != nil {
		// If load fails, start with empty cache (file might not exist yet)
		// This is not an error condition
	}

	return cache, nil
}

// normalizeKey converts a location query to a cache key
func normalizeKey(query string) string {
	return strings.ToLower(strings.TrimSpace(query))
}

// Lookup checks if a location exists in the cache
func (c *Cache) Lookup(query string) (Result, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := normalizeKey(query)
	result, exists := c.data.Locations[key]
	return result, exists
}

// Store saves a geocoding result to the cache
func (c *Cache) Store(result Result) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := normalizeKey(result.Query)
	c.data.Locations[key] = result

	// Save to disk
	return c.save()
}

// load reads the cache from disk
func (c *Cache) load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, not an error
		}
		return fmt.Errorf("read cache file: %w", err)
	}

	if err := json.Unmarshal(data, &c.data); err != nil {
		return fmt.Errorf("parse cache file: %w", err)
	}

	// Ensure locations map is initialized
	if c.data.Locations == nil {
		c.data.Locations = make(map[string]Result)
	}

	return nil
}

// save writes the cache to disk
func (c *Cache) save() error {
	data, err := json.MarshalIndent(c.data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}

	if err := os.WriteFile(c.filePath, data, 0644); err != nil {
		return fmt.Errorf("write cache file: %w", err)
	}

	return nil
}
