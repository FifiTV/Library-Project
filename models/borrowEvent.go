package models

import (
	"time"
)

type BorrowEvent struct {
	UserID          int       `firestore:"user_id"`
	InventoryNumber int       `firestore:"inventory_number"`
	BookID          int       `firestore:"book_id"`
	BorrowStart     time.Time `firestore:"borrow_start"`
	BorrowEnd       time.Time `firestore:"borrow_end"`
}
