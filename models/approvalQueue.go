package models

type ApprovalQueue struct {
	UserID          string `firestore:"user_id"`
	BookID          string `firestore:"book_id"`
	InventoryNumber string `firestore:"inventory_number"`
}
