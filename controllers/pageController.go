package controllers

import (
	"fmt"
	"my-firebase-project/middleware"
	"my-firebase-project/models"
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func GetMainPage(c *fiber.Ctx) error {
	sess, _ := middleware.GetSession(c)
	loginMessage := sess.Get("loginMessage")
	if loginMessage != nil {
		c.Locals("loginMessage", loginMessage) // Pass to template
		sess.Delete("loginMessage")            // Remove the message after use
		sess.Save()                            // Save session changes
	}
	return middleware.Render("index", c, fiber.Map{
		"loginMessage": loginMessage,
	})
}

func GetRegistrationPage(c *fiber.Ctx) error {
	return middleware.Render("registration", c, fiber.Map{})
}

func GetAboutUsPage(c *fiber.Ctx) error {
	return middleware.Render("aboutUs", c, fiber.Map{})
}

func GetLoginPage(c *fiber.Ctx) error {
	return middleware.Render("login", c, fiber.Map{})
}

func GetListBookPage(c *fiber.Ctx) error {

	books := GetAllBooks(c)

	return middleware.Render("booklist", c, fiber.Map{
		"Title": "BookList Page",
		"Books": books,
	})
}

func GetBookDetailsPage(c *fiber.Ctx) error {

	Id, err := strconv.Atoi(c.Params("id"))
	//decodedTitle, err := url.QueryUnescape(Title)
	if err != nil {
		// Handle the error if URL decoding fails
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error decoding title: %v", err))
	}

	book := GetOneBook(c, Id)
	books, err1 := GetCopiesOfBook(c, &book, true)
	if err1 != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Error during finding avaliable books: %v", err))
	}

	return middleware.Render("bookdetails", c, fiber.Map{
		"Title":                  "BookList Page",
		"Book":                   book,
		"NumberOfAvaliableBooks": len(books),
	})
}

func GetAddBookPage(c *fiber.Ctx) error {
	genres := GetGenres(c)
	return middleware.Render("forms/addBook", c, fiber.Map{
		"genres": genres,
	})
}

func GetProposeNewBookPage(c *fiber.Ctx) error {
	return middleware.Render("forms/proposeBook", c, fiber.Map{})
}

func GetHistoryPage(c *fiber.Ctx) error {
	// Check if the "show_current" query parameter is present
	showCurrentOnly := c.Query("show_current") == "true"

	// Get the filtered borrow events for the user along with the book details
	borrowEventsWithBooks, err := GetAllBorrowEventsForUser(c, showCurrentOnly)
	if err != nil {
		return err
	}

	// Sort borrowEventsWithBooks by the BorrowEnd field
	sort.Slice(borrowEventsWithBooks, func(i, j int) bool {
		return borrowEventsWithBooks[i].BorrowEvent.BorrowEnd.After(borrowEventsWithBooks[j].BorrowEvent.BorrowEnd)
	})

	now := time.Now()

	// Get user's approval queue items
	userApprovalItems := GetApprovalQueueItemsForUser(c)

	return middleware.Render("history", c, fiber.Map{
		"Title":         "Historia wypożyczeń",
		"BorrowEvents":  borrowEventsWithBooks,
		"CurrentTime":   now,
		"ApprovalItems": userApprovalItems,
	})
}

func GetApprovalQueuePage(c *fiber.Ctx) error {
	approvalItems := GetApprovalItems(c)

	return middleware.Render("approvalQueue", c, fiber.Map{
		"Title":         "Wypożyczenia do potwierdzenia",
		"ApprovalItems": approvalItems,
	})
}

func GetProposedBookItemsPage(c *fiber.Ctx) error {
	proposedBooksItems := ReturnProposedBooks(c)

	return middleware.Render("proposedBooksList", c, fiber.Map{
		"Title":             "Propozycje użytkowników",
		"ProposedBookItems": proposedBooksItems,
	})
}

func GetBooksToReturnPage(c *fiber.Ctx) error {
	booksToReturn := GetBooksToReturn(c)
	now := time.Now()

	return middleware.Render("booksToReturn", c, fiber.Map{
		"Title":         "Książki w wypożyczeniu",
		"BooksToReturn": booksToReturn,
		"CurrentTime":   now,
	})
}

func GetApprovalQueueItemsForUser(c *fiber.Ctx) []models.ApprovalItem {
	// Get the approval items to process
	approvalItems := GetApprovalItems(c)

	// Retrieve userID from session
	sess, _ := middleware.GetSession(c)
	userID := sess.Get("userID").(int)

	// Get the user details from the database or any other data source
	user := GetOneUser(c, userID)

	// Create a new slice to store the items that match the user's ID
	var userApprovalItems []models.ApprovalItem

	// Iterate through each approval item and check if it belongs to the current user
	for _, item := range approvalItems {
		if item.User.Id == user.Id {
			// If the user IDs match, add the item to the userApprovalItems collection
			userApprovalItems = append(userApprovalItems, item)
		}
	}

	// Return filtered approval items
	return userApprovalItems
}
