package models

import "time"

// CacherStat is representing information about a single object size in cache.
type CacherStat struct {
	Size    int64
	Created time.Time
}
