package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/twmb/franz-go/pkg/kadm"
	"os"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// Config is all configuration we need to build a new Kafka Client.
//
// franz-go uses functional options for the same purpose, but we're simplifying this
// to just the values in this config struct for the purposes of this demo. If you use
// franz-go in production code, we'd recommend using the functional options directly.
type Config struct {
	// BootstrapServers are the bootstrap servers to call.
	BootstrapServers []string
	RootCAPath       string
	Group            string
	ClientID         string
	Topic            string
	RecreateTopic    bool
	TopicConfig      []string
	TopicPartitions  int
	ValidateMode     string
}

// NewKafkaClient returns a new franz-go Kafka Client for the given Config.
func NewKafkaClient(config Config, consumer bool) (*kgo.Client, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(config.BootstrapServers...),
		kgo.ClientID(config.ClientID),
	}

	if consumer {
		opts = append(opts,
			kgo.ConsumerGroup(config.Group),
			kgo.ConsumeTopics(config.Topic),
			kgo.FetchMaxWait(time.Second),
			kgo.FetchIsolationLevel(kgo.ReadCommitted()),
			kgo.RequireStableFetchOffsets(),
		)
	}

	if config.RootCAPath != "" {
		dialerTLSConfig, err := buildDialerTLSConfig(config.RootCAPath)
		if err != nil {
			return nil, fmt.Errorf("build dial tls config: %w", err)
		}

		opts = append(opts, kgo.DialTLSConfig(dialerTLSConfig))
	}

	return kgo.NewClient(opts...)
}

// NewAdminClient returns an franz-go admin client for the given Config.
func NewAdminClient(config Config) (*kadm.Client, error) {
	client, err := NewKafkaClient(config, false)
	if err != nil {
		return nil, err
	}

	return kadm.NewClient(client), nil
}

func ConfigureBroker(ctx context.Context, client *kadm.Client, config Config) error {
	alterConfigs := make([]kadm.AlterConfig, 1)
	alterConfigs[0] = kadm.AlterConfig{
		Name:  "bufstream.validate.mode",
		Value: &config.ValidateMode,
	}
	_, err := client.AlterBrokerConfigs(ctx, alterConfigs)
	if err != nil {
		return err
	}

	return nil
}

func buildDialerTLSConfig(rootCAPath string) (*tls.Config, error) {
	pool := x509.NewCertPool()

	caCert, err := os.ReadFile(rootCAPath)
	if err != nil {
		return nil, err
	}

	if !pool.AppendCertsFromPEM(caCert) {
		return nil, errors.New("parse CA cert failed")
	}

	tlsCfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    pool,
	}

	return tlsCfg, nil
}
