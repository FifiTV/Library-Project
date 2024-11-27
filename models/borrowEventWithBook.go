package models

// BorrowEventWithBook combines the BorrowEvent and Book structs
type BorrowEventWithBook struct {
	BorrowEvent          BorrowEvent `json:"borrow_event"`
	Book                 Book        `json:"book"`
	FormattedBorrowStart string      `json:"formatted_borrow_start"`
	FormattedBorrowEnd   string      `json:"formatted_borrow_end"`
}
