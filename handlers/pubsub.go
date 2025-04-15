package handlers

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/phonghaido/cloud-data-migration/internal/config"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type PubSubClient struct {
	SystemConfig       config.SystemConfig
	PubSubClientConfig config.PubSubClientConfig
	PubSubClient       *pubsub.Client
}

type Message struct {
	Key  string `json:"key"`
	ETag string `json:"etag"`
}

func NewPubSucClient(c config.PubSubClientConfig, sc config.SystemConfig) (PubSubClient, error) {
	err := os.Setenv("PUBSUB_EMULATOR_HOST", c.PubSubHost)
	if err != nil {
		return PubSubClient{}, err
	}
	ctx := context.Background()
	pubsubClient, err := pubsub.NewClient(ctx, c.ProjectID, option.WithCredentialsFile(c.GCPCredentials))
	if err != nil {
		return PubSubClient{}, err
	}

	return PubSubClient{
		SystemConfig:       sc,
		PubSubClientConfig: c,
		PubSubClient:       pubsubClient,
	}, nil
}

func (ps PubSubClient) VerifyTopic() (*pubsub.Topic, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	topic := ps.PubSubClient.Topic(ps.PubSubClientConfig.TopicID)
	ok, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if !ok {
		topic, err = ps.PubSubClient.CreateTopic(ctx, ps.PubSubClientConfig.TopicID)
		if err != nil {
			return nil, err
		}
		logrus.Infof("Successfully created Pub/Sub topic: %s", ps.PubSubClientConfig.TopicID)
	}

	return topic, nil
}

func (ps PubSubClient) VerifySubscription(topic *pubsub.Topic) (*pubsub.Subscription, error) {
	ctx := context.Background()
	sub := ps.PubSubClient.Subscription(ps.PubSubClientConfig.SubScriptionID)
	ok, err := sub.Exists(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		sub, err = ps.PubSubClient.CreateSubscription(ctx, ps.PubSubClientConfig.SubScriptionID, pubsub.SubscriptionConfig{
			Topic: topic,
		})
		if err != nil {
			return nil, err
		}
		logrus.Infof("Successfully created Pub/Sub subscription: %s", ps.PubSubClientConfig.SubScriptionID)
	}

	return sub, nil
}

func (ps PubSubClient) ProcessMessage(awsClient AWSClient, gcsClient GCSClient, redisClient RedisClient) error {
	ctx := context.Background()
	topic, err := ps.VerifyTopic()
	if err != nil {
		return err
	}

	sub, err := ps.VerifySubscription(topic)
	if err != nil {
		return err
	}

	sem := make(chan struct{}, ps.SystemConfig.MaxWorker)
	logrus.Info("Waiting for message...")

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		sem <- struct{}{}
		go func(msg *pubsub.Message) {
			defer func() { <-sem }()
			var fileMsg Message
			if err := json.Unmarshal(msg.Data, &fileMsg); err != nil {
				logrus.Warnf("Skipping invalid message: %s - error: %v", string(msg.Data), err)
				msg.Ack()
				return
			}

			body, err := awsClient.DownloadFromS3(fileMsg.Key)
			if err != nil {
				logrus.Errorf("Download failed: %v", err)
				msg.Nack()
				return
			}

			err = gcsClient.UploadFile(body, fileMsg.Key)
			if err != nil {
				logrus.Errorf("Uploaded failed: %v", err)
				msg.Nack()
				return
			}

			logrus.Infof("Successfully processed: %s", fileMsg.Key)
			msg.Ack()
		}(msg)
	})

	return err
}

func (ps PubSubClient) PublishMessage(msg Message) error {
	ctx := context.Background()
	topic, err := ps.VerifyTopic()
	if err != nil {
		return err
	}

	msgData, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	result := topic.Publish(ctx, &pubsub.Message{
		Data: msgData,
	})

	if _, err = result.Get(ctx); err != nil {
		return err
	}

	logrus.Infof("Successfully published message for file %s with ETag %s", msg.Key, msg.ETag)

	return nil
}
