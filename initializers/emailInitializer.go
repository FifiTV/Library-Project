package initializers

import (
	"my-firebase-project/models"
	"os"
)

// Create a global email configuration (or load it from environment variables)
var EmailConfig models.EmailConfig

func InitMail() {
	EmailConfig = models.EmailConfig{
		SMTPHost:    "smtp.gmail.com",
		SMTPPort:    587,
		SenderEmail: os.Getenv("SENDEREMAIL"),
		SenderPass:  os.Getenv("SENDERPASS"),
	}
}
