package handlers

import (
	"context"
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

func (ps PubSubClient) VerifyTopicAndSubScription() (*pubsub.Subscription, error) {
	logrus.Infof("Verifying PubSub topic (%s) and subscription (%s)", ps.PubSubClientConfig.TopicID, ps.PubSubClientConfig.SubScriptionID)
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
	} else {
		logrus.Infof("Topic %s already existed", ps.PubSubClientConfig.TopicID)
	}

	sub := ps.PubSubClient.Subscription(ps.PubSubClientConfig.SubScriptionID)
	ok, err = sub.Exists(ctx)
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
	} else {
		logrus.Infof("Subscription %s already existed", ps.PubSubClientConfig.SubScriptionID)
	}

	return sub, nil
}

func (ps PubSubClient) PubSubHandler(awsClient AWSClient, gcsClient GCSClient) error {
	ctx := context.Background()
	sub, err := ps.VerifyTopicAndSubScription()
	if err != nil {
		return err
	}

	sem := make(chan struct{}, ps.SystemConfig.MaxWorker)
	logrus.Info("Waiting for message")

	err = sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		sem <- struct{}{}
		go func(msg *pubsub.Message) {
			defer func() { <-sem }()
			key := string(msg.Data)

			body, err := awsClient.DownloadFromS3(key)
			if err != nil {
				logrus.Errorf("Download failed: %v", err)
				msg.Nack()
				return
			}

			err = gcsClient.UploadFile(body, key)
			if err != nil {
				logrus.Errorf("Uploaded failed: %v", err)
				msg.Nack()
				return
			}
			logrus.Infof("Successfully processed: %s", key)
			msg.Ack()
		}(msg)
	})

	return err
}
