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
	"time"
)

type ICognitoService interface {
	Login(l *model.LoginRequest) (*model.AuthResponse, error)
	SignUp(r *model.RegistrationRequest) (*model.SignupResponse, error)
	ConfirmSignUp(v *model.VerificationRequest) (bool, error)
	ResendCode(email string) (*cognito.ResendConfirmationCodeOutput, error)
	ForgetPassword(email string) (*cognito.ForgotPasswordOutput, error)
	ConfirmForgotPassword(p *model.ConfirmForgotPasswordRequest) (bool, error)
	ChangePassword(p *model.ChangePassword) (bool, error)
	UpdateUsername(ua *model.UpdateUsername) error
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
			"USERNAME":    l.Username,
			"PASSWORD":    l.Password,
			"SECRET_HASH": c.generateCognitoSecretHash(l.Username),
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
	params := &cognito.SignUpInput{
		ClientId:   aws.String(c.config.CognitoAppClientID),
		Password:   aws.String(r.Password),
		Username:   aws.String(r.Email),
		SecretHash: aws.String(c.generateCognitoSecretHash(r.Email)),
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
		Username:         aws.String(v.Email),
		ConfirmationCode: aws.String(v.Code),
		ClientId:         aws.String(c.config.CognitoAppClientID),
		SecretHash:       aws.String(c.generateCognitoSecretHash(v.Email)),
	}

	_, err := c.cognitoClient.ConfirmSignUp(context.TODO(), params)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c CognitoService) ResendCode(email string) (*cognito.ResendConfirmationCodeOutput, error) {
	params := &cognito.ResendConfirmationCodeInput{
		ClientId:   aws.String(c.config.CognitoAppClientID),
		Username:   aws.String(email),
		SecretHash: aws.String(c.generateCognitoSecretHash(email)),
	}

	resp, err := c.cognitoClient.ResendConfirmationCode(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c CognitoService) ForgetPassword(email string) (*cognito.ForgotPasswordOutput, error) {
	params := &cognito.ForgotPasswordInput{
		ClientId:   aws.String(c.config.CognitoAppClientID),
		Username:   aws.String(email),
		SecretHash: aws.String(c.generateCognitoSecretHash(email)),
	}
	resp, err := c.cognitoClient.ForgotPassword(context.TODO(), params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c CognitoService) ConfirmForgotPassword(p *model.ConfirmForgotPasswordRequest) (bool, error) {
	params := &cognito.ConfirmForgotPasswordInput{
		ClientId:         aws.String(c.config.CognitoAppClientID),
		ConfirmationCode: aws.String(p.Code),
		Password:         aws.String(p.ProposedPassword),
		Username:         aws.String(p.Username),
		SecretHash:       aws.String(c.generateCognitoSecretHash(p.Username)),
	}
	_, err := c.cognitoClient.ConfirmForgotPassword(context.TODO(), params)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c CognitoService) ChangePassword(p *model.ChangePassword) (bool, error) {
	params := &cognito.ChangePasswordInput{
		AccessToken:      aws.String(p.AccessToken),
		PreviousPassword: aws.String(p.PreviousPassword),
		ProposedPassword: aws.String(p.ProposedPassword),
	}
	_, err := c.cognitoClient.ChangePassword(context.TODO(), params)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c CognitoService) UpdateUsername(ua *model.UpdateUsername) error {
	params := &cognito.UpdateUserAttributesInput{
		AccessToken: aws.String(ua.AccessToken),
		UserAttributes: []types.AttributeType{
			{
				Name:  aws.String("preferred_username"),
				Value: aws.String(ua.PreferredUsername),
			},
		},
	}
	_, err := c.cognitoClient.UpdateUserAttributes(context.TODO(), params)
	if err != nil {
		return err
	}
	return nil
}
