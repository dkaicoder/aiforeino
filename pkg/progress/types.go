package progress

// Kind identifies a progress fact produced by tools/graphs (no transport framing).
type Kind int

const (
	KindStepStart Kind = iota
	KindStepEnd
	KindExportComplete
)

// ProgressEvent is domain-only data for UI / logs. HTTP layer maps it to SSE.
type ProgressEvent struct {
	Kind Kind

	TaskID string
	Node   string // graph node name for step events

	// Time is formatted for display (e.g. 15:04:05), used on step events.
	Time string

	// ExportStatus is "completed" or "error" for KindExportComplete.
	ExportStatus string
	Message      string
	URL          string
}
