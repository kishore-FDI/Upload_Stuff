package api

// ProgressMessage represents upload progress data
type ProgressMessage struct {
	Type      string  `json:"type"` // "progress", "complete", "created", "error"
	UploadID  string  `json:"upload_id"`
	Progress  float64 `json:"progress"` // 0-100
	BytesSent int64   `json:"bytes_sent"`
	TotalSize int64   `json:"total_size"`
	Status    string  `json:"status"` // "uploading", "completed", "failed", "created"
	Message   string  `json:"message,omitempty"`
}
