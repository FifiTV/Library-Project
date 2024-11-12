package helpers

import (
	"my-firebase-project/controllers"
	"my-firebase-project/initializers"

	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App) {
	app.Get("/registration", controllers.GetRegistrationPage)
	app.Post("/registration", func(c *fiber.Ctx) error {
		return controllers.RegisterHandler(c, initializers.Client)
	})

	app.Get("/login", controllers.GetLoginPage)
	app.Post("/login", func(c *fiber.Ctx) error {
		return controllers.LoginHandler(c, initializers.Client)
	})

	app.Get("/", controllers.GetMainPage)
}
