package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Object struct {
	Key           string    `json:"key"`
	Size          int64     `json:"size"`
	StorageClasss string    `json:"storage_classs"`
	LastModified  time.Time `json:"last_modified"`
}

type S3Client struct {
	Bucket   string
	api      *s3.S3
	MaxFiles int64
}

func NewS3Client(bucket string, maxFiles int64) S3Client {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	return S3Client{bucket, s3.New(sess), maxFiles}
}

func (client S3Client) GetFiles(continuationToken *string) ([]S3Object, error, *string) {
	var objects = make([]S3Object, 0)

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(client.Bucket),
		MaxKeys: aws.Int64(client.MaxFiles),
	}

	if continuationToken != nil {
		input.ContinuationToken = continuationToken
	}

	resp, err := client.api.ListObjectsV2(input)
	if err != nil {
		return objects, err, nil
	}

	for _, item := range resp.Contents {
		obj := S3Object{
			Key:           *item.Key,
			LastModified:  *item.LastModified,
			Size:          *item.Size,
			StorageClasss: *item.StorageClass,
		}
		objects = append(objects, obj)
	}

	return objects, nil, resp.NextContinuationToken
}
