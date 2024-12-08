package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func Render(name string, c *fiber.Ctx, data fiber.Map) error {
	// Sprawdzenie czy jest zalogowany
	isLoggedIn := IsLogged(c)

	if data == nil {
		data = fiber.Map{}
	}

	// Dodanie danych do widoku
	data["isLoggedIn"] = isLoggedIn

	return c.Render(name, data)
}
