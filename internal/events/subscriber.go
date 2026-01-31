package events

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-aws/sqs"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/trenchesdeveloper/go-ai-store/internal/config"
	"github.com/trenchesdeveloper/go-ai-store/internal/providers"
)

type WatermillSubscriber struct {
	subscriber message.Subscriber
	queueName  string
	logger     watermill.LoggerAdapter
}

func NewEventSubscriber(ctx context.Context, cfg *config.Config) (*WatermillSubscriber, error) {
	logger := watermill.NewStdLogger(false, false)

	// Create AWS config
	awsConfig, err := providers.CreateAWSConfig(ctx, cfg.AWS.S3Endpoint, cfg.AWS.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create aws config: %w", err)
	}

	// Create SQS subscriber
	subscriberConfig := sqs.SubscriberConfig{
		AWSConfig: awsConfig,
	}

	subscriber, err := sqs.NewSubscriber(subscriberConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqs subscriber: %w", err)
	}

	return &WatermillSubscriber{
		subscriber: subscriber,
		queueName:  cfg.AWS.EventQueueName,
		logger:     logger,
	}, nil
}

// Subscribe returns a channel of messages from the SQS queue
func (w *WatermillSubscriber) Subscribe(ctx context.Context) (<-chan *message.Message, error) {
	return w.subscriber.Subscribe(ctx, w.queueName)
}

// Close closes the subscriber
func (w *WatermillSubscriber) Close() error {
	return w.subscriber.Close()
}
