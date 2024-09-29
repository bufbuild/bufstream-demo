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

type Config struct {
	Kafka kafka.Config
	CSR   csr.Config
}

func Main(do func(context.Context, Config) error) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
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
	config := Config{
		Kafka: kafka.Config{
			BootstrapServers: []string{"0.0.0.0:9092"},
			Topic:            "email-updated",
			Group:            "email-verifier",
			ClientID:         "bufstream-demo",
		},
		CSR: csr.Config{
			URL:      "",
			Username: "",
			Password: "",
		},
	}

	// TODO: Why do all of these need short names?
	flagSet.StringArrayVarP(
		&config.Kafka.BootstrapServers,
		"bootstrap",
		"b",
		config.Kafka.BootstrapServers,
		"The Bufstream bootstrap server addresses.",
	)
	flagSet.StringVarP(
		&config.Kafka.Topic,
		"topic",
		"t",
		config.Kafka.Topic,
		"The Kafka topic name to use.",
	)
	flagSet.StringVarP(
		&config.Kafka.Group,
		"group",
		"g",
		config.Kafka.Group,
		"The Kafka consumer group ID.",
	)
	flagSet.StringVarP(
		&config.CSR.URL,
		"csr-url",
		"c",
		config.CSR.URL,
		"The Confluent Schema Registry URL.",
	)
	flagSet.StringVarP(
		&config.CSR.Username,
		"csr-user",
		"u",
		config.CSR.Username,
		"The Confluent Schema Registry username, if authentication is needed.",
	)
	flagSet.StringVarP(
		&config.CSR.Password,
		"csr-pass",
		"p",
		config.CSR.Password,
		"The Confluent Schema Registry password/token, if authentication is needed.",
	)
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return Config{}, err
	}
	return config, nil
}
