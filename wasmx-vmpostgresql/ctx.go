package vmpostgresql

import (
	"context"
	"fmt"
)

func WithSqlEmptyContext(ctx context.Context) context.Context {
	vctx := &SqlContext{dbConnections: map[string]*SqlOpenConnection{}}
	return context.WithValue(ctx, SqlContextKey, vctx)
}

func WithSqlContext(ctx context.Context, vctx *SqlContext) context.Context {
	return context.WithValue(ctx, SqlContextKey, vctx)
}

func GetSqlContext(goContextParent context.Context) (*SqlContext, error) {
	vctx_ := goContextParent.Value(SqlContextKey)
	vctx := (vctx_).(*SqlContext)
	if vctx == nil {
		return nil, fmt.Errorf("postgresql context not set")
	}
	return vctx, nil
}
