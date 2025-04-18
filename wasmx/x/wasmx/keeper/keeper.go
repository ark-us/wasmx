package keeper

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	address "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	baseapp "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	cw8 "github.com/loredanacirstea/wasmx/x/wasmx/cw8"
	cw8types "github.com/loredanacirstea/wasmx/x/wasmx/cw8/types"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	cchtypes "github.com/loredanacirstea/wasmx/x/wasmx/types/contract_handler"
	"github.com/loredanacirstea/wasmx/x/wasmx/types/contract_handler/alias"
	memc "github.com/loredanacirstea/wasmx/x/wasmx/vm/memory/common"
)

// contractMemoryLimit is the memory limit of each contract execution (in MiB)
// constant value so all nodes run with the same limit.
const contractMemoryLimit = 32

type (
	Keeper struct {
		cdc                   codec.Codec
		txConfig              client.TxConfig
		storeKey              storetypes.StoreKey
		memKey                storetypes.StoreKey
		tKey                  storetypes.StoreKey
		metaConsKey           storetypes.StoreKey
		singleConsKey         storetypes.StoreKey
		paramstore            paramtypes.Subspace
		interfaceRegistry     cdctypes.InterfaceRegistry
		msgRouter             *baseapp.MsgServiceRouter
		grpcQueryRouter       *baseapp.GRPCQueryRouter
		wasmVMResponseHandler cw8types.WasmVMResponseHandler
		wasmVMQueryHandler    cw8.WasmVMQueryHandler
		validatorAddressCodec address.Codec
		consensusAddressCodec address.Codec
		addressCodec          address.Codec
		accBech32Codec        mcodec.AccBech32Codec
		ak                    types.AccountKeeper
		bank                  types.BankKeeperWasmx

		WasmRuntime memc.IWasmVmMeta

		cch *cchtypes.ContractHandlerMap
		// queryGasLimit is the max wasmvm gas that can be spent on executing a query with a contract
		queryGasLimit uint64
		gasRegister   types.GasRegister
		denom         string
		permAddrs     map[string]authtypes.PermissionsForAddress

		wasmvm  *WasmxEngine
		tempDir string
		binDir  string

		// the address capable of executing messages through governance. Typically, this
		// should be the x/gov module account.
		authority string
		app       types.Application
	}
)

func NewKeeper(
	goRoutineGroup *errgroup.Group,
	goContextParent context.Context,
	cdc codec.Codec,
	txConfig client.TxConfig,
	storeKey storetypes.StoreKey,
	memKey storetypes.StoreKey,
	tKey storetypes.StoreKey,
	metaConsKey storetypes.StoreKey,
	singleConsKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	// TODO
	// portSource cw8types.ICS20TransferPortSource,
	// stakingKeeper types.StakingKeeper,
	// distrKeeper types.DistributionKeeper,
	// channelKeeper types.ChannelKeeper,
	wasmConfig types.WasmConfig,
	homeDir string,
	denom string,
	permAddrs map[string]authtypes.PermissionsForAddress,
	interfaceRegistry cdctypes.InterfaceRegistry,
	msgRouter *baseapp.MsgServiceRouter,
	grpcQueryRouter *baseapp.GRPCQueryRouter,
	authority string,
	validatorAddressCodec address.Codec,
	consensusAddressCodec address.Codec,
	addressCodec address.Codec,
	app types.Application,
	wasmRuntime memc.IWasmVmMeta,
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

	wasmvm, err := NewVM(goRoutineGroup, goContextParent, contractsPath, sourcesDir, contractMemoryLimit, wasmConfig.ContractDebugMode, wasmConfig.MemoryCacheSize, app, GetLogger, wasmRuntime)
	if err != nil {
		panic(err)
	}

	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	accBech32Codec := mcodec.MustUnwrapAccBech32Codec(addressCodec)

	keeper := &Keeper{
		cdc:                   cdc,
		txConfig:              txConfig,
		storeKey:              storeKey,
		memKey:                memKey,
		tKey:                  tKey,
		metaConsKey:           metaConsKey,
		singleConsKey:         singleConsKey,
		paramstore:            ps,
		interfaceRegistry:     interfaceRegistry,
		msgRouter:             msgRouter,
		grpcQueryRouter:       grpcQueryRouter,
		denom:                 denom,
		permAddrs:             permAddrs,
		validatorAddressCodec: validatorAddressCodec,
		consensusAddressCodec: consensusAddressCodec,
		addressCodec:          addressCodec,
		accBech32Codec:        accBech32Codec,

		queryGasLimit: wasmConfig.SmartQueryGasLimit,
		gasRegister:   NewDefaultWasmGasRegister(),
		wasmvm:        wasmvm,
		tempDir:       tempDir,
		binDir:        binDir,
		authority:     authority,
		app:           app,
		WasmRuntime:   wasmRuntime,
	}

	// cosmwasm support
	// handler := cw8.NewMessageDispatcher(keeper, cdc, portSource)
	// keeper.wasmVMResponseHandler = handler
	// qhandler := cw8.DefaultQueryPlugins(bankKeeper, stakingKeeper, distrKeeper, channelKeeper, keeper)
	// keeper.wasmVMQueryHandler = qhandler
	return keeper
}

func (k *Keeper) SetAccountKeeper(ak types.AccountKeeper) {
	k.ak = ak
}

func (k *Keeper) SetBankKeeper(bank types.BankKeeperWasmx) {
	k.bank = bank
}

func (k *Keeper) SetContractHandlerMap() {
	// Register core contracts after the cw8 handlers are attached to the keeper
	cch := cchtypes.NewContractHandlerMap(k)
	cch.Register(types.ROLE_ALIAS, alias.NewAliasHandler())
	k.cch = &cch
}

func (k *Keeper) GetAccountKeeper() types.AccountKeeper {
	if k.ak == nil {
		panic("no account keeper found on wasmx keeper")
	}
	return k.ak
}

func (k *Keeper) GetBankKeeper() types.BankKeeperWasmx {
	if k.bank == nil {
		panic("no bank keeper found on wasmx keeper")
	}
	return k.bank
}

func (k *Keeper) Logger(ctx sdk.Context) log.Logger {
	return GetLogger(ctx)
}

// GetAuthority returns the module's authority.
func (k *Keeper) GetAuthority() string {
	return k.authority
}

func (k *Keeper) ContractHandler() *cchtypes.ContractHandlerMap {
	return k.cch
}

func (k *Keeper) WasmVMResponseHandler() cw8types.WasmVMResponseHandler {
	return k.wasmVMResponseHandler
}

func (k *Keeper) AddressCodec() address.Codec {
	return k.addressCodec
}

func (k *Keeper) ValidatorAddressCodec() address.Codec {
	return k.validatorAddressCodec
}

func (k *Keeper) ConsensusAddressCodec() address.Codec {
	return k.consensusAddressCodec
}

func (k *Keeper) AccBech32Codec() mcodec.AccBech32Codec {
	return k.accBech32Codec
}

func GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With(log.ModuleKey, fmt.Sprintf("x/%s", types.ModuleName), "chain_id", ctx.ChainID())
}

// 0755 = User:rwx Group:r-x World:r-x
// 0750 = User:rwx Group:r-x World:---
// 0770
const nodeDirPerm = 0o755

func createDirsIfNotExist(dirpath string) error {
	return os.MkdirAll(dirpath, nodeDirPerm)
}
