package services

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/Sneh16Shah/ai-visibility-tracker/db"
	"github.com/Sneh16Shah/ai-visibility-tracker/models"
)

// EmailService handles sending alert emails
type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	fromEmail    string
	enabled      bool
}

// Global email service instance
var emailService *EmailService

// InitEmailService initializes the email service from environment
func InitEmailService() *EmailService {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	fromEmail := os.Getenv("SMTP_FROM")

	enabled := smtpHost != "" && smtpUser != "" && smtpPassword != ""

	if enabled {
		log.Printf("📧 Email alerts enabled (SMTP: %s:%s)", smtpHost, smtpPort)
	} else {
		log.Println("📧 Email alerts disabled (SMTP not configured)")
	}

	emailService = &EmailService{
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUser:     smtpUser,
		smtpPassword: smtpPassword,
		fromEmail:    fromEmail,
		enabled:      enabled,
	}

	return emailService
}

// GetEmailService returns the global email service
func GetEmailService() *EmailService {
	return emailService
}

// IsEnabled returns whether email is configured
func (e *EmailService) IsEnabled() bool {
	return e.enabled
}

// SendAlert sends a visibility alert email
func (e *EmailService) SendAlert(toEmail string, brand *models.Brand, currentScore, threshold float64) error {
	if !e.enabled {
		log.Println("Email not configured, skipping alert")
		return nil
	}

	subject := fmt.Sprintf("⚠️ AI Visibility Alert: %s score dropped below %d", brand.Name, int(threshold))

	body := fmt.Sprintf(`
AI Visibility Alert for %s

Your brand's AI visibility score has dropped below your configured threshold.

Current Score: %.1f
Alert Threshold: %.1f

This means AI assistants are mentioning your brand less frequently than expected.

Recommended Actions:
• Improve SEO for AI-related content
• Update product descriptions with natural language
• Create more FAQ content that AI can reference
• Check competitor strategies

View Dashboard: http://localhost:5173/

---
AI Visibility Tracker
`, brand.Name, currentScore, threshold)

	return e.sendEmail(toEmail, subject, body)
}

// sendEmail sends a generic email
func (e *EmailService) sendEmail(to, subject, body string) error {
	from := e.fromEmail
	if from == "" {
		from = e.smtpUser
	}

	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", from, to, subject, body))

	auth := smtp.PlainAuth("", e.smtpUser, e.smtpPassword, e.smtpHost)

	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)
	err := smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Printf("📧 Alert email sent to %s", to)
	return nil
}

// CheckAndSendAlerts checks all brands and sends alerts if needed
func (e *EmailService) CheckAndSendAlerts() {
	if !e.enabled {
		return
	}

	brandRepo := db.NewBrandRepository()
	brands, err := brandRepo.GetAllBrands()
	if err != nil {
		log.Printf("Error fetching brands for alerts: %v", err)
		return
	}

	metricRepo := db.NewMetricRepository()

	for _, brand := range brands {
		// Check if brand has alert threshold set
		if brand.AlertThreshold <= 0 {
			continue
		}

		// Get latest metrics
		latest, err := metricRepo.GetLatestByBrandID(brand.ID)
		if err != nil {
			continue
		}

		// Check if score is below threshold
		if latest.VisibilityScore < brand.AlertThreshold {
			// Get user email (from brand owner)
			userEmail := getAlertEmail(brand.UserID)
			if userEmail != "" {
				e.SendAlert(userEmail, &brand, latest.VisibilityScore, brand.AlertThreshold)
			}
		}
	}
}

// getAlertEmail gets the email for a user
func getAlertEmail(userID int) string {
	userRepo := db.NewUserRepository()
	user, err := userRepo.GetByID(userID)
	if err != nil {
		return ""
	}
	// Check if email contains @ to validate
	if strings.Contains(user.Email, "@") {
		return user.Email
	}
	return ""
}
