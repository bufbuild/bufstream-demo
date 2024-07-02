package main

import (
	"context"
	"fmt"
	"log/slog"

	bufstream_demov1 "github.com/bufbuild/bufstream-demo/gen/bufbuild/bufstream_demo/v1"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Consumer struct {
	client       *kgo.Client
	topic        string
	deserializer serde.Deserializer
}

func NewConsumer(
	client *kgo.Client,
	topic string,
	deserializer serde.Deserializer,
) *Consumer {
	return &Consumer{
		client:       client,
		topic:        topic,
		deserializer: deserializer,
	}
}

func (d *Consumer) Run(ctx context.Context) (retErr error) {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			if err := d.consume(ctx); err != nil {
				return err
			}
		}
	}
}

func (d *Consumer) consume(ctx context.Context) error {
	fetches := d.client.PollFetches(ctx)
	if errs := fetches.Errors(); len(errs) > 0 {
		return fmt.Errorf("failed to poll fetches: %v", errs)
	}
	for _, record := range fetches.Records() {
		data, err := d.deserializer.Deserialize(record.Topic, record.Value)
		if err != nil {
			slog.ErrorContext(ctx, "failed to deserialize record", "error", err)
			continue
		}
		msg, ok := data.(*bufstream_demov1.EmailUpdated)
		if !ok {
			slog.ErrorContext(ctx, "received unexpected record type", "type", fmt.Sprintf("%T", msg))
			continue
		}
		slog.InfoContext(ctx, "received record",
			"old_address", msg.GetOldAddress(),
			"new_address", msg.GetNewAddress(),
		)
	}
	return nil
}
