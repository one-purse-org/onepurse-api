package services

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"mime/multipart"
)

const partSize = 2 * 1024 * 1024

type IS3Service interface {
	Upload(filename string, file multipart.File) (string, error)
}

type S3Service struct {
	config     *config.Config
	s3Client   *s3.Client
	Uploader   *manager.Uploader
	Downloader *manager.Downloader
}

func NewS3Service(cfg *config.Config) (IS3Service, error) {
	conf, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		logrus.Fatalf("[COGNITO]: unable to loaad SDK config")
		return nil, errors.Wrap(err, "unable to load SDK config")
	}
	s3Client := s3.NewFromConfig(conf)

	svc := S3Service{
		config:   cfg,
		s3Client: s3Client,
		Uploader: manager.NewUploader(s3Client, func(u *manager.Uploader) {
			u.PartSize = partSize
		}),
		Downloader: manager.NewDownloader(s3Client, func(d *manager.Downloader) {
			d.PartSize = partSize
		}),
	}
	return svc, nil
}

func (s S3Service) Upload(filename string, file multipart.File) (string, error) {
	params := &s3.PutObjectInput{
		Bucket: aws.String(s.config.S3Bucket),
		Key:    aws.String(filename),
		ACL:    "public-read",
		Body:   file,
	}
	n, err := s.Uploader.Upload(context.TODO(), params)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to %s, %v", filename, err)
	}
	return n.Location, nil
}
