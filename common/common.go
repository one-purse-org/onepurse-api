package common

import (
	"fmt"
)

type ContentType int

const (
	ContentTypeJSON ContentType = 0
)

type ContextKey string

func (c ContextKey) String() string {
	return fmt.Sprintf("onepurse-api-context-key-%v", string(c))
}
