package kafkapubsub

import (
	"context"

	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/kafkapubsub"
)

const (
	PublisherType = "kafka"
)

type KafkaPubSub struct {
	topic *pubsub.Topic
}

type Config struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
}

func New(ctx context.Context, brokers []string, topic string) (*KafkaPubSub, error) {
	config := kafkapubsub.MinimalConfig()

	pubSubTopic, err := kafkapubsub.OpenTopic(brokers, config, topic, nil)
	if err != nil {
		return nil, err
	}
	return &KafkaPubSub{
		topic: pubSubTopic,
	}, nil
}

func FromConfig(ctx context.Context, config Config) (*KafkaPubSub, error) {
	return New(ctx, config.Brokers, config.Topic)
}

func (pub *KafkaPubSub) Name() string {
	return PublisherType
}

func (pub *KafkaPubSub) Send(ctx context.Context, body []byte) error {
	return pub.topic.Send(ctx, &pubsub.Message{
		Body: body,
	})
}
