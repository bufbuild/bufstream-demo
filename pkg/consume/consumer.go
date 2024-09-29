// Package consume implements a toy consumer.
package consume

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

// Consumer is an example consumer of a given topic using given Protobuf message type.
//
// A Consume takes a Kafka client and a topic, and expects to recieve Protobuf messages
// of the given type. Upon every received message, a handler is invoked. If malformed
// data is recieved (data that can not be deserialized into the given Protobuf message type),
// a malformed data handler is invoked.
//
// This is a toy example, but shows the basics you need to receive Protobuf messages
// from Kafka using franz-go.
type Consumer[M proto.Message] struct {
	client               *kgo.Client
	deserializer         serde.Deserializer
	topic                string
	messageHandler       func(M) error
	malformedDataHandler func([]byte, error) error
}

// NewConsumer returns a new Consumer.
//
// Always use this constructor to construct Consumers.
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

// ConsumerOption is an option when constructing a new Consumer.
//
// All parameters except options are required. ConsumerOptions allow
// for optional parameters.
type ConsumerOption[M proto.Message] func(*Consumer[M])

// WithMessageHandler returns a new ConsumerOption that overrides the default
// handler of received messages.
//
// The default handler uses slog to log incoming messages.
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

func (c *Consumer[M]) Consume(ctx context.Context) error {
	fetches := c.client.PollFetches(ctx)
	if errs := fetches.Errors(); len(errs) > 0 {
		return fmt.Errorf("failed to fetch records: %v", errs)
	}
	for _, record := range fetches.Records() {
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
