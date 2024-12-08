package controllers

import (
	"context"
	"log"
	"my-firebase-project/models"
	"regexp"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(c *fiber.Ctx, client *firestore.Client) error {
	var user models.User

	// Parse request body
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).Render("registration", fiber.Map{
			"errorMessage": "Nieprawidłowe dane wejściowe.",
		})
	}

	// Parse birth_date
	dateOfBirth := c.FormValue("birth_date")
	if dateOfBirth == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "Data urodzenia jest wymagana.",
		})
	}
	parsedDate, err := time.Parse("2006-01-02", dateOfBirth)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "Nieprawidłowy format daty. Użyj RRRR-MM-DD.",
		})
	}
	user.BirthDate = parsedDate

	// Validate age
	currentDate := time.Now()
	age := currentDate.Year() - user.BirthDate.Year()
	if currentDate.YearDay() < user.BirthDate.YearDay() {
		age--
	}
	if age <= 0 || age > 120 {
		return c.Status(fiber.StatusBadRequest).Render("registration", fiber.Map{
			"errorMessage": "Wiek musi być większy niż 0 i mniejszy lub równy 120.",
		})
	}

	// Validate email
	if user.Email == "" || !isValidEmail(user.Email) {
		return c.Status(fiber.StatusBadRequest).Render("registration", fiber.Map{
			"errorMessage": "Nieprawidłowy lub brakujący email.",
		})
	}

	// Validate password
	if user.Password == "" || !isValidPassword(user.Password) {
		return c.Status(fiber.StatusBadRequest).Render("registration", fiber.Map{
			"errorMessage": "Nieprawidłowe hasło. Hasło musi mieć co najmniej 8 znaków.",
		})
	}

	// Check for existing user
	iter := client.Collection("users").Where("email", "==", user.Email).Documents(context.Background())
	_, docErr := iter.Next()
	if docErr == nil {
		return c.Status(fiber.StatusConflict).Render("registration", fiber.Map{
			"errorMessage": "Użytkownik z tym adresem email już istnieje.",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.Status(fiber.StatusInternalServerError).Render("registration", fiber.Map{
			"errorMessage": "Błąd podczas przetwarzania hasła.",
		})
	}
	user.Password = string(hashedPassword)

	// Generate new user ID
	user.Id, err = GetUserCountPlusOne(client)
	if err != nil {
		log.Printf("Failed to get user count: %v", err)
		return c.Status(fiber.StatusInternalServerError).Render("registration", fiber.Map{
			"errorMessage": "Błąd podczas generowania identyfikatora użytkownika.",
		})
	}

	// Save user
	_, _, err = client.Collection("users").Add(context.Background(), user)
	if err != nil {
		log.Printf("Firestore error: %v", err)
		return c.Status(fiber.StatusInternalServerError).Render("registration", fiber.Map{
			"errorMessage": "Nie udało się zarejestrować użytkownika.",
		})
	}

	return c.Render("index", fiber.Map{
		"successMessage": "Rejestracja zakończona sukcesem!",
	})
}

// Helper function to validate email
func isValidEmail(email string) bool {
	var emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return regexp.MustCompile(emailRegex).MatchString(email)
}

// Helper function to validate password
func isValidPassword(password string) bool {
	return len(password) >= 8 &&
		regexp.MustCompile(`[0-9a-zA-Z]`).MatchString(password) &&
		regexp.MustCompile(`[!@#$%^&*]`).MatchString(password)
}

// Helper function to retrieve the number of documents in the "users" collection.
func GetUserCountPlusOne(client *firestore.Client) (int, error) {

	ctx := context.Background()

	// Use Firestore's collection reference to get all documents
	users := client.Collection("users")

	// Get all documents to count them
	docs, err := users.Documents(ctx).GetAll()
	if err != nil {
		log.Fatalf("Failed to get documents: %v", err)
	}
	count := len(docs) + 1
	return count, nil
}
