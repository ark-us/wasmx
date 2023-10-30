package keeper

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	baseapp "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	cw8 "mythos/v1/x/wasmx/cw8"
	cw8types "mythos/v1/x/wasmx/cw8/types"
	"mythos/v1/x/wasmx/types"
	cchtypes "mythos/v1/x/wasmx/types/contract_handler"
	"mythos/v1/x/wasmx/types/contract_handler/alias"
)

// contractMemoryLimit is the memory limit of each contract execution (in MiB)
// constant value so all nodes run with the same limit.
const contractMemoryLimit = 32

type (
	Keeper struct {
		cdc                   codec.Codec
		storeKey              storetypes.StoreKey
		memKey                storetypes.StoreKey
		tKey                  storetypes.StoreKey
		paramstore            paramtypes.Subspace
		interfaceRegistry     cdctypes.InterfaceRegistry
		msgRouter             *baseapp.MsgServiceRouter
		grpcQueryRouter       *baseapp.GRPCQueryRouter
		wasmVMResponseHandler cw8types.WasmVMResponseHandler
		wasmVMQueryHandler    cw8.WasmVMQueryHandler

		accountKeeper types.AccountKeeper
		bank          types.BankKeeper
		cch           *cchtypes.ContractHandlerMap
		// queryGasLimit is the max wasmvm gas that can be spent on executing a query with a contract
		queryGasLimit uint64
		gasRegister   GasRegister
		denom         string

		wasmvm  WasmxEngine
		tempDir string
		binDir  string

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	tKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	portSource cw8types.ICS20TransferPortSource,
	stakingKeeper types.StakingKeeper,
	distrKeeper types.DistributionKeeper,
	channelKeeper types.ChannelKeeper,
	wasmConfig types.WasmConfig,
	homeDir string,
	denom string,
	interfaceRegistry cdctypes.InterfaceRegistry,
	msgRouter *baseapp.MsgServiceRouter,
	grpcQueryRouter *baseapp.GRPCQueryRouter,
	authority string,
) *Keeper {
	contractsPath := filepath.Join(homeDir, types.ContractsDir)
	err := createDirsIfNotExist(contractsPath)
	if err != nil {
		panic(err)
	}
	err = createDirsIfNotExist(path.Join(contractsPath, types.PINNED_FOLDER))
	if err != nil {
		panic(err)
	}

	// for interpreted source codes (e.g. python)
	sourcesDir := path.Join(contractsPath, types.SourceCodeDir)
	err = createDirsIfNotExist(sourcesDir)
	if err != nil {
		panic(err)
	}
	sourcesPyDir := path.Join(sourcesDir, types.FILE_EXTENSIONS[types.ROLE_INTERPRETER_PYTHON])
	err = createDirsIfNotExist(sourcesPyDir)
	if err != nil {
		panic(err)
	}
	sourcesJsDir := path.Join(sourcesDir, types.FILE_EXTENSIONS[types.ROLE_INTERPRETER_JS])
	err = createDirsIfNotExist(sourcesJsDir)
	if err != nil {
		panic(err)
	}

	tempDir := path.Join(homeDir, types.TempDir)
	err = createDirsIfNotExist(tempDir)
	if err != nil {
		panic(err)
	}
	binDir := path.Join(homeDir, types.BinDir)
	err = createDirsIfNotExist(tempDir)
	if err != nil {
		panic(err)
	}

	wasmvm, err := NewVM(contractsPath, sourcesDir, contractMemoryLimit, wasmConfig.ContractDebugMode, wasmConfig.MemoryCacheSize)
	if err != nil {
		panic(err)
	}

	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	keeper := &Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		memKey:            memKey,
		tKey:              tKey,
		paramstore:        ps,
		interfaceRegistry: interfaceRegistry,
		msgRouter:         msgRouter,
		grpcQueryRouter:   grpcQueryRouter,
		denom:             denom,

		accountKeeper: accountKeeper,
		bank:          bankKeeper,
		queryGasLimit: wasmConfig.SmartQueryGasLimit,
		gasRegister:   NewDefaultWasmGasRegister(),
		wasmvm:        *wasmvm,
		tempDir:       tempDir,
		binDir:        binDir,
		authority:     authority,
	}

	// cosmwasm support
	handler := cw8.NewMessageDispatcher(keeper, cdc, portSource)
	keeper.wasmVMResponseHandler = handler
	qhandler := cw8.DefaultQueryPlugins(bankKeeper, stakingKeeper, distrKeeper, channelKeeper, keeper)
	keeper.wasmVMQueryHandler = qhandler

	// Register core contracts after the cw8 handlers are attached to the keeper
	cch := cchtypes.NewContractHandlerMap(*keeper)
	cch.Register(types.ROLE_ALIAS, alias.NewAliasHandler())
	keeper.cch = &cch

	return keeper
}

// TODO remove
// used in system contracts
func (k Keeper) CloneWithStoreKey(storeKey storetypes.StoreKey, memKey storetypes.StoreKey) Keeper {
	return Keeper{
		cdc:               k.cdc,
		storeKey:          storeKey,
		memKey:            memKey,
		tKey:              k.tKey,
		paramstore:        k.paramstore,
		interfaceRegistry: k.interfaceRegistry,
		msgRouter:         k.msgRouter,
		grpcQueryRouter:   k.grpcQueryRouter,
		denom:             k.denom,

		accountKeeper:         k.accountKeeper,
		bank:                  k.bank,
		queryGasLimit:         k.queryGasLimit,
		gasRegister:           k.gasRegister,
		wasmvm:                k.wasmvm,
		tempDir:               k.tempDir,
		binDir:                k.binDir,
		authority:             k.authority,
		wasmVMResponseHandler: k.wasmVMResponseHandler,
		wasmVMQueryHandler:    k.wasmVMQueryHandler,
		cch:                   k.cch,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) ContractHandler() *cchtypes.ContractHandlerMap {
	return k.cch
}

func (k Keeper) WasmVMResponseHandler() cw8types.WasmVMResponseHandler {
	return k.wasmVMResponseHandler
}

// 0755 = User:rwx Group:r-x World:r-x
// 0750 = User:rwx Group:r-x World:---
// 0770
const nodeDirPerm = 0o755

func createDirsIfNotExist(dirpath string) error {
	return os.MkdirAll(dirpath, nodeDirPerm)
}

func createFileIfNotExist(filepath string) error {
	return os.WriteFile(filepath, []byte{}, 0644)
}
