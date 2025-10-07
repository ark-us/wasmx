package types

import (
	"context"
	"fmt"
	"sync"

	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

type ContextKey string

const BackgroundProcessesContextKey ContextKey = "background-context"

type BackgroundProcess struct {
	Label          string
	RuntimeHandler memc.RuntimeHandler
	ExecuteHandler func(funcName string) ([]byte, error)
}

type BackgroundProcesses struct {
	processes    map[string]*BackgroundProcess
	mtxProcesses sync.Mutex
}

func (v *BackgroundProcesses) GetProcess(key string) (*BackgroundProcess, bool) {
	v.mtxProcesses.Lock()
	defer v.mtxProcesses.Unlock()
	value, found := v.processes[key]
	return value, found
}

func (v *BackgroundProcesses) SetProcess(key string, value *BackgroundProcess) {
	v.mtxProcesses.Lock()
	defer v.mtxProcesses.Unlock()
	v.processes[key] = value
}

func ContextWithBackgroundProcesses(ctx context.Context) context.Context {
	procc := &BackgroundProcesses{processes: map[string]*BackgroundProcess{}}
	return context.WithValue(ctx, BackgroundProcessesContextKey, procc)
}

func AddBackgroundProcesses(ctx context.Context, proc *BackgroundProcess) error {
	procc, err := GetBackgroundProcesses(ctx)
	if err != nil {
		return err
	}
	procc.SetProcess(proc.Label, proc)
	return nil
}

func GetBackgroundProcesses(ctx context.Context) (*BackgroundProcesses, error) {
	procc_ := ctx.Value(BackgroundProcessesContextKey)
	if procc_ == nil {
		return nil, fmt.Errorf("background processes not set on context")
	}
	procc := (procc_).(*BackgroundProcesses)
	if procc == nil {
		return nil, fmt.Errorf("background processes not set on context")
	}
	return procc, nil
}
