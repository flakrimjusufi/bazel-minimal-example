package guiver

/*
	handles formatting data and sending it to space
*/

import (
	"log"
	"os"

	"encoding/json"
	"time"

	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
)

type (
	// TopicInt is the interface we use on a pub/sub topic
	TopicInt interface {
		doPublish(ctx context.Context, msg *pubsub.Message) (string, error)
	}

	// PublishResultInt is the interface of *pubsub.PublishResult
	PublishResultInt interface {
		Get(ctx context.Context) (serverID string, err error)
		Ready() <-chan struct{}
	}
)

var (
	backendTopic TopicInt

	// publishBatchSize is the # of events to send to pubsub at a time
	publishBatchSize = 500
)

const (
	publishTimeout = time.Millisecond * 4000
)

func init() {
	if os.Getenv("TESTING") != "" {
		backendTopic = &testTopic{}
		return
	}

	publishBatchSize, _ = GetInt("PUBLISH_BATCH_SIZE", publishBatchSize)

	projectID, _ := GetString("GUIVER_CLIENT_PROJECT_ID", "whisper-infra")
	client, err := pubsub.NewClient(context.Background(), projectID)
	if err != nil {
		log.Println(err)
		return
	}

	topicName, _ := GetString("GUIVER_CLIENT_TOPIC", "weaver_events")
	topic := client.Topic(topicName)

	exists, _ := topic.Exists(context.Background())

	// if it doesnt exist and we are in a test then re-create it
	if !exists {
		topic, err = client.CreateTopic(context.Background(), topicName)
		if err != nil {
			log.Printf("%s/%s topic doesnt exist and we can't make one: %s", projectID, topicName, err)
		} else {
			log.Printf("recreated topic %s/%s", projectID, topicName)
		}
	}

	// some configs taken from other configs, some i tweaked
	topic.PublishSettings.DelayThreshold = 1000 * time.Millisecond
	topic.PublishSettings.Timeout = publishTimeout
	topic.PublishSettings.NumGoroutines = 20

	// weaver batcher handles batching so we want every publish
	// to actually publish
	topic.PublishSettings.CountThreshold = 8e6

	backendTopic = newPubsubTopic(topic)
}

func publishToPubSub(events []*WeaverEvent) error {
	data, _ := json.Marshal(events)

	ctx, cancel := context.WithTimeout(context.Background(), publishTimeout)
	defer cancel()

	sid, err := backendTopic.doPublish(ctx, &pubsub.Message{
		Data: data,
	})

	switch err {
	case nil:
		postProcess(events)
		return nil
	case context.DeadlineExceeded:
		log.Printf("guiver timeout writing to pubsub %s %s %d %d bytes", sid, err, len(events), len(data))
		return err
	default:
		log.Printf("guiver error writing to pubsub %s %s %d %d bytes", sid, err, len(events), len(data))
		return err
	}
}
