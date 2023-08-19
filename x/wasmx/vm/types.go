package vm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dbm "github.com/tendermint/tm-db"

	"github.com/second-state/WasmEdge-go/wasmedge"

	cw8types "mythos/v1/x/wasmx/cw8/types"
	"mythos/v1/x/wasmx/types"
)

var (
	// 0, 1, 2 are used by wasmedge for success, terminate, fail
	Result_OutOfGas = wasmedge.NewResult(wasmedge.ErrCategory_UserLevel, 10)

	LOG_TYPE_WASMX = "wasmx"
)

type NativePrecompileHandler interface {
	IsPrecompile(contractAddress sdk.AccAddress) bool
	Execute(context *Context, contractAddress sdk.AccAddress, input []byte) ([]byte, error)
}

type ContractContext struct {
	FilePath         string
	Vm               *wasmedge.VM
	VmAst            *wasmedge.AST
	VmExecutor       *wasmedge.Executor
	ContractStoreKey []byte
	Context          *Context
	SystemDeps       []types.SystemDep
	Bytecode         []byte // runtime bytecode
	CodeHash         []byte
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
	filepath := c.FilePath
	if types.HasUtf8SystemDep(c.SystemDeps) {
		filepath = ""
	}
	contractVm, cleanups, err := InitiateWasm(newctx, filepath, nil, c.SystemDeps)
	if err != nil {
		runCleanups(cleanups)
		return nil, err
	}
	setExecutionBytecode(newctx, contractVm, types.ENTRY_POINT_EXECUTE)
	newctx.ContractRouter[newctx.Env.Contract.Address.String()].Vm = contractVm

	executeHandler := GetExecuteFunctionHandler(c.SystemDeps)
	_, err = executeHandler(newctx, contractVm, types.ENTRY_POINT_EXECUTE)
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
	Ctx            sdk.Context
	GasMeter       types.GasMeter
	Env            *types.Env
	ContractRouter ContractRouter
	ContractStore  types.KVStore
	CosmosHandler  types.WasmxCosmosHandler
	NativeHandler  NativePrecompileHandler
	ReturnData     []byte
	CurrentCallId  uint32
	Logs           []WasmxLog
	Messages       []cw8types.SubMsg `json:"messages"`
	dbIterators    map[int32]dbm.Iterator
}

func (context *Context) GetCosmosHandler() types.WasmxCosmosHandler {
	return context.CosmosHandler
}

func (context *Context) GetContext() sdk.Context {
	return context.Ctx
}

type EwasmFunctionWrapper struct {
	Name string
	Vm   *wasmedge.VM
}

type WasmxLog struct {
	ContractAddress sdk.AccAddress
	Data            []byte
	Topics          [][32]byte
	Type            string
}
