package controllers

import (
	"my-firebase-project/middleware"

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
