package cacher

import (
	"fmt"
	"sync"

	"source.hodakov.me/hdkv/faketunes/internal/application"
	"source.hodakov.me/hdkv/faketunes/internal/domains"
	"source.hodakov.me/hdkv/faketunes/internal/domains/cacher/models"
)

var (
	_ domains.Cacher = new(Cacher)
	_ domains.Domain = new(Cacher)
)

type Cacher struct {
	app *application.App

	transcoder domains.Transcoder

	cacheDir    string
	cacheMutex  sync.RWMutex
	currentSize int64
	maxSize     int64
	items       map[string]*models.CacheItem
	stat        map[string]*models.CacherStat
}

func New(app *application.App) *Cacher {
	return &Cacher{
		app:      app,
		cacheDir: app.Config().Paths.Destination + "./.cache",
		maxSize:  app.Config().FakeTunes.CacheSize * 1024 * 1024,
		items:    make(map[string]*models.CacheItem, 0),
		stat:     make(map[string]*models.CacherStat, 0),
	}
}

func (c *Cacher) ConnectDependencies() error {
	transcoder, ok := c.app.RetrieveDomain(domains.TranscoderName).(domains.Transcoder)
	if !ok {
		return fmt.Errorf(
			"%w: %w (%s)", ErrCacher, ErrConnectDependencies,
			"transcoder domain interface conversion failed",
		)
	}

	c.transcoder = transcoder

	return nil
}

func (c *Cacher) Start() error {
	return nil
}
