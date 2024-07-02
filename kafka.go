package main

import (
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func NewKafkaClient(bootstrapServers []string, groupID, topic string) (*kgo.Client, error) {
	return kgo.NewClient(
		kgo.SeedBrokers(bootstrapServers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
		kgo.AllowAutoTopicCreation(),
		kgo.ClientID("bufstream-demo"),
		kgo.FetchMaxWait(time.Second),
		kgo.SoftwareNameAndVersion("bufstream-demo", "0.1.0"),
	)
}
