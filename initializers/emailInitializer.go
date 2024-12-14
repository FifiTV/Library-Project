package initializers

import (
	"my-firebase-project/models"
	"os"
)

// Create a global email configuration (or load it from environment variables)
var EmailConfig = models.EmailConfig{
	SMTPHost:    "smtp.gmail.com",
	SMTPPort:    587,
	SenderEmail: "student.paigr4@gmail.com",
	SenderPass:  os.Getenv("SENDERPASS"), // Use app-specific passwords if needed
}
