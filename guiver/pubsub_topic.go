package guiver

import (
	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
)

type (
	pubsubTopic struct {
		topic *pubsub.Topic
	}
)

func newPubsubTopic(topic *pubsub.Topic) *pubsubTopic {
	return &pubsubTopic{
		topic,
	}
}

func (pst *pubsubTopic) doPublish(ctx context.Context, msg *pubsub.Message) (string, error) {
	res := pst.topic.Publish(ctx, msg)

	return res.Get(ctx)
}
