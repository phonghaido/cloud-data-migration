package main

import (
	"github.com/phonghaido/cloud-data-migration/handlers"
	"github.com/phonghaido/cloud-data-migration/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	awsClientConfig := config.GetAWSConfig()

	awsClient := handlers.NewAWSClient(awsClientConfig)
	key := "test-1/CV - PhongHaiDo.pdf"

	err := awsClient.WriteToLocal(key)
	if err != nil {
		logrus.Fatalf("Error: %v", err)
	}
}
