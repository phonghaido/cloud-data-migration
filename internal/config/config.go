package config

import (
	"errors"
	"strconv"

	"github.com/spf13/viper"
)

type AWSClientConfig struct {
	AWSAccessKeyID     string `json:"aws_access_key_id"`
	AWSSecretAccessKey string `json:"aws_secret_access_key"`
	AWSRegion          string `json:"aws_region"`
	S3Bucket           string `json:"s3_bucket"`
}

type GCSClientConfig struct {
	ProjectID      string
	BucketName     string
	GCPCredentials string
}

type PubSubClientConfig struct {
	ProjectID      string
	GCPCredentials string
	SubScriptionID string
	TopicID        string
	PubSubHost     string
}

type SystemConfig struct {
	MaxWorker     int
	AdminUsername string
	AdminPassword string
}

type RedisConfig struct {
	Address  string
	Password string
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

func GetGCPConfig() GCSClientConfig {
	viper.AutomaticEnv()

	viper.SetDefault("GCP_CREDENTIALS", "")
	viper.SetDefault("GCP_PROJECT_ID", "")
	viper.SetDefault("GCP_BUCKET_NAME", "")

	return GCSClientConfig{
		ProjectID:      viper.GetString("GCP_PROJECT_ID"),
		BucketName:     viper.GetString("GCP_BUCKET_NAME"),
		GCPCredentials: viper.GetString("GCP_CREDENTIALS"),
	}
}

func GetPubSubConfig() PubSubClientConfig {
	viper.AutomaticEnv()

	viper.SetDefault("GCP_CREDENTIALS", "")
	viper.SetDefault("GCP_PROJECT_ID", "")
	viper.SetDefault("GCP_PUBSUB_HOST", "")
	viper.SetDefault("GCP_SUBSCRIPTION_ID", "")
	viper.SetDefault("GCP_TOPIC_ID", "")

	return PubSubClientConfig{
		ProjectID:      viper.GetString("GCP_PROJECT_ID"),
		GCPCredentials: viper.GetString("GCP_CREDENTIALS"),
		SubScriptionID: viper.GetString("GCP_SUBSCRIPTION_ID"),
		TopicID:        viper.GetString("GCP_TOPIC_ID"),
		PubSubHost:     viper.GetString("GCP_PUBSUB_HOST"),
	}
}

func GetSystemConfig() (SystemConfig, error) {
	viper.AutomaticEnv()

	viper.SetDefault("SYSTEM_MAX_WORKERS", 5)
	viper.SetDefault("SYSTEM_ADMIN_USERNAME", "")
	viper.SetDefault("SYSTEM_ADMIN_PASSWORD", "")

	maxWorkersInterface := viper.Get("SYSTEM_MAX_WORKERS")
	maxWorkers, ok := maxWorkersInterface.(int)
	if !ok {
		if str, ok := maxWorkersInterface.(string); ok {
			parsedValue, err := strconv.Atoi(str)
			if err != nil {
				return SystemConfig{}, errors.New("SYSTEM_MAX_WORKERS must be an integer")
			}
			maxWorkers = parsedValue
		} else {
			return SystemConfig{}, errors.New("SYSTEM_MAX_WORKERS must be an integer")
		}
	}

	return SystemConfig{
		MaxWorker:     maxWorkers,
		AdminUsername: viper.GetString("SYSTEM_ADMIN_USERNAME"),
		AdminPassword: viper.GetString("SYSTEM_ADMIN_PASSWORD"),
	}, nil
}

func GetRedisConfig() RedisConfig {
	viper.AutomaticEnv()

	viper.SetDefault("REDIS_ADDRESS", "")
	viper.SetDefault("REDIS_PASSWORD", "")

	return RedisConfig{
		Address:  viper.GetString("REDIS_ADDRESS"),
		Password: viper.GetString("REDIS_PASSWORD"),
	}
}
