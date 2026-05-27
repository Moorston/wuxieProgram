package minio

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	inner *minio.Client
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

func NewClient(cfg Config) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	return &Client{inner: client}, nil
}

func (c *Client) PresignPutURL(ctx context.Context, bucket, objectName string, expires time.Duration) (string, error) {
	presignedURL, err := c.inner.PresignedPutObject(ctx, bucket, objectName, expires)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

func (c *Client) PresignGetURL(ctx context.Context, bucket, objectName string, expires time.Duration) (string, error) {
	reqParams := make(map[string][]string)
	presignedURL, err := c.inner.PresignedGetObject(ctx, bucket, objectName, expires, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}

func (c *Client) FGetObject(ctx context.Context, bucket, objectName, localPath string) error {
	return c.inner.FGetObject(ctx, bucket, objectName, localPath, minio.GetObjectOptions{})
}

func (c *Client) FPutObject(ctx context.Context, bucket, objectName, localPath, contentType string) error {
	_, err := c.inner.FPutObject(ctx, bucket, objectName, localPath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (c *Client) StatObject(ctx context.Context, bucket, objectName string) (minio.ObjectInfo, error) {
	return c.inner.StatObject(ctx, bucket, objectName, minio.GetObjectOptions{})
}

func (c *Client) EnsureBuckets(ctx context.Context, buckets ...string) error {
	for _, bucket := range buckets {
		exists, err := c.inner.BucketExists(ctx, bucket)
		if err != nil {
			return err
		}
		if !exists {
			if err := c.inner.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) RemoveObject(ctx context.Context, bucket, objectName string) error {
	return c.inner.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

func (c *Client) ListObjects(ctx context.Context, bucket, prefix string) ([]string, int64, error) {
	var names []string
	var totalSize int64

	objects := c.inner.ListObjects(ctx, bucket, prefix, true, nil)
	for obj := range objects {
		if obj.Err != nil {
			return nil, 0, obj.Err
		}
		names = append(names, obj.Key)
		totalSize += obj.Size
	}

	return names, totalSize, nil
}
