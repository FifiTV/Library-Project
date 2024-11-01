package loginManager

import (
	"context"
	"encoding/json"
	"html/template"
	"net/http"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gorilla/sessions"
)

var (
	authClient *auth.Client
	store      = sessions.NewCookieStore([]byte("your-secret-key")) // Use a secure random key in production
)

// func SetAuthClient(client *auth.Client) {
// 	authClient = client
// }

func SetAuthClient(app *firebase.App, ctx context.Context) {
	client, err := app.Auth(ctx)

	if err != nil {
		// fmt.Printf("Error initializing Auth client: %v\n", err)
		return
	}

	authClient = client
}

// check if the user is logged in
func LoginGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session-name") // Fetch the session
		if err != nil {
			http.Error(w, "Could not get session", http.StatusInternalServerError)
			return
		}

		// Check if the user ID is set in the session
		userID, ok := session.Values["userID"].(string)
		if !ok || userID == "" {
			// User is not logged in, redirect to login page
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r) // Call the next handler
	})
}

// Login function updated to set session on successful login
func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		// Use Firebase Authentication to sign in the user
		user, err := authClient.GetUserByEmail(context.Background(), req.Email)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Set user ID in session
		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, "Could not get session", http.StatusInternalServerError)
			return
		}
		session.Values["userID"] = user.UID // Store user ID
		err = session.Save(r, w)            // Save session
		if err != nil {
			http.Error(w, "Could not save session", http.StatusInternalServerError)
			return
		}

		// Respond with success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful", "uid": user.UID})
		return
	}

	// Display login form if not a POST request
	tmpl, err := template.ParseFiles("pages/login.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}
