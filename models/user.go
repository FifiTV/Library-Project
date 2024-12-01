package models

import "time"

// Roles:
//
// 0 -> Banned user
//
// 1 -> Default user
//
// 2 -> Librarian
type User struct {
	Id        int       `firestore:"id"`
	FirstName string    `firestore:"firstname"`
	LastName  string    `firestore:"lastname"`
	Email     string    `firestore:"email"`
	Password  string    `firestore:"password"`
	Role      int       `firestore:"role"`
	BirthDate time.Time `firestore:"birth_date"`
}
