package vmkv

import (
	"context"
	"fmt"
)

func WithKvDbEmptyContext(ctx context.Context) context.Context {
	vctx := &KvDbContext{DbConnections: map[string]*KvOpenConnection{}}
	return context.WithValue(ctx, KvDbContextKey, vctx)
}

func WithKvDbContext(ctx context.Context, vctx *KvDbContext) context.Context {
	return context.WithValue(ctx, KvDbContextKey, vctx)
}

func GetKvDbContext(goContextParent context.Context) (*KvDbContext, error) {
	vctx_ := goContextParent.Value(KvDbContextKey)
	vctx := (vctx_).(*KvDbContext)
	if vctx == nil {
		return nil, fmt.Errorf("kv db context not set")
	}
	return vctx, nil
}
