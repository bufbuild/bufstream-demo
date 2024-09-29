package kafka

import (
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

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

type Config struct {
	BootstrapServers []string
	Group            string
	Topic            string
	ClientID         string
}
