package controllers

import (
	"context"
	"fmt"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/models"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
)

func GetAllBooks(c *fiber.Ctx) []models.Book {

	searchTitle := c.Query("title", "")
	searchAuthor := c.Query("author", "")
	searchGenre := c.Query("genre", "")
	searchYear := c.Query("year", "")
	searchPublisher := c.Query("publisher", "")

	// Reference the "books" collection
	booksCollection := initializers.Client.Collection("books")
	query := booksCollection.Query

	if searchTitle != "" {
		// Perform case-insensitive search by title
		query = query.Where("title", ">=", strings.Title(searchTitle)).Where("title", "<", strings.Title(searchTitle)+"\uf8ff")
	}
	if searchAuthor != "" {
		query = query.Where("author", ">=", strings.Title(searchAuthor)).Where("author", "<", strings.Title(searchAuthor)+"\uf8ff")
	}
	if searchGenre != "" {
		query = query.Where("genre", "==", searchGenre)
	}
	if searchPublisher != "" {
		query = query.Where("publisher", ">=", strings.Title(searchPublisher)).Where("publisher", "<", strings.Title(searchPublisher)+"\uf8ff")
	}
	if searchYear != "" {
		year, _ := strconv.Atoi(searchYear)
		startOfYear := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		endOfYear := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)

		query = query.Where("published_at", ">=", startOfYear).Where("published_at", "<", endOfYear)
	}

	// Get all documents in the collection
	docs, err := query.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)
	}

	// Slice to store books
	var books []models.Book

	// Loop through documents and decode into Book structs
	for _, doc := range docs {
		var book models.Book
		if err := doc.DataTo(&book); err != nil {
			log.Printf("Error decoding document: %v", err)
		}
		books = append(books, book)
	}

	searchTitle = ""
	searchAuthor = ""

	// Return books in JSON format
	return books
}
func GetOneBook(c *fiber.Ctx, bookId int) models.Book {
	booksCollection := initializers.Client.Collection("books")

	docs, err := booksCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)

	}

	var bookReturn models.Book

	for _, doc := range docs {
		var book models.Book
		if err := doc.DataTo(&book); err != nil {
			log.Printf("Error decoding document: %v", err)
		}
		if book.Id == bookId {
			bookReturn = book
		}
	}

	return bookReturn
}

func AddBook(c *fiber.Ctx, client *firestore.Client) error {
	var book models.Book

	// Parse request body
	if err := c.BodyParser(&book); err != nil {
		return c.Status(fiber.StatusBadRequest).Render("forms/addBook", fiber.Map{
			"errorMessage": "Nieprawidłowe dane wejściowe.",
		})
	}

	fmt.Println(book)
	return nil
}
