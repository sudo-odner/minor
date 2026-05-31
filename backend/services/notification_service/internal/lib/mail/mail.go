package mail

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type SMTPProvider struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewSMTPProvider() *SMTPProvider {
	return &SMTPProvider{
		host: os.Getenv("SMTP_HOST"),
		port: os.Getenv("SMTP_PORT"),
		user: os.Getenv("SMTP_USER"),
		pass: os.Getenv("SMTP_PASS"),
		from: os.Getenv("SMTP_FROM"),
	}
}

// Send реализует общий интерфейс EmailProvider
func (s *SMTPProvider) Send(ctx context.Context, toEmail string, title, body string) error {
	if s.host == "" || toEmail == "" {
		return fmt.Errorf("SMTP host or recipient email is empty")
	}

	header := fmt.Sprintf("From: %s\r\n", s.from) +
		fmt.Sprintf("To: %s\r\n", toEmail) +
		fmt.Sprintf("Subject: %s\r\n", title) +
		"MIME-version: 1.0;\r\n" +
		"Content-Type: text/plain; charset=\"UTF-8\";\r\n\r\n"

	msg := []byte(header + body)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)

	err := smtp.SendMail(addr, auth, s.user, []string{toEmail}, msg)
	if err != nil {
		return err
	}

	log.Printf("Email '%s' sent to %s", title, toEmail)
	return nil
}