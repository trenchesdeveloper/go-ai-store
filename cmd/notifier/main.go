package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/trenchesdeveloper/go-ai-store/internal/config"
	"github.com/trenchesdeveloper/go-ai-store/internal/events"
	"github.com/trenchesdeveloper/go-ai-store/internal/notifications"
)

func main() {
	// Setup logger
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	log.Info().Msg("Starting notifier service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Create email service
	emailService := notifications.NewEmailServiceWithConfig(notifications.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		From:     cfg.SMTP.From,
	})

	log.Info().
		Str("smtp_host", cfg.SMTP.Host).
		Int("smtp_port", cfg.SMTP.Port).
		Str("smtp_from", cfg.SMTP.From).
		Msg("Email service configured")

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Create SQS subscriber
	subscriber, err := events.NewEventSubscriber(ctx, cfg)
	if err != nil {
		cancel()
		log.Fatal().Err(err).Msg("Failed to create event subscriber")
	}

	// Subscribe to messages
	messages, err := subscriber.Subscribe(ctx)
	if err != nil {
		cancel()
		_ = subscriber.Close()
		log.Fatal().Err(err).Msg("Failed to subscribe to events")
	}

	// Setup defer for cleanup (after all potential Fatal exits)
	defer cancel()
	defer func() {
		if err := subscriber.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close subscriber")
		}
	}()

	log.Info().
		Str("queue", cfg.AWS.EventQueueName).
		Msg("Subscribed to SQS queue, waiting for messages...")

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Process messages
	go func() {
		for msg := range messages {
			log.Info().
				Str("message_id", msg.UUID).
				Msg("Processing message")

			// Parse the notification
			var notification notifications.Notification
			if err := json.Unmarshal(msg.Payload, &notification); err != nil {
				log.Error().
					Err(err).
					Str("message_id", msg.UUID).
					Msg("Failed to unmarshal notification")
				msg.Nack()
				continue
			}

			// Get event type from metadata if not in payload
			eventType := notification.Type
			if eventType == "" {
				eventType = notifications.NotificationType(msg.Metadata.Get("type"))
			}

			// Send email based on notification type
			var sendErr error
			switch eventType {
			case notifications.NotificationTypeWelcome:
				log.Info().
					Str("type", string(eventType)).
					Str("email", notification.Email).
					Msg("Sending welcome email")
				sendErr = emailService.SendWelcomeEmail(notification.Email, notification.Username)

			case notifications.NotificationTypePasswordReset:
				log.Info().
					Str("type", string(eventType)).
					Str("email", notification.Email).
					Msg("Sending password reset email")
				sendErr = emailService.SendPasswordResetEmail(notification.Email, notification.ResetToken)

			case notifications.NotificationTypeOrderConfirmation:
				log.Info().
					Str("type", string(eventType)).
					Str("email", notification.Email).
					Str("order_id", notification.OrderID).
					Msg("Sending order confirmation email")
				sendErr = emailService.SendOrderConfirmationEmail(notification.Email, notification.OrderID, notification.Total)

			case notifications.NotificationTypeLoginNotification:
				log.Info().
					Str("type", string(eventType)).
					Str("email", notification.Email).
					Msg("Sending login notification email")
				sendErr = emailService.SendLoginNotificationEmail(
					notification.Email,
					notification.Username,
					notification.IPAddress,
					notification.UserAgent,
					notification.LoginTime,
				)

			case notifications.NotificationTypeUserLoggedIn:
				log.Info().
					Str("type", string(eventType)).
					Str("email", notification.Email).
					Int64("user_id", notification.UserID).
					Msg("Sending user logged in notification email")
				// Use login notification email with basic info
				sendErr = emailService.SendLoginNotificationEmail(
					notification.Email,
					notification.Username,
					"", // IP not available in this event
					"", // User agent not available
					time.Now().Format(time.RFC3339),
				)

			default:
				log.Warn().
					Str("type", string(eventType)).
					Msg("Unknown notification type")
				msg.Ack()
				continue
			}

			if sendErr != nil {
				log.Error().
					Err(sendErr).
					Str("message_id", msg.UUID).
					Str("type", string(notification.Type)).
					Msg("Failed to send email")
				msg.Nack()
				continue
			}

			log.Info().
				Str("message_id", msg.UUID).
				Str("type", string(notification.Type)).
				Str("email", notification.Email).
				Msg("Email sent successfully")

			msg.Ack()
		}
	}()

	// Wait for shutdown signal
	<-quit
	log.Info().Msg("Shutting down notifier service...")
	cancel()
	log.Info().Msg("Notifier service stopped")
}
