package kafka

import (
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
	//
	// TODO: A good explanation and links to docs as to what this does.
	BootstrapServers []string
	Group            string
	Topic            string
	ClientID         string
}

// NewKafkaClient returns a new franz-go Kafka Client for the given Config.
func NewKafkaClient(config Config) (*kgo.Client, error) {
	return kgo.NewClient(
		kgo.SeedBrokers(config.BootstrapServers...),
		kgo.ConsumerGroup(config.Group),
		kgo.ConsumeTopics(config.Topic),
		kgo.ClientID(config.ClientID),
		kgo.AllowAutoTopicCreation(),
		kgo.FetchMaxWait(time.Second),
		// TODO: Why did we set this?
		//kgo.SoftwareNameAndVersion("bufstream-demo", "0.1.0"),
	)
}
