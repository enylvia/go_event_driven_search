// pkg/model/event.go
package model

import "time"

type NewsEvent struct {
	Type      string       `json:"type"`      // "CREATED", "UPDATED", "DELETED"
	Timestamp time.Time    `json:"timestamp"` // Waktu event terjadi
	Payload   DocumentNews `json:"payload"`   // Data berita yang terkait
}

type DocumentNews struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Author      string     `json:"author"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	PublishedAt time.Time  `json:"published_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}
