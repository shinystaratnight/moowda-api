package services

import (
	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"

	"net/mail"
)

// This file should be moved somewhere more appropriate and functionality should be refactored.
// App.config should probably be called from server.go and injected into a mailing service.
type EmailSender interface {
	SendEmail(from mail.Address, to mail.Address, subj, text, html string) error
}

// EmailService describes a dialer object ready for sending email.
type EmailService struct {
	client *sendgrid.Client
}

// SendMail sends a pre-constructed email.
func (s *EmailService) SendEmail(from mail.Address, to mail.Address, subj, text, html string) error {
	m := sgmail.NewSingleEmail(
		sgmail.NewEmail(from.Name, from.Address),
		subj,
		sgmail.NewEmail(to.Name, to.Address),
		text,
		html,
	)

	_, err := s.client.Send(m)
	if err != nil {
		return err
	}
	return nil
}

// NewEmailService returns a configured EmailService object. However, errors won't be thrown until send-time.
func NewEmailService(sendgridAPIKey string) *EmailService {
	client := sendgrid.NewSendClient(sendgridAPIKey)
	return &EmailService{client}
}
