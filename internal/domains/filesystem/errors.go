package filesystem

import "errors"

var (
	ErrFilesystem                         = errors.New("filesystem")
	ErrConnectDependencies                = errors.New("failed to connect dependencies")
	ErrFailedToPrepareDirectories         = errors.New("failed to prepare directories")
	ErrNoSource                           = errors.New("source does not exist")
	ErrFailedToCleanupDestination         = errors.New("failed to clean up destination directory")
	ErrFailedToCreateDestinationDirectory = errors.New("failed to create destination directory")
)
