package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"my-firebase-project/loginManager"

	"google.golang.org/api/iterator"
)

func getMain(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("pages/index.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	ctx := context.Background()
	app, err := initializeApp()
	if err != nil {
		http.Error(w, "Could not initialize app", http.StatusInternalServerError)
		return
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		http.Error(w, "Could not connect to Firestore", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	iter := client.Collection("books").Documents(ctx)
	var books []map[string]interface{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, "Error fetching documents: "+err.Error(), http.StatusInternalServerError)
			return
		}
		books = append(books, doc.Data())
	}

	tmpl.Execute(w, books)
}

// Example protected route
func protected(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "This is a protected page. User is logged in.")
}

func main() {
	ctx := context.Background()
	app, err := initializeApp()
	if err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		return
	}

	// Initialize Firebase Auth Client
	// authClient, err := app.Auth(ctx)
	// if err != nil {
	// 	fmt.Printf("Error initializing Auth client: %v\n", err)
	// 	return
	// }
	loginManager.SetAuthClient(app, ctx)

	http.HandleFunc("/", getMain)
	http.HandleFunc("/login", loginManager.Login)
	http.Handle("/protected", loginManager.LoginGuard(http.HandlerFunc(protected)))

	http.ListenAndServe(":8000", nil)
}
