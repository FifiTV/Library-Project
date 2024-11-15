package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

func GetSession(c *fiber.Ctx) (*session.Session, error) {
	sess, err := store.Get(c)
	if err != nil {
		return nil, c.Status(fiber.StatusInternalServerError).SendString("Session error")
	}
	return sess, nil
}
