package options

import (
	"context"
	"time"

	"github.com/webitel/call_audit/auth"
)

type CreateOptions interface {
	context.Context
	GetAuthOpts() auth.Auther
	RequestTime() time.Time
	GetFields() []string
	GetDerivedSearchOpts() map[string]*SearchOptions
	GetUnknownFields() []string
	GetIDs() []int64
	GetParentID() int64
	GetChildID() int64
}
