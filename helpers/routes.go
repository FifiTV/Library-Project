package helpers

import (
	"my-firebase-project/controllers"

	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App) {
	app.Get("/registration", controllers.GetRegistrationPage)
	app.Get("/", controllers.GetMainPage)
	app.Get("/booklist", controllers.GetListBookPage)

}
