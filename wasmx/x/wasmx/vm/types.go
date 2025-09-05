package vm

import (
	"context"
	"strings"

	"golang.org/x/sync/errgroup"

	log "cosmossdk.io/log"
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	cw8types "github.com/loredanacirstea/wasmx/x/wasmx/cw8/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

var (
	LOG_TYPE_WASMX = "wasmx"
)

type NativePrecompileHandler interface {
	IsPrecompile(contractAddress sdk.AccAddress) bool
	Execute(ctx *Context, contractAddress mcodec.AccAddressPrefixed, input []byte) ([]byte, error)
}

// key is a bech32 string
type ContractRouter = map[string]*Context

type Context struct {
	GoRoutineGroup           *errgroup.Group
	GoContextParent          context.Context
	Ctx                      sdk.Context
	Logger                   func(ctx sdk.Context) log.Logger
	GasMeter                 types.GasMeter
	Env                      *types.Env
	ContractRouter           ContractRouter
	ContractStore            prefix.Store
	CosmosHandler            types.WasmxCosmosHandler
	App                      types.Application
	NativeHandler            NativePrecompileHandler
	ReturnData               []byte
	FinishData               []byte
	CurrentSubCallLevel      uint32
	CurrentSubCallId         uint32
	CurrentSubCallLevelCount uint32
	Logs                     []WasmxLog
	CosmosEvents             []types.Event
	Messages                 []cw8types.SubMsg `json:"messages"`
	dbIterators              map[int32]types.Iterator
	RuntimeHandler           memc.RuntimeHandler
	ContractInfo             *types.ContractDependency
	newIVmFn                 memc.NewIVmFn
}

// not used at this point
func (c *Context) Clone() *Context {
	return &Context{
		GoRoutineGroup:           c.GoRoutineGroup,
		GoContextParent:          c.GoContextParent,
		Ctx:                      c.Ctx,
		Logger:                   c.Logger,
		GasMeter:                 c.GasMeter,
		Env:                      c.Env.Clone(),
		ContractRouter:           c.ContractRouter,
		ContractStore:            c.ContractStore,
		CosmosHandler:            c.CosmosHandler,
		App:                      c.App,
		NativeHandler:            c.NativeHandler,
		ReturnData:               []byte{},
		FinishData:               []byte{},
		CurrentSubCallLevel:      c.CurrentSubCallLevel,
		CurrentSubCallId:         c.CurrentSubCallId,
		CurrentSubCallLevelCount: c.CurrentSubCallLevelCount,
		Logs:                     []WasmxLog{},
		CosmosEvents:             []types.Event{},
		Messages:                 []cw8types.SubMsg{},
		dbIterators:              c.dbIterators,
		RuntimeHandler:           c.RuntimeHandler,
		ContractInfo:             c.ContractInfo.Clone(),
		newIVmFn:                 c.newIVmFn,
	}
}

func (c *Context) Execute() ([]byte, error) {
	found := c.NativeHandler.IsPrecompile(c.Env.Contract.Address.Bytes())
	if found {
		data, err := c.NativeHandler.Execute(c, c.Env.Contract.Address, c.Env.CurrentCall.CallData)
		if err != nil {
			return nil, err
		}
		c.ReturnData = data
		return data, nil
	}
	filepath := c.ContractInfo.CodeFilePath
	if types.HasUtf8SystemDep(c.ContractInfo.SystemDeps) {
		filepath = ""
	}
	rnh := getRuntimeHandler(c.newIVmFn, c.Ctx, c.ContractInfo.SystemDeps, c.ContractInfo.Pinned)
	defer func() {
		rnh.GetVm().Cleanup()
	}()
	err := InitiateWasm(c, rnh, filepath, c.ContractInfo.AotFilePath, nil, c.ContractInfo.SystemDeps, c.ContractInfo.Role != "")
	if err != nil {
		return nil, err
	}
	isEvm := isEvmInterpreted(c.ContractInfo.SystemDepsRaw)
	setExecutionBytecode(c, rnh, types.ENTRY_POINT_EXECUTE, isEvm)

	c.ContractRouter[c.Env.Contract.Address.String()].RuntimeHandler = rnh

	executeHandler := GetExecuteFunctionHandler(c.ContractInfo.SystemDeps)
	_, err = executeHandler(c, rnh.GetVm(), types.ENTRY_POINT_EXECUTE, make([]interface{}, 0), false)
	if err != nil {
		rnh.GetVm().Cleanup()
		return nil, err
	}
	rnh.GetVm().Cleanup()
	return c.ReturnData, nil
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

type WasmxLog struct {
	ContractAddress  mcodec.AccAddressPrefixed
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

type StorageDeleteRange struct {
	StartKey []byte `json:"start_key"`
	EndKey   []byte `json:"end_key"`
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

type AppWithSubCallHook interface {
	BeginSubCall(ctx context.Context, level uint32, index uint32, isquery bool) error
	EndSubCall(ctx context.Context, level uint32, index uint32, isquery bool, err error) error
}

type HasBeginSubCall interface {
	BeginSubCall(ctx context.Context, level uint32, index uint32, isquery bool) error
}

type HasEndSubCall interface {
	EndSubCall(ctx context.Context, level uint32, index uint32, isquery bool, err error) error
}

func isEvmInterpreted(deps []string) bool {
	isEvm := false
	for _, d := range deps {
		if strings.Contains(d, types.INTERPRETER_EVM_ROOT) {
			isEvm = true
		}
	}
	return isEvm
}
