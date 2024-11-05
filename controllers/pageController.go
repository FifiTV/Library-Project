package controllers

import "github.com/gofiber/fiber/v2"

func GetMainPage(c *fiber.Ctx) error {

	// Example how to send data into template
	// return c.Render("partials/index", fiber.Map{
	// 	"Title": "Aye!",
	// })

	return c.Render("partials/index", fiber.Map{})
}
