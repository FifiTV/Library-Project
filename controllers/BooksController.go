package controllers

import (
	"context"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/models"

	"github.com/gofiber/fiber/v2"
)

func GetAllBooks(c *fiber.Ctx) error {
	// Reference the "books" collection
	booksCollection := initializers.Client.Collection("books")
	// Get all documents in the collection
	docs, err := booksCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch books",
		})
	}

	// Slice to store books
	var books []models.Book

	// Loop through documents and decode into Book structs
	for _, doc := range docs {
		var book models.Book
		if err := doc.DataTo(&book); err != nil {
			log.Printf("Error decoding document: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to decode book data",
			})
		}
		books = append(books, book)
	}

	// Return books in JSON format
	return c.JSON(books)
}
