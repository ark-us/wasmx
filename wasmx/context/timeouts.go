package config

import (
	"context"
	"fmt"
	"sync"
)

type TimeoutGoroutinesContextKey string

const TimeoutGoroutinesKey TimeoutGoroutinesContextKey = "TimeoutGoroutines"

type TimeoutGoroutinesInfo struct {
	GoRoutines map[string]context.CancelFunc
	mu         sync.Mutex
}

func NewTimeoutGoroutinesInfo(data map[string]context.CancelFunc) *TimeoutGoroutinesInfo {
	return &TimeoutGoroutinesInfo{
		GoRoutines: data,
	}
}

func WithTimeoutGoroutinesInfoEmpty(ctx context.Context) (context.Context, *TimeoutGoroutinesInfo) {
	data := &TimeoutGoroutinesInfo{GoRoutines: map[string]context.CancelFunc{}}
	return context.WithValue(ctx, TimeoutGoroutinesKey, data), data
}

func SetTimeoutGoroutine(ctx context.Context, key string, cancelfn context.CancelFunc) error {
	datai := ctx.Value(TimeoutGoroutinesKey)
	data, ok := (datai).(*TimeoutGoroutinesInfo)
	if !ok {
		return fmt.Errorf("TimeoutGoroutinesInfo not set on context")
	}
	if data == nil {
		return fmt.Errorf("TimeoutGoroutinesInfo not set on context")
	}
	data.mu.Lock()
	data.GoRoutines[key] = cancelfn
	data.mu.Unlock()
	return nil
}

func GetTimeoutGoroutine(ctx context.Context, key string) (context.CancelFunc, error) {
	datai := ctx.Value(TimeoutGoroutinesKey)
	data, ok := (datai).(*TimeoutGoroutinesInfo)
	if !ok {
		return nil, fmt.Errorf("TimeoutGoroutinesInfo not set on context")
	}
	if data == nil {
		return nil, fmt.Errorf("TimeoutGoroutinesInfo not set on context")
	}
	data.mu.Lock()
	gorout, found := data.GoRoutines[key]
	data.mu.Unlock()
	if !found {
		return nil, nil
	}
	return gorout, nil
}

func RemoveTimeoutGoroutine(ctx context.Context, key string) error {
	datai := ctx.Value(TimeoutGoroutinesKey)
	data, ok := (datai).(*TimeoutGoroutinesInfo)
	if !ok {
		return fmt.Errorf("TimeoutGoroutinesInfo not set on context")
	}
	if data == nil {
		return fmt.Errorf("TimeoutGoroutinesInfo not set on context")
	}
	data.mu.Lock()
	delete(data.GoRoutines, key)
	data.mu.Unlock()
	return nil
}
