package runbook

import "time"

type Status string

const (
	Draft     Status = "draft"
	Review    Status = "review"
	Published Status = "published"
)

type Runbook struct {
	ID          string     `json:"id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	Status      Status     `json:"status"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	FolderID    *string    `json:"folder_id,omitempty"`
}
