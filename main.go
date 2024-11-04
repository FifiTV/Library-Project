package main

import (
	"context"
	"fmt"
	"my-firebase-project/initializers"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
)

// Example protected route
func protected(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is a protected page. User is logged in.")
}

func init() {
	initializers.LoadEnvvariable()
	initializers.ConnectToDb(context.Background())
}

func main() {
	app := fiber.New()

	Routes(app)
	app.Listen(":" + os.Getenv("PORT"))
}
