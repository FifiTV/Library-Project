package helpers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func CreateApp() *fiber.App {

	engine := html.New("views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/", "./public")

	Routes(app)
	return app
}
