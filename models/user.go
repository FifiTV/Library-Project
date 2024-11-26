package models

type User struct {
	Id        int    `firestore:"id"`
	Email     string `firestore:"email"`
	Password  string `firestore:"password"`
	FirstName string `firestore:"firstname"`
	LastName  string `firestore:"lastname"`
	Age       int    `firestore:"age"`
	Role      int    `firestore:"role"`
}
