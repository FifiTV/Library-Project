package models

import "time"

// Notification represents a single notification
type Notification struct {
	ID          string    `json:"id"`
	RecipientID string    `json:"recipientId"`
	BookTitle   string    `json:"bookTitle"`
	Message     string    `json:"message"`
	Role        int       `json:"role"`
	Status      bool      `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
}
