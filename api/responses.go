package api

import (
	"context"
	"github.com/isongjosiah/work/onepurse-api/common"
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


