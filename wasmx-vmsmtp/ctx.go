package vmsmtp

import (
	"context"
	"fmt"
)

func WithSmtpEmptyContext(ctx context.Context) context.Context {
	vctx := &SmtpContext{
		DbConnections:     map[string]*SmtpOpenConnection{},
		ServerConnections: map[string]*SmtpServerConnection{},
	}
	return context.WithValue(ctx, SmtpContextKey, vctx)
}

func WithSmtpContext(ctx context.Context, vctx *SmtpContext) context.Context {
	return context.WithValue(ctx, SmtpContextKey, vctx)
}

func GetSmtpContext(goContextParent context.Context) (*SmtpContext, error) {
	vctx_ := goContextParent.Value(SmtpContextKey)
	vctx := (vctx_).(*SmtpContext)
	if vctx == nil {
		return nil, fmt.Errorf("SMTP context not set")
	}
	return vctx, nil
}
