package main

import (
	"context"
	"fmt"
	"my-firebase-project/helpers"
	"my-firebase-project/initializers"
	"my-firebase-project/workers"

	"net/http"
	"os"
)

// Example protected route
func protected(w http.ResponseWriter) {
	fmt.Fprintln(w, "This is a protected page. User is logged in.")
}

func init() {
	initializers.LoadEnvvariable()
	initializers.ConnectToDb(context.Background())
}

func main() {

	app := helpers.CreateApp()
	workers.SetWorkers()
	// localhost for remove a windowds defeder ask
	app.Listen("localhost:" + os.Getenv("PORT"))
	// controllers.SendEmail("", "Test", "Test body")
}
