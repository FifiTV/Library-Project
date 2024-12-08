package models

// BorrowEventWithBook combines the BorrowEvent and Book structs
type BorrowEventWithBook struct {
	BorrowEvent BorrowEvent `json:"borrow_event"`
	Book        Book        `json:"book"`
}
