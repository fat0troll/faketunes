package transcoder

func (t *Transcoder) QueueChannel() chan struct{} {
	return t.transcodeQueue
}
