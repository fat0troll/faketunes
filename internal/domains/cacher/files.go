package cacher

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"source.hodakov.me/hdkv/faketunes/internal/domains/cacher/dto"
	"source.hodakov.me/hdkv/faketunes/internal/domains/cacher/models"
)

// GetFileDTO gets the ALAC file from cache or transcodes one with transcoder if needed.
func (c *Cacher) GetFileDTO(sourcePath string) (*dto.CacheItem, error) {
	item, err := c.getFile(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w (%w)", ErrCacher, ErrFailedToGetSourceFile, err)
	}

	c.app.Logger().WithField("item", item).Debug("Retrieved cache item")

	return models.CacheItemModelToDTO(item), nil
}

func (c *Cacher) getFile(sourcePath string) (*models.CacheItem, error) {
	sourceFileInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w (%w)", ErrCacher, ErrFailedToGetSourceFile, err)
	}

	keyData := fmt.Sprintf("%s:%d", sourcePath, sourceFileInfo.ModTime().UnixNano())
	hash := md5.Sum([]byte(keyData))
	cacheKey := hex.EncodeToString(hash[:])
	cacheFilePath := filepath.Join(c.cacheDir, cacheKey+".m4a")

	c.itemsMutex.Lock()
	defer c.itemsMutex.Unlock()

	// Check if file information exists in cache
	if item, ok := c.items[cacheKey]; ok {
		if _, err := os.Stat(item.Path); err != nil {
			// File exists in cache and on disk
			item.Updated = time.Now().UTC()

			c.updateCachedStat(sourcePath, item.Size)

			return item, nil
		}
	}

	// Check if file exists on disk but information about it doesn't exist in
	// the memory (for example, after application restart).
	if cachedFileInfo, err := os.Stat(cacheFilePath); err == nil {
		// Verify that the file on disk is newer than the source file and has content.
		// If that's the case, return the item information and store it in memory.
		if cachedFileInfo.ModTime().After(sourceFileInfo.ModTime()) &&
			cachedFileInfo.Size() > 1024 {
			item := &models.CacheItem{
				Path:    cacheFilePath,
				Size:    cachedFileInfo.Size(),
				Updated: time.Now().UTC(),
			}
			c.items[cacheKey] = item
			c.currentSize += cachedFileInfo.Size()

			c.updateCachedStat(sourcePath, item.Size)

			return item, nil
		}
	}

	// File does not exist on disk, need to transcode.
	// Register in the queue
	c.transcoder.QueueChannel() <- struct{}{}

	defer func() {
		<-c.transcoder.QueueChannel()
	}()

	// Convert file
	size, err := c.transcoder.Convert(sourcePath, cacheFilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w (%w)", ErrCacher, ErrFailedToTranscodeFile, err)
	}

	// Add converted file information to cache
	item := &models.CacheItem{
		Path:    cacheFilePath,
		Size:    size,
		Updated: time.Now(),
	}
	c.items[cacheKey] = item
	c.currentSize += size

	c.updateCachedStat(sourcePath, size)
	// TODO: run cleanup on inotify events.
	c.cleanup()

	return item, nil
}
