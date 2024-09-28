package main

import (
	"context"
	"fmt"
	"log/slog"

	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
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

func (d *Consumer) Run(ctx context.Context, expect int) error {
	return d.consume(ctx, expect)
}

func (d *Consumer) consume(ctx context.Context, expect int) error {
	var got int
	for got < expect {
		fetches := d.client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			return fmt.Errorf("failed to fetch records: %v", errs)
		}
		for _, record := range fetches.Records() {
			got++
			data, err := d.deserializer.Deserialize(record.Topic, record.Value)
			if err != nil {
				slog.Info(fmt.Sprintf("received malformed message: %v", err))
				continue
			}
			msg, ok := data.(*demov1.EmailUpdated)
			if !ok {
				slog.Error("received unexpected record type", "type", fmt.Sprintf("%T", msg))
				continue
			}
			var suffix string
			if old := msg.GetOldEmailAddress(); old == "" {
				suffix = "redacted old email"
			} else {
				suffix = fmt.Sprintf("old email %s", old)
			}
			slog.Info(fmt.Sprintf("received message with new email %s and %s", msg.GetNewEmailAddress(), suffix))
		}
	}
	return nil
}
