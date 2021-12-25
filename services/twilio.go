package services

import (
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/sirupsen/logrus"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type Twilio struct {
	config *config.Config
	client *twilio.RestClient
}

func NewTwilioService(cfg *config.Config) (*Twilio, error) {
	twilioClient := twilio.NewRestClient()
	return &Twilio{
		config: cfg,
		client: twilioClient,
	}, nil
}

func (t Twilio) SendMessage(toNumber string, message string) error {
	fromPhone := t.config.TwilioPhoneNumber
	toPhone := toNumber

	params := &openapi.CreateMessageParams{
		Body: &message,
		From: &fromPhone,
		To:   &toPhone,
	}
	res, err := t.client.ApiV2010.CreateMessage(params)
	if err != nil {
		logrus.Errorf("[Twilio]: error sending message: %s", err.Error())
		return err
	}
	fmt.Println(res)
	return nil

}
