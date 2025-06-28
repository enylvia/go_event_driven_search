package model

import "time"

type NewsEvent struct {
	Type      string       `json:"type"`
	Timestamp time.Time    `json:"timestamp"`
	Payload   DocumentNews `json:"payload"`
}
