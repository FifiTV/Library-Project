package controllers

import (
	"context"
	"fmt"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"
	"my-firebase-project/models"
	"sort"
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
		Select()

	agg, err := countQuery.Documents(c.Context()).GetAll()
	if err != nil {
		return -1
	}
	return len(agg)
}

func AddNewBookToLibrary(c *fiber.Ctx, client *firestore.Client) error {
	title := c.FormValue("title")
	isNewBook := c.FormValue("newBook") == "on"
	isBookExists := GetCountOfRecords(c, client, "books", "title", title)
	title = strings.TrimSpace(title)
	if title == "" {
		return middleware.Render("forms/addBook", c, fiber.Map{
			"errorMessage": "Proszę podać tytuł!",
		})
	}

	title = strings.Title(title)

	var book *models.Book
	if isBookExists == 0 {
		book = &models.Book{
			Title:       title,
			Author:      c.FormValue("author"),
			Description: c.FormValue("description"),
			Genre:       c.FormValue("genre"),
			Cover:       c.FormValue("coverLink"),
		}

		pagess := c.FormValue("pages")
		pages, _ := strconv.Atoi(pagess)
		book.Pages = pages

		//publishedAt
		publishedAt := c.FormValue("publishedAt")
		const dateFormat = "2006-01-02"
		parsedTime, _ := time.Parse(dateFormat, publishedAt)
		book.PublishedAt = parsedTime

		err := AddBook(c, client, book)
		if err != nil {
			return middleware.Render("forms/addBook", c, fiber.Map{
				"errorMessage": "Wprowadź poprawne dane!",
			})
		}
	} else {
		book, _ = GetBookByTitle(c, client, title)
	}

	if isNewBook && isBookExists > 0 {
		return middleware.Render("forms/addBook", c, fiber.Map{
			"errorMessage": "Ta książka już jest dodana do bazy! Musisz dodać egemplarz!",
		})
	}
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
		ExtendDate:      1,
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
	bookTitle := GetOneBook(c, bookID).Title

	err = CreateNotification(
		strconv.Itoa(userID),
		bookTitle,
		fmt.Sprintf("Twoja prośba o wypozyczenie książki: %s została zaakceptowana.", bookTitle),
		1,
		false,
	)
	if err != nil {
		log.Printf("Error creating user notification: %v", err)
	}

	return c.Redirect("/approvalQueue")
}

func Cancel(c *fiber.Ctx) error {

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

	approvalQueueCollection := initializers.Client.Collection("approvalQueue")
	bookCopiesCollection := initializers.Client.Collection("bookCopies")

	queryApproval := approvalQueueCollection.Where("inventory_number", "==", inventoryNumber).Limit(1)
	queryCopies := bookCopiesCollection.Where("inventory_number", "==", inventoryNumber).Limit(1)

	docCopies, err := queryCopies.Documents(context.Background()).GetAll()
	if err != nil {
		return nil
	}

	if _, err = docCopies[0].Ref.Update(context.Background(), []firestore.Update{
		{
			Path:  "available",
			Value: true,
		},
	}); err != nil {
		return nil
	}

	docApproval, err := queryApproval.Documents(context.Background()).GetAll()
	if err != nil {
		return nil
	}

	if _, err = docApproval[0].Ref.Delete(context.Background()); err != nil {
		return nil
	}
	bookTitle := GetOneBook(c, bookID).Title

	err = CreateNotification(
		strconv.Itoa(userID),
		bookTitle,
		fmt.Sprintf("Twoja prośba o wypozyczenie książki: %s została odrzucona.", bookTitle),
		1,
		false,
	)
	if err != nil {
		log.Printf("Error creating user notification: %v", err)
	}

	return c.Redirect("/approvalQueue")
}

func GetBooksToReturn(c *fiber.Ctx) []models.ReturnBook {
	borrowEventsCollection := initializers.Client.Collection("borrowEvents")

	docs, err := borrowEventsCollection.Documents(context.Background()).GetAll()
	if err != nil {
		return nil
	}

	var borrowEvents []models.BorrowEvent

	for _, doc := range docs {
		var event models.BorrowEvent
		if err := doc.DataTo(&event); err != nil {
			log.Printf("Error decoding document: %v", err)
		}

		borrowEvents = append(borrowEvents, event)
	}

	var booksToReturn []models.ReturnBook

	for _, item := range borrowEvents {
		book := GetOneBook(c, item.BookID)
		user := GetOneUser(c, item.UserID)
		bookCopy := GetCopyOfBook(c, item.InventoryNumber)

		bookToReturn := models.ReturnBook{
			BorrowEvent: item,
			Book:        book,
			BookCopy:    bookCopy,
			User:        user,
		}

		booksToReturn = append(booksToReturn, bookToReturn)
	}

	return booksToReturn

}

func BookReturned(c *fiber.Ctx) error {
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

	borrowEventsCollection := initializers.Client.Collection("borrowEvents")
	bookCopiesCollection := initializers.Client.Collection("bookCopies")

	queryCopies := bookCopiesCollection.Where("inventory_number", "==", inventoryNumber).Limit(1)

	docCopies, err := queryCopies.Documents(context.Background()).GetAll()
	if err != nil {
		return nil
	}

	if _, err = docCopies[0].Ref.Update(context.Background(), []firestore.Update{
		{
			Path:  "available",
			Value: true,
		},
	}); err != nil {
		return nil
	}

	queryEvents := borrowEventsCollection.Where("inventory_number", "==", inventoryNumber)

	docEvents, err := queryEvents.Documents(context.Background()).GetAll()
	if err != nil {
		return nil
	}

	if _, err = docEvents[0].Ref.Delete(context.Background()); err != nil {
		return nil
	}

	book := GetOneBook(c, bookID)
	message := "Dziękujemy za oddanie książki " + book.Title

	err = CreateNotification(
		strconv.Itoa(userID),
		book.Title,
		message,
		1,
		false,
	)
	if err != nil {
		log.Printf("Error creating user notification: %v", err)
	}

	return c.Redirect("/booksToReturn")
}

func ReturnProposedBooks(c *fiber.Ctx) []models.PropsedBookItem {
	bookPropositionsCollection := initializers.Client.Collection("bookPropositions")

	docs, err := bookPropositionsCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Println("Error reading documents: %v", err)
	}

	var proposeBooks []models.ProposedBook

	for _, doc := range docs {
		var propBook models.ProposedBook
		if err := doc.DataTo(&propBook); err != nil {
			log.Println("Error decoding document: %v", err)
		}
		proposeBooks = append(proposeBooks, propBook)
	}

	var proposeBookItems []models.PropsedBookItem

	for _, item := range proposeBooks {
		var commentItems []models.CommentItem
		for _, com := range item.Comments {
			user := GetOneUser(c, com.UserID)
			commentItem := models.CommentItem{
				UserName:     user.FirstName,
				UserLastName: user.LastName,
				Content:      com.Content,
			}

			commentItems = append(commentItems, commentItem)
		}
		proposedBookItem := models.PropsedBookItem{
			Title:    item.Title,
			Author:   item.Author,
			UpVotes:  item.UpVoutes,
			Comments: commentItems,
		}
		proposeBookItems = append(proposeBookItems, proposedBookItem)
	}
	sort.Slice(proposeBookItems, func(i, j int) bool {
		return proposeBookItems[i].UpVotes > proposeBookItems[j].UpVotes
	})

	return proposeBookItems
}
