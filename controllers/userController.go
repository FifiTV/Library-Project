package controllers

import (
	"context"
	"fmt"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"
	"my-firebase-project/models"
	"regexp"
	"strconv"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/iterator"
)

func GetAllBorrowEvents(c *fiber.Ctx) []models.BorrowEvent {
	// Reference the "borrowEvents" collection
	borrowEventsCollection := initializers.Client.Collection("borrowEvents")
	// Get all documents in the collection
	docs, err := borrowEventsCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)
	}

	// Slice to store borrowEvents
	var borrowEvents []models.BorrowEvent

	// Loop through documents and decode into borrowEvents structs
	for _, doc := range docs {
		var borrowEvent models.BorrowEvent
		if err := doc.DataTo(&borrowEvent); err != nil {
			log.Printf("Error decoding document: %v", err)
		}
		borrowEvents = append(borrowEvents, borrowEvent)
	}

	// Return borrowEvents in JSON format
	return borrowEvents
}

func GetAllBorrowEventsForUser(c *fiber.Ctx, showCurrentOnly bool) ([]models.BorrowEventWithBook, error) {
	// Get all borrow events for the user
	borrowEvents := GetAllBorrowEvents(c)

	// Retrieve userID from session
	sess, _ := middleware.GetSession(c)
	userID := sess.Get("userID").(int)

	// Filter borrow events for the user
	var filteredBorrowEvents []models.BorrowEvent
	for _, event := range borrowEvents {
		if event.UserID == userID {
			// Apply additional filtering if `showCurrentOnly` is true
			if showCurrentOnly {
				if event.BorrowEnd.After(time.Now()) {
					filteredBorrowEvents = append(filteredBorrowEvents, event)
				}
			} else {
				filteredBorrowEvents = append(filteredBorrowEvents, event)
			}
		}
	}

	// Prepare the result by adding the book details to each borrow event
	var borrowEventsWithBooks []models.BorrowEventWithBook
	for _, event := range filteredBorrowEvents {
		// Fetch the book details by BookID
		book := GetOneBook(c, event.BookID)

		// Combine the borrow event with the book details
		borrowEventWithBook := models.BorrowEventWithBook{
			BorrowEvent: event,
			Book:        book,
		}

		// Add the combined data to the result
		borrowEventsWithBooks = append(borrowEventsWithBooks, borrowEventWithBook)
	}

	// Return the combined list
	return borrowEventsWithBooks, nil
}

func GetOneUser(c *fiber.Ctx, userId int) models.User {
	usersCollection := initializers.Client.Collection("users")

	docs, err := usersCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)

	}

	var userReturn models.User

	for _, doc := range docs {
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			log.Printf("Error decoding document: %v", err)
		}
		if user.Id == userId {
			userReturn = user
		}
	}

	return userReturn
}

func GetLibrarians(ctx context.Context, client *firestore.Client) ([]string, error) {
	// Pobierz użytkowników z rolą "Librarian"
	iter := client.Collection("users").Where("role", "==", 2).Documents(ctx)
	defer iter.Stop()

	var librarianIDs []string
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Error fetching librarian: %v", err)
			return nil, err
		}

		// Pobierz ID użytkownika
		id := doc.Ref.ID
		librarianIDs = append(librarianIDs, id)
	}

	return librarianIDs, nil
}

func sendReminders(c *fiber.Ctx) error {
	// Retrieve userID from session
	// sess, _ := middleware.GetSession(c)
	// userID := sess.Get("userID").(int)

	// Get user
	// user:= GetOneUser(c,userID)

	// Get his borrowEvents
	borrowEventsWithBooks, _ := GetAllBorrowEventsForUser(c, true)

	var titlesDueSoon []models.Book
	now := time.Now()

	for _, item := range borrowEventsWithBooks {
		if item.BorrowEvent.BorrowEnd.After(now) && item.BorrowEvent.BorrowEnd.Before(now.Add(7*24*time.Hour)) {
			titlesDueSoon = append(titlesDueSoon, item.Book)
		}
	}

	// fmt.Println("Books due within 7 days:", titlesDueSoon)
	// body :="You should return your books: ID"
	// SendEmail(user.Email,"You have 7 days left to read your books",body)
	return nil
}

func ExtendDate(c *fiber.Ctx) error {

	inventoryNumber, err := strconv.Atoi(c.Params("inventoryNumber"))
	if err != nil {
		return err
	}

	borrowEventsCollection := initializers.Client.Collection("borrowEvents")

	queryEvents := borrowEventsCollection.Where("inventory_number", "==", inventoryNumber).Limit(1)

	docEvents, err := queryEvents.Documents(context.Background()).GetAll()
	if err != nil {
		return nil
	}

	var event models.BorrowEvent
	if err := docEvents[0].DataTo(&event); err != nil {
		log.Printf("Error decoding document: %v", err)
	}

	if event.ExtendDate != 0 {

		newValue := event.ExtendDate - 1
		newDate := event.BorrowEnd.AddDate(0, 0, 7)

		if _, err = docEvents[0].Ref.Update(context.Background(), []firestore.Update{
			{
				Path:  "extend_date",
				Value: newValue,
			},
			{
				Path:  "borrow_end",
				Value: newDate,
			},
		}); err != nil {
			return nil
		}
	}

	return c.Redirect("/history")
}
func DeleteAccount(c *fiber.Ctx) error {
	// Pobierz e-mail z formularza
	email := c.FormValue("email")
	log.Printf("Podany e-mail: %s", email)

	// Sprawdź, czy e-mail został podany
	if email == "" {
		log.Println("Błąd: Nie podano e-maila")
		return c.Status(fiber.StatusBadRequest).Render("deleteAccount", fiber.Map{
			"errorMessage": "Proszę podać adres e-mail!",
		})
	}

	// Pobierz dane użytkownika z sesji
	sess, err := middleware.GetSession(c)
	if err != nil {
		log.Printf("Błąd pobierania sesji: %v", err)
		return c.Status(fiber.StatusInternalServerError).Render("deleteAccount", fiber.Map{
			"errorMessage": "Nie udało się pobrać danych sesji. Spróbuj ponownie później.",
		})
	}

	loggedInEmail, ok := sess.Get("email").(string)
	log.Printf("E-mail z sesji: %s", loggedInEmail)
	if !ok || loggedInEmail == "" {
		log.Println("Błąd: E-mail w sesji jest pusty")
		return c.Status(fiber.StatusUnauthorized).Render("deleteAccount", fiber.Map{
			"errorMessage": "Nie jesteś zalogowany. Zaloguj się, aby usunąć swoje konto.",
		})
	}

	// Porównaj podany e-mail z zalogowanym
	if email != loggedInEmail {
		log.Printf("Błąd: Podany e-mail (%s) różni się od zalogowanego (%s)", email, loggedInEmail)
		log.Println("Sesja NIE została zniszczona. Użytkownik pozostaje zalogowany.")
		return c.Render("deleteAccount", fiber.Map{
			"errorMessage": "Podany adres e-mail nie pasuje do zalogowanego użytkownika. Spróbuj ponownie.",
		})
	}

	// Pobierz użytkownika z Firestore na podstawie e-maila
	userDocs, err := initializers.Client.Collection("users").Where("email", "==", email).Documents(c.Context()).GetAll()
	if err != nil || len(userDocs) == 0 {
		log.Printf("Błąd: Nie znaleziono użytkownika w Firestore z e-mailem: %s", email)
		return c.Render("deleteAccount", fiber.Map{
			"errorMessage": "Nie znaleziono konta z tym adresem e-mail.",
		})
	}

	// Usuń konto użytkownika
	if _, err := userDocs[0].Ref.Delete(c.Context()); err != nil {
		log.Printf("Błąd: Nie udało się usunąć użytkownika z Firestore: %v", err)
		return c.Render("deleteAccount", fiber.Map{
			"errorMessage": "Wystąpił błąd podczas usuwania konta.",
		})
	}

	// Usuń sesję użytkownika tylko po poprawnym usunięciu konta
	if err := sess.Destroy(); err != nil {
		log.Printf("Błąd podczas niszczenia sesji: %v", err)
		return c.Status(fiber.StatusInternalServerError).Render("deleteAccount", fiber.Map{
			"errorMessage": "Wystąpił błąd podczas usuwania sesji. Konto zostało usunięte.",
		})
	}

	log.Println("Konto zostało pomyślnie usunięte.")
	return c.Redirect("/")
}

func ProposeNewBook(c *fiber.Ctx) error {
	bookPropositionsCollection := initializers.Client.Collection("bookPropositions")
	title := c.FormValue("title")
	author := c.FormValue("author")
	comment := c.FormValue("comment")
	sess, _ := middleware.GetSession(c)
	userID := sess.Get("userID").(int)
	newComment := models.Comment{
		UserID:  userID,
		Content: comment,
	}

	queryIsExisiting := bookPropositionsCollection.
		Where("title", "==", title).
		Where("author", "==", author)

	docs, err := queryIsExisiting.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)
	}

	if len(docs) == 1 {
		docID := docs[0].Ref.ID

		doc := docs[0]

		comments, ok := doc.Data()["comments"].([]interface{})
		if !ok {
			log.Printf("Error casting comments field")
		}

		var commentsCheck []models.Comment
		for _, comment := range comments {
			commentMap, ok := comment.(map[string]interface{})
			if !ok {
				continue
			}
			commentsCheck = append(commentsCheck, models.Comment{
				UserID:  int(commentMap["userId"].(int64)),
				Content: commentMap["content"].(string),
			})
		}

		for _, existingComment := range commentsCheck {
			if existingComment.UserID == userID {
				return middleware.Render("forms/proposeBook", c, fiber.Map{
					"errorMessage": "Już złożyłeś propozycję tej książki!",
				})
			}
		}
		_, err := bookPropositionsCollection.Doc(docID).Update(context.Background(), []firestore.Update{
			{
				Path:  "comments",
				Value: firestore.ArrayUnion(newComment),
			},
			{
				Path:  "upVotes",
				Value: firestore.Increment(1),
			},
		})
		if err != nil {
			return nil
		}
	} else {
		newProposedBook := models.ProposedBook{
			Title:    title,
			Author:   author,
			Comments: []models.Comment{newComment},
			UpVoutes: 1,
		}
		_, _, err := bookPropositionsCollection.Add(context.Background(), newProposedBook)
		if err != nil {
			log.Printf("Error adding new ProposalBook: %v", err)
		}
	}

	return c.Redirect("/proposeBook")
}

func GetEmailFormForResetPasswd(c *fiber.Ctx) error {
	return middleware.Render("forms/resetPasswdFormEmail", c, fiber.Map{})
}

func SetNewPasswd(c *fiber.Ctx) error {
	type RequestBody struct {
		Email string `json:"email"`
	}
	var body RequestBody
	if err := c.BodyParser(&body); err != nil {
		// fmt.Println("Error parsing request body:", err)
		return c.Status(400).SendString("Invalid request body")
	}

	userMail := body.Email
	if userMail == "" {
		// fmt.Println("Email not provided")
		return c.Status(400).SendString("Email is required")
	}

	err := SendResetPasswdEMail(userMail)
	if err != nil {
		fmt.Println("Error sending reset email:", err)
		return c.Status(500).SendString("Failed to send reset email")
	}

	return middleware.Render("pageAfterSendRequestForResetPasswd", c, fiber.Map{})
}

func GetResetPasswdForm(c *fiber.Ctx) error {
	mail := c.Params("id")
	return middleware.Render("forms/resetPasswd", c, fiber.Map{
		"Email": mail,
	})
}

func GetOneUserByMail(c *fiber.Ctx, mail string) models.User {
	usersCollection := initializers.Client.Collection("users")

	docs, err := usersCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)

	}

	var userReturn models.User

	for _, doc := range docs {
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			log.Printf("Error decoding document: %v", err)
		}
		if user.Email == mail {
			userReturn = user
		}
	}

	return userReturn
}

func updatePasswdForUser(c *fiber.Ctx, userId int, newPasswd string) error {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPasswd), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.Status(500).SendString("Internal Server Error")
	}

	usersCollection := initializers.Client.Collection("users")
	userQuery := usersCollection.Where("id", "==", userId).Limit(1)
	iter := userQuery.Documents(c.Context())

	doc, err := iter.Next()
	if err != nil {
		log.Printf("Error finding user with id %d: %v", userId, err)
		return c.Status(404).SendString("User not found")
	}

	userDocRef := doc.Ref

	_, err = userDocRef.Update(c.Context(), []firestore.Update{
		{
			Path:  "password",
			Value: string(hashedPassword),
		},
	})

	if err != nil {
		log.Printf("Error updating password in Firestore: %v", err)
		return c.Status(500).SendString("Error updating password")
	}

	return nil
}

func ResetPasswd(c *fiber.Ctx) error {
	mail := c.Params("id")
	newPasswd := c.FormValue("newPasswd")
	newPasswdR := c.FormValue("newPasswdR")

	if !(len(newPasswd) >= 8 &&
		regexp.MustCompile(`[0-9a-zA-Z]`).MatchString(newPasswd) &&
		regexp.MustCompile(`[!@#$%^&*]`).MatchString(newPasswd)) {

		return middleware.Render("forms/resetPasswd", c, fiber.Map{
			"Email":   mail,
			"Message": "Hasło nie spełnia wymagań!",
		})
	}

	user := GetOneUserByMail(c, mail)

	if newPasswd != newPasswdR {
		// Return a 400 status with a message that passwords do not match
		return middleware.Render("forms/resetPasswd", c, fiber.Map{
			"Email":   mail,
			"Message": "Hasła muszą być takie same!",
		})
	}

	updatePasswdForUser(c, user.Id, newPasswd)
	return middleware.Render("pageAfterChangingPasswd", c, fiber.Map{})
}
