package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/models"

	"cloud.google.com/go/firestore"
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

func GetCopyOfBook(c *fiber.Ctx, inventoryNumber int) models.BookCopy {
	booksCopiesCollection := initializers.Client.Collection("bookCopies")

	docs, err := booksCopiesCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)

	}

	var bookCopyReturn models.BookCopy

	for _, doc := range docs {
		var bookCopy models.BookCopy
		if err := doc.DataTo(&bookCopy); err != nil {
			log.Printf("Error decoding document: %v", err)
		}
		if bookCopy.InventoryNumber == inventoryNumber {
			bookCopyReturn = bookCopy
		}
	}

	return bookCopyReturn
}

func ifInventoryNumberExist(c *fiber.Ctx, client *firestore.Client, inventoryNumber int) bool {
	countQuery := client.Collection("bookCopies").
		Where("inventory_number", "==", inventoryNumber).
		Select() // No fields needed; we just need the count.

	// Perform the count aggregation
	agg, _ := countQuery.Documents(c.Context()).GetAll()

	return len(agg) != 0
}

func AddBookCopy(c *fiber.Ctx, client *firestore.Client, bookCopy *models.BookCopy) error {
	ifExist := ifInventoryNumberExist(c, client, bookCopy.InventoryNumber)

	if ifExist {
		return fmt.Errorf("Książka z tym numerem inwentarza już istnieje")
	}

	if bookCopy.BookID == 0 {
		return fmt.Errorf("Podaj poprawny tytuł")
	}

	_, _, err := client.Collection("bookCopies").Add(context.Background(), bookCopy)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func getNextInventoryNumber(c *fiber.Ctx, client *firestore.Client) (int, error) {
	// Query the collection, ordering by `inventory_number` in descending order
	maxQuery := client.Collection("bookCopies").
		OrderBy("inventory_number", firestore.Desc).
		Limit(1)

	// Execute the query
	docIterator := maxQuery.Documents(c.Context())
	defer docIterator.Stop()

	// Fetch the first document
	doc, err := docIterator.Next()
	if err == iterator.Done {

		return 1, nil
	} else if err != nil {
		// Handle other errors
		return 0, err
	}

	// Parse the `inventory_number` field
	var data map[string]interface{}
	if err := doc.DataTo(&data); err != nil {
		return 0, err
	}

	// Safely get `inventory_number` from the document
	if inventoryNumber, ok := data["inventory_number"].(int64); ok {
		return int(inventoryNumber) + 1, nil
	}

	return 0, fmt.Errorf("unable to parse inventory_number field")
}

func GetNextInventoryNumber(c *fiber.Ctx) error {
	nextInventory, err := getNextInventoryNumber(c, initializers.Client)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch next inventory number",
		})
	}

	// Respond with the next inventory number
	return c.JSON(fiber.Map{
		"next_inventory_number": nextInventory,
	})
}
