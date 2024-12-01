package controllers

import (
	"context"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"
	"my-firebase-project/models"

	"github.com/gofiber/fiber/v2"
)

func GetAllBorrowEvents(c *fiber.Ctx) []models.BorrowEvent {
	// Reference the "borrowEvents" collection
	borrowEventsCollection := initializers.Client.Collection("borrowEvents")
	// Get all documents in the collection
	docs, err := borrowEventsCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)
	}

	// Slice to store borrowEvents
	var borrowEvents []models.BorrowEvent

	// Loop through documents and decode into borrowEvents structs
	for _, doc := range docs {
		var borrowEvent models.BorrowEvent
		if err := doc.DataTo(&borrowEvent); err != nil {
			log.Printf("Error decoding document: %v", err)
		}
		borrowEvents = append(borrowEvents, borrowEvent)
	}

	// Return borrowEvents in JSON format
	return borrowEvents
}

func GetAllBorrowEventsForUser(c *fiber.Ctx) ([]models.BorrowEventWithBook, error) {
	// Get all borrow events for the user
	borrowEvents := GetAllBorrowEvents(c)

	// Retrieve userID from session
	sess, _ := middleware.GetSession(c)
	userID := sess.Get("userID").(int)

	// Filter borrow events for the user
	var filteredBorrowEvents []models.BorrowEvent
	for _, event := range borrowEvents {
		if event.UserID == userID {
			// Add matching events to the filtered list
			filteredBorrowEvents = append(filteredBorrowEvents, event)
		}
	}

	// Prepare the result by adding the book details to each borrow event
	var borrowEventsWithBooks []models.BorrowEventWithBook
	for _, event := range filteredBorrowEvents {
		// Fetch the book details by BookID
		book := GetOneBook(c, event.BookID)

		// Combine the borrow event with the book details
		borrowEventWithBook := models.BorrowEventWithBook{
			BorrowEvent: event,
			Book:        book,
		}

		// Add the combined data to the result
		borrowEventsWithBooks = append(borrowEventsWithBooks, borrowEventWithBook)
	}

	// Return the combined list
	return borrowEventsWithBooks, nil
}
