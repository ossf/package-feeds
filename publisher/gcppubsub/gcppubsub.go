package gcppubsub

import (
	"context"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
)

const (
	PublisherType = "gcp_pubsub"
)

type PubSub struct {
	topic *pubsub.Topic
}

type PubSubConfig struct {
	URL string `mapstructure:"url"`
}

func New(ctx context.Context, url string) (*PubSub, error) {
	topic, err := pubsub.OpenTopic(context.TODO(), url)
	if err != nil {
		return nil, err
	}
	pub := &PubSub{
		topic: topic,
	}
	return pub, nil
}

func FromConfig(ctx context.Context, config PubSubConfig) (*PubSub, error) {
	return New(ctx, config.URL)
}

func (pub *PubSub) Name() string {
	return PublisherType
}

func (pub *PubSub) Send(ctx context.Context, body []byte) error {
	return pub.topic.Send(ctx, &pubsub.Message{
		Body: body,
	})
}
