package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type NotificationService struct {
	logger       *Logger
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	adminEmail   string
	telegramBot  string
	telegramChat string
	enabled      bool
}

type TelegramMessage struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func NewNotificationService(logger *Logger) *NotificationService {
	return &NotificationService{
		logger:       logger,
		smtpHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		smtpPort:     getEnv("SMTP_PORT", "587"),
		smtpUser:     getEnv("SMTP_USER", ""),
		smtpPassword: getEnv("SMTP_PASSWORD", ""),
		adminEmail:   getEnv("ADMIN_EMAIL", ""),
		telegramBot:  getEnv("TELEGRAM_BOT_TOKEN", ""),
		telegramChat: getEnv("TELEGRAM_CHAT_ID", ""),
		enabled:      getEnv("NOTIFICATIONS_ENABLED", "true") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SendAdminAlert sends notification to admin about new order
func (n *NotificationService) SendAdminAlert(orderID, jobID string) error {
	if !n.enabled {
		n.logger.Log("Notifications disabled, skipping admin alert")
		return nil
	}

	message := fmt.Sprintf(`🚨 *New Order Processing Complete*

Order ID: %s
Job ID: %s
Status: Ready for Review
Admin Panel: http://localhost:8090

Please review and approve for delivery.`, orderID, jobID)

	// Send Telegram notification
	if n.telegramBot != "" && n.telegramChat != "" {
		if err := n.sendTelegramMessage(message); err != nil {
			n.logger.Error(fmt.Sprintf("Failed to send Telegram notification: %v", err))
		} else {
			n.logger.Log("Telegram notification sent successfully")
		}
	}

	// Send Email notification
	if n.adminEmail != "" && n.smtpHost != "" {
		subject := fmt.Sprintf("New Order #%s Ready for Review", orderID)
		emailBody := fmt.Sprintf(`
Dear Admin,

A new order has been processed and is ready for review.

Order ID: %s
Job ID: %s
Status: Ready for Review

Please visit the admin panel to review and approve for delivery:
http://localhost:8090

Best regards,
ENDECode System
`, orderID, jobID)

		if err := n.sendEmail(n.adminEmail, subject, emailBody); err != nil {
			n.logger.Error(fmt.Sprintf("Failed to send email notification: %v", err))
		} else {
			n.logger.Log("Email notification sent successfully")
		}
	}

	return nil
}

// SendCustomerDownloadLink sends download link to customer
func (n *NotificationService) SendCustomerDownloadLink(orderID, email, downloadLink string, expiryDays int) error {
	if !n.enabled {
		n.logger.Log("Notifications disabled, skipping customer notification")
		return nil
	}

	subject := "Your Photos Are Ready for Download"
	emailBody := fmt.Sprintf(`
Dear Customer,

Your photo processing order has been completed and approved!

Order ID: %s
Download Link: %s

Important Information:
- This link will expire in %d days
- Maximum 3 downloads allowed
- Please download your photos as soon as possible

If you have any questions, please contact our support team.

Best regards,
Photo Processing Team
`, orderID, downloadLink, expiryDays)

	if err := n.sendEmail(email, subject, emailBody); err != nil {
		n.logger.Error(fmt.Sprintf("Failed to send customer email: %v", err))
		return err
	}

	n.logger.Log(fmt.Sprintf("Download link sent to customer: %s", email))
	return nil
}

// SendProcessingStatus sends status update to customer
func (n *NotificationService) SendProcessingStatus(orderID, email, status string) error {
	if !n.enabled {
		return nil
	}

	var subject, body string

	switch status {
	case "processing":
		subject = "Your Order is Being Processed"
		body = fmt.Sprintf(`
Dear Customer,

Your order #%s has been received and is currently being processed.

Status: Processing photos
Expected completion: Within 24 hours

You will receive another email when your photos are ready for download.

Best regards,
Photo Processing Team
`, orderID)

	case "pending_approval":
		subject = "Your Order is Pending Approval"
		body = fmt.Sprintf(`
Dear Customer,

Your order #%s has been processed and is now pending approval.

Status: Awaiting admin approval
Expected approval: Within 12 hours

You will receive a download link once the order is approved.

Best regards,
Photo Processing Team
`, orderID)

	default:
		return nil
	}

	return n.sendEmail(email, subject, body)
}

// sendTelegramMessage sends message to Telegram
func (n *NotificationService) sendTelegramMessage(message string) error {
	telegramAPI := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.telegramBot)

	msgData := TelegramMessage{
		ChatID:    n.telegramChat,
		Text:      message,
		ParseMode: "Markdown",
	}

	jsonData, err := json.Marshal(msgData)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram message: %v", err)
	}

	resp, err := http.Post(telegramAPI, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned status: %d", resp.StatusCode)
	}

	return nil
}

// sendEmail sends email via SMTP
func (n *NotificationService) sendEmail(to, subject, body string) error {
	if n.smtpHost == "" || n.smtpUser == "" {
		return fmt.Errorf("SMTP not configured")
	}

	auth := smtp.PlainAuth("", n.smtpUser, n.smtpPassword, n.smtpHost)
	
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", to, subject, body)
	
	addr := fmt.Sprintf("%s:%s", n.smtpHost, n.smtpPort)
	return smtp.SendMail(addr, auth, n.smtpUser, []string{to}, []byte(msg))
}

// SendOrderStatusWebhook sends webhook to WordPress about order status change
func (n *NotificationService) SendOrderStatusWebhook(orderID, status, downloadLink string) error {
	webhookData := map[string]interface{}{
		"order_id":      orderID,
		"status":        status,
		"download_link": downloadLink,
		"timestamp":     time.Now().Unix(),
	}

	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook data: %v", err)
	}

	// In real implementation, this would be configurable WordPress webhook URL
	webhookURL := "http://endecode-wordpress:80/wp-json/endecode/v1/order-status"
	
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		n.logger.Error(fmt.Sprintf("Failed to send webhook: %v", err))
		return err
	}
	defer resp.Body.Close()

	n.logger.Log(fmt.Sprintf("Webhook sent for order %s, status: %s", orderID, status))
	return nil
} 