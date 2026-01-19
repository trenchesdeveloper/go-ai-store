package events

import "context"

type EventPublisher interface {
	Publish(ctx context.Context, eventType string, payload interface{}, metadata map[string]string) error
	Close() error
}
