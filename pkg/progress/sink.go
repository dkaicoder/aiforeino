package progress

import "context"

// ProgressSink receives domain progress events. Tools call Publish only;
// transport (SSE/WebSocket) is implemented outside.
type ProgressSink interface {
	Publish(ev ProgressEvent) error
}

type sinkKey struct{}

// WithSink returns a child context that carries sink for GetSink.
func WithSink(ctx context.Context, sink ProgressSink) context.Context {
	return context.WithValue(ctx, sinkKey{}, sink)
}

// GetSink returns the sink bound to ctx, if any.
func GetSink(ctx context.Context) (ProgressSink, bool) {
	s, ok := ctx.Value(sinkKey{}).(ProgressSink)
	return s, ok
}

// TryPublish publishes ev if ctx carries a sink; otherwise no-op.
func TryPublish(ctx context.Context, ev ProgressEvent) {
	s, ok := GetSink(ctx)
	if !ok {
		return
	}
	_ = s.Publish(ev)
}
