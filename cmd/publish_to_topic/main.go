package main

import (
	"time"

	"github.com/phonghaido/cloud-data-migration/handlers"
	"github.com/phonghaido/cloud-data-migration/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	systemConfig, err := config.GetSystemConfig()
	if err != nil {
		logrus.Fatalf("Error while parsing system config: %v", err)
	}

	awsClientConfig := config.GetAWSConfig()
	awsClient := handlers.NewAWSClient(awsClientConfig)

	redisConfig := config.GetRedisConfig()
	redisClient := handlers.NewRedisClient(redisConfig)
	defer redisClient.RedisClient.Close()

	pubsubClientConfig := config.GetPubSubConfig()
	pubsubClient, err := handlers.NewPubSucClient(pubsubClientConfig, systemConfig)
	if err != nil {
		logrus.Fatalf("Error while creating PubSub client: %v", err)
	}
	defer pubsubClient.PubSubClient.Close()

	for {
		err := awsClient.PublishS3Keys(redisClient, pubsubClient)
		if err != nil {
			logrus.Fatalf("Error: %v", err)
		}
		time.Sleep(10 * time.Minute)
	}
}
