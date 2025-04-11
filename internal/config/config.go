package config

import (
	"github.com/spf13/viper"
)

type AWSClientConfig struct {
	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
	AWSRegion          string `json:"aws_region"`
	S3Bucket           string `json:"s3_bucket"`
}

type GCPClientConfig struct {
	ProjectID      string
	BucketName     string
	GCPCredentials string
	SubScriptionID string
	TopicID        string
	PubSubHost     string
}

func GetAWSConfig() AWSClientConfig {
	viper.AutomaticEnv()

	viper.SetDefault("AWS_ACCESS_KEY_ID", "")
	viper.SetDefault("AWS_SECRET_ACCESS_KEY", "")
	viper.SetDefault("AWS_REGION", "")
	viper.SetDefault("AWS_S3_BUCKET", "")

	return AWSClientConfig{
		AWSAccessKeyID:     viper.GetString("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: viper.GetString("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          viper.GetString("AWS_REGION"),
		S3Bucket:           viper.GetString("AWS_S3_BUCKET"),
	}
}

func GetGCPConfig() GCPClientConfig {
	viper.AutomaticEnv()

	viper.SetDefault("GCP_CREDENTIALS", "")
	viper.SetDefault("GCP_PROJECT_ID", "")
	viper.SetDefault("GCP_BUCKET_NAME", "")
	viper.SetDefault("GCP_PUBSUB_HOST", "")
	viper.SetDefault("GCP_SUBSCRIPTION_ID", "")
	viper.SetDefault("GCP_TOPIC_ID", "")

	return GCPClientConfig{
		ProjectID:      viper.GetString("GCP_PROJECT_ID"),
		BucketName:     viper.GetString("GCP_BUCKET_NAME"),
		GCPCredentials: viper.GetString("GCP_CREDENTIALS"),
		SubScriptionID: viper.GetString("GCP_SUBSCRIPTION_ID"),
		TopicID:        viper.GetString("GCP_TOPIC_ID"),
		PubSubHost:     viper.GetString("GCP_PUBSUB_HOST"),
	}
}
