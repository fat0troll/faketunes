package transcoder

import "errors"

var (
	ErrTranscoder               = errors.New("transcoder")
	ErrTranscodeError           = errors.New("transcode error")
	ErrTranscodedFileIsTooSmall = errors.New("transcoded file is too small")
	ErrTranscodedFileNotFound   = errors.New("transcoded file not found")
)
