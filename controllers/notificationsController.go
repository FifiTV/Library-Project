package controllers

import (
	"context"
	"fmt"
	"my-firebase-project/initializers"
	"my-firebase-project/models"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
)

// FetchNotifications retrieves notifications for a specific user
func FetchNotifications(c *fiber.Ctx) error {
	userID := c.Query("userId") // Pobierz ID użytkownika z zapytania
	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing userId parameter",
		})
	}

	ctx := context.Background()
	notifications := []models.Notification{}

	// Pobierz powiadomienia z Firestore
	snapshot, err := initializers.Client.Collection("notifications").
		Where("recipientId", "==", userID).
		OrderBy("timestamp", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
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

	return c.JSON(notifications)
}

// CreateNotification creates a new notification
func CreateNotification(recipientID, bookTitle, message string, role int, status bool) error {
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
	return err
}
func AddTestNotifications(c *fiber.Ctx) error {
	ctx := context.Background()

	// Przykładowe dane testowe
	testNotifications := []map[string]interface{}{
		{
			"recipientId": "test-user-1",
			"bookTitle":   "Książka A",
			"message":     "Powiadomienie 1: Wypożyczyłeś książkę A.",
			"role":        1,
			"status":      false,
			"timestamp":   firestore.ServerTimestamp,
		},
		{
			"recipientId": "test-user-1",
			"bookTitle":   "Książka B",
			"message":     "Powiadomienie 2: Oddaj książkę B.",
			"role":        1,
			"status":      false,
			"timestamp":   firestore.ServerTimestamp,
		},
		{
			"recipientId": "test-user-2",
			"bookTitle":   "Książka C",
			"message":     "Powiadomienie 3: Wypożyczyłeś książkę C.",
			"role":        2,
			"status":      false,
			"timestamp":   firestore.ServerTimestamp,
		},
	}

	for _, notification := range testNotifications {
		_, _, err := initializers.Client.Collection("notifications").Add(ctx, notification)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to add test notification: %v", err),
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Test notifications added successfully.",
	})
}
