package vmsql

import (
	"context"
	"database/sql"
	"fmt"
)

func WithSqlEmptyContext(ctx context.Context) context.Context {
	vctx := &SqlContext{DbConnections: map[string]*sql.DB{}}
	return context.WithValue(ctx, SqlContextKey, vctx)
}

func WithSqlContext(ctx context.Context, vctx *SqlContext) context.Context {
	return context.WithValue(ctx, SqlContextKey, vctx)
}

func GetSqlContext(goContextParent context.Context) (*SqlContext, error) {
	vctx_ := goContextParent.Value(SqlContextKey)
	vctx := (vctx_).(*SqlContext)
	if vctx == nil {
		return nil, fmt.Errorf("sql context not set")
	}
	return vctx, nil
}
