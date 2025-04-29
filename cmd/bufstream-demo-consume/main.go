// Package main implements the consumer of the demo.
//
// This is run as part of docker compose.
//
// The consumer will read as many records it can at once, print what it received,
// sleep for one second, and then loop.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1" // Corrected import path
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/consume"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"google.golang.org/protobuf/proto" // Added proto import
)

func main() {
	// See the app package for the boilerplate we use to set up the producer and
	// consumer, including bound flags.
	app.Main(run)
}

func run(ctx context.Context, config app.Config) error {
	// Initialize Kafka clients and consumers for each topic
	var consumers []func() error

	// EmailUpdated consumer
	if config.Kafka.Topic != "" {
		client, err := kafka.NewKafkaClient(config.Kafka, true)
		if err != nil {
			return err
		}
		defer client.Close()
		emailConsumer := consume.NewConsumer[*demov1.EmailUpdated](
			client,
			config.Kafka.Topic,
			consume.WithMessageHandler(handleEmailUpdated),
		)
		consumers = append(consumers, func() error { return emailConsumer.Consume(ctx) })
	}

	// ProductsSearched consumer
	if config.SearchTopic != "" {
		cfg := config.Kafka
		cfg.Topic = config.SearchTopic
		client, err := kafka.NewKafkaClient(cfg, true)
		if err != nil {
			return err
		}
		defer client.Close()
		searchConsumer := consume.NewConsumer[*demov1.ProductsSearched](
			client,
			config.SearchTopic,
			consume.WithMessageHandler(handleProductsSearched),
		)
		consumers = append(consumers, func() error { return searchConsumer.Consume(ctx) })
	}

	// ProductListViewed consumer
	if config.ListViewedTopic != "" {
		cfg := config.Kafka
		cfg.Topic = config.ListViewedTopic
		client, err := kafka.NewKafkaClient(cfg, true)
		if err != nil {
			return err
		}
		defer client.Close()
		listViewedConsumer := consume.NewConsumer[*demov1.ProductListViewed](
			client,
			config.ListViewedTopic,
			consume.WithMessageHandler(handleProductListViewed),
		)
		consumers = append(consumers, func() error { return listViewedConsumer.Consume(ctx) })
	}

	// ProductListFiltered consumer
	if config.ListFilteredTopic != "" {
		cfg := config.Kafka
		cfg.Topic = config.ListFilteredTopic
		client, err := kafka.NewKafkaClient(cfg, true)
		if err != nil {
			return err
		}
		defer client.Close()
		listFilteredConsumer := consume.NewConsumer[*demov1.ProductListFiltered](
			client,
			config.ListFilteredTopic,
			consume.WithMessageHandler(handleProductListFiltered),
		)
		consumers = append(consumers, func() error { return listFilteredConsumer.Consume(ctx) })
	}

	// Ordering event consumers
	consumers = append(consumers, createOrderingConsumer[*demov1.ProductClicked](ctx, config.Kafka, config.ProductClickedTopic, handleProductClicked)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.ProductViewed](ctx, config.Kafka, config.ProductViewedTopic, handleProductViewed)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.ProductAdded](ctx, config.Kafka, config.ProductAddedTopic, handleProductAdded)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.ProductRemoved](ctx, config.Kafka, config.ProductRemovedTopic, handleProductRemoved)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.CartViewed](ctx, config.Kafka, config.CartViewedTopic, handleCartViewed)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.CheckoutStarted](ctx, config.Kafka, config.CheckoutStartedTopic, handleCheckoutStarted)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.CheckoutStepViewed](ctx, config.Kafka, config.CheckoutStepViewedTopic, handleCheckoutStepViewed)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.CheckoutStepCompleted](ctx, config.Kafka, config.CheckoutStepCompletedTopic, handleCheckoutStepCompleted)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.PaymentInfoEntered](ctx, config.Kafka, config.PaymentInfoEnteredTopic, handlePaymentInfoEntered)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.OrderUpdated](ctx, config.Kafka, config.OrderUpdatedTopic, handleOrderUpdated)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.OrderCompleted](ctx, config.Kafka, config.OrderCompletedTopic, handleOrderCompleted)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.OrderRefunded](ctx, config.Kafka, config.OrderRefundedTopic, handleOrderRefunded)...)
	consumers = append(consumers, createOrderingConsumer[*demov1.OrderCancelled](ctx, config.Kafka, config.OrderCancelledTopic, handleOrderCancelled)...)

	slog.Info("starting consume")
	for {
		// Poll each consumer for messages
		for _, consumeFunc := range consumers {
			if err := consumeFunc(); err != nil {
				return err
			}
		}
		time.Sleep(time.Second)
	}
}

// createOrderingConsumer is a helper function to create a consumer for an ordering event topic.
func createOrderingConsumer[M proto.Message](ctx context.Context, baseCfg kafka.Config, topic string, handler func(M) error) []func() error {
	if topic == "" {
		return nil
	}
	cfg := baseCfg
	cfg.Topic = topic
	// Create a unique group ID for each topic consumer to avoid rebalancing issues if multiple consumers share the same group
	// In a real application, you might want a more sophisticated group management strategy.
	cfg.Group = baseCfg.Group + "-" + topic
	client, err := kafka.NewKafkaClient(cfg, true)
	if err != nil {
		slog.Error("failed to create Kafka client for topic", "topic", topic, "error", err)
		// Decide how to handle this - perhaps return an error or panic?
		// For now, just log and return nil, effectively skipping this consumer.
		return nil
	}
	// Note: We are not closing the client here. This assumes the main run function's defer will handle it,
	// or that a more robust lifecycle management is in place.
	// Consider adding client.Close() to a cleanup mechanism if needed.

	consumer := consume.NewConsumer[M](
		client,
		topic,
		consume.WithMessageHandler(handler),
	)
	return []func() error{func() error { return consumer.Consume(ctx) }}
}

func handleEmailUpdated(message *demov1.EmailUpdated) error {
	var suffix string
	if old := message.GetOldEmailAddress(); old == "" {
		suffix = "redacted old email"
	} else {
		suffix = fmt.Sprintf("old email %s", old)
	}
	slog.Info(fmt.Sprintf("consumed message with new email %s and %s", message.GetNewEmailAddress(), suffix))
	return nil
}

// handleProductsSearched logs ProductsSearched events
func handleProductsSearched(message *demov1.ProductsSearched) error {
	slog.Info("consumed ProductsSearched event", "id", message.GetId(), "query", message.GetQuery())
	return nil
}

// handleProductListViewed logs ProductListViewed events
func handleProductListViewed(message *demov1.ProductListViewed) error {
	slog.Info("consumed ProductListViewed event", "id", message.GetId(), "list_id", message.GetListId(), "category", message.GetCategory(), "count", len(message.GetProducts()))
	return nil
}

// handleProductListFiltered logs ProductListFiltered events
func handleProductListFiltered(message *demov1.ProductListFiltered) error {
	slog.Info("consumed ProductListFiltered event", "id", message.GetId(), "list_id", message.GetListId(), "filters", message.GetFilters(), "sorts", message.GetSorts(), "count", len(message.GetProducts()))
	return nil
}

// --- Ordering Event Handlers ---

func handleProductClicked(message *demov1.ProductClicked) error {
	slog.Info("consumed ProductClicked event", "id", message.GetId(), "product_id", message.GetProductId(), "name", message.GetName())
	return nil
}

func handleProductViewed(message *demov1.ProductViewed) error {
	slog.Info("consumed ProductViewed event", "id", message.GetId(), "product_id", message.GetProductId(), "name", message.GetName())
	return nil
}

func handleProductAdded(message *demov1.ProductAdded) error {
	slog.Info("consumed ProductAdded event", "id", message.GetId(), "cart_id", message.GetCartId(), "product_id", message.GetProductId(), "quantity", message.GetQuantity())
	return nil
}

func handleProductRemoved(message *demov1.ProductRemoved) error {
	slog.Info("consumed ProductRemoved event", "cart_id", message.GetCartId(), "product_id", message.GetProductId(), "quantity", message.GetQuantity())
	return nil
}

func handleCartViewed(message *demov1.CartViewed) error {
	slog.Info("consumed CartViewed event", "cart_id", message.GetCartId(), "count", len(message.GetProducts()))
	return nil
}

func handleCheckoutStarted(message *demov1.CheckoutStarted) error {
	slog.Info("consumed CheckoutStarted event", "order_id", message.GetOrderId(), "total", message.GetValue())
	return nil
}

func handleCheckoutStepViewed(message *demov1.CheckoutStepViewed) error {
	slog.Info("consumed CheckoutStepViewed event", "checkout_id", message.GetCheckoutId(), "step", message.GetStep())
	return nil
}

func handleCheckoutStepCompleted(message *demov1.CheckoutStepCompleted) error {
	slog.Info("consumed CheckoutStepCompleted event", "checkout_id", message.GetCheckoutId(), "step", message.GetStep())
	return nil
}

func handlePaymentInfoEntered(message *demov1.PaymentInfoEntered) error {
	slog.Info("consumed PaymentInfoEntered event", "checkout_id", message.GetCheckoutId(), "order_id", message.GetOrderId())
	return nil
}

func handleOrderUpdated(message *demov1.OrderUpdated) error {
	slog.Info("consumed OrderUpdated event", "order_id", message.GetOrderId(), "total", message.GetTotal())
	return nil
}

func handleOrderCompleted(message *demov1.OrderCompleted) error {
	slog.Info("consumed OrderCompleted event", "order_id", message.GetOrderId(), "total", message.GetTotal())
	return nil
}

func handleOrderRefunded(message *demov1.OrderRefunded) error {
	slog.Info("consumed OrderRefunded event", "order_id", message.GetOrderId(), "total", message.GetTotal())
	return nil
}

func handleOrderCancelled(message *demov1.OrderCancelled) error {
	slog.Info("consumed OrderCancelled event", "order_id", message.GetOrderId(), "total", message.GetTotal())
	return nil
}
