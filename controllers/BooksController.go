package controllers

import (
	"context"
	"fmt"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"
	"my-firebase-project/models"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/api/iterator"
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

func getNewIDForBook(c *fiber.Ctx, client *firestore.Client) (int, error) {
	// Set up context for Firestore query
	ctx := c.Context() // use the context from Fiber's request

	// Query all documents in the "books" collection
	iter := client.Collection("books").Documents(ctx) // get the iterator for documents in the collection
	defer iter.Stop()

	var maxID int
	// Iterate through documents and find the max ID value
	for {
		doc, err := iter.Next()
		if err != nil {
			// If there is any other error, return it
			break
		}

		// Get the ID field value from the document (it could be int or float64)
		idValue := doc.Data()["id"]
		if idValue == nil {
			// If idValue is nil, print the document data for debugging
			fmt.Println("No ID found for document:", doc.Data())
			continue // skip documents without an ID field
		}

		var idInt int
		// Convert the idValue to an int, if possible
		switch v := idValue.(type) {
		case int:
			idInt = v
		case int64:
			idInt = int(v) // converting int64 to int
		case float64:
			idInt = int(v) // converting float64 to int
		default:
			continue
		}
		if idInt > maxID {
			maxID = idInt
		}
	}
	return maxID + 1, nil
}

func GetBookByTitle(c *fiber.Ctx, client *firestore.Client, title string) (*models.Book, error) {
	// Set up context for Firestore query
	ctx := c.Context()

	query := client.Collection("books").Where("title", "==", title).Limit(1) // Limit to 1 result

	iter := query.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			return nil, fmt.Errorf("nie ma")
		}
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

		var book models.Book
		if err := doc.DataTo(&book); err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		return &book, nil
	}
}

func AddBook(c *fiber.Ctx, client *firestore.Client, book *models.Book) error {

	// Validate Title
	title := c.FormValue("title")
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("title is required")
	}
	book.Title = title

	// Validate Author
	author := c.FormValue("author")
	authorRegex := `^[A-Za-zÀ-ÖØ-öø-ÿ]+(([' -][A-Za-zÀ-ÖØ-öø-ÿ]+)*)$`
	matched, _ := regexp.MatchString(authorRegex, author)
	if strings.TrimSpace(author) == "" || !matched {
		return fmt.Errorf("please provide a valid author")
	}
	book.Author = author

	// Validate Pages
	pagesStr := c.FormValue("pages")
	pages, err := strconv.Atoi(pagesStr)
	if err != nil || pages <= 0 {
		return fmt.Errorf("please provide a valid number of pages")
	}
	book.Pages = pages

	// Get New ID
	bookId, err := getNewIDForBook(c, client)
	if err != nil {
		return fmt.Errorf("failed to generate new book ID: %v", err)
	}
	book.Id = bookId

	// Validate PublishedAt (Date)
	publishedAt := c.FormValue("publishedAt")
	const dateFormat = "2006-01-02"
	parsedTime, err := time.Parse(dateFormat, publishedAt)
	if err != nil {
		return fmt.Errorf("failed to parse published date: %v", err)
	}
	book.PublishedAt = parsedTime

	// Validate Description
	description := c.FormValue("description")
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description is required")
	}
	book.Description = description

	// Validate Cover URL
	coverLink := c.FormValue("coverLink")
	if strings.TrimSpace(coverLink) == "" {
		return fmt.Errorf("cover link is required")
	}
	book.Cover = coverLink

	// Add the Book to Firestore
	_, _, err = client.Collection("books").Add(context.Background(), book)
	if err != nil {
		return fmt.Errorf("failed to add book to Firestore: %v", err)
	}

	// Success message (optional, can be omitted for a clean API response)
	// fmt.Println("Book added successfully:", book)

	return nil // No error means success
}

func BorrowBook(c *fiber.Ctx, client *firestore.Client) error {
	bookID, _ := strconv.Atoi(c.Params("id"))
	sess, _ := middleware.GetSession(c)
	userID := sess.Get("userID")
	ctx := context.Background()

	book := GetOneBook(c, bookID)
	bookCopies, _ := GetCopiesOfBook(c, &book, true)
	if len(bookCopies) == 0 {
		return fmt.Errorf("nie ma dostepnej kopii")
	}

	if bookCopies[0].Available {
		// Add the entry to the approvalQueue collection
		_, _, err := client.Collection("approvalQueue").Add(ctx, map[string]interface{}{
			"user_id":          userID,
			"book_id":          bookID,
			"inventory_number": bookCopies[0].InventoryNumber,
		})
		if err != nil {
			return err
		}
		log.Println("Entry added to approvalQueue successfully")
	} else {
		log.Println("The book is not available")
	}

	return c.Render("bookdetails", fiber.Map{
		"Book":                   book,
		"NumberOfAvaliableBooks": len(bookCopies),
		"successMessage":         "Wysłano prośbę o wypożyczenie!",
	})

}
