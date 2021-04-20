package gcppubsub

import (
	"context"

	"gocloud.dev/pubsub"

	// Load gcp driver.
	_ "gocloud.dev/pubsub/gcppubsub"
)

const (
	PublisherType = "gcp_pubsub"
)

type GCPPubSub struct {
	topic *pubsub.Topic
}

type Config struct {
	URL string `mapstructure:"url"`
}

func New(ctx context.Context, url string) (*GCPPubSub, error) {
	topic, err := pubsub.OpenTopic(context.TODO(), url)
	if err != nil {
		return nil, err
	}
	pub := &GCPPubSub{
		topic: topic,
	}
	return pub, nil
}

func FromConfig(ctx context.Context, config Config) (*GCPPubSub, error) {
	return New(ctx, config.URL)
}

func (pub *GCPPubSub) Name() string {
	return PublisherType
}

func (pub *GCPPubSub) Send(ctx context.Context, body []byte) error {
	return pub.topic.Send(ctx, &pubsub.Message{
		Body: body,
	})
}
