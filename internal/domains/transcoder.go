package domains

const TranscoderName = "transcoder"

type Transcoder interface {
	Convert(sourcePath, destinationPath string) (int64, error)
	QueueChannel() chan struct{}
}
