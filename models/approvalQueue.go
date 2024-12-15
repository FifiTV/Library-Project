package models

type ApprovalQueue struct {
	UserID          int `firestore:"user_id"`
	BookID          int `firestore:"book_id"`
	InventoryNumber int `firestore:"inventory_number"`
}
