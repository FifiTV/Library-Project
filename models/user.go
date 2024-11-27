package models

type User struct {
	Id        int    `firestore:"id"`
	Email     string `firestore:"email"`
	Password  string `firestore:"password"`
	FirstName string `firestore:"firs_tname"`
	LastName  string `firestore:"last_name"`
	Age       int    `firestore:"age"`
	Role      int    `firestore:"role"`
}
