package services

import (
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/pkg/errors"
)

type AWS struct {
	Cognito ICognitoService
	S3      IS3Service
	SNS     ISNSService
}

func NewAWS(cfg *config.Config) (*AWS, error) {
	s3, err := NewS3Service(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("%v S3", "Failed to setup service:"))
	}

	cognito, err := NewCognitoService(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("%v Cognito", "Failed to setup service:"))
	}

	sns, err := NewSNSService(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("%v SNS", "Failed to setup service:"))
	}

	aws := &AWS{
		Cognito: cognito,
		S3:      s3,
		SNS:     sns,
	}

	return aws, nil
}
