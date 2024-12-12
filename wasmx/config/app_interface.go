package config

import (
	"context"
	"encoding/json"
	_ "net/http/pprof"

	"golang.org/x/sync/errgroup"

	address "cosmossdk.io/core/address"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	// capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtcfg "github.com/cometbft/cometbft/config"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	mctx "github.com/loredanacirstea/wasmx/v1/context"
	srvconfig "github.com/loredanacirstea/wasmx/v1/server/config"
	networktypes "github.com/loredanacirstea/wasmx/v1/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/v1/x/wasmx/types"

	mcodec "github.com/loredanacirstea/wasmx/v1/codec"
	menc "github.com/loredanacirstea/wasmx/v1/encoding"
)

type ActionExecutor interface {
	Execute(goCtx context.Context, height int64, cb func(goctx context.Context) (any, error)) (any, error)
	ExecuteWithHeader(goCtx context.Context, header cmtproto.Header, cb func(goctx context.Context) (any, error)) (any, error)
	ExecuteWithMockHeader(goCtx context.Context, cb func(goctx context.Context) (any, error)) (any, error)
	GetApp() MythosApp
	GetBaseApp() BaseApp
	GetLogger() log.Logger
}

type NetworkKeeper interface {
	// WasmxWrapper
	ExecuteContract(ctx sdk.Context, msg *networktypes.MsgExecuteContract) (*networktypes.MsgExecuteContractResponse, error)
	ExecuteCosmosMsg(ctx sdk.Context, msg sdk.Msg, owner mcodec.AccAddressPrefixed) ([]sdk.Event, []byte, error)
	QueryContract(ctx sdk.Context, req *networktypes.MsgQueryContract) (*networktypes.MsgQueryContractResponse, error)

	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmxtypes.ContractInfo
	Codec() codec.Codec

	GetHeaderByHeight(app MythosApp, logger log.Logger, height int64, prove bool) (*cmtproto.Header, error)
}

type WasmxKeeper interface {
	GetCodeInfo(ctx sdk.Context, codeID uint64) *wasmxtypes.CodeInfo
	IterateContractInfo(ctx sdk.Context, cb func(sdk.AccAddress, wasmxtypes.ContractInfo) bool)
	IterateCodeInfos(ctx sdk.Context, cb func(uint64, wasmxtypes.CodeInfo) bool)
	ExecuteContractInstantiationInternal(
		ctx sdk.Context,
		codeID uint64,
		codeInfo *wasmxtypes.CodeInfo,
		creator mcodec.AccAddressPrefixed,
		contractAddress mcodec.AccAddressPrefixed,
		storageType wasmxtypes.ContractStorageType,
		initMsg []byte,
		deposit sdk.Coins,
		label string,
	) (wasmxtypes.ContractResponse, uint64, error)
}

type MythosApp interface {
	AddressCodec() address.Codec
	AppCodec() codec.Codec
	JSONCodec() codec.JSONCodec
	BeginBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (sdk.BeginBlock, error)
	BlockedModuleAccountAddrs() map[string]bool
	ConsensusAddressCodec() address.Codec
	DefaultGenesis() map[string]json.RawMessage
	EndBlocker(ctx sdk.Context, metadata []byte) (sdk.EndBlock, error)
	GetActionExecutor() ActionExecutor
	GetBaseApp() *baseapp.BaseApp
	GetCLessKey(storeKey string) *storetypes.ConsensuslessStoreKey
	GetCMetaKey(storeKey string) *storetypes.ConsensusMetaStoreKey
	GetChainCfg() *menc.ChainConfig
	GetGoContextParent() context.Context
	GetGoRoutineGroup() *errgroup.Group
	GetIBCKeeper() *ibckeeper.Keeper
	GetKey(storeKey string) *storetypes.KVStoreKey
	GetMKey(storeKey string) *storetypes.MemoryStoreKey
	GetMemKey(storeKey string) *storetypes.MemoryStoreKey
	GetMultiChainApp() (*MultiChainApp, error)
	GetNetworkKeeper() NetworkKeeper
	GetWasmxKeeper() WasmxKeeper
	// GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper
	GetStakingKeeper() ibctestingtypes.StakingKeeper
	GetSubspace(moduleName string) paramstypes.Subspace
	GetTKey(storeKey string) *storetypes.TransientStoreKey
	GetTxConfig() client.TxConfig
	InterfaceRegistry() types.InterfaceRegistry
	LegacyAmino() *codec.LegacyAmino
	LoadHeight(height int64) error
	MinGasPrices() sdk.DecCoins
	ModuleAccountAddrs() map[string]bool
	Name() string
	RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig)
	RegisterNodeService(clientCtx client.Context, cfg config.Config)
	RegisterTendermintService(clientCtx client.Context)
	RegisterTxService(clientCtx client.Context)
	RegisterUpgradeHandlers()
	SimulationManager() *module.SimulationManager
	TxConfig() client.TxConfig
	ValidatorAddressCodec() address.Codec

	GetServerConfig() *srvconfig.Config
	GetTendermintConfig() *cmtcfg.Config
	GetRpcClient() client.CometRPC

	// baseapp
	Query(context.Context, *abci.RequestQuery) (*abci.ResponseQuery, error)
	GRPCQueryRouter() *baseapp.GRPCQueryRouter
	MsgServiceRouter() *baseapp.MsgServiceRouter

	// statesync new app
	NonDeterministicGetNodePorts() mctx.NodePorts
	NonDeterministicGetNodePortsInitial() mctx.NodePorts
	NonDeterministicSetNodePortsInitial(mctx.NodePorts)

	// debugging
	Db() dbm.DB
	DebugDb()
}
