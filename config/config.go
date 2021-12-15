package config

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gopkg.in/caarlos0/env.v2"
	"log"
	"time"
)

type key int

const (
	KeyServiceName key = iota + 1
)

const (
	AppSrvName = "onepurse-api"
)

const (
	HTTPClientTimout              = 10 * time.Second
	HTTPClientMaxIdleConns        = 100
	HTTPClientMaxIdleConnsPerHost = 100
)

const (
	HeaderRequestSource      = "X-Request-Source"
	HeaderRequestID          = "X-Request-ID"
	HTTPHeaderStrictTransSec = "Strict-Transport-Security"
)

type Config struct {
	ServiceName            string
	AWSRegion              string `env:"AWS_REGION" required:"true"`
	S3Bucket               string `env:"S3_BUCKET" required:"true"`
	Port                   int    `env:"PORT" required:"true"`
	CognitoUserPoolID      string `env:"COGNITO_USER_POOL_ID" required:"true"`
	CognitoAppClientID     string `env:"COGNITO_APP_CLIENT_ID" required:"true"`
	CognitoAppClientSecret string `env:"COGNITO_APP_CLIENT_SECRET" required:"true"`
	PlaidClientId          string `env:"PLAID_CLIENT_ID" required:"true"`
	PlaidClientName        string `env:"PLAID_CLIENT_NAME" required:"true"`
	PlaidSecret            string `env:"PLAID_SECRET" required:"true"`
	PlaidEnv               string `env:"PLAID_ENV" required:"true"`
	PlaidProducts          string `env:"PLAID_PRODUCTS" required:"true"`
	PlaidCountryCodes      string `env:"PLAID_COUNTRY_CODE" required:"true"`
	PlaidRedirectUri       string `env:"PLAID_REDIRECT_URI" required:"true"`
	OkraToken              string `env:"OKRA_TOKEN" required:"true"`
	MongoURI               string `env:"MONGO_URI" required:"true"` // TODO: set up a database properly before production deployment
	Environment            string `env:"ENVIRONMENT" envDefault:"development"`
	Debug                  bool
}

// New returns a pointer to a config struct
func New() *Config {
	var cfg Config

	cfg.ServiceName = AppSrvName
	if err := godotenv.Load("./.env"); err != nil {
		logrus.Warnf("Error loading the environment: %s. Ignore this warning if you are in the production environment", err.Error())
	}
	logrus.Info(".env file loaded successfully")

	if err := env.Parse(&cfg); err != nil {
		log.Fatal(err.Error())
	}

	return &cfg
}
