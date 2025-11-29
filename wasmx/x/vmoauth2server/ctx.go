package vmoauth2server

import (
	"context"
	"fmt"
)

func GetOAuth2ServerContext(ctx context.Context) (*OAuth2ServerContext, error) {
	val := ctx.Value(OAuth2ServerContextKey)
	if val == nil {
		return nil, fmt.Errorf("oauth2server context not found")
	}
	oauth2Ctx, ok := val.(*OAuth2ServerContext)
	if !ok {
		return nil, fmt.Errorf("oauth2server context has incorrect type")
	}
	return oauth2Ctx, nil
}

func WithOAuth2ServerEmptyContext(ctx context.Context) context.Context {
	vctx := &OAuth2ServerContext{Instances: map[string]*OAuth2ServerInstance{}}
	return context.WithValue(ctx, OAuth2ServerContextKey, vctx)
}

func WithOAuth2ServerContext(ctx context.Context, vctx *OAuth2ServerContext) context.Context {
	return context.WithValue(ctx, OAuth2ServerContextKey, vctx)
}
