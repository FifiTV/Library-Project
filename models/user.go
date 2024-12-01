package models

import "time"

type User struct {
	Id        int       `firestore:"id"`
	FirstName string    `firestore:"firstname"`
	LastName  string    `firestore:"lastname"`
	Email     string    `firestore:"email"`
	Password  string    `firestore:"password"`
	Role      int       `firestore:"role"`
	BirthDate time.Time `firestore:"birth_date"`
}
