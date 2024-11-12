package controllers

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"my-firebase-project/models"
)

func RegisterHandler(c *fiber.Ctx, client *firestore.Client) error {
	var user models.User

	// Parse request body
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request payload")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to hash password")
	}
	user.Password = string(hashedPassword)

	// Save user in Firestore
	_, _, err = client.Collection("users").Add(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create user")
	}

	return c.Status(fiber.StatusCreated).SendString("User registered successfully")
}
