package services

import (
	"net/mail"
)

// NotificationService handles notifications.
type NotificationService struct {
	// Email
	EmailService EmailSender
}

// NewNotificationService returns a new NotificationService responsible for handling pushes.
func NewNotificationService(mailService *EmailService) *NotificationService {
	return &NotificationService{EmailService: mailService}
}

// SendMail sends an email message by calling the internal email service.
func (s *NotificationService) SendEmail(from string, to string, subj, text, html string) error {
	return s.EmailService.SendEmail(
		mail.Address{Name: "Moowda", Address: from},
		mail.Address{Name: "To", Address: to},
		subj,
		text,
		html,
	)
}
