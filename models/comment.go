package models

type Comment struct {
	UserID  int    `firestore:"userId"`
	Content string `firestore:"content"`
}
