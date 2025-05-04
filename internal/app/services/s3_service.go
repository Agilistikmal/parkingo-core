package services

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type S3Service struct {
	Client     *minio.Client
	Endpoint   string
	BucketName string
}

func NewS3Service() *S3Service {
	endpoint := viper.GetString("s3.endpoint")
	accessKey := viper.GetString("s3.access_key")
	secretKey := viper.GetString("s3.secret_key")
	bucketName := viper.GetString("s3.bucket_name")
	token := viper.GetString("s3.token")

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, token),
		Secure: true,
		Region: "auto",
	})
	if err != nil {
		logrus.Fatal("Failed to create S3 client: ", err)
	}

	return &S3Service{
		Client:     client,
		Endpoint:   endpoint,
		BucketName: bucketName,
	}
}

func (s *S3Service) UploadFile(filePath string, objectName string, contentType string) error {
	_, err := s.Client.FPutObject(context.Background(), s.BucketName, objectName, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		logrus.Error("Failed to upload file: ", err)
		return err
	}

	return nil
}
