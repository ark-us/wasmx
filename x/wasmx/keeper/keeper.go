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

	"wasmx/x/wasmx/types"
)

// contractMemoryLimit is the memory limit of each contract execution (in MiB)
// constant value so all nodes run with the same limit.
const contractMemoryLimit = 32

type (
	Keeper struct {
		cdc               codec.Codec
		storeKey          storetypes.StoreKey
		memKey            storetypes.StoreKey
		paramstore        paramtypes.Subspace
		interfaceRegistry cdctypes.InterfaceRegistry
		msgRouter         *baseapp.MsgServiceRouter
		grpcQueryRouter   *baseapp.GRPCQueryRouter

		accountKeeper types.AccountKeeper
		bank          types.BankKeeper
		// queryGasLimit is the max wasmvm gas that can be spent on executing a query with a contract
		queryGasLimit uint64
		gasRegister   GasRegister
		denom         string

		wasmvm WasmxEngine
	}
)

func NewKeeper(
	cdc codec.Codec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
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

	wasmvm, err := NewVM(contractsPath, contractMemoryLimit, wasmConfig.ContractDebugMode, wasmConfig.MemoryCacheSize)
	if err != nil {
		panic(err)
	}

	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
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
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func createDirsIfNotExist(dirpath string) error {
	return os.MkdirAll(dirpath, 0770)
}
