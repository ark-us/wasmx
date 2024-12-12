package types

import (
	"context"
	"fmt"

	memc "github.com/loredanacirstea/wasmx/v1/x/wasmx/vm/memory/common"
)

type ContextKey string

const BackgroundProcessesContextKey ContextKey = "background-context"

type BackgroundProcess struct {
	Label          string
	RuntimeHandler memc.RuntimeHandler
	ExecuteHandler func(funcName string) ([]byte, error)
}

type BackgroundProcesses struct {
	Processes map[string]*BackgroundProcess
}

func ContextWithBackgroundProcesses(ctx context.Context) context.Context {
	procc := &BackgroundProcesses{Processes: map[string]*BackgroundProcess{}}
	return context.WithValue(ctx, BackgroundProcessesContextKey, procc)
}

func AddBackgroundProcesses(ctx context.Context, proc *BackgroundProcess) error {
	procc, err := GetBackgroundProcesses(ctx)
	if err != nil {
		return err
	}
	procc.Processes[proc.Label] = proc
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
