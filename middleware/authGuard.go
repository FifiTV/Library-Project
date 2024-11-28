package middleware

import (
	_ "fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

var store = session.New() // Utworzenie magazynu sesji

// Use this when you want to create a route, and the user must be logged in to get there
//
// Parameters:
//
// - c (*fiber.Ctx): The Fiber context, used to manage the HTTP request and response.
//
// Usage:
//
// app.Get("/logout", middleware.AuthGuard, controllers.LogoutHandler)
//
// Returns:
//
// - error: Any error encountered during the rendering process.
func AuthGuard(c *fiber.Ctx) error {
	if !IsLogged(c) {
		return c.Redirect("/login")
	}
	// Przejście do następnego handlera
	return c.Next()
}

func SessionChecker(c *fiber.Ctx) error {
	// Pobierz sesję użytkownika
	sess, err := store.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Session error")
	}

	// Sprawdź, czy użytkownik jest zalogowany
	isLoggedIn, ok := sess.Get("isLoggedIn").(bool)

	if ok && isLoggedIn {
		// Ustaw `isLoggedIn` w kontekście, aby inne trasy miały do niego dostęp
		c.Locals("isLoggedIn", true)
	} else {
		c.Locals("isLoggedIn", false)
	}

	return c.Next() // Przejście do następnego handlera
}

func IsLogged(c *fiber.Ctx) bool {
	isLoggedIn, ok := c.Locals("isLoggedIn").(bool)
	if !ok {
		isLoggedIn = false
	}
	return isLoggedIn
}
