package models

import "time"

type Book struct {
	Title         string    `firestore:"title"`
	Author        string    `firestore:"author"`
	Pages         int       `firestore:"pages"`
	Id            int       `firestore:"id"`
	Description   string    `firebase:"description"`
	Cover         string    `firebase:"cover"`
	DateOfRelease time.Time `firebase:"date_of_release"`
}
