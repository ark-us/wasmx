package vmimap

import (
	"context"
	"fmt"
)

func WithImapEmptyContext(ctx context.Context) context.Context {
	vctx := &ImapContext{DbConnections: map[string]*ImapOpenConnection{}}
	return context.WithValue(ctx, ImapContextKey, vctx)
}

func WithImapContext(ctx context.Context, vctx *ImapContext) context.Context {
	return context.WithValue(ctx, ImapContextKey, vctx)
}

func GetImapContext(goContextParent context.Context) (*ImapContext, error) {
	vctx_ := goContextParent.Value(ImapContextKey)
	vctx := (vctx_).(*ImapContext)
	if vctx == nil {
		return nil, fmt.Errorf("IMAP context not set")
	}
	return vctx, nil
}
