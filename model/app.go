package model

import (
	"context"

	"github.com/VoroniakPavlo/call_audit/auth"
	"github.com/VoroniakPavlo/call_audit/internal/server/interceptor"
)

const (
	AppServiceName = "call_audit"
	NamespaceName  = "webitel"
)

func GetAutherOutOfContext(ctx context.Context) auth.Auther {
	return ctx.Value(interceptor.SessionHeader).(auth.Auther)
}
