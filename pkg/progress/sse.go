package progress

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// WriteSSE encodes a ProgressEvent as one SSE frame matching static/home/chat.html listeners.
func WriteSSE(w http.ResponseWriter, flusher http.Flusher, ev ProgressEvent) error {
	var eventName string
	var payload any

	switch ev.Kind {
	case KindStepStart:
		eventName = "startprogress"
		payload = map[string]string{
			"task_id": ev.TaskID,
			"node":    ev.Node,
			"status":  "start",
			"time":    ev.Time,
		}
	case KindStepEnd:
		eventName = "endprogress"
		payload = map[string]string{
			"task_id": ev.TaskID,
			"node":    ev.Node,
			"status":  "end",
			"time":    ev.Time,
		}
	case KindExportComplete:
		eventName = "progress"
		m := map[string]string{
			"task_id": ev.TaskID,
			"status":  ev.ExportStatus,
		}
		if ev.URL != "" {
			m["url"] = ev.URL
		}
		if ev.Message != "" {
			m["msg"] = ev.Message
		}
		payload = m
	default:
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, string(data)); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}
