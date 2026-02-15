package dispatcher

import (
	"fmt"
	"net/smtp"

	"github.com/rs/zerolog/log"
)

// EmailSender sends email notifications via SMTP (MailHog for development).
type EmailSender struct {
	host string
	port int
	from string
}

// NewEmailSender creates a new EmailSender.
func NewEmailSender(host string, port int) *EmailSender {
	return &EmailSender{
		host: host,
		port: port,
		from: "noreply@streamvault.com",
	}
}

// Send sends an HTML email. Returns an error if the send fails.
func (s *EmailSender) Send(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n",
		s.from, to, subject)
	msg := []byte(headers + htmlBody)

	if err := smtp.SendMail(addr, nil, s.from, []string{to}, msg); err != nil {
		log.Error().Err(err).Str("to", to).Str("subject", subject).Msg("failed to send email")
		return fmt.Errorf("sending email to %s: %w", to, err)
	}

	log.Info().Str("to", to).Str("subject", subject).Msg("email sent")
	return nil
}
