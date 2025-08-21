package options

import (
	"context"
	"time"

	"github.com/webitel/call_audit/auth"
)

type DeleteOptions interface {
	context.Context
	GetAuthOpts() auth.Auther
	RequestTime() time.Time

	// Additional filtering

	GetFilters() map[string]any
	RemoveFilter(string)
	AddFilter(string, any)
	GetFilter(string) any

	// If connection to parent object required
	GetParentID() int64

	// ID filtering
	GetIDs() []int64
}
