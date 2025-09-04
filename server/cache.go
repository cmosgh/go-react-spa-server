package server

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

type cachedAsset struct {
	Content  []byte
	ModTime  time.Time
	Size     int64
	MimeType string // To store content type
}

var (
	inMemoryCache = make(map[string]cachedAsset)
)

// LoadCriticalAssetsIntoCache reads specified critical assets into memory.
func LoadCriticalAssetsIntoCache(staticDir string) error {
	// Re-initialize the cache to ensure it's completely empty
	inMemoryCache = make(map[string]cachedAsset)

	assetsToCache := []struct {
		Name     string
		MimeType string
	}{
		{"index.html", "text/html; charset=utf-8"},
		{"vite.svg", "image/svg+xml"},
	}

	for _, asset := range assetsToCache {
		filePath := filepath.Join(staticDir, asset.Name)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Warning: Could not load critical asset %s into cache: %v", filePath, err)
			continue
		}
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			log.Printf("Warning: Could not get file info for %s: %v", filePath, err)
			continue
		}
		inMemoryCache["/"+asset.Name] = cachedAsset{
			Content:  content,
			ModTime:  fileInfo.ModTime(),
			Size:     fileInfo.Size(),
			MimeType: asset.MimeType,
		}
	}
	log.Printf("Loaded %d critical assets into in-memory cache.", len(inMemoryCache))
	return nil
}

// GetCachedAsset retrieves a cached asset by its URL path.
func GetCachedAsset(urlPath string) (cachedAsset, bool) {
	asset, ok := inMemoryCache[urlPath]
	return asset, ok
}
