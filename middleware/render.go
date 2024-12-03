package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// Render renders a template with the provided data and adds user session flags.
//
// Parameters:
// - name (string): The name of the template to be rendered.
// - c (*fiber.Ctx): The Fiber context, used to manage the HTTP request and response.
// - data (fiber.Map): A map containing data to be passed to the template.
//
// Returns:
// - error: Any error encountered during the rendering process.
func Render(name string, c *fiber.Ctx, data fiber.Map) error {
	// Sprawdzenie czy jest zalogowany
	isLoggedIn := IsLogged(c)

	// Pobranie sesji użytkownika
	sess, err := GetSession(c)
	if err != nil {
		log.Printf("Error fetching session: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Session error")
	}

	// Pobranie danych z sesji
	userID, _ := sess.Get("userID").(int)
	userRole, _ := sess.Get("userRole").(int)

	if data == nil {
		data = fiber.Map{}
	}

	// Dodanie danych do widoku
	data["isLoggedIn"] = isLoggedIn
	data["userID"] = userID
	data["userRole"] = userRole

	log.Printf("Render called with userID: %d, userRole: %d, isLoggedIn: %v", userID, userRole, isLoggedIn)

	return c.Render(name, data)
}
