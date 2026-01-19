package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/trenchesdeveloper/go-ai-store/internal/config"
	"github.com/trenchesdeveloper/go-ai-store/internal/providers"
)

type WatermillPublisher struct {
	publisher message.Publisher
	queueName string
	logger    watermill.LoggerAdapter
}

func NewEventPublisher(ctx context.Context, cfg *config.Config) (*WatermillPublisher, error) {
	logger := watermill.NewStdLogger(false, false)

	// create aws config
	awsConfig, err := providers.CreateAWSConfig(ctx, cfg.AWS.S3Endpoint, cfg.AWS.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws config: %w", err)
	}

	// create sqs publisher
	publisherConfig := sqs.PublisherConfig{
		AWSConfig: awsConfig,
		Marshaler: nil,
	}

	// create the publisher with custom config
	publisher, err := sqs.NewPublisher(publisherConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqs publisher: %w", err)
	}

	return &WatermillPublisher{
		publisher: publisher,
		queueName: cfg.AWS.EventQueueName,
		logger:    logger,
	}, nil
}

func (w *WatermillPublisher) Publish(ctx context.Context, eventType string, payload interface{}, metadata map[string]string) error {
	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Create message using Watermill's NewMessage constructor
	msg := message.NewMessage(watermill.NewUUID(), payloadBytes)

	// Set metadata
	msg.Metadata.Set("type", eventType)
	msg.Metadata.Set("time", time.Now().Format(time.RFC3339))

	// Add custom metadata
	for k, v := range metadata {
		msg.Metadata.Set(k, v)
	}

	// Publish to SQS (topic/queueName is the first arg, messages follow)
	return w.publisher.Publish(w.queueName, msg)
}

func (w *WatermillPublisher) Close() error {
	return w.publisher.Close()
}
