package models

import "time"

type Book struct {
	Author       string    `json:"author"`
	Pages        int       `json:"pages"`
	Published_at time.Time `json:"published_at"`
	Title        string    `json:"title"`
}
