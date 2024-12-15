package models

import "time"

// type Book struct {
// 	Author       string    `json:"author"`
// 	Pages        int       `json:"pages"`
// 	Published_at time.Time `json:"published_at"`
// 	Title        string    `json:"title"`
// }

type Book struct {
	Title       string    `firestore:"title"`
	Author      string    `firestore:"author"`
	Publisher   string    `firestore:"publisher"`
	Genre       string    `firestore:"genre"`
	Pages       int       `firestore:"pages"`
	Id          int       `firestore:"id"`
	Description string    `firestore:"description"`
	Cover       string    `firestore:"cover"`
	PublishedAt time.Time `firestore:"published_at"`
}
