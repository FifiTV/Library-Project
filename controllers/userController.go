package controllers

import (
	"context"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"
	"my-firebase-project/models"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/iterator"
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

func GetOneUser(c *fiber.Ctx, userId int) models.User {
	usersCollection := initializers.Client.Collection("users")

	docs, err := usersCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)

	}

	var userReturn models.User

	for _, doc := range docs {
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			log.Printf("Error decoding document: %v", err)
		}
		if user.Id == userId {
			userReturn = user
		}
	}

	return userReturn
}
func GetLibrarians(ctx context.Context, client *firestore.Client) ([]string, error) {
	// Pobierz użytkowników z rolą "Librarian"
	iter := client.Collection("users").Where("role", "==", 2).Documents(ctx)
	defer iter.Stop()

	var librarianIDs []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error fetching librarian: %v", err)
			return nil, err
		}

		// Pobierz ID użytkownika
		id := doc.Ref.ID
		librarianIDs = append(librarianIDs, id)
	}

	return librarianIDs, nil
}
