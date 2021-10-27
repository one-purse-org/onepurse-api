package api_test

import (
	"context"
	"fmt"
	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/isongjosiah/work/onepurse-api/api"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/isongjosiah/work/onepurse-api/dal/model"
	"github.com/isongjosiah/work/onepurse-api/deps"
	"github.com/isongjosiah/work/onepurse-api/services"
	"testing"
)

type mockSignUpAPI func(ctx context.Context, params *cognito.SignUpInput, optFns ...func(options *cognito.Options)) (*cognito.SignUpOutput, error)
type mockConfirmSignUpAPI func(ctx context.Context, params *cognito.ConfirmSignUpInput, optFns ...func(options *cognito.Options)) (*cognito.SignUpOutput, error)

func (m mockSignUpAPI) SignUp(ctx context.Context, params *cognito.SignUpInput, optFns ...func(options *cognito.Options)) (*cognito.SignUpOutput, error) {
	return m(ctx, params, optFns...)
}

func init() {
	fmt.Println("setting up testing function")
}

func TestSignUpToCognito(t *testing.T) {
	var (
		a *api.API
	)
	cfg := &config.Config{
		AWSRegion:              "us-east-1",
		CognitoUserPoolID:      "us-east-1_05PaYPw56",
		CognitoAppClientID:     "6aq08nni11hgfduaroht6ummk4",
		CognitoAppClientSecret: "1a3452pc4gbkve13hjjrlk138atvtjd97o3g5pilk65onsbmu44p",
	}
	cog, _ := services.NewCognitoService(cfg)
	dependencies := &deps.Dependencies{
		AWS: &services.AWS{
			Cognito: cog,
		},
		DAL: nil,
	}
	a = &api.API{
		Config: cfg,
		Deps:   dependencies,
	}

	cases := []struct {
		email    string
		fullName string
		phone    string
		password string
		expect   *model.SignupResponse
	}{
		{"random@email.com", "random guy", "+1 582-202-3524", "&fjfM,bJ", &model.SignupResponse{IsConfirmed: true, DeliveryMedium: "EMAIL", Destination: "random@email.com"}},
		{"random2@email.com", "random dude", "+1 582-529-4839", "3XwA[q;W", &model.SignupResponse{IsConfirmed: true, DeliveryMedium: "EMAIL", Destination: "random@email.com"}},
		{"random3@email.com", "random chick", "+1 518-861-1309", "F]4@Spd5", &model.SignupResponse{IsConfirmed: true, DeliveryMedium: "EMAIL", Destination: "random@email.com"}},
		{"random4@email.com", "random babe", "+1 206-555-6525", "34d#rdNb", &model.SignupResponse{IsConfirmed: true, DeliveryMedium: "EMAIL", Destination: "random@email.com"}},
	}

	for _, tt := range cases {
		t.Run("signup to cognito", func(t *testing.T) {
			temp := &tt
			var user *model.RegistrationRequest
			user = &model.RegistrationRequest{
				Email:    temp.email,
				FullName: temp.fullName,
				Password: temp.password,
				Phone:    temp.phone,
			}
			fmt.Println()
			content, err := a.Deps.AWS.Cognito.SignUp(user)
			if err != nil {
				t.Fatalf("expect no error, got %v", err)
			}
			if e, a := tt.expect, content; e != a {
				t.Errorf("expect %v, got %v", e, a)
			}
		})
	}
}
