package middleware

import (
	"errors"
	_ "fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	Banned    int = 0
	User      int = 1
	Librarian int = 2
)

// getUserRole retrieves the user's role from the session.
// This function fetches the session associated with the current request and extracts the user's role.
// It returns the role as an integer or an error if the role is not found or invalid.
//
// Parameters:
//   - c: *fiber.Ctx - The Fiber context, which holds information about the current HTTP request.
//
// Returns:
//   - int: The user's role as an integer.
//   - error: An error if the role is not found or cannot be cast to an integer.
func getUserRole(c *fiber.Ctx) (int, error) {
	sess, _ := GetSession(c)
	userRole, ok := sess.Get("userRole").(int)
	if !ok {
		// Return an error with a meaningful message
		return -1, errors.New("Nie znaleziono roli użytkownika, lub jest nieprawidłowa")
	}
	return userRole, nil
}

// RoleGuard is a middleware that restricts access to routes based on user roles.
// It checks if the user's role (retrieved from the session) meets the minimum required role for accessing the route.
//
// Parameters:
//   - role: int - The required role level for the route. For example:
//   - middleware.Banned
//   - middleware.User
//   - middleware.Librarian
//
// Usage:
//   - app.Get("/history", middleware.AuthGuard, middleware.RoleGuard(middleware.User), controllers.GetHistoryPage)
//
// Behavior:
//   - If the user's role is not found or invalid, it returns a 403 Forbidden status with an error message.
//   - If the user's role is lower than the required role, access is denied with a 403 status.
//   - If the user has sufficient permissions, the middleware calls the next handler in the chain.
func RoleGuard(role int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		actualRole, err := getUserRole(c)
		if err != nil {
			return c.Status(fiber.StatusForbidden).Render("errors/error", fiber.Map{
				"errorMessage": "Wystąpił błąd.",
			})
		}

		if actualRole < role {
			return c.Status(fiber.StatusForbidden).Render("errors/error", fiber.Map{
				"errorMessage": "Użytkownik nie ma dostępu do zasobu.",
			})
		}

		return c.Next()
	}
}
