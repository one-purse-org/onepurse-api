package api

import (
	"context"
	"github.com/isongjosiah/work/onepurse-api/common"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/isongjosiah/work/onepurse-api/tracing"
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
