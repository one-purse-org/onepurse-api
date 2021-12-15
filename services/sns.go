package services

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ISNSService interface {
	CreatePlatformEndpoint(token string) (*sns.CreatePlatformEndpointOutput, error)
	SendPushNotification(endpoint, message, subject string) (*sns.PublishOutput, error)
}

type SNSService struct {
	config    *config.Config
	snsClient *sns.Client
}

func NewSNSService(cfg *config.Config) (ISNSService, error) {
	conf, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		logrus.Fatalf("[SNS]: unable to load SDK config")
		return nil, errors.Wrap(err, "unable to load SDK config")
	}
	snsClient := sns.NewFromConfig(conf)

	svc := SNSService{
		config:    cfg,
		snsClient: snsClient,
	}
	return svc, nil
}

func (s SNSService) CreatePlatformEndpoint(token string) (*sns.CreatePlatformEndpointOutput, error) {
	params := &sns.CreatePlatformEndpointInput{
		PlatformApplicationArn: aws.String(s.config.SNSPlatformApplicationArn),
		Token:                  aws.String(token),
		Attributes:             nil,
		CustomUserData:         nil,
	}
	output, err := s.snsClient.CreatePlatformEndpoint(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func (s SNSService) SendPushNotification(endpoint, message, subject string) (*sns.PublishOutput, error) {
	params := &sns.PublishInput{
		Message:           aws.String(message),
		MessageAttributes: nil,
		Subject:           aws.String(subject),
		TargetArn:         aws.String(endpoint),
	}
	output, err := s.snsClient.Publish(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	return output, nil
}
