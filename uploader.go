package makaroni

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"strings"
)

type UploadFunc func(key string, content string, contentType string) error

func NewUploader(endpoint string, region string, bucket string, keyID string, secret string) (UploadFunc, error) {
	awsSession, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(keyID, secret, ""),
		Endpoint:    &endpoint,
		Region:      &region,
	})
	if err != nil {
		return nil, err
	}
	uploader := s3manager.NewUploader(awsSession)
	upload := func(key string, content string, contentType string) error {
		_, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:      &bucket,
			Key:         &key,
			Body:        strings.NewReader(content),
			ContentType: &contentType,
		})
		return err
	}
	return upload, nil
}
