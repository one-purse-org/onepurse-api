package services

import (
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/pkg/errors"
)

type AWS struct {
	Cognito ICognitoService
}

func NewAWS(cfg *config.Config) (*AWS, error) {
	cognito, err := NewCognitoService(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("%v Cognito", "Failed to setup service:"))
	}

	aws := &AWS{
		Cognito: cognito,
	}

	return aws, nil
}
