package dto

import "time"

type CacheItem struct {
	Path    string
	Size    int64
	Updated time.Time
}
