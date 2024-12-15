package controllers

import (
	"my-firebase-project/middleware"

	"github.com/gofiber/fiber/v2"
)

func GetError404Page(c *fiber.Ctx) error {
	return middleware.Render("errors/error", c, fiber.Map{
		"errorMessage": "Nie udało nam się znaleźć podstrony.",
		"errorCode":    404,
	})
}
