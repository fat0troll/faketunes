package cacher

import "errors"

var (
	ErrCacher                   = errors.New("cacher")
	ErrConnectDependencies      = errors.New("failed to connect dependencies")
	ErrFailedToDeleteCachedFile = errors.New("failed to delete cached file")
	ErrFailedToGetSourceFile    = errors.New("failed to get source file")
	ErrFailedToTranscodeFile    = errors.New("failed to transcode file")
)
