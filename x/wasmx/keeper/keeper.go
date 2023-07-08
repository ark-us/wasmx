package keeper

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	baseapp "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

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
		paramstore            paramtypes.Subspace
		interfaceRegistry     cdctypes.InterfaceRegistry
		msgRouter             *baseapp.MsgServiceRouter
		grpcQueryRouter       *baseapp.GRPCQueryRouter
		wasmVMResponseHandler cw8types.WasmVMResponseHandler

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
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	portSource cw8types.ICS20TransferPortSource,
	wasmConfig types.WasmConfig,
	homeDir string,
	denom string,
	interfaceRegistry cdctypes.InterfaceRegistry,
	msgRouter *baseapp.MsgServiceRouter,
	grpcQueryRouter *baseapp.GRPCQueryRouter,
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

	wasmvm, err := NewVM(contractsPath, contractMemoryLimit, wasmConfig.ContractDebugMode, wasmConfig.MemoryCacheSize)
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
	}

	// Register core contracts
	cch := cchtypes.NewContractHandlerMap(*keeper)
	cch.Register(types.ROLE_ALIAS, alias.NewAliasHandler())
	keeper.cch = &cch

	handler := cw8.NewMessageDispatcher(keeper, cdc, portSource)
	keeper.wasmVMResponseHandler = handler

	return keeper
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) ContractHandler() *cchtypes.ContractHandlerMap {
	return k.cch
}

// 0755 = User:rwx Group:r-x World:r-x
// 0750 = User:rwx Group:r-x World:---
// 0770
const nodeDirPerm = 0o755

func createDirsIfNotExist(dirpath string) error {
	return os.MkdirAll(dirpath, nodeDirPerm)
}
