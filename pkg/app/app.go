// Package app implements boilerplate code shared by the producer and consumer.
//
// It implements Main, which both the producer and consumer use within their main functions.
// It also binds all relevant flags.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"github.com/spf13/pflag"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/bufbuild/bufstream-demo/pkg/csr"
)

const (
	defaultKafkaClientID = "bufstream-demo"
)

var (
	defaultKafkaBootstrapServers = []string{"localhost:9092"}
)

// Config contains all application configuration needed by the producer and consumer.
type Config struct {
	Kafka kafka.Config
	// Topics for browsing events
	SearchTopic      string
	ListViewedTopic  string
	ListFilteredTopic string
	// Confluent Schema Registry config
	CSR csr.Config
}

// Main is used by the consumer's main function.
//
// It sets up logging, interrupt handling, and binds and parses all flags. Afterwards, it calls
// action to invoke the application logic.
func Main(action func(context.Context, Config) error) {
	doMain(false, action)
}

// MainAutoCreateTopic is used by the producer's main function. It is just like [Main] except
// that it will also create the topic if necessary. The producer defines the topic and provides
// the data, so the consumer should not be the one auto-creating it.
//
// Note that in a real production workload, neither producer nor consumer applications should
// ever create topics. This should be considered an infrastructure concern, and the topic
// should be provisioned with correct configuration before a producer ever tries to send
// messages to it. If the topic does not exist, this should be a failure in the producer
// since it means a likely misconfiguration.
//
// This demo workload creates the topic, despite it not being a typical good practice, just
// for simplicity, so there are fewer steps to get the demo running.
func MainAutoCreateTopic(action func(context.Context, Config) error) {
	doMain(true, action)
}

func doMain(autoCreateTopic bool, action func(context.Context, Config) error) { // Set up slog. We use the global logger throughout this demo.
	// Set up slog. We use the global logger throughout this demo.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	// Cancel the context on interrupt, i.e. ctrl+c for our purposes.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, autoCreateTopic, action); err != nil {
		slog.Error("program error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, autoCreateTopic bool, action func(context.Context, Config) error) error {
	config, err := parseConfig(autoCreateTopic)
	if err != nil {
		return err
	}
	if autoCreateTopic {
		// Create configured topics
		// Default topic
		if config.Kafka.Topic != "" {
			if err := maybeCreateTopic(ctx, config.Kafka); err != nil {
				return err
			}
		}
		// Browsing event topics
		for _, t := range []string{config.SearchTopic, config.ListViewedTopic, config.ListFilteredTopic} {
			if t == "" {
				continue
			}
			cfg := config.Kafka
			cfg.Topic = t
			if err := maybeCreateTopic(ctx, cfg); err != nil {
				return err
			}
		}
	}
	return action(ctx, config)
}

func parseConfig(canCreateTopic bool) (Config, error) {
	flagSet := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	config := Config{}
	flagSet.StringArrayVar(
		&config.Kafka.BootstrapServers,
		"bootstrap",
		defaultKafkaBootstrapServers,
		"The Bufstream bootstrap server addresses.",
	)
	flagSet.StringVar(
		&config.Kafka.ClientID,
		"client-id",
		defaultKafkaClientID,
		"The Kafka client ID.",
	)
	flagSet.StringVar(
		&config.Kafka.Topic,
		"topic",
		"",
		"The Kafka topic name to use.",
	)
	// Browsing event topics
	flagSet.StringVar(
		&config.SearchTopic,
		"search-topic",
		"",
		"Topic for ProductsSearched events",
	)
	flagSet.StringVar(
		&config.ListViewedTopic,
		"list-viewed-topic",
		"",
		"Topic for ProductListViewed events",
	)
	flagSet.StringVar(
		&config.ListFilteredTopic,
		"list-filtered-topic",
		"",
		"Topic for ProductListFiltered events",
	)
	// Schema Registry flags
	flagSet.StringVar(
		&config.CSR.URL,
		"csr-url",
		"",
		"URL of the Confluent Schema Registry",
	)
	flagSet.StringVar(
		&config.CSR.Username,
		"csr-username",
		"",
		"Username for the Confluent Schema Registry (optional)",
	)
	flagSet.StringVar(
		&config.CSR.Password,
		"csr-password",
		"",
		"Password for the Confluent Schema Registry (optional)",
	)
	if canCreateTopic {
		flagSet.BoolVar(
			&config.Kafka.RecreateTopic,
			"recreate-topic",
			false,
			"If true, the topic will be recreated even if it already exists.",
		)
		flagSet.IntVar(
			&config.Kafka.TopicPartitions,
			"topic-partitions",
			1,
			"The number of partitions to use when creating the topic.",
		)
		flagSet.StringSliceVar(
			&config.Kafka.TopicConfig,
			"topic-config",
			nil,
			"Topic config parameters to use when creating the topic.",
		)
	}
	flagSet.StringVar(
		&config.Kafka.Group,
		"group",
		"",
		"The Kafka consumer group ID.",
	)
	flagSet.StringVar(
		&config.Kafka.RootCAPath,
		"tls-root-ca-path",
		"",
		"A path to root CA certificate for kafka TLS.",
	)
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return Config{}, err
	}
	return config, nil
}

func maybeCreateTopic(ctx context.Context, config kafka.Config) error {
	client, err := kafka.NewKafkaClient(config, false)
	if err != nil {
		return err
	}
	defer client.Close()

	admClient := kadm.NewClient(client)
	if config.RecreateTopic {
		resp, err := admClient.DeleteTopic(ctx, config.Topic)
		if err == nil {
			err = resp.Err
		}
		if !isUnknownTopic(err) {
			return err // something went wrong
		}
	} else {
		resp, err := admClient.DescribeTopicConfigs(ctx, config.Topic)
		if err == nil {
			if len(resp) != 1 {
				return fmt.Errorf("expected 1 topic config, got %d", len(resp))
			}
			err = resp[0].Err
		}
		if err == nil {
			return nil // topic exists; nothing to create
		}
		if !isUnknownTopic(err) {
			return err // something went wrong
		}
		// Else, topic does not exist, so we fall through to create it.
	}
	configs := make(map[string]*string, len(config.TopicConfig))
	for _, conf := range config.TopicConfig {
		k, v, _ := strings.Cut(conf, "=")
		if v == "" {
			configs[k] = nil
		} else {
			configs[k] = &v
		}
	}
	resp, err := admClient.CreateTopic(ctx, int32(config.TopicPartitions), 1, configs, config.Topic)
	if err == nil {
		err = resp.Err
	}
	return err
}

func isUnknownTopic(err error) bool {
	var kError *kerr.Error
	return errors.As(err, &kError) && kError.Code == kerr.UnknownTopicOrPartition.Code
}
