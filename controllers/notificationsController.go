package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"my-firebase-project/initializers"
	"my-firebase-project/middleware"
	"my-firebase-project/models"
	"strings"
	"text/template"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/gomail.v2"
)

// Pobranie danych z sesji
func FetchNotifications(c *fiber.Ctx) error {
	sess, err := middleware.GetSession(c)
	if err != nil {
		log.Println("Error fetching session:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch session",
		})
	}

	// Pobierz userID z sesji
	userID, ok := sess.Get("userID").(int)
	if !ok || userID == 0 {
		log.Println("User ID not found or invalid in session")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized: Missing or invalid user ID",
		})
	}

	// log.Printf("Fetching notifications for userId: %d", userID)

	ctx := context.Background()
	notifications := []models.Notification{}

	// Pobranie powiadomień
	snapshot, err := initializers.Client.Collection("notifications").
		Where("recipientId", "==", fmt.Sprintf("%d", userID)).
		OrderBy("timestamp", firestore.Desc).
		Documents(ctx).GetAll()
	if err != nil {
		// log.Printf("Error fetching notifications: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch notifications: %v", err),
		})
	}

	for _, doc := range snapshot {
		var notification models.Notification
		doc.DataTo(&notification)
		notification.ID = doc.Ref.ID
		notifications = append(notifications, notification)
	}

	return c.JSON(notifications)
}

// Tworzenie powiadomienia
func CreateNotification(recipientID, bookTitle, message string, role int, status bool) error {
	// log.Printf("Creating notification for user: %s, book: %s", recipientID, bookTitle)

	ctx := context.Background()

	notification := map[string]interface{}{
		"recipientId": recipientID,
		"bookTitle":   bookTitle,
		"message":     message,
		"role":        role,
		"status":      status,
		"timestamp":   firestore.ServerTimestamp,
	}

	_, _, err := initializers.Client.Collection("notifications").Add(ctx, notification)
	if err != nil {
		log.Printf("Error creating notification: %v", err)
		return err
	}

	return nil
}

// TEST POWIADOMIEŃ RĘCZNE DODANIE
func AddTestNotifications(c *fiber.Ctx) error {
	log.Println("Adding test notifications...")

	ctx := context.Background()

	testNotifications := []map[string]interface{}{
		{
			"recipientId": "5",
			"bookTitle":   "Book A",
			"message":     "Poprosiłeś o wypożyczenie",
			"role":        1,
			"status":      false,
			"timestamp":   firestore.ServerTimestamp,
		},
		{
			"recipientId": "5",
			"bookTitle":   "Book B",
			"message":     "Prośba o zwrot książki",
			"role":        1,
			"status":      false,
			"timestamp":   firestore.ServerTimestamp,
		},
		{
			"recipientId": "5",
			"bookTitle":   "Book C",
			"message":     "Prośba o wypożyczenie",
			"role":        2,
			"status":      false,
			"timestamp":   firestore.ServerTimestamp,
		},
	}

	for _, notification := range testNotifications {
		_, _, err := initializers.Client.Collection("notifications").Add(ctx, notification)
		if err != nil {
			log.Printf("Error adding test notification: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to add test notification: %v", err),
			})
		}
		log.Printf("Added notification: %+v", notification)
	}

	log.Println("Test notifications added successfully.")
	return c.JSON(fiber.Map{
		"message": "Test notifications added successfully.",
	})
}

// sendEmail sends an email using the SMTP configuration
func SendEmail(to, subject, body string) error {

	m := gomail.NewMessage()
	m.SetHeader("From", initializers.EmailConfig.SenderEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Set up the dialer with SMTP credentials
	d := gomail.NewDialer(initializers.EmailConfig.SMTPHost, initializers.EmailConfig.SMTPPort, initializers.EmailConfig.SenderEmail, initializers.EmailConfig.SenderPass)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

func loadTemplate(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func SendResetPasswdEMail(to string) error {
	templateContent, err := loadTemplate("views/email/passwdReset.html")
	if err != nil {
		return fmt.Errorf("could not load template: %v", err)
	}
	sub := "Nowe Hasło"
	data := map[string]interface{}{
		"Subject":  sub,
		"UserMail": to,
	}

	tmpl, err := template.New("email").Parse(templateContent)
	if err != nil {
		return fmt.Errorf("could not parse template: %v", err)
	}

	var body strings.Builder

	err = tmpl.Execute(&body, data)
	if err != nil {
		return fmt.Errorf("could not execute template: %v", err)
	}
	go SendEmail(to, sub, body.String())

	// fmt.Println("Email sent successfully!")
	return nil
}

func SendUpcomingBorrowEventsEmail(c *fiber.Ctx) error {
	// Retrieve borrow events for the user
	borrowEventsForUser, err := GetAllBorrowEventsForUser(c, true)
	if err != nil {
		return fmt.Errorf("could not get borrow events: %v", err)
	}

	// Get the current time and define the cutoff time for 7 days in the future
	now := time.Now()
	cutoff := now.Add(7 * 24 * time.Hour)

	// Filter events that are within the next 7 days
	var upcomingEvents []models.BorrowEvent
	var bookDetails []string
	for _, eventWrapper := range borrowEventsForUser {
		borrowEvent := eventWrapper.BorrowEvent
		if borrowEvent.BorrowEnd.After(now) && borrowEvent.BorrowEnd.Before(cutoff) {
			upcomingEvents = append(upcomingEvents, borrowEvent)
			bookDetails = append(bookDetails, fmt.Sprintf("%s (zwróć do: %s)", eventWrapper.Book.Title, borrowEvent.BorrowEnd.Format("02/01/2006")))
		}
	}

	// If no upcoming events, return early
	if len(upcomingEvents) == 0 {
		return nil
	}

	sess, _ := middleware.GetSession(c)
	email := sess.Get("mail").(string)

	// Prepare the email body
	bookList := strings.Join(bookDetails, "<br>")
	emailBody := fmt.Sprintf("<html><body><p>Powinieneś oddać następujące ksiązki w ciągu następnych 7 dni:</p><ul><li>%s</li></ul></body></html>", bookList)

	// Send the email
	sub := "Zwrot książek do biblioteki"
	go SendEmail(email, sub, emailBody)

	return nil
}
