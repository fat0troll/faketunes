package filesystem

import (
	"fmt"

	"source.hodakov.me/hdkv/faketunes/internal/application"
	"source.hodakov.me/hdkv/faketunes/internal/domains"
)

var (
	_ domains.Filesystem = new(FS)
	_ domains.Domain     = new(FS)
)

type FS struct {
	app *application.App

	cacher domains.Cacher

	sourceDir      string
	destinationDir string
	cacheDir       string
	metadataDir    string

	inodeCounter uint64
}

func New(app *application.App) *FS {
	return &FS{
		app: app,

		sourceDir:      app.Config().Paths.Source,
		destinationDir: app.Config().Paths.Destination + "/Music",
		cacheDir:       app.Config().Paths.Destination + "/.cache",
		metadataDir:    app.Config().Paths.Destination + "/.metadata",

		inodeCounter: 1000, // Start counting inodes after the reserved ones
	}
}

func (f *FS) ConnectDependencies() error {
	cacher, ok := f.app.RetrieveDomain(domains.CacherName).(domains.Cacher)
	if !ok {
		return fmt.Errorf(
			"%w: %w (%s)", ErrFilesystem, ErrConnectDependencies,
			"cacher domain interface conversion failed",
		)
	}

	f.cacher = cacher

	return nil
}

func (f *FS) Start() error {
	err := f.prepareDirectories()
	if err != nil {
		return fmt.Errorf("%w: %w (%w)", ErrFilesystem, ErrFailedToPrepareDirectories, err)
	}

	go func() {
		f.mount()
	}()

	return nil
}
