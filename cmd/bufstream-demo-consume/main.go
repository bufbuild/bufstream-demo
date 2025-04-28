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

	demov1 "github.com/bufbuild/bufstream-demo/gen/bufstream/demo/v1"
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/consume"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
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
	emailClient, err := kafka.NewKafkaClient(config.Kafka, true)
	if err != nil {
		return err
	}
	defer emailClient.Close()
	emailConsumer := consume.NewConsumer[*demov1.EmailUpdated](
		emailClient,
		config.Kafka.Topic,
		consume.WithMessageHandler(handleEmailUpdated),
	)
	consumers = append(consumers, func() error { return emailConsumer.Consume(ctx) })

	// ProductsSearched consumer
	if config.SearchTopic != "" {
		searchCfg := config.Kafka
		searchCfg.Topic = config.SearchTopic
		searchClient, err := kafka.NewKafkaClient(searchCfg, true)
		if err != nil {
			return err
		}
		defer searchClient.Close()
		searchConsumer := consume.NewConsumer[*demov1.ProductsSearched](
			searchClient,
			config.SearchTopic,
			consume.WithMessageHandler(handleProductsSearched),
		)
		consumers = append(consumers, func() error { return searchConsumer.Consume(ctx) })
	}

	// ProductListViewed consumer
	if config.ListViewedTopic != "" {
		viewedCfg := config.Kafka
		viewedCfg.Topic = config.ListViewedTopic
		viewedClient, err := kafka.NewKafkaClient(viewedCfg, true)
		if err != nil {
			return err
		}
		defer viewedClient.Close()
		viewedConsumer := consume.NewConsumer[*demov1.ProductListViewed](
			viewedClient,
			config.ListViewedTopic,
			consume.WithMessageHandler(handleProductListViewed),
		)
		consumers = append(consumers, func() error { return viewedConsumer.Consume(ctx) })
	}

	// ProductListFiltered consumer
	if config.ListFilteredTopic != "" {
		filteredCfg := config.Kafka
		filteredCfg.Topic = config.ListFilteredTopic
		filteredClient, err := kafka.NewKafkaClient(filteredCfg, true)
		if err != nil {
			return err
		}
		defer filteredClient.Close()
		filteredConsumer := consume.NewConsumer[*demov1.ProductListFiltered](
			filteredClient,
			config.ListFilteredTopic,
			consume.WithMessageHandler(handleProductListFiltered),
		)
		consumers = append(consumers, func() error { return filteredConsumer.Consume(ctx) })
	}

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
