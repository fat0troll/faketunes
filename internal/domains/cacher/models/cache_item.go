package models

import (
	"time"

	"source.hodakov.me/hdkv/faketunes/internal/domains/cacher/dto"
)

type CacheItem struct {
	Path    string
	Size    int64
	Updated time.Time
}

func CacheItemModelToDTO(item *CacheItem) *dto.CacheItem {
	return &dto.CacheItem{
		Path:    item.Path,
		Size:    item.Size,
		Updated: item.Updated,
	}
}
