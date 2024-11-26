package controllers

import (
	"context"
	"my-firebase-project/models"
	"regexp"

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

	// Validate age
	if user.Age <= 0 || user.Age > 120 {
		return c.Status(fiber.StatusBadRequest).Render("registration", fiber.Map{
			"errorMessage": "Wiek musi być większy niż 0 i mniejszy niż 120.",
		})
	}

	// Validate email
	if user.Email == "" {
		return c.Status(fiber.StatusBadRequest).Render("registration", fiber.Map{
			"errorMessage": "Email jest wymagany.",
		})
	}
	if !isValidEmail(user.Email) {
		return c.Status(fiber.StatusBadRequest).Render("registration", fiber.Map{
			"errorMessage": "Nieprawidłowy format email.",
		})
	}

	// Validate password
	if user.Password == "" {
		return c.Status(fiber.StatusBadRequest).Render("registration", fiber.Map{
			"errorMessage": "Hasło jest wymagane.",
		})
	}
	if !isValidPassword(user.Password) {
		return c.Status(fiber.StatusBadRequest).Render("registration", fiber.Map{
			"errorMessage": "Hasło musi zawierać co najmniej 8 znaków, jedną cyfrę, jeden znak specjalny, jedną małą i jedną wielką literę.",
		})
	}

	// Check if the user already exists
	iter := client.Collection("users").Where("email", "==", user.Email).Documents(context.Background())
	_, err := iter.Next()
	if err == nil {
		return c.Status(fiber.StatusConflict).Render("registration", fiber.Map{
			"errorMessage": "Użytkownik z tym adresem email już istnieje.",
		})
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("registration", fiber.Map{
			"errorMessage": "Błąd podczas tworzenia konta.",
		})
	}
	user.Password = string(hashedPassword)

	// Save user in Firestore
	_, _, err = client.Collection("users").Add(context.Background(), user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("registration", fiber.Map{
			"errorMessage": "Nie udało się zarejestrować użytkownika.",
		})
	}

	// Success: Render the confirmation page or redirect to login
	return c.Render("index", fiber.Map{
		"successMessage": "Rejestracja zakończona sukcesem! Możesz się teraz zalogować.",
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
		regexp.MustCompile(`[0-9]`).MatchString(password) &&
		regexp.MustCompile(`[!@#$%^&*]`).MatchString(password) &&
		regexp.MustCompile(`[a-z]`).MatchString(password) &&
		regexp.MustCompile(`[A-Z]`).MatchString(password)
}
