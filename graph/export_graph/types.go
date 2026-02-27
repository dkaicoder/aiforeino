package export_graph

import "main/internal/repository"

type MyGraphState struct {
	Query        string
	ExportTaskID string
	DownloadRepo repository.DownloadListRepository
}
