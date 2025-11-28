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
