package api

import (
	"context"
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/common"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/sirupsen/logrus"
	"net/http"
)

const ContextKeyRequestSource = common.ContextKey("header-request-source")

func RequestTracing(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestSource := r.Header.Get(config.HeaderRequestSource)
		if requestSource == "" {
			writeErrorResponse(w, http.StatusBadRequest, "X-Request-Source header is required")
			return
		}
		requestID := r.Header.Get(config.HeaderRequestID)
		if requestID == "" {
			writeErrorResponse(w, http.StatusBadRequest, "X-Request-ID header is missing")
			return
		}

		tracingContext := tracing.Context{
			RequestID:     requestID,
			RequestSource: requestSource,
		}

		ctx = context.WithValue(ctx, tracing.ContextKeyTracing, tracingContext)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

func Authorization(next http.Handler) http.Handler {
	cfg := config.New() //TODO: Find a way to skip this, makes no sense doing this twice

	fn := func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", cfg.AWSRegion, cfg.CognitoUserPoolID)
		keyset, err := jwk.Fetch(context.Background(), url)
		if err != nil {
			logrus.Fatalf("failed to get JWKS from provided resource: %s", err.Error())
			writeErrorResponse(w, http.StatusUnauthorized, "Not authorized")
			return
		}

		awsJwt := r.Header.Get("Authorization")
		if awsJwt == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "Not authorized")
			return
		}
		token, err := jwt.ParseString(
			awsJwt,
			jwt.WithKeySet(keyset),
			jwt.WithValidate(true),
		)
		if err != nil {
			logrus.Warnf("error parsing token %s", err.Error())
			writeErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		fmt.Println(token)

	}

	return http.HandlerFunc(fn)
}
