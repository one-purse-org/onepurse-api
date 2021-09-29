package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/isongjosiah/work/onepurse-api/common"
	"github.com/isongjosiah/work/onepurse-api/tracing"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	requestIDKey     = "request_id"
	requestSourceKey = "request_source"
)

// ServerResponse represents the structure of the response sent to the client
type ServerResponse struct {
	Err         error
	Message     string
	StatusCode  int
	Context     context.Context
	ContentType common.ContentType
	Payload     interface{}
}

type ErrorResponse struct {
	ErrorMessage string `json:"error_message"`
	ErrorCode    int    `json:"error_code"`
}

func (r *ServerResponse) Error() string {
	return r.Err.Error()
}

func WithContext(tracingContext *tracing.Context) *logrus.Entry {
	fields := logrus.Fields{}

	if tracingContext != nil {
		fields[requestIDKey] = tracingContext.RequestID
		fields[requestSourceKey] = tracingContext.RequestSource
	}

	return logrus.WithFields(fields)
}

func NewUserFacingError(message string, tracingContext *tracing.Context) error {
	return fmt.Errorf("%v, request-id: %v", message, tracingContext.RequestID)
}

func RespondWithError(err error, message string, httpStatusCode int, tracingContext *tracing.Context) *ServerResponse {
	var wrappedErr error
	if err != nil {
		wrappedErr = errors.Wrap(err, message)
	} else {
		wrappedErr = errors.New(message)
	}

	WithContext(tracingContext).WithFields(logrus.Fields{
		"err": wrappedErr,
	}).Warn(message)

	return &ServerResponse{
		Err:        NewUserFacingError(message, tracingContext),
		StatusCode: httpStatusCode,
		Message:    message,
	}
}

func RespondWithJSONPayload(ctx *tracing.Context, data interface{}) *ServerResponse {
	result, err := json.Marshal(data)

	if err != nil {
		message := "Error creating json response"

		WithContext(ctx).WithFields(logrus.Fields{
			"err": err,
		}).Error(message)

		return RespondWithError(err, message, http.StatusInternalServerError, ctx)
	}

	return &ServerResponse{Payload: result}
}

func RespondWithWarning(err error, message string, httpStatusCode int, tracingContext *tracing.Context) *ServerResponse {
	wrappedErr := errors.Wrap(err, message)

	WithContext(tracingContext).WithFields(logrus.Fields{
		"err": wrappedErr,
	}).Warn(message)

	return &ServerResponse{
		Err:        NewUserFacingError(message, tracingContext),
		StatusCode: httpStatusCode,
		Message:    message,
	}
}

func WriteJSONResponse(rw http.ResponseWriter, statusCode int, content []byte) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)
	rw.Write(content)
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, errString string) {
	r := RespondWithError(nil, errString, http.StatusBadRequest, nil)
	errorResponse, _ := json.Marshal(r)
	WriteJSONResponse(w, statusCode, errorResponse)
}
