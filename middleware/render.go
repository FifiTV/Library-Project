package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// Render renders a template with the provided data and adds a login status flag.
//
// If you want to render something please, use this function.
// If you want to add some flags for use in multiple html templates, add them here.
//
// Parameters:
// - name (string): The name of the template to be rendered.
// - c (*fiber.Ctx): The Fiber context, used to manage the HTTP request and response.
// - data (fiber.Map): A map containing data to be passed to the template. If nil, an empty map is created.
//
// Returns:
// - error: Any error encountered during the rendering process.
func Render(name string, c *fiber.Ctx, data fiber.Map) error {

	// Sprawdzenie czy jest zalogowany
	isLoggedIn := IsLogged(c)

	if data == nil {
		data = fiber.Map{}
	}
	// Dodanie flagi isLoggedIn
	data["isLoggedIn"] = isLoggedIn

	return c.Render(name, data)
}
