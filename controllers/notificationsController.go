package controllers

import (
	"context"
	"fmt"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"
	"my-firebase-project/models"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
)

// FetchNotifications retrieves notifications for the logged-in user
func FetchNotifications(c *fiber.Ctx) error {
	sess, err := middleware.GetSession(c)
	if err != nil {
		log.Println("Error fetching session:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch session",
		})
	}

	// Pobierz userID z sesji
	userID, ok := sess.Get("userID").(int)
	if !ok || userID == 0 {
		log.Println("User ID not found or invalid in session")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Missing or invalid user ID",
		})
	}

	log.Printf("Fetching notifications for userId: %d", userID)

	ctx := context.Background()
	notifications := []models.Notification{}

	// Pobierz powiadomienia z Firestore
	snapshot, err := initializers.Client.Collection("notifications").
		Where("recipientId", "==", fmt.Sprintf("%d", userID)). // Konwersja ID na string
		OrderBy("timestamp", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		log.Printf("Error fetching notifications: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch notifications: %v", err),
		})
	}

	for _, doc := range snapshot {
		var notification models.Notification
		doc.DataTo(&notification)
		notification.ID = doc.Ref.ID
		notifications = append(notifications, notification)
	}

	log.Printf("Returning notifications: %+v", notifications)
	return c.JSON(notifications)
}


// CreateNotification creates a new notification
func CreateNotification(recipientID, bookTitle, message string, role int, status bool) error {
	log.Printf("Creating notification for user: %s, book: %s", recipientID, bookTitle)

	ctx := context.Background()

	notification := map[string]interface{}{
		"recipientId": recipientID,
		"bookTitle":   bookTitle,
		"message":     message,
		"role":        role,
		"status":      status,
		"timestamp":   firestore.ServerTimestamp,
	}

	_, _, err := initializers.Client.Collection("notifications").Add(ctx, notification)
	if err != nil {
		log.Printf("Error creating notification: %v", err)
		return err
	}

	log.Println("Notification created successfully.")
	return nil
}

// AddTestNotifications adds test notifications to Firestore
func AddTestNotifications(c *fiber.Ctx) error {
	log.Println("Adding test notifications...")

	ctx := context.Background()

	// Example test data
	testNotifications := []map[string]interface{}{
		{
			"recipientId": "5",
			"bookTitle":   "Book A",
			"message":     "Notification 1: You borrowed Book A.",
			"role":        2,
			"status":      false,
			"timestamp":   firestore.ServerTimestamp,
		},
		{
			"recipientId": "5",
			"bookTitle":   "Book B",
			"message":     "Notification 2: Return Book B.",
			"role":        2,
			"status":      false,
			"timestamp":   firestore.ServerTimestamp,
		},
		{
			"recipientId": "5",
			"bookTitle":   "Book C",
			"message":     "Notification 3: You borrowed Book C.",
			"role":        2,
			"status":      false,
			"timestamp":   firestore.ServerTimestamp,
		},
	}

	for _, notification := range testNotifications {
		_, _, err := initializers.Client.Collection("notifications").Add(ctx, notification)
		if err != nil {
			log.Printf("Error adding test notification: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to add test notification: %v", err),
			})
		}
		log.Printf("Added notification: %+v", notification)
	}

	log.Println("Test notifications added successfully.")
	return c.JSON(fiber.Map{
		"message": "Test notifications added successfully.",
	})
}
