package kafka

import (
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type ClientConfig struct {
	BootstrapServers []string
	Group            string
	Topic            string
	ClientID         string
}

func NewKafkaClient(clientConfig ClientConfig) (*kgo.Client, error) {
	return kgo.NewClient(
		kgo.SeedBrokers(clientConfig.BootstrapServers...),
		kgo.ConsumerGroup(clientConfig.Group),
		kgo.ConsumeTopics(clientConfig.Topic),
		kgo.ClientID(clientConfig.ClientID),
		kgo.AllowAutoTopicCreation(),
		kgo.FetchMaxWait(time.Second),
		// TODO: Why did we set this?
		//kgo.SoftwareNameAndVersion("bufstream-demo", "0.1.0"),
	)
}
