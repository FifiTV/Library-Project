package models

type EmailConfig struct {
	SMTPHost    string
	SMTPPort    int
	SenderEmail string
	SenderPass  string
}
