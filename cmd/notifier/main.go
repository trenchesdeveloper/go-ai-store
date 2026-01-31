package main

import (
	"fmt"
)

func main() {
	fmt.Println("Notifier service started")
	// Load AWS configuration
	// cfg, err := config.LoadDefaultConfig(context.TODO())
	// if err != nil {
	// 	log.Fatalf("failed to load AWS config: %v", err)
	// }

	// // Create SQS client
	// sqsClient := sqs.NewFromConfig(cfg)

	// // Get SQS queue URL from environment variable
	// queueURL := os.Getenv("SQS_QUEUE_URL")
	// if queueURL == "" {
	// 	log.Fatal("SQS_QUEUE_URL environment variable is not set")
	// }

	// // Create email service
	// emailService := notifications.NewEmailServiceWithConfig(notifications.SMTPConfig{
	// 	Host:     os.Getenv("SMTP_HOST"),
	// 	Port:     587,
	// 	Username: os.Getenv("SMTP_USERNAME"),
	// 	Password: os.Getenv("SMTP_PASSWORD"),
	// 	From:     os.Getenv("SMTP_FROM"),
	// })

	// log.Println("Notifier service started, waiting for messages...")

	// // Poll SQS queue for messages
	// for {
	// 	result, err := sqsClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
	// 		QueueUrl:            &queueURL,
	// 		MaxNumberOfMessages: 1,
	// 		WaitTimeSeconds:     20,
	// 	})
	// 	if err != nil {
	// 		log.Printf("failed to receive message: %v", err)
	// 		time.Sleep(5 * time.Second)
	// 		continue
	// 	}

	// 	for _, message := range result.Messages {
	// 		log.Printf("processing message: %s", *message.MessageId)

	// 		// Parse the message body
	// 		var notification notifications.Notification
	// 		if err := json.Unmarshal([]byte(*message.Body), &notification); err != nil {
	// 			log.Printf("failed to unmarshal notification: %v", err)
	// 			continue
	// 		}

	// 		// Send the email based on notification type
	// 		var err error
	// 		switch notification.Type {
	// 		case notifications.NotificationTypeWelcome:
	// 			err = emailService.SendWelcomeEmail(notification.Email, notification.Username)
	// 		case notifications.NotificationTypePasswordReset:
	// 			err = emailService.SendPasswordResetEmail(notification.Email, notification.ResetToken)
	// 		case notifications.NotificationTypeOrderConfirmation:
	// 			err = emailService.SendOrderConfirmationEmail(notification.Email, notification.OrderID, notification.Total)
	// 		case notifications.NotificationTypeLoginNotification:
	// 			err = emailService.SendLoginNotificationEmail(notification.Email, notification.Username, notification.IPAddress, notification.UserAgent, notification.LoginTime)
	// 		default:
	// 			log.Printf("unknown notification type: %s", notification.Type)
	// 		}

	// 		if err != nil {
	// 			log.Printf("failed to send notification: %v", err)
	// 			continue
	// 		}

	// 		log.Printf("successfully sent notification: %s", *message.MessageId)

	// 		// Delete the message from the queue
	// 		_, err = sqsClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
	// 			QueueUrl:      &queueURL,
	// 			ReceiptHandle: message.ReceiptHandle,
	// 		})
	// 		if err != nil {
	// 			log.Printf("failed to delete message: %v", err)
	// 		}
	// 	}
	// }
}
