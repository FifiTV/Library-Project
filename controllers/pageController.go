package controllers

import (
	"github.com/gofiber/fiber/v2"
)

func GetMainPage(c *fiber.Ctx) error {

	// Example how to send data into template
	// return c.Render("partials/index", fiber.Map{
	// 	"Title": "Aye!",
	// })

	return c.Render("partials/index", fiber.Map{})
}

func GetRegistrationPage(c *fiber.Ctx) error {
	return c.Render("registration", fiber.Map{})
}

func GetListBookPage(c *fiber.Ctx) error {

	books := GetAllBooks(c)
	//fmt.Print(books)

	return c.Render("booklist", fiber.Map{
		"Title": "BookList Page",
		"Books": books,
	})
}
