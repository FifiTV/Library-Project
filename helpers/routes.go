package helpers

import (
	"my-firebase-project/controllers"

	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App) {
	// app.Get("/", controllers.GetAllBooks)
	app.Get("/", controllers.GetMainPage)
}
