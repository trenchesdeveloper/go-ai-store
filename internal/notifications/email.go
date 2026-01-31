package notifications

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
)

// SMTPConfig holds the SMTP server configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// EmailService handles email sending operations
type EmailService struct {
	config SMTPConfig
}

type Email struct {
	To      []string
	Subject string
	Body    string
	IsHTML  bool
}

// NewEmailServiceWithConfig creates a new EmailService with a custom configuration
func NewEmailServiceWithConfig(config SMTPConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// Send sends an email using the configured SMTP server
func (s *EmailService) Send(email Email) error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	// Connect to the SMTP server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// Send EHLO/HELO
	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("failed to send HELO: %w", err)
	}

	// Check if STARTTLS is available and upgrade if possible
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: s.config.Host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate if credentials are provided
	if s.config.Username != "" && s.config.Password != "" {
		auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Set the sender
	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set the recipients
	for _, recipient := range email.To {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Get the data writer
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	// Build and write the email message
	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", s.config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", email.To[0]))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", email.Subject))

	if email.IsHTML {
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	} else {
		msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	}

	msg.WriteString("\r\n")
	msg.WriteString(email.Body)

	if _, err := wc.Write(msg.Bytes()); err != nil {
		wc.Close()
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	// Send QUIT command
	return client.Quit()
}

// SendTemplate sends an email using an HTML template
func (s *EmailService) SendTemplate(to []string, subject string, templatePath string, data interface{}) error {
	// Parse the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute the template with the provided data
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	return s.Send(Email{
		To:      to,
		Subject: subject,
		Body:    body.String(),
		IsHTML:  true,
	})
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *EmailService) SendWelcomeEmail(to string, username string) error {
	body := fmt.Sprintf(`
		<h1>Welcome to Go AI Store!</h1>
		<p>Hello %s,</p>
		<p>Thank you for registering with us. We're excited to have you on board!</p>
		<p>Best regards,<br>The Go AI Store Team</p>
	`, username)

	return s.Send(Email{
		To:      []string{to},
		Subject: "Welcome to Go AI Store!",
		Body:    body,
		IsHTML:  true,
	})
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(to string, resetToken string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s",  "http://localhost:8000", resetToken)

	body := fmt.Sprintf(`
		<h1>Password Reset Request</h1>
		<p>You have requested to reset your password.</p>
		<p>Click the link below to reset your password:</p>
		<p><a href="%s">Reset Password</a></p>
		<p>If you did not request this, please ignore this email.</p>
		<p>This link will expire in 1 hour.</p>
		<p>Best regards,<br>The Go AI Store Team</p>
	`, resetURL)

	return s.Send(Email{
		To:      []string{to},
		Subject: "Password Reset Request",
		Body:    body,
		IsHTML:  true,
	})
}

// SendOrderConfirmationEmail sends an order confirmation email
func (s *EmailService) SendOrderConfirmationEmail(to string, orderID string, total float64) error {
	body := fmt.Sprintf(`
		<h1>Order Confirmation</h1>
		<p>Thank you for your order!</p>
		<p><strong>Order ID:</strong> %s</p>
		<p><strong>Total:</strong> $%.2f</p>
		<p>We will notify you when your order ships.</p>
		<p>Best regards,<br>The Go AI Store Team</p>
	`, orderID, total)

	return s.Send(Email{
		To:      []string{to},
		Subject: fmt.Sprintf("Order Confirmation - %s", orderID),
		Body:    body,
		IsHTML:  true,
	})
}

// SendLoginNotificationEmail sends a login notification email to alert users of new sign-ins
func (s *EmailService) SendLoginNotificationEmail(to string, username string, ipAddress string, userAgent string, loginTime string) error {
	body := fmt.Sprintf(`
		<h1>New Login Detected</h1>
		<p>Hello %s,</p>
		<p>We detected a new login to your account:</p>
		<table style="border-collapse: collapse; margin: 20px 0;">
			<tr>
				<td style="padding: 8px; border: 1px solid #ddd;"><strong>Time:</strong></td>
				<td style="padding: 8px; border: 1px solid #ddd;">%s</td>
			</tr>
			<tr>
				<td style="padding: 8px; border: 1px solid #ddd;"><strong>IP Address:</strong></td>
				<td style="padding: 8px; border: 1px solid #ddd;">%s</td>
			</tr>
			<tr>
				<td style="padding: 8px; border: 1px solid #ddd;"><strong>Device:</strong></td>
				<td style="padding: 8px; border: 1px solid #ddd;">%s</td>
			</tr>
		</table>
		<p>If this was you, no action is needed.</p>
		<p>If you did not perform this login, please <a href="%s/reset-password">reset your password</a> immediately and contact our support team.</p>
		<p>Best regards,<br>The Go AI Store Team</p>
	`, username, loginTime, ipAddress, userAgent, "http://localhost:8000")

	return s.Send(Email{
		To:      []string{to},
		Subject: "New Login to Your Account",
		Body:    body,
		IsHTML:  true,
	})
}
