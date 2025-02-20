package controllers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"
	"my-firebase-project/models"
	"net/http"
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
			user.FirestoreDocID = doc.Ref.ID
			userReturn = user
		}
	}

	return userReturn
}

func GetLibrarians(ctx context.Context, client *firestore.Client) ([]string, error) {
	log.Println("Rozpoczynanie pobierania bibliotekarzy...")

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
			log.Printf("Błąd pobierania dokumentu użytkownika: %v", err)
			return nil, err
		}

		// Wyświetl dane dokumentu dla debugowania
		/*log.Printf("Przetwarzanie dokumentu ID: %s", doc.Ref.ID)
		log.Printf("Dane użytkownika: %v", doc.Data())*/

		// Pobierz pole "id" (userID) z dokumentu
		userID, ok := doc.Data()["id"].(int64) // Lub float64 w zależności od struktury danych
		if !ok {
			log.Printf("Nie udało się pobrać userID z dokumentu: %s", doc.Ref.ID)
			continue
		}

		// Dodaj `userID` do listy
		librarianIDs = append(librarianIDs, fmt.Sprintf("%d", userID))
	}

	log.Printf("Iteracja zakończona. Znaleziono %d bibliotekarzy.", len(librarianIDs))
	log.Printf("Lista bibliotekarzy: %v", librarianIDs)
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
		// Pobierz tytuł książki i dane użytkownika
		book := GetOneBook(c, event.BookID)
		user := GetOneUser(c, event.UserID)

		// Wyślij powiadomienie do użytkownika
		err := CreateNotification(
			strconv.Itoa(event.UserID),
			book.Title, // Tytuł książki
			"Przedłużyłeś termin oddania książki o tydzień.",
			1, // Powiadomienie dla użytkownika
			false,
		)
		if err != nil {
			log.Printf("Błąd podczas tworzenia powiadomienia dla użytkownika: %v", err)
		}

		// Pobierz bibliotekarzy
		librarians, err := GetLibrarians(context.Background(), initializers.Client)
		if err != nil {
			log.Printf("Błąd podczas pobierania bibliotekarzy: %v", err)
		} else {
			// Wyślij powiadomienie do każdego bibliotekarza
			for _, librarianID := range librarians {
				err := CreateNotification(
					librarianID,
					book.Title, // Tytuł książki
					fmt.Sprintf("Użytkownik %s przedłużył termin oddania książki o tydzień.", user.Email),
					2, // Powiadomienie dla bibliotekarzy
					false,
				)
				if err != nil {
					log.Printf("Błąd podczas tworzenia powiadomienia dla bibliotekarza: %v", err)
				}
			}
		}
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errorMessage": "Nie można przedłużyć terminu oddania książki. Limit został wyczerpany.",
		})
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
		// Pobierz bibliotekarzy
		librarians, err := GetLibrarians(context.Background(), initializers.Client)
		if err != nil {
			log.Printf("Błąd podczas pobierania bibliotekarzy: %v", err)
		} else {
			// Wyślij powiadomienie do każdego bibliotekarza
			for _, librarianID := range librarians {
				err := CreateNotification(
					librarianID,
					fmt.Sprintf("Dodano nową propozycję książki: %s", title),
					fmt.Sprintf("Użytkownik zgłosił propozycję książki '%s' autorstwa '%s'.", title, author),
					2, // Powiadomienie dla bibliotekarzy
					false,
				)
				if err != nil {
					log.Printf("Błąd podczas tworzenia powiadomienia dla bibliotekarza: %v", err)
				}
			}
		}
	}

	return c.Redirect("/proposeBook")
}

func GetEmailFormForResetPasswd(c *fiber.Ctx) error {
	return middleware.Render("forms/resetPasswdFormEmail", c, fiber.Map{})
}

func createResetPasswdEvent(c *fiber.Ctx, mail string) (string, error) {
	user, err := GetOneUserByMail(c, mail)
	if err != nil {
		return "", fmt.Errorf("user with email %s not found: %v", mail, err)
	}

	usersCollection := initializers.Client.Collection("users")
	userDocRef := usersCollection.Doc(user.FirestoreDocID) // This gives you the document reference
	if userDocRef == nil {
		return "", fmt.Errorf("could not get document reference for user with email %s", mail)
	}

	keyValue := generateRandomString(32)

	resetEvent := models.ResetPasswdEvent{
		UserId:   user.Id,
		Start:    time.Now(),
		End:      time.Now().Add(10 * time.Minute),
		KeyValue: keyValue,
	}

	eventsCollection := initializers.Client.Collection("resetPasswdEvent")
	_, _, err1 := eventsCollection.Add(context.Background(), resetEvent)
	if err1 != nil {
		return "", fmt.Errorf("error saving reset password event: %v", err1)
	}

	return keyValue, nil
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
	hash, err := createResetPasswdEvent(c, userMail)
	if err != nil {
		fmt.Println("Error during creating Reset Password Event:", err)
		return c.Status(500).SendString("Failed to send reset email")
	}

	err = SendResetPasswdEMail(userMail, hash)
	if err != nil {
		fmt.Println("Error sending reset email:", err)
		return c.Status(500).SendString("Failed to send reset email")
	}

	return middleware.Render("pageAfterSendRequestForResetPasswd", c, fiber.Map{})
}

func GetResetPasswdForm(c *fiber.Ctx) error {
	randKey := c.Params("randkey")
	return middleware.Render("forms/resetPasswd", c, fiber.Map{
		"RandKey": randKey,
	})
}

func GetOneUserByMail(c *fiber.Ctx, mail string) (models.User, error) {
	usersCollection := initializers.Client.Collection("users")

	docs, err := usersCollection.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)
		return models.User{}, fmt.Errorf("failed to get documents: %w", err)
	}

	var userReturn models.User

	for _, doc := range docs {
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			log.Printf("Error decoding document: %v", err)
			continue
		}

		if user.Email == mail {
			user.FirestoreDocID = doc.Ref.ID
			userReturn = user
			break
		}
	}

	if userReturn.FirestoreDocID == "" {
		return models.User{}, fmt.Errorf("user with email %s not found", mail)
	}

	return userReturn, nil
}

// func updatePasswdForUser(c *fiber.Ctx, userId int, newPasswd string) error {
// 	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPasswd), bcrypt.DefaultCost)
// 	if err != nil {
// 		log.Printf("Error hashing password: %v", err)
// 		return c.Status(500).SendString("Internal Server Error")
// 	}
// 	usersCollection := initializers.Client.Collection("users")
// 	userQuery := usersCollection.Where("id", "==", userId).Limit(1)
// 	iter := userQuery.Documents(c.Context())
// 	doc, err := iter.Next()
// 	if err != nil {
// 		log.Printf("Error finding user with id %d: %v", userId, err)
// 		return c.Status(404).SendString("User not found")
// 	}
// 	userDocRef := doc.Ref
// 	_, err = userDocRef.Update(c.Context(), []firestore.Update{
// 		{
// 			Path:  "password",
// 			Value: string(hashedPassword),
// 		},
// 	})
// 	if err != nil {
// 		log.Printf("Error updating password in Firestore: %v", err)
// 		return c.Status(500).SendString("Error updating password")
// 	}

// 	return nil
// }

func updatePasswdForUser(c *fiber.Ctx, resetPasswdEvent models.ResetPasswdEvent, newPasswd string) error {
	// Hash the new password before storing it
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPasswd), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return c.Status(500).SendString("Internal Server Error")
	}

	userId := resetPasswdEvent.UserId
	if userId == 0 {
		log.Printf("Invalid user ID in reset password event")
		return c.Status(400).SendString("Invalid reset password event")
	}

	user := GetOneUser(c, userId)
	if err != nil {
		log.Printf("Error fetching user with ID %d: %v", userId, err)
		return c.Status(500).SendString("Error fetching user details")
	}

	usersCollection := initializers.Client.Collection("users")
	userDocRef := usersCollection.Doc(user.FirestoreDocID)
	fmt.Println(user)
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

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano()) // No casting needed
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func getResetPasswdEventByKeyValue(c *fiber.Ctx, keyValue string) (models.ResetPasswdEvent, error) {
	eventsCollection := initializers.Client.Collection("resetPasswdEvent")

	query := eventsCollection.Where("key_value", "==", keyValue)

	docs, err := query.Documents(context.Background()).GetAll()
	if err != nil {
		return models.ResetPasswdEvent{}, fmt.Errorf("error querying reset password event: %v", err)
	}

	if len(docs) == 0 {
		return models.ResetPasswdEvent{}, fmt.Errorf("no reset password event found for key_value: %s", keyValue)
	}

	var resetEvent models.ResetPasswdEvent

	if err := docs[0].DataTo(&resetEvent); err != nil {
		return models.ResetPasswdEvent{}, fmt.Errorf("error decoding reset password event: %v", err)
	}
	return resetEvent, nil
}

func ResetPasswd(c *fiber.Ctx) error {
	keyVal := c.Params("randkey")
	newPasswd := c.FormValue("newPasswd")
	newPasswdR := c.FormValue("newPasswdR")

	resetPasswdEvent, err := getResetPasswdEventByKeyValue(c, keyVal)
	currentTime := time.Now()
	if currentTime.Before(resetPasswdEvent.Start) || currentTime.After(resetPasswdEvent.End) {
		return middleware.Render("forms/resetPasswd", c, fiber.Map{
			"Message": "Link do resetowania hasła wygasł. Proszę spróbować ponownie.",
		})
	}

	if !(len(newPasswd) >= 8 &&
		regexp.MustCompile(`[0-9a-zA-Z]`).MatchString(newPasswd) &&
		regexp.MustCompile(`[!@#$%^&*]`).MatchString(newPasswd)) {

		return middleware.Render("forms/resetPasswd", c, fiber.Map{
			"Message": "Hasło nie spełnia wymagań!",
		})
	}

	if err != nil {
		fmt.Println(err)
		return middleware.Render("forms/resetPasswd", c, fiber.Map{
			"Message": "Wystąpił błąd proszę spróbować ponownie!",
		})
	}

	if newPasswd != newPasswdR {
		return middleware.Render("forms/resetPasswd", c, fiber.Map{
			"Message": "Hasła muszą być takie same!",
		})
	}

	updatePasswdForUser(c, resetPasswdEvent, newPasswd)
	return middleware.Render("pageAfterChangingPasswd", c, fiber.Map{})
}

func GetAllUsers(c *fiber.Ctx) ([]models.User, error) {

	usersCollection := initializers.Client.Collection("users")

	query := usersCollection.OrderBy("id", firestore.Asc)

	docs, err := query.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("Error reading documents: %v", err)
		return nil, err
	}
	var users []models.User
	for _, doc := range docs {
		var user models.User
		if err := doc.DataTo(&user); err != nil {
			log.Printf("Error decoding document: %v", err)
			continue
		}
		users = append(users, user)
	}

	return users, nil
}

func GetAllUsersPage(c *fiber.Ctx) error {
	users, _ := GetAllUsers(c)
	return middleware.Render("AllUserPage", c, fiber.Map{
		"Users": users,
	})
}

func changeUserRole(c *fiber.Ctx, userId int, role int) error {
	user := GetOneUser(c, userId)
	if user.Id == 0 {
		return c.Status(http.StatusNotFound).SendString("User not found")
	}

	usersCollection := initializers.Client.Collection("users")
	_, err := usersCollection.Doc(user.FirestoreDocID).Update(context.Background(), []firestore.Update{
		{
			Path:  "role",
			Value: role,
		},
	})
	if err != nil {
		log.Printf("Error updating user role: %v", err)
		return c.Status(http.StatusInternalServerError).SendString("Error updating user role")
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "User role updated successfully",
		"userId":  userId,
		"newRole": role,
	})
}

func SetNewRoleForUser(c *fiber.Ctx) error {
	// Get user ID from form values and convert it to int
	userIdStr := c.FormValue("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid user ID")
	}
	roleStr := c.FormValue("role")
	newRole, err := strconv.Atoi(roleStr)
	if err != nil {
		return c.Status(http.StatusBadRequest).SendString("Invalid role")
	}
	err = changeUserRole(c, userId, newRole)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to change user role")
	}

	return c.Redirect("/get-all-users")
}
