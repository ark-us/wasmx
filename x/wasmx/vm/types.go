package vm

import (
	"context"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"golang.org/x/sync/errgroup"

	dbm "github.com/cometbft/cometbft-db"

	"github.com/second-state/WasmEdge-go/wasmedge"

	cw8types "mythos/v1/x/wasmx/cw8/types"
	"mythos/v1/x/wasmx/types"
)

var (
	// 0, 1, 2 are used by wasmedge for success, terminate, fail
	// Result_OutOfGas = wasmedge.NewResult(wasmedge.ErrCategory_UserLevel, 10)

	LOG_TYPE_WASMX = "wasmx"
)

type NativePrecompileHandler interface {
	IsPrecompile(contractAddress sdk.AccAddress) bool
	Execute(ctx *Context, contractAddress sdk.AccAddress, input []byte) ([]byte, error)
}

type ContractContext struct {
	Vm           *wasmedge.VM
	VmAst        *wasmedge.AST
	VmExecutor   *wasmedge.Executor
	Context      *Context
	ContractInfo types.ContractDependency
}

// TODO deeper clone - this may be rewritten on each nested call
// we don't want to rewrite the original data, which acts as a cache
func (c ContractContext) CloneShallow() *ContractContext {
	return &ContractContext{
		Vm:           c.Vm,
		VmAst:        c.VmAst,
		VmExecutor:   c.VmExecutor,
		Context:      c.Context,
		ContractInfo: c.ContractInfo,
	}
}

func (c ContractContext) Execute(newctx *Context) ([]byte, error) {
	found := newctx.NativeHandler.IsPrecompile(newctx.Env.Contract.Address)
	if found {
		data, err := newctx.NativeHandler.Execute(newctx, newctx.Env.Contract.Address, newctx.Env.CurrentCall.CallData)
		if err != nil {
			return nil, err
		}
		newctx.ReturnData = data
		return data, nil
	}
	filepath := c.ContractInfo.FilePath
	if types.HasUtf8SystemDep(c.ContractInfo.SystemDeps) {
		filepath = ""
	}
	contractVm, cleanups, err := InitiateWasm(newctx, filepath, nil, c.ContractInfo.SystemDeps)
	if err != nil {
		runCleanups(cleanups)
		return nil, err
	}
	setExecutionBytecode(newctx, contractVm, types.ENTRY_POINT_EXECUTE)
	newctx.ContractRouter[newctx.Env.Contract.Address.String()].Vm = contractVm

	executeHandler := GetExecuteFunctionHandler(c.ContractInfo.SystemDeps)
	_, err = executeHandler(newctx, contractVm, types.ENTRY_POINT_EXECUTE, make([]interface{}, 0))
	if err != nil {
		runCleanups(cleanups)
		return nil, err
	}
	runCleanups(cleanups)
	return newctx.ReturnData, nil
}

// key is a bech32 string
type ContractRouter = map[string]*ContractContext

type Context struct {
	GoRoutineGroup  *errgroup.Group
	GoContextParent context.Context
	Ctx             sdk.Context
	GasMeter        types.GasMeter
	Env             *types.Env
	ContractRouter  ContractRouter
	ContractStore   prefix.Store
	CosmosHandler   types.WasmxCosmosHandler
	App             types.Application
	NativeHandler   NativePrecompileHandler
	ReturnData      []byte
	FinishData      []byte
	CurrentCallId   uint32
	Logs            []WasmxLog
	Messages        []cw8types.SubMsg `json:"messages"`
	dbIterators     map[int32]dbm.Iterator
}

func (ctx *Context) GetCosmosHandler() types.WasmxCosmosHandler {
	return ctx.CosmosHandler
}

func (ctx *Context) GetApplication() types.Application {
	return ctx.App
}

func (ctx *Context) GetContext() sdk.Context {
	return ctx.Ctx
}

func (ctx *Context) GetVmFromContext() (*wasmedge.VM, error) {
	addr := ctx.Env.Contract.Address
	contractCtx, ok := ctx.ContractRouter[addr.String()]
	if !ok {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "contract context not found for address %s", addr.String())
	}
	return contractCtx.Vm, nil
}

func (ctx *Context) MustGetVmFromContext() *wasmedge.VM {
	vm, err := ctx.GetVmFromContext()
	if err != nil {
		panic(err.Error())
	}
	return vm
}

type EwasmFunctionWrapper struct {
	Name string
	Vm   *wasmedge.VM
}

type WasmxLog struct {
	ContractAddress  sdk.AccAddress
	SystemDependency string
	Data             []byte
	Topics           [][32]byte
	Type             string
}

type StorageRange struct {
	StartKey []byte `json:"start_key"`
	EndKey   []byte `json:"end_key"`
	Reverse  bool   `json:"reverse"`
}

type StoragePair struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type StoragePairs struct {
	Values []StoragePair `json:"values"`
}

type VerifyCosmosTxResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error"`
}
