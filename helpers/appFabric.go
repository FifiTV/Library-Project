package helpers

import (
	"my-firebase-project/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func CreateApp() *fiber.App {

	engine := html.New("views", ".html")
	engine.Reload(true)

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/", "./public")

	app.Use(middleware.SessionChecker)

	Routes(app)
	return app
}
