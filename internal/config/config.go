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

func GetAWSConfig() AWSClientConfig {
	viper.SetEnvPrefix("CLOUD_MIGRATION")
	viper.AutomaticEnv()

	viper.SetDefault("AWS_ACCESS_KEY_ID", "")
	viper.SetDefault("AWS_SECRET_ACCESS_KEY", "")
	viper.SetDefault("AWS_REGION", "")
	viper.SetDefault("S3_BUCKET", "")

	return AWSClientConfig{
		AWSAccessKeyID:     viper.GetString("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: viper.GetString("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          viper.GetString("AWS_REGION"),
		S3Bucket:           viper.GetString("S3_BUCKET"),
	}
}
