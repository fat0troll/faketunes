package cacher

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"source.hodakov.me/hdkv/faketunes/internal/domains/cacher/models"
)

// getStat returns file size without triggering conversion (for ls/stat).
func (c *Cacher) GetStat(sourcePath string) (int64, error) {
	c.statMutex.RLock()
	defer c.statMutex.RUnlock()

	// First check cache
	if size, ok := c.getCachedStat(sourcePath); ok {
		return size, nil
	}

	// Check if we have a cached converted file
	info, err := os.Stat(sourcePath)
	if err != nil {
		return 0, err
	}

	keyData := fmt.Sprintf("%s:%d", sourcePath, info.ModTime().UnixNano())
	hash := md5.Sum([]byte(keyData))
	key := hex.EncodeToString(hash[:])
	cachePath := filepath.Join(c.cacheDir, key+".m4a")

	// Check if converted file exists and is valid
	if cacheInfo, err := os.Stat(cachePath); err == nil {
		if cacheInfo.ModTime().After(info.ModTime()) && cacheInfo.Size() > 1024 {
			c.updateCachedStat(sourcePath, cacheInfo.Size())

			return cacheInfo.Size(), nil
		}
	}

	// Return estimated size (FLAC file size as placeholder)
	return info.Size(), nil
}

// updateCachedStat updates the stat cache.
func (c *Cacher) updateCachedStat(sourcePath string, size int64) {
	c.statMutex.Lock()
	defer c.statMutex.Unlock()

	c.stat[sourcePath] = &models.CacherStat{
		Size:    size,
		Created: time.Now(),
	}
}

// getCachedStat returns cached file stats.
func (c *Cacher) getCachedStat(sourcePath string) (int64, bool) {
	c.statMutex.RLock()
	defer c.statMutex.RUnlock()

	if stat, ok := c.stat[sourcePath]; ok {
		return stat.Size, true
	}

	return 0, false
}
