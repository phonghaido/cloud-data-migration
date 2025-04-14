package main

import (
	"github.com/phonghaido/cloud-data-migration/handlers"
	"github.com/phonghaido/cloud-data-migration/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	awsClientConfig := config.GetAWSConfig()
	awsClient := handlers.NewAWSClient(awsClientConfig)

	redisConfig := config.GetRedisConfig()
	redisClient := handlers.NewRedisClient(redisConfig)
	for {
		err := awsClient.ListAllObject(redisClient)
		if err != nil {
			logrus.Fatalf("Error: %v", err)
		}
	}
}
