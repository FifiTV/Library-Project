package controllers

import (
	"context"
	"my-firebase-project/middleware"
	"my-firebase-project/models"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(c *fiber.Ctx, client *firestore.Client) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).Render("login", fiber.Map{
			"errorMessage": "Invalid request payload.",
		})
	}

	iter := client.Collection("users").Where("email", "==", user.Email).Documents(context.Background())
	doc, err := iter.Next()
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).Render("login", fiber.Map{
			"errorMessage": "Nie ma takiego użytkownika. Zarejestruj się.",
		})
	}

	var storedUser models.User
	if err := doc.DataTo(&storedUser); err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("login", fiber.Map{
			"errorMessage": "Błąd podczas połączenia.",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).Render("login", fiber.Map{
			"errorMessage": "Niepoprawne hasło.",
		})
	}

	sess, _ := middleware.GetSession(c)

	sess.Set("isLoggedIn", true)
	sess.Set("email", storedUser.Email)
	sess.Set("userID", storedUser.Id)
	sess.Set("userRole", storedUser.Role)
	sess.Set("loginMessage", "Udało Ci się zalogować!")
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("login", fiber.Map{
			"errorMessage": "Could not save session.",
		})
	}

	return c.Redirect("/")
}

func LogoutHandler(c *fiber.Ctx) error {
	sess, _ := middleware.GetSession(c)

	// Usuwanie sesji
	if err := sess.Destroy(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Could not destroy session")
	}

	return c.Redirect("/") // Powrót na stronę główną po wylogowaniu
}

func UserAuth(c *fiber.Ctx, client *firestore.Client) error {
	return LoginHandler(c, client)
}

func UserLogoff(c *fiber.Ctx) error {
	return LogoutHandler(c)
}
