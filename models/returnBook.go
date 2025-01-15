package models

type ReturnBook struct {
	BorrowEvent BorrowEvent
	Book        Book
	BookCopy    BookCopy
	User        User
}
