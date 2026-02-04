package common

import (
	"context"
	"fmt"
	"net/http"
)

type ProgressEmitter interface {
	Emit(msg string)
}
type progressKeyType struct{}

var progressKey = progressKeyType{}

type LogEmitter struct {
	W       http.ResponseWriter
	Flusher http.Flusher
}

func WithProgressEmitter(
	ctx context.Context,
	emitter ProgressEmitter,
) context.Context {
	return context.WithValue(ctx, progressKey, emitter)
}
func GetProgressEmitter(ctx context.Context) (ProgressEmitter, bool) {
	emitter, ok := ctx.Value(progressKey).(ProgressEmitter)
	return emitter, ok
}
func (l *LogEmitter) Emit(msg string) {
	fmt.Fprint(l.W, msg)
	l.Flusher.Flush()
}
