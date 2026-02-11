package application

import "errors"

var (
	ErrApplication               = errors.New("application")
	ErrConfigInitializationError = errors.New("config initialization error")
	ErrConnectDependencies       = errors.New("failed to connect dependencies")
	ErrDomainInit                = errors.New("failed to initialize domains")
)
