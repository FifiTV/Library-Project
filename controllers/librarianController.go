package controllers

import (
	"context"
	"fmt"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"
	"my-firebase-project/models"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	_ "golang.org/x/text/cases"
)

func GetCountOfRecords(c *fiber.Ctx, client *firestore.Client, collection string, key string, value string) int {
	countQuery := client.Collection(collection).
		Where(key, "==", value).
		Select() // No fields needed; we just need the count.

	// Perform the count aggregation
	agg, err := countQuery.Documents(c.Context()).GetAll()
	if err != nil {
		return -1
	}
	return len(agg)
}

func AddNewBookToLibrary(c *fiber.Ctx, client *firestore.Client) error {
	// var book models.Book
	// var bookCopy models.BookCopy
	title := c.FormValue("title")
	isNewBook := c.FormValue("newBook") == "on"
	isBookExists := GetCountOfRecords(c, client, "books", "title", title)

	if strings.TrimSpace(title) == "" {
		return c.Status(fiber.StatusBadRequest).Render("forms/addBook", fiber.Map{
			"errorMessage": "Proszę podać tytuł!",
		})
	}

	title = strings.Title(title)
	book, _ := GetBookByTitle(c, client, title)

	if isNewBook && isBookExists == 0 && book == nil {
		book = &models.Book{}
		book.Title = title

		author := c.FormValue("author")
		book.Author = author

		pagess := c.FormValue("pages")
		pages, _ := strconv.Atoi(pagess)
		book.Pages = pages

		//publishedAt
		publishedAt := c.FormValue("publishedAt")
		const dateFormat = "2006-01-02"
		parsedTime, _ := time.Parse(dateFormat, publishedAt)
		book.PublishedAt = parsedTime

		// description
		desc := c.FormValue("description")
		book.Description = desc

		//Genre
		genre := c.FormValue("genre")
		book.Genre = genre

		url := c.FormValue("coverLink")
		book.Cover = url

		err := AddBook(c, client, book)
		if err != nil {
			return middleware.Render("forms/addBook", c, fiber.Map{
				"errorMessage": "Wprowadź poprawne dane!",
			})
		}
	} else if book == nil {
		fmt.Println("Copy")
		return middleware.Render("forms/addBook", c, fiber.Map{
			"errorMessage": "Musisz dodać tę książkę do zbioru!",
		})
	}
	fmt.Println("Copy")
	//Check the logic of this part
	// if book.title is correct
	if isNewBook && isBookExists > 0 {
		return middleware.Render("forms/addBook", c, fiber.Map{
			"errorMessage": "Ta książka już jest dodana do bazy! Musisz dodać egemplarz!",
		})
	}
	fmt.Println("Copy")
	// Add Copy
	var bookCopy models.BookCopy

	inventoryNumber, _ := strconv.Atoi(c.FormValue("inventoryNumber"))

	bookCopy.AddedOn = time.Now()
	bookCopy.Available = true
	bookCopy.BookID = book.Id
	bookCopy.InventoryNumber = inventoryNumber

	bookCopy.Location = c.FormValue("location")

	errABC := AddBookCopy(c, client, &bookCopy)
	if errABC != nil {
		return middleware.Render("forms/addBook", c, fiber.Map{
			"errorMessage": errABC.Error(),
		})
	}
	return c.Redirect("/addBook")

}

func GetApprovalItems(c *fiber.Ctx) []models.ApprovalItem {
	approvalQueueCollection := initializers.Client.Collection("approvalQueue")

	docs, err := approvalQueueCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)
	}

	var approvalQueue []models.ApprovalQueue

	for _, doc := range docs {
		var approval models.ApprovalQueue
		if err := doc.DataTo(&approval); err != nil {
			log.Printf("Error decoding document: %v", err)
		}
		approvalQueue = append(approvalQueue, approval)
	}

	var approvalItems []models.ApprovalItem

	for _, item := range approvalQueue {
		book := GetOneBook(c, item.BookID)
		user := GetOneUser(c, item.UserID)
		bookCopy := GetCopyOfBook(c, item.InventoryNumber)

		approvalItem := models.ApprovalItem{
			ApprovalQueue: item,
			Book:          book,
			BookCopy:      bookCopy,
			User:          user,
		}

		approvalItems = append(approvalItems, approvalItem)
	}

	// Return borrowEvents in JSON format
	return approvalItems

}

func ChangeStatus(c *fiber.Ctx) error {

	inventoryNumber, err := strconv.Atoi(c.Params("inventoryNumber"))
	if err != nil {
		return err
	}
	bookID, err := strconv.Atoi(c.Params("bookID"))
	if err != nil {
		return err
	}
	userID, err := strconv.Atoi(c.Params("userID"))
	if err != nil {
		return err
	}

	newBorrowEvent := models.BorrowEvent{

		UserID:          userID,
		InventoryNumber: inventoryNumber,
		BookID:          bookID,
		BorrowStart:     time.Now(),
		BorrowEnd:       time.Now().AddDate(0, 1, 0),
	}

	approvalQueueCollection := initializers.Client.Collection("approvalQueue")
	bookCopiesCollection := initializers.Client.Collection("bookCopies")
	borrowEventsCollection := initializers.Client.Collection("borrowEvents")

	queryCopies := bookCopiesCollection.Where("inventory_number", "==", inventoryNumber).Limit(1)
	queryApproval := approvalQueueCollection.Where("inventory_number", "==", inventoryNumber).Limit(1)

	docCopies, err := queryCopies.Documents(context.Background()).GetAll()
	if err != nil {
		return nil
	}

	docApproval, err := queryApproval.Documents(context.Background()).GetAll()
	if err != nil {
		return nil
	}

	if _, err = docCopies[0].Ref.Update(context.Background(), []firestore.Update{
		{
			Path:  "available",
			Value: false,
		},
	}); err != nil {
		return nil
	}

	_, _, err = borrowEventsCollection.Add(context.Background(), newBorrowEvent)
	if err != nil {
		return nil
	}

	if _, err = docApproval[0].Ref.Delete(context.Background()); err != nil {
		return nil
	}

	return c.Redirect("/approvalQueue")
}

func Cancel(c *fiber.Ctx) error {

	inventoryNumber, err := strconv.Atoi(c.Params("inventoryNumber"))
	if err != nil {
		return err
	}

	approvalQueueCollection := initializers.Client.Collection("approvalQueue")

	queryApproval := approvalQueueCollection.Where("inventory_number", "==", inventoryNumber).Limit(1)

	docApproval, err := queryApproval.Documents(context.Background()).GetAll()
	if err != nil {
		return nil
	}

	if _, err = docApproval[0].Ref.Delete(context.Background()); err != nil {
		return nil
	}

	return c.Redirect("/approvalQueue")
}
