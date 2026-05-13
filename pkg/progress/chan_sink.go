package progress

// ChanSink sends events to a channel; non-blocking when ch is full (drops).
type ChanSink struct {
	ch chan<- ProgressEvent
}

func NewChanSink(ch chan<- ProgressEvent) *ChanSink {
	return &ChanSink{ch: ch}
}

func (s *ChanSink) Publish(ev ProgressEvent) error {
	select {
	case s.ch <- ev:
	default:
		// buffer full: drop rather than block graph/tool execution
	}
	return nil
}
