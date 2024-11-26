package models

type User struct {
	Email     string `firestore:"email"`
	Password  string `firestore:"password"`
	FirstName string `firestore:"firstname"`
	LastName  string `firestore:"lastname"`
	Age       int    `firestore:"age"`
	Role      int    `firestore:"role"`
}
