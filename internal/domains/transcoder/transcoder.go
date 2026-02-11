package transcoder

import (
	"source.hodakov.me/hdkv/faketunes/internal/application"
	"source.hodakov.me/hdkv/faketunes/internal/domains"
)

var (
	_ domains.Transcoder = new(Transcoder)
	_ domains.Domain     = new(Transcoder)
)

type Transcoder struct {
	app            *application.App
	transcodeQueue chan struct{}
}

func New(app *application.App) *Transcoder {
	return &Transcoder{
		app:            app,
		transcodeQueue: make(chan struct{}, app.Config().Transcoding.Parallel),
	}
}

func (t *Transcoder) ConnectDependencies() error {
	return nil
}

func (t *Transcoder) Start() error {
	return nil
}
