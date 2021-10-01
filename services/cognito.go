package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type ICognitoService interface {
	Login(l *model.LoginRequest) (*model.AuthResponse, error)
	SignUp(r *model.RegistrationRequest) (*model.SignupResponse, error)
	ConfirmSignUp(v *model.VerificationRequest) (bool, error)
	ResendConfirmationCode(r *model.ResendConfirmationCodeRequest) (bool, error)
}

type CognitoService struct {
	config        *config.Config
	cognitoClient *cognito.Client
}

func NewCognitoService(cfg *config.Config) (ICognitoService, error) {
	conf, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		logrus.Fatalf("[COGNITO]: unable to load SDK config")
		return nil, errors.Wrap(err, "unable to load SDK config")
	}
	cognitoClient := cognito.NewFromConfig(conf)

	svc := CognitoService{
		config:        cfg,
		cognitoClient: cognitoClient,
	}

	return svc, nil
}

func (c CognitoService) generateCognitoSecretHash(username string) string {
	mac := hmac.New(sha256.New, []byte(c.config.CognitoAppClientSecret))
	mac.Write([]byte(username + c.config.CognitoAppClientID))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (c CognitoService) Login(l *model.LoginRequest) (*model.AuthResponse, error) {
	params := &cognito.InitiateAuthInput{
		AuthFlow: "USER_PASSWORD_AUTH",
		AuthParameters: map[string]string{
			"USERNAME":    l.UserName,
			"PASSWORD":    l.Password,
			"SECRET_HASH": c.generateCognitoSecretHash(l.UserName),
		},
		ClientId: aws.String(c.config.CognitoAppClientID),
	}

	now := time.Now()
	cognitoResponse, err := c.cognitoClient.InitiateAuth(context.TODO(), params)
	if err != nil {
		return nil, err
	}

	authResponse := &model.AuthResponse{
		AccessToken:  *cognitoResponse.AuthenticationResult.AccessToken,
		RefreshToken: *cognitoResponse.AuthenticationResult.RefreshToken,
		ExpiresAt:    now.Add(time.Second * time.Duration(cognitoResponse.AuthenticationResult.ExpiresIn)),
	}

	return authResponse, nil
}

func (c CognitoService) SignUp(r *model.RegistrationRequest) (*model.SignupResponse, error) {
	temp := strings.Split(r.FullName, " ")
	params := &cognito.SignUpInput{
		ClientId:   aws.String(c.config.CognitoAppClientID),
		Username:   aws.String(temp[1]),
		Password:   aws.String(r.Password),
		SecretHash: aws.String(c.generateCognitoSecretHash(temp[1])),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("phone_number"),
				Value: aws.String(r.Phone),
			},
			{
				Name:  aws.String("name"),
				Value: aws.String(r.FullName),
			},
			{
				Name:  aws.String("email"),
				Value: aws.String(r.Email),
			},
		},
	}

	cognitoResponse, err := c.cognitoClient.SignUp(context.TODO(), params)
	if err != nil {
		return nil, err
	}

	signupResponse := &model.SignupResponse{
		IsConfirmed:    cognitoResponse.UserConfirmed,
		DeliveryMedium: string(cognitoResponse.CodeDeliveryDetails.DeliveryMedium),
		Destination:    *cognitoResponse.CodeDeliveryDetails.Destination,
	}
	return signupResponse, nil
}

func (c CognitoService) ConfirmSignUp(v *model.VerificationRequest) (bool, error) {
	params := &cognito.ConfirmSignUpInput{
		Username:         aws.String(v.UserName),
		ConfirmationCode: aws.String(v.Code),
		ClientId:         aws.String(c.config.CognitoAppClientID),
		SecretHash:       aws.String(c.generateCognitoSecretHash(v.UserName)),
	}

	_, err := c.cognitoClient.ConfirmSignUp(context.TODO(), params)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c CognitoService) ResendConfirmationCode(r *model.ResendConfirmationCodeRequest) (bool, error) {
	params := &cognito.ResendConfirmationCodeInput{
		ClientId:   aws.String(c.config.CognitoAppClientID),
		Username:   aws.String(r.UserName),
		SecretHash: aws.String(c.generateCognitoSecretHash(r.UserName)),
	}
	_, err := c.cognitoClient.ResendConfirmationCode(context.TODO(), params)
	if err != nil {
		return false, err
	}

	return true, nil
}
