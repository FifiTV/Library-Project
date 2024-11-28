package models

import "time"

type Book struct {
	Title       string    `firestore:"title"`
	Author      string    `firestore:"author"`
	Pages       int       `firestore:"pages"`
	Id          int       `firestore:"id"`
	Description string    `firestore:"description"`
	Cover       string    `firestore:"cover"`
	PublishedAt time.Time `firestore:"published_at"`
}
