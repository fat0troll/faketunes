package application

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"github.com/sirupsen/logrus"
	"source.hodakov.me/hdkv/faketunes/internal/configuration"
	"source.hodakov.me/hdkv/faketunes/internal/domains"
)

type App struct {
	ctx    context.Context
	logger *logrus.Entry
	config *configuration.Config

	domains      map[string]domains.Domain
	domainsMutex sync.RWMutex
}

func (a *App) Config() *configuration.Config {
	return a.config
}

func (a *App) Context() context.Context {
	return a.ctx
}

func (a *App) Logger() *logrus.Entry {
	return a.logger
}

func New(ctx context.Context) *App {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	app := new(App)

	// Initialize standard logger with memory stats and context attached permanently.
	logger := logrus.StandardLogger()

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	app.logger = logger.WithContext(ctx).WithFields(logrus.Fields{
		"memalloc": fmt.Sprintf("%dMB", m.Alloc/1024/1024),
		"memsys":   fmt.Sprintf("%dMB", m.Sys/1024/1024),
		"numgc":    fmt.Sprintf("%d", m.NumGC),
	})

	app.ctx = ctx

	app.domains = make(map[string]domains.Domain)

	return app
}

func (a *App) InitConfig() error {
	config, err := configuration.New()
	if err != nil {
		return fmt.Errorf("%w: %w (%w)", ErrApplication, ErrConfigInitializationError, err)
	}

	a.config = config

	return nil
}

func (a *App) InitLogger() {
	a.logger.Logger.SetLevel(a.config.FakeTunes.LogLevel)

	a.logger.WithField("log level", a.config.FakeTunes.LogLevel).Debug("Set log level")
}

func (a *App) RegisterDomain(name string, implementation domains.Domain) {
	a.domainsMutex.Lock()
	defer a.domainsMutex.Unlock()

	a.domains[name] = implementation
}

func (a *App) RetrieveDomain(name string) any {
	a.domainsMutex.RLock()
	defer a.domainsMutex.RUnlock()

	return a.domains[name]
}

func (a *App) ConnectDependencies() error {
	a.domainsMutex.RLock()
	defer a.domainsMutex.RUnlock()

	for _, domain := range a.domains {
		err := domain.ConnectDependencies()
		if err != nil {
			return fmt.Errorf("%w: %w (%w)", ErrApplication, ErrConnectDependencies, err)
		}
	}

	return nil
}

func (a *App) StartDomains() error {
	a.domainsMutex.RLock()
	defer a.domainsMutex.RUnlock()

	for _, domain := range a.domains {
		err := domain.Start()
		if err != nil {
			return fmt.Errorf("%w: %w (%w)", ErrApplication, ErrDomainInit, err)
		}
	}

	return nil
}
