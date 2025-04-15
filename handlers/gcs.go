package handlers

import (
	"context"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"github.com/phonghaido/cloud-data-migration/internal/config"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type GCSClient struct {
	GCSClientConfig config.GCSClientConfig
	StorageClient   *storage.Client
}

func NewGCSClient(c config.GCSClientConfig) (GCSClient, error) {
	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsFile(c.GCPCredentials))
	if err != nil {
		return GCSClient{}, err
	}

	return GCSClient{
		GCSClientConfig: c,
		StorageClient:   storageClient,
	}, nil
}

func (g GCSClient) UploadFile(reader io.ReadCloser, objectName string) error {
	defer reader.Close()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Minute*15)
	defer cancel()

	wc := g.StorageClient.Bucket(g.GCSClientConfig.BucketName).Object(objectName).NewWriter(ctx)
	if _, err := io.Copy(wc, reader); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	logrus.Infof("Upload complete: %s", objectName)

	return nil
}
