module github.com/isongjosiah/work/onepurse-api

// +heroku goVersion go1.15
go 1.15

require (
	github.com/Uchencho/OkraGo v0.0.0-20200816211114-9b04bc8cf993
	github.com/aws/aws-sdk-go-v2 v1.11.2
	github.com/aws/aws-sdk-go-v2/config v1.11.0
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.7.4
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.6.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.21.0
	github.com/aws/aws-sdk-go-v2/service/sns v1.12.1
	github.com/aws/smithy-go v1.9.0
	github.com/caarlos0/env v3.5.0+incompatible // indirect
	github.com/go-chi/chi v1.5.4
	github.com/go-chi/cors v1.2.0
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/joho/godotenv v1.4.0
	github.com/lestrrat-go/jwx v1.2.8
	github.com/lucsky/cuid v1.2.1
	github.com/pkg/errors v0.9.1
	github.com/plaid/plaid-go v1.6.0
	github.com/pquerna/otp v1.3.0
	github.com/sethvargo/go-password v0.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/twilio/twilio-go v0.19.0
	go.mongodb.org/mongo-driver v1.7.2
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sys v0.0.0-20211205182925-97ca703d548d // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/caarlos0/env.v2 v2.0.0-20161013201842-d0de832ed2fb
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
