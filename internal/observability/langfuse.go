package observability

import (
	"sync"

	"main/config"

	"github.com/cloudwego/eino-ext/callbacks/langfuse"
	"github.com/cloudwego/eino/callbacks"
)

var (
	langfuseFlush func()
	langfuseOnce  sync.Once
)

// InitLangfuseEino registers the Langfuse callback as a global Eino handler once at process startup.
// It must not live inside business graphs/tools.
func InitLangfuseEino(cfg *config.ParamsConfig) {
	if cfg == nil {
		return
	}
	langfuseOnce.Do(func() {
		cbh, flusher := langfuse.NewLangfuseHandler(&langfuse.Config{
			Host:      cfg.Langfuse.Host,
			PublicKey: cfg.Langfuse.PublicKey,
			SecretKey: cfg.Langfuse.SecretKey,
		})
		callbacks.AppendGlobalHandlers(cbh)
		langfuseFlush = flusher
	})
}

// FlushLangfuse flushes buffered Langfuse payloads (e.g. after an HTTP stream completes).
func FlushLangfuse() {
	if langfuseFlush != nil {
		langfuseFlush()
	}
}
