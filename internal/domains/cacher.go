package domains

import "source.hodakov.me/hdkv/faketunes/internal/domains/cacher/dto"

const CacherName = "cacher"

type Cacher interface {
	GetStat(sourcePath string) (int64, error)
	GetFileDTO(sourcePath string) (*dto.CacheItem, error)
}
