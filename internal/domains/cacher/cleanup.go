package cacher

import (
	"fmt"
	"os"
	"time"
)

func (c *Cacher) cleanup() error {
	for c.currentSize > c.maxSize && len(c.items) > 0 {
		var (
			itemKey    string
			itemSize   int64
			oldestTime time.Time
		)

		for key, item := range c.items {
			if itemKey == "" || item.Updated.Before(oldestTime) {
				itemKey = key
				oldestTime = item.Updated
				itemSize = item.Size
			}
		}

		if itemKey != "" {
			err := os.Remove(c.items[itemKey].Path)
			if err != nil {
				return fmt.Errorf("%w: %w (%w)", ErrCacher, ErrFailedToDeleteCachedFile, err)
			}

			delete(c.items, itemKey)
			c.currentSize -= itemSize
		}
	}

	return nil
}
