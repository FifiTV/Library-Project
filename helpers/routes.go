package helpers

import (
	"my-firebase-project/controllers"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"

	"github.com/gofiber/fiber/v2"
)

func Routes(app *fiber.App) {
	app.Get("/registration", controllers.GetRegistrationPage)
	app.Post("/registration", func(c *fiber.Ctx) error {
		return controllers.RegisterHandler(c, initializers.Client)
	})

	app.Get("/login", controllers.GetLoginPage)
	app.Post("/login", func(c *fiber.Ctx) error {
		return controllers.UserAuth(c, initializers.Client)
	})
	app.Get("/logout", middleware.AuthGuard, controllers.LogoutHandler)

	app.Get("/", controllers.GetMainPage)
	app.Get("/booklist", controllers.GetListBookPage)
	app.Get("/bookdetails/:id", controllers.GetBookDetailsPage)

	app.Get("/history", middleware.AuthGuard,
		middleware.RoleGuard(middleware.User),
		controllers.GetHistoryPage)

	app.Get("/addBook", controllers.GetAddBookPage)
	app.Post("/addBook", func(c *fiber.Ctx) error {
		return controllers.AddNewBookToLibrary(c, initializers.Client)
	})

	app.Post("/borrowBook/:id", func(c *fiber.Ctx) error {
		return controllers.BorrowBook(c, initializers.Client)
	})
}
