package publisher

import (
	"context"

	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/gcppubsub"
)

type PubSub struct {
	topic *pubsub.Topic
}

func NewPubSub(ctx context.Context, url string) (*PubSub, error) {
	topic, err := pubsub.OpenTopic(context.TODO(), url)
	if err != nil {
		return nil, err
	}
	pub := &PubSub{
		topic: topic,
	}
	return pub, nil
}

func (pub *PubSub) Name() string {
	return "gcp-pubsub"
}

func (pub *PubSub) Send(ctx context.Context, body []byte) error {
	return pub.topic.Send(ctx, &pubsub.Message{
		Body: body,
	})
}
