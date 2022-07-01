package guiver

import (
	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
)

type (
	testTopic struct{}
)

func (*testTopic) doPublish(ctx context.Context, msg *pubsub.Message) (string, error) {
	return "mocked123", nil
}
