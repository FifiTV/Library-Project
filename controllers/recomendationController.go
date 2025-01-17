package controllers

import (
	"context"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/models"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
)

var currentTopBooks []models.Book

func getNewTopBooks(n int) {
	ctx := context.Background()

	client := initializers.Client
	query := client.Collection("books").OrderBy("avg_score", firestore.Desc).Limit(n)

	docs, err := query.Documents(ctx).GetAll()

	if err != nil {
		log.Fatalf("Error fetching top books: %v", err)
	}

	topBooks := make([]models.Book, len(docs))
	for i, doc := range docs {
		var book models.Book
		err := doc.DataTo(&book)
		if err != nil {
			log.Fatalf("Error parsing document: %v", err)
		}
		topBooks[i] = book
	}
	currentTopBooks = topBooks
}

func GetTopBooks(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(currentTopBooks)
}

func StartWorker() {
	getNewTopBooks(10)
}
