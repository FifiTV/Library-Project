package controllers

import (
	"errors"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/models"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/iterator"
)

// GetCopiesOfBook retrieves a list of book copies from the database based on the specified availability status.
//
// Parameters:
//   - c: *fiber.Ctx - The Fiber context used for request handling and obtaining the Firestore context.
//   - b: *models.Book - The book object for which copies are being retrieved.
//   - availability: bool - A flag to filter only available book copies. If true, only available copies are returned.
//
// Returns:
//   - []*models.BookCopy: A slice of pointers to BookCopy objects representing the retrieved book copies.
//   - error: An error object if any issues occur during the query or data retrieval process.
//
// Usage:
//
//	copies, err := GetCopiesOfBook(ctx, book, true)
//	if err != nil {
//	    log.Println("Error fetching book copies:", err)
//	}
//
// Notes:
//   - This function queries the "bookCopies" collection in Firestore.
//   - If availability is set to true, it filters only available copies.
func GetCopiesOfBook(c *fiber.Ctx, b *models.Book, availability bool) ([]*models.BookCopy, error) {
	var copies []*models.BookCopy
	ctx := c.Context()

	query := initializers.Client.Collection("bookCopies").Where("book_id", "==", b.Id)

	if availability {
		query = query.Where("available", "==", true)
	}

	iter := query.Documents(ctx)

	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			log.Println("Error iterating over documents:", err)
			return nil, err
		}

		var copy models.BookCopy
		if err := doc.DataTo(&copy); err != nil {
			log.Println("Error decoding document:", err)
			continue
		}
		copies = append(copies, &copy)
	}

	return copies, nil
}
