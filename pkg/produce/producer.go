package produce

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

type Producer[M proto.Message] struct {
	client     *kgo.Client
	serializer serde.Serializer
	topic      string
}

func NewProducer[M proto.Message](
	client *kgo.Client,
	serializer serde.Serializer,
	topic string,
) *Producer[M] {
	return &Producer[M]{
		client:     client,
		topic:      topic,
		serializer: serializer,
	}
}

func (p *Producer[M]) ProduceProtobufMessage(ctx context.Context, key string, message M) error {
	payload, err := p.serializer.Serialize(p.topic, message)
	if err != nil {
		return err
	}
	if err := p.produce(ctx, key, payload); err != nil {
		return err
	}
	slog.Info("produced protobuf message", "key", key)
	return nil
}

func (p *Producer[M]) ProduceInvalid(ctx context.Context, key string) error {
	if err := p.produce(ctx, key, []byte("\x00foobar")); err != nil {
		return err
	}
	slog.Info("produced invalid data", "key", key)
	return nil
}

func (p *Producer[M]) produce(ctx context.Context, key string, payload []byte) error {
	produceResults := p.client.ProduceSync(
		ctx,
		&kgo.Record{
			Key:   []byte(key),
			Value: payload,
			Topic: p.topic,
		},
	)
	if err := produceResults.FirstErr(); err != nil {
		return fmt.Errorf("failed to publish: %w", err)
	}
	return nil
}
