// Package main implements the consumer of the demo's DLQ.
//
// The consumer will read as many DLQ records it can at once, print what it
// received, and then loop.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	dlqv1beta1 "buf.build/gen/go/bufbuild/bufstream/protocolbuffers/go/buf/bufstream/dlq/v1beta1"
	"buf.build/go/protovalidate"
	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/consume"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"google.golang.org/protobuf/proto"
)

func main() {
	// See the app package for the boilerplate we use to set up the producer and
	// consumer, including bound flags.
	app.Main(run)
}

func run(ctx context.Context, config app.Config) error {
	client, err := kafka.NewKafkaClient(config.Kafka, true)
	if err != nil {
		return err
	}
	defer client.Close()

	consumer := consume.NewConsumer(
		client,
		config.Kafka.Topic,
		consume.WithMessageHandler(handleDlqRecord),
	)

	slog.InfoContext(ctx, "starting consume")
	for {
		// Read as many messages as we can.
		//
		// Only return error if there is an unexpected system error. Of note, an error is not
		// returned if the data that the consumer receives is malformed.
		if err := consumer.Consume(ctx); err != nil {
			return err
		}
	}
}

func handleDlqRecord(_ context.Context, record *dlqv1beta1.Record) error {
	// Reconstruct the original message: we expect a Cart in this toy example.
	cart := &demov1.Cart{}
	if err := proto.Unmarshal(record.GetValue(), cart); err != nil {
		return fmt.Errorf("failed to unmarshal dlq value into shoppingv1.Cart: %w", err)
	}

	// Try to use Protovalidate to determine what was wrong with the cart!
	if err := protovalidate.Validate(cart); err != nil {
		slog.Info("DLQ received a cart that failed due to validation errors:", "ID", cart.GetCartId(), "error", err)
		return nil
	}

	// If we can't explain why the Cart failed, that's an error.
	return errors.New("DLQ received a cart for an unknown reason")
}
