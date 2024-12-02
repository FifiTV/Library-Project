package controllers

import (
	"fmt"
	"my-firebase-project/middleware"
	"strconv"

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
	return middleware.Render("forms/addBook", c, fiber.Map{})
}

func GetHistoryPage(c *fiber.Ctx) error {
	// Get the filtered borrow events for the user along with the book details
	borrowEventsWithBooks, err := GetAllBorrowEventsForUser(c)
	if err != nil {
		return err
	}

	return middleware.Render("history", c, fiber.Map{
		"Title":        "Historia wypożyczeń",
		"BorrowEvents": borrowEventsWithBooks,
	})
}
