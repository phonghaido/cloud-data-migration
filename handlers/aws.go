package handlers

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/phonghaido/cloud-data-migration/internal/config"
	"github.com/sirupsen/logrus"
)

type AWSClient struct {
	S3ClientConfig config.AWSClientConfig
	S3Client       *s3.S3
}

func NewAWSClient(c config.AWSClientConfig) AWSClient {
	session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(c.AWSRegion),
		Credentials: credentials.NewStaticCredentials(
			c.AWSAccessKeyID,
			c.AWSSecretAccessKey,
			"",
		),
	}))

	s3Client := s3.New(session)

	return AWSClient{
		S3ClientConfig: c,
		S3Client:       s3Client,
	}
}

func (a AWSClient) DownloadFromS3(key string) (io.ReadCloser, error) {
	out, err := a.S3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(a.S3ClientConfig.S3Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return out.Body, nil
}

func (a AWSClient) PublishS3Keys(redisClient RedisClient, pubsubClient PubSubClient) error {
	ctx := context.Background()
	resp, err := a.S3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(a.S3ClientConfig.S3Bucket),
	})
	if err != nil {
		return err
	}

	for _, item := range resp.Contents {
		key := *item.Key
		if *item.Size == 0 && strings.HasSuffix(key, "/") {
			continue
		}

		isPublished, err := redisClient.IsPublished(ctx, key)
		if err != nil {
			return err
		}

		if isPublished == 0 {
			pubID, err := pubsubClient.PublishKey(key)
			if err != nil {
				return err
			}
			redisClient.MarkAsPublished(ctx, key, pubID)
			logrus.Infof("Successfully published for the key: %s", key)
		}
	}

	return nil
}

func (a AWSClient) WriteToLocal(key string) error {
	dir := filepath.Dir(key)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	file, err := os.Create(key)
	if err != nil {
		return err
	}
	defer file.Close()

	output, err := a.DownloadFromS3(key)
	if err != nil {
		return err
	}
	defer output.Close()

	_, err = io.Copy(file, output)
	if err != nil {
		return err
	}

	return nil
}
