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
	app.Post("/history/extendDate/:inventoryNumber", middleware.AuthGuard,
		middleware.RoleGuard(middleware.User),
		controllers.ExtendDate)

	app.Get("/approvalQueue", middleware.AuthGuard,
		middleware.RoleGuard(middleware.Librarian),
		controllers.GetApprovalQueuePage)
	app.Post("/approvalQueue/approved/:inventoryNumber/:bookID/:userID", middleware.AuthGuard,
		middleware.RoleGuard(middleware.Librarian),
		controllers.ChangeStatus)
	app.Post("/approvalQueue/rejected/:inventoryNumber/:bookID/:userID", middleware.AuthGuard,
		middleware.RoleGuard(middleware.Librarian),
		controllers.Cancel)

	app.Get("/addBook",
		middleware.AuthGuard,
		middleware.RoleGuard(middleware.Librarian),
		controllers.GetAddBookPage)
	app.Post("/addBook",
		middleware.AuthGuard,
		middleware.RoleGuard(middleware.Librarian),
		func(c *fiber.Ctx) error {
			return controllers.AddNewBookToLibrary(c, initializers.Client)
		})

	app.Post("/bookdetails/:id", middleware.AuthGuard, func(c *fiber.Ctx) error {
		return controllers.BorrowBook(c, initializers.Client)
	})
	app.Get("/notifications", controllers.FetchNotifications)
	app.Get("/add-test-notifications", controllers.AddTestNotifications)

	app.Get("/api/next-inventory-number", controllers.GetNextInventoryNumber)
	app.Post("/api/score-book/:user/:book", controllers.ScoreBook)
	app.Get("/api/get-score-book/:user/:book", controllers.GetBookScore)
	// Add here routes
	//

	// Handle 404 errors (Not Found)
	app.Use(func(c *fiber.Ctx) error {
		return controllers.GetError404Page(c)
	})
}
