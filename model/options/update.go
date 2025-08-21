package options

import (
	"context"
	"time"

	"github.com/VoroniakPavlo/call_audit/auth"

)

type UpdateOptions interface {
	context.Context
	GetAuthOpts() auth.Auther
	GetFields() []string
	GetUnknownFields() []string
	GetDerivedSearchOpts() map[string]*SearchOptions
	RequestTime() time.Time
	GetMask() []string
	GetParentID() int64
	GetIDs() []int64
}
