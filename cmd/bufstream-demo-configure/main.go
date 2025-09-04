// Package main implements runtime configuration of the demo.
//
// This is run as part of docker compose.
//
// TODO: describe what it does
package main

import (
	"context"
	"github.com/bufbuild/bufstream-demo/pkg/app"
	"github.com/bufbuild/bufstream-demo/pkg/kafka"
	"github.com/twmb/franz-go/pkg/kadm"
	"log/slog"
)

func main() {
	// See the app package for the boilerplate we use to set up the producer and
	// consumer, including bound flags.
	app.Main(run)
}

func run(ctx context.Context, config app.Config) error {
	// Unlike the producer or consumer, create an admin client.
	client, err := kafka.NewAdminClient(config.Kafka)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := configureBroker(ctx, client, config); err != nil {
		return err
	}
	return nil
}

func configureBroker(ctx context.Context, client *kadm.Client, config app.Config) error {
	slog.InfoContext(ctx, "configuring Bufstream broker")
	if err := kafka.ConfigureBroker(ctx, client, config.Kafka); err != nil {
		return err
	}
	slog.InfoContext(ctx, "configured Bufstream broker", "ValidateMode", config.Kafka.ValidateMode)
	return nil
}
