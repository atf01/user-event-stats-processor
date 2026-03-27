package models

import "time"

type Event struct {
	UserID    string    `json:"user_id"`
	EventType string    `json:"event_type"`
	Value     int64     `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type UserStats map[string]int64
