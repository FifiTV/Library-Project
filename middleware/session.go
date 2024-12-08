package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

// Return a current session
//
// Parameters:
//
// - c (*fiber.Ctx): The Fiber context, used to manage the HTTP request and response.
//
// Returns:
//
// - session: A current session.
// - error: Any error encountered during the rendering process.
func GetSession(c *fiber.Ctx) (*session.Session, error) {
	sess, err := store.Get(c)
	if err != nil {
		return nil, c.Status(fiber.StatusInternalServerError).SendString("Session error")
	}
	return sess, nil
}
