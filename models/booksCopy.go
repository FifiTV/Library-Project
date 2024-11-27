package models

import (
	"time"
)

type BookCopy struct {
	InventoryNumber int       `firestore:"inventory_number"`
	Available       bool      `firestore:"available"`
	BookID          int       `firestore:"book_id"`
	Location        string    `firestore:"location"`
	AddedOn         time.Time `firestore:"added_on"`
}
