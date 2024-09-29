package consume

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

type Consumer[M proto.Message] struct {
	client               *kgo.Client
	deserializer         serde.Deserializer
	topic                string
	messageHandler       func(M) error
	malformedDataHandler func([]byte, error) error
}

type ConsumerOption[M proto.Message] func(*Consumer[M])

func WithMessageHandler[M proto.Message](messageHandler func(M) error) ConsumerOption[M] {
	return func(consumer *Consumer[M]) {
		consumer.messageHandler = messageHandler
	}
}

func WithMalformedDataHandler[M proto.Message](malformedDataHandler func([]byte, error) error) ConsumerOption[M] {
	return func(consumer *Consumer[M]) {
		consumer.malformedDataHandler = malformedDataHandler
	}
}

func NewConsumer[M proto.Message](
	client *kgo.Client,
	deserializer serde.Deserializer,
	topic string,
	options ...ConsumerOption[M],
) *Consumer[M] {
	consumer := &Consumer[M]{
		client:               client,
		deserializer:         deserializer,
		topic:                topic,
		messageHandler:       defaultMessageHandler[M],
		malformedDataHandler: defaultMalformedDataHandler,
	}
	for _, option := range options {
		option(consumer)
	}
	return consumer
}

func (c *Consumer[M]) Consume(ctx context.Context, expectedMessageCount int) error {
	var got int
	for got < expectedMessageCount {
		fetches := c.client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			return fmt.Errorf("failed to fetch records: %v", errs)
		}
		for _, record := range fetches.Records() {
			got++
			data, err := c.deserializer.Deserialize(record.Topic, record.Value)
			if err != nil {
				if err := c.malformedDataHandler(record.Value, err); err != nil {
					return err
				}
				continue
			}
			message, ok := data.(M)
			if !ok {
				// TODO: This was a log only, why? This should never happen in our code.
				return fmt.Errorf("received unexpected message type: %T", data)
			}
			if err := c.messageHandler(message); err != nil {
				return err
			}
		}
	}
	return nil
}

func defaultMessageHandler[M proto.Message](message M) error {
	slog.Info("consumed message", "message", message)
	return nil
}

func defaultMalformedDataHandler(payload []byte, err error) error {
	slog.Info("consumed malformed data", "error", err)
	return nil
}
