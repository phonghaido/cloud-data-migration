package main

import (
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

	gcsClientConfig := config.GetGCPConfig()
	gcsClient, err := handlers.NewGCSClient(gcsClientConfig)
	if err != nil {
		logrus.Fatalf("Error while creating GCS client: %v", err)
	}
	defer gcsClient.StorageClient.Close()

	pubsubClientConfig := config.GetPubSubConfig()
	pubsubClient, err := handlers.NewPubSucClient(pubsubClientConfig, systemConfig)
	if err != nil {
		logrus.Fatalf("Error while creating PubSub client: %v", err)
	}
	defer pubsubClient.PubSubClient.Close()

	redisConfig := config.GetRedisConfig()
	redisClient := handlers.NewRedisClient(redisConfig)
	defer redisClient.RedisClient.Close()

	pubsubClient.ProcessMessage(awsClient, gcsClient, redisClient)
}
