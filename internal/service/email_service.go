package service

import (
	"fmt"
	"net/smtp"

	"starter-kit-grpc-golang/config"
)

type EmailService interface {
	SendEmail(to, subject, body string) error
	SendResetPasswordEmail(to, token string) error
	SendVerificationEmail(to, token string) error
}

type emailService struct {
	cfg *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{cfg: cfg}
}

func (s *emailService) SendEmail(to, subject, body string) error {
	// In test/dev, we might skip actual sending if not configured
	if s.cfg.SMTP.Host == "" {
		return nil 
	}

	auth := smtp.PlainAuth("", s.cfg.SMTP.Username, s.cfg.SMTP.Password, s.cfg.SMTP.Host)

	msg := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=\"utf-8\"\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body))

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTP.Host, s.cfg.SMTP.Port)

	return smtp.SendMail(addr, auth, s.cfg.SMTP.From, []string{to}, msg)
}

func (s *emailService) SendResetPasswordEmail(to, token string) error {
	subject := "Reset Password"
	// Ensure this URL points to your Frontend
	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", token)
	text := fmt.Sprintf("Dear user,\n\nTo reset your password, click on this link: %s\n\nIf you did not request this, please ignore this email.", resetURL)
	return s.SendEmail(to, subject, text)
}

func (s *emailService) SendVerificationEmail(to, token string) error {
	subject := "Email Verification"
	// Ensure this URL points to your Frontend
	verifyURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", token)
	text := fmt.Sprintf("Dear user,\n\nTo verify your email, click on this link: %s\n\nIf you did not create an account, please ignore this email.", verifyURL)
	return s.SendEmail(to, subject, text)
}