package configuration

import "errors"

var (
	ErrConfiguration               = errors.New("configuration")
	ErrCantReadConfigFile          = errors.New("can't read config file")
	ErrCantParseConfigFile         = errors.New("can't parse config file")
	ErrSourceDirectoryDoesNotExist = errors.New("source directory does not exist")
)
