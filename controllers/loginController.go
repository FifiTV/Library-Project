package controllers

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"my-firebase-project/models"
)

func LoginHandler(c *fiber.Ctx, client *firestore.Client) error {
	var user models.User

	// Parse request body
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request payload")
	}

	// Query Firestore for user
	iter := client.Collection("users").Where("email", "==", user.Email).Documents(context.Background())
	doc, err := iter.Next()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("User not found")
	}

	// Fetch the stored user data
	var storedUser models.User
	if err := doc.DataTo(&storedUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Error retrieving user")
	}

	// Compare passwords
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid password")
	}

	return c.Status(fiber.StatusOK).SendString("Login successful")
}
