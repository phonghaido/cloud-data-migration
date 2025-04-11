package handlers

import (
	"context"
	"io"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/phonghaido/cloud-data-migration/internal/config"
	"google.golang.org/api/option"
)

type GCPClient struct {
	GCPClientConfig config.GCPClientConfig
	StorageClient   *storage.Client
	PubSubClient    *pubsub.Client
}

func NewGCPClient(c config.GCPClientConfig) (GCPClient, error) {
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsFile(c.GCPCredentials))
	if err != nil {
		return GCPClient{}, err
	}

	pubsubClient, err := pubsub.NewClient(ctx, c.ProjectID, option.WithCredentialsFile(c.GCPCredentials))
	if err != nil {
		return GCPClient{}, err
	}

	return GCPClient{
		GCPClientConfig: c,
		StorageClient:   storageClient,
		PubSubClient:    pubsubClient,
	}, nil
}

func (g GCPClient) UploadFile(reader io.ReadCloser, objectName string) error {
	defer reader.Close()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*15)
	defer cancel()

	wc := g.StorageClient.Bucket(g.GCPClientConfig.BucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(wc, reader); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}
