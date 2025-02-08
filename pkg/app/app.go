// Package app implements boilerplate code shared by the producer and consumer.
//
// It implements Main, which both the producer and consumer use within their main functions.
// It also binds all relevant flags.
package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/bufbuild/bufstream-demo/pkg/csr"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"github.com/spf13/pflag"
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
	CSR   csr.Config
}

// Main is used by the producer and consumer within their main functions.
//
// It sets up logging, interrupt handling, and binds and parses all flags. Afterwards, it calls
// do to invoke the application logic.
func Main(do func(context.Context, Config) error) {
	// Set up slog. We use the global logger throughout this demo.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	// Cancel the context on interrupt, i.e. ctrl+c for our purposes.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, do); err != nil {
		slog.Error("program error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, do func(context.Context, Config) error) error {
	config, err := parseConfig()
	if err != nil {
		return err
	}
	return do(ctx, config)
}

func parseConfig() (Config, error) {
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
	flagSet.StringVar(
		&config.Kafka.Group,
		"group",
		"",
		"The Kafka consumer group ID.",
	)
	flagSet.StringVar(
		&config.CSR.URL,
		"csr-url",
		"",
		"The Confluent Schema Registry URL.",
	)
	flagSet.StringVar(
		&config.CSR.Username,
		"csr-user",
		"",
		"The Confluent Schema Registry username, if authentication is needed.",
	)
	flagSet.StringVar(
		&config.CSR.Password,
		"csr-pass",
		"",
		"The Confluent Schema Registry password/token, if authentication is needed.",
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
