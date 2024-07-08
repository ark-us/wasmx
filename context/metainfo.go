package config

import (
	"context"
	"fmt"
)

type ExecutionMetaInfoContextKey string

const ExecutionMetaInfoKey ExecutionMetaInfoContextKey = "ExecutionMetaInfo"

type ExecutionMetaInfo struct {
	Data map[string]interface{}
}

func NewExecutionMetaInfo(data map[string]interface{}) *ExecutionMetaInfo {
	return &ExecutionMetaInfo{
		Data: data,
	}
}

func WithExecutionMetaInfoEmpty(ctx context.Context) (context.Context, *ExecutionMetaInfo) {
	data := &ExecutionMetaInfo{Data: map[string]interface{}{}}
	return context.WithValue(ctx, ExecutionMetaInfoKey, data), data
}

func WithExecutionMetaInfo(ctx context.Context, data *ExecutionMetaInfo) context.Context {
	return context.WithValue(ctx, ExecutionMetaInfoKey, data)
}

func GetExecutionMetaInfo(ctx context.Context) (*ExecutionMetaInfo, error) {
	datai := ctx.Value(ExecutionMetaInfoKey)
	data, ok := (datai).(*ExecutionMetaInfo)
	if !ok {
		return nil, fmt.Errorf("ExecutionMetaInfo not set on context")
	}
	if data == nil {
		return nil, fmt.Errorf("ExecutionMetaInfo not set on context")
	}
	return data, nil
}

func ResetExecutionMetaInfo(ctx context.Context) {
	datai := ctx.Value(ExecutionMetaInfoKey)
	data, ok := (datai).(*ExecutionMetaInfo)
	if ok && data != nil {
		data.Data = map[string]interface{}{}
	}
}
