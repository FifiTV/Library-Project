package models

type User struct {
	Email    string `firestore:"email"`
	Password string `firestore:"password"`
}
