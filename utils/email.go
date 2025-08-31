package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

	func SendEmail(subject, body string) error {
		smtpHost := os.Getenv("SMTP_HOST")
		smtpPort := os.Getenv("SMTP_PORT")
		smtpUser := os.Getenv("SMTP_USER")
		smtpPass := os.Getenv("SMTP_PASS")
		notifyEmail := os.Getenv("NOTIFY_EMAIL")

		if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" || notifyEmail == "" {
			return fmt.Errorf("missing SMTP configuration in environment variables")
		}

		auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		msg := []byte("To: " + notifyEmail + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"\r\n" +
			body + "\r\n")

		return smtp.SendMail(smtpHost+":"+smtpPort, auth, smtpUser, []string{notifyEmail}, msg)
	}
