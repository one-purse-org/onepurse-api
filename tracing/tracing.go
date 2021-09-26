package tracing

import (
	"github.com/isongjosiah/work/onepurse-api/common"
	"github.com/isongjosiah/work/onepurse-api/config"
	"github.com/lucsky/cuid"
)

const ContextKeyTracing = common.ContextKey("tracing-context")

// Context represents a tracing context
type Context struct {
	RequestID     string
	RequestSource string
}

// NewContext creates a new tracing context
func NewContext() *Context {
	return &Context{
		RequestID:     cuid.New(),
		RequestSource: config.AppSrvName,
	}
}


func (t *Context) GetOutgoingHeaders() map[string]string {
	return map[string]string{
		config.HeaderRequestID:     t.RequestID,
		config.HeaderRequestSource: t.RequestSource,
	}
}
