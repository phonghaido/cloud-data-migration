package main

import (
	"github.com/phonghaido/cloud-data-migration/handlers"
	"github.com/phonghaido/cloud-data-migration/internal/config"
	"github.com/sirupsen/logrus"
)

func main() {
	awsClientConfig := config.GetAWSConfig()
	awsClient := handlers.NewAWSClient(awsClientConfig)

	gcpClientConfig := config.GetGCPConfig()
	gcpClient, err := handlers.NewGCPClient(gcpClientConfig)
	if err != nil {
		logrus.Fatalf("Error while creating GCP client: %v", err)
	}

	key := "test-1/CV - PhongHaiDo.pdf"

	output, err := awsClient.DownloadFromS3(key)
	if err != nil {
		logrus.Fatalf("Error while downloading file from AWS S3: %v", err)
	}

	err = gcpClient.UploadFile(output, key)
	if err != nil {
		logrus.Fatalf("Error while uploading file to GCS: %v", err)
	}

}
