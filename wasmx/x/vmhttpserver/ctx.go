package vmhttpserver

import (
	"context"
	"fmt"
)

func WithHttpServerEmptyContext(ctx context.Context) context.Context {
	vctx := &HttpServerContext{}
	return context.WithValue(ctx, HttpServerContextKey, vctx)
}

func WithHttpServerContext(ctx context.Context, vctx *HttpServerContext) context.Context {
	return context.WithValue(ctx, HttpServerContextKey, vctx)
}

func GetHttpServerContext(goContextParent context.Context) (*HttpServerContext, error) {
	vctx_ := goContextParent.Value(HttpServerContextKey)
	vctx := (vctx_).(*HttpServerContext)
	if vctx == nil {
		return nil, fmt.Errorf("kv db context not set")
	}
	return vctx, nil
}
