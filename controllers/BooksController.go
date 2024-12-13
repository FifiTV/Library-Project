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
	// Pobierz ID książki z parametrów
	bookID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Invalid book ID: %v", err))
	}

	// Pobierz sesję użytkownika
	sess, err := middleware.GetSession(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve session")
	}

	// Pobierz userID z sesji
	userID, ok := sess.Get("userID").(int)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).SendString("User not logged in")
	}

	ctx := context.Background()

	// Pobierz informacje o książce
	book := GetOneBook(c, bookID)
	if book.Id == 0 {
		return c.Status(fiber.StatusNotFound).SendString("Book not found")
	}

	// Pobierz dostępne egzemplarze książki
	bookCopies, err := GetCopiesOfBook(c, &book, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve book copies")
	}
	if len(bookCopies) == 0 {
		return c.Status(fiber.StatusConflict).SendString("No available copies")
	}

	// Pobierz wszystkie dokumenty z kolekcji approvalQueue
	approvalQueueDocs, err := client.Collection("approvalQueue").Documents(ctx).GetAll()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve approval queue")
	}

	// Stwórz mapę istniejących numerów inwentarzowych w approvalQueue
	existingInventoryNumbers := make(map[int]bool)
	for _, doc := range approvalQueueDocs {
		data := doc.Data()
		if inventoryNumber, ok := data["inventory_number"].(int64); ok {
			existingInventoryNumbers[int(inventoryNumber)] = true
		} else {
			fmt.Printf("Skipping document with invalid or missing inventory_number: %+v\n", data)
		}
	}

	// Znajdź pierwszy egzemplarz książki, którego numer inwentarzowy nie jest w kolejce
	var selectedCopy *models.BookCopy
	for _, copy := range bookCopies {
		if !existingInventoryNumbers[copy.InventoryNumber] {
			selectedCopy = copy
			break
		}
	}
	if selectedCopy == nil {
		return c.Status(fiber.StatusConflict).SendString("No available copies not already in approval queue")
	}

	// Dodaj wybrany egzemplarz do kolejki
	_, _, err = client.Collection("approvalQueue").Add(ctx, map[string]interface{}{
		"user_id":          userID,
		"book_id":          bookID,
		"inventory_number": selectedCopy.InventoryNumber,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to add request to approvalQueue")
	}

	var doc *firestore.DocumentSnapshot

	// Create the query
	query := client.Collection("bookCopies").
		Where("inventory_number", "==", selectedCopy.InventoryNumber).
		Limit(1) // We expect only one document, so we limit the results to 1

	// Execute the query
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return fmt.Errorf("failed to get documents: %v", err)
	}

	// If no documents are found, return an error
	if len(docs) == 0 {
		return fmt.Errorf("no document found with inventory_number: %d", selectedCopy.InventoryNumber)
	}

	// Get the first document from the result (since we limited to 1)
	doc = docs[0]

	fmt.Printf("Document data: %+v\n", doc.Data())

	// Zaktualizuj pole available dla wybranego egzemplarza książki
	bookCopyDocRef := client.Collection("bookCopies").Doc(doc.Ref.ID) // Use the document ID of the selected copy
	_, err = bookCopyDocRef.Update(ctx, []firestore.Update{
		{
			Path:  "available",
			Value: false,
		},
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update book copy availability")
	}

	// // Jeśli książka jest dostępna, dodaj ją do kolejki
	if bookCopyDocRef != nil {
		_, _, err := client.Collection("approvalQueue").Add(ctx, map[string]interface{}{
			"user_id":          userID,
			"book_id":          bookID,
			"inventory_number": bookCopies[0].InventoryNumber,
		})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to add request to approvalQueue")
		}

		// Tworzenie powiadomienia dla użytkownika
		err = CreateNotification(
			fmt.Sprintf("%d", userID), // recipientId
			book.Title,                // bookTitle
			"Twoja prośba o wypożyczenie książki została wysłana.", // message
			1,     // role (użytkownik)
			false, // status
		)
		if err != nil {
			log.Printf("Error creating user notification: %v", err)
		}

		// Pobranie ID wszystkich bibliotekarzy
		librarians, err := GetLibrarians(ctx, client)
		if err != nil {
			log.Printf("Error fetching librarians: %v", err)
		}

		// Tworzenie powiadomienia dla każdego bibliotekarza
		for _, librarian := range librarians {
			err = CreateNotification(
				librarian,  // recipientId
				book.Title, // bookTitle
				fmt.Sprintf("Nowa prośba o wypożyczenie książki: %s.", book.Title),
				2,     // role (bibliotekarz)
				false, // status
			)
			if err != nil {
				log.Printf("Error creating librarian notification: %v", err)
			}
		}

		log.Println("Entry added to approvalQueue successfully")
	} else {
		log.Println("The book is not available")
	}

	return middleware.Render("bookdetails", c, fiber.Map{
		"Book":                   book,
		"NumberOfAvaliableBooks": len(bookCopies) - 1,
		"successMessage":         "Wysłano prośbę o wypożyczenie!",
	})
}
