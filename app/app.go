package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	_ "net/http/pprof"

	"os"
	"path/filepath"

	"github.com/spf13/cast"
	"golang.org/x/sync/errgroup"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"

	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/circuit"

	circuitkeeper "cosmossdk.io/x/circuit/keeper"

	circuittypes "cosmossdk.io/x/circuit/types"
	"cosmossdk.io/x/evidence"

	evidencekeeper "cosmossdk.io/x/evidence/keeper"

	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"

	feegrantkeeper "cosmossdk.io/x/feegrant/keeper"

	"cosmossdk.io/x/upgrade"

	dbm "github.com/cosmos/cosmos-db"
	// "github.com/cosmos/gogoproto/proto"

	// upgradeclient "cosmossdk.io/x/upgrade/client"
	upgradekeeper "cosmossdk.io/x/upgrade/keeper"

	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"

	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"

	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/std"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	// "github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/version"

	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"

	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/x/authz"

	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"

	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"

	"github.com/cosmos/cosmos-sdk/x/crisis"

	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"

	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"

	// distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"

	testdata_pulsar "github.com/cosmos/cosmos-sdk/testutil/testdata/testpb"

	consensus "github.com/cosmos/cosmos-sdk/x/consensus"

	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"

	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"

	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"

	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"

	"github.com/cosmos/cosmos-sdk/x/group"

	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"

	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"

	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"

	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/ibc-go/modules/capability"

	capabilitykeeper "github.com/cosmos/ibc-go/modules/capability/keeper"

	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"

	ica "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts"

	icacontrollerkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/keeper"

	icacontrollertypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/controller/types"

	icahost "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host"

	icahosttypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/types"

	icatypes "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/types"
	"github.com/cosmos/ibc-go/v8/modules/apps/transfer"

	ibctransferkeeper "github.com/cosmos/ibc-go/v8/modules/apps/transfer/keeper"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	ibc "github.com/cosmos/ibc-go/v8/modules/core"

	// ibcclientclient "github.com/cosmos/ibc-go/v8/modules/core/02-client/client"
	icahostkeeper "github.com/cosmos/ibc-go/v8/modules/apps/27-interchain-accounts/host/keeper"

	ibcporttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"

	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	ibckeeper "github.com/cosmos/ibc-go/v8/modules/core/keeper"

	ibctesting "github.com/cosmos/ibc-go/v8/testing"

	ibctestingtypes "github.com/cosmos/ibc-go/v8/testing/types"

	abci "github.com/cometbft/cometbft/abci/types"

	ante "mythos/v1/app/ante"

	appparams "mythos/v1/app/params"

	docs "mythos/v1/docs"

	networkmodule "mythos/v1/x/network"

	networkmodulekeeper "mythos/v1/x/network/keeper"

	networkmoduletypes "mythos/v1/x/network/types"

	wasmxmodule "mythos/v1/x/wasmx"

	wasmxmodulekeeper "mythos/v1/x/wasmx/keeper"

	wasmxmoduletypes "mythos/v1/x/wasmx/types"

	networktypes "mythos/v1/x/network/types"

	websrvmodule "mythos/v1/x/websrv"

	websrvmodulekeeper "mythos/v1/x/websrv/keeper"

	websrvmoduletypes "mythos/v1/x/websrv/types"

	cosmosmodkeeper "mythos/v1/x/cosmosmod/keeper"

	cosmosmodtypes "mythos/v1/x/cosmosmod/types"

	cosmosmod "mythos/v1/x/cosmosmod"
)

// this line is used by starport scaffolding # stargate/app/moduleImport

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// module account permissions
	// TODO remove/replace this ?
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:       nil,
		distrtypes.ModuleName:            nil,
		icatypes.ModuleName:              nil,
		minttypes.ModuleName:             {authtypes.Minter},
		stakingtypes.BondedPoolName:      {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:   {authtypes.Burner, authtypes.Staking},
		wasmxmoduletypes.ROLE_GOVERNANCE: {authtypes.Burner},
		ibctransfertypes.ModuleName:      {authtypes.Minter, authtypes.Burner},
		wasmxmoduletypes.ModuleName:      {authtypes.Minter, authtypes.Burner},
		// this line is used by starport scaffolding # stargate/app/maccPerms
	}
)

var (
	_ servertypes.Application = (*App)(nil)
	_ ibctesting.TestingApp   = (*App)(nil)
	_ runtime.AppI            = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+Name)
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	goRoutineGroup  *errgroup.Group
	goContextParent context.Context

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	txConfig          client.TxConfig
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys      map[string]*storetypes.KVStoreKey
	tkeys     map[string]*storetypes.TransientStoreKey
	memKeys   map[string]*storetypes.MemoryStoreKey
	clessKeys map[string]*storetypes.ConsensuslessStoreKey

	// keepers
	AccountKeeper    *cosmosmodkeeper.KeeperAuth
	CapabilityKeeper *capabilitykeeper.Keeper
	BankKeeper       *cosmosmodkeeper.KeeperBank
	StakingKeeper    *cosmosmodkeeper.KeeperStaking
	GovKeeper        *cosmosmodkeeper.KeeperGov
	SlashingKeeper   *cosmosmodkeeper.KeeperSlashing
	DistrKeeper      *cosmosmodkeeper.KeeperDistribution

	AuthzKeeper authzkeeper.Keeper
	MintKeeper  mintkeeper.Keeper

	CrisisKeeper          *crisiskeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	IBCKeeper             *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	EvidenceKeeper        evidencekeeper.Keeper
	TransferKeeper        ibctransferkeeper.Keeper
	ICAHostKeeper         icahostkeeper.Keeper
	ICAControllerKeeper   icacontrollerkeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper
	CircuitKeeper         circuitkeeper.Keeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper

	WasmxKeeper     wasmxmodulekeeper.Keeper
	CosmosmodKeeper *cosmosmodkeeper.Keeper
	WebsrvKeeper    websrvmodulekeeper.Keeper

	NetworkKeeper  networkmodulekeeper.Keeper
	actionExecutor *networkmodulekeeper.ActionExecutor
	// this line is used by starport scaffolding # stargate/app/keeperDeclaration

	// mm is the module manager
	mm                 *module.Manager
	BasicModuleManager module.BasicManager

	// sm is the simulation manager
	sm           *module.SimulationManager
	configurator module.Configurator
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	goRoutineGroup := appOpts.Get("goroutineGroup").(*errgroup.Group)
	goContextParent := appOpts.Get("goContextParent").(context.Context)

	// TODO - do we need this?
	// std.RegisterLegacyAminoCodec(cdc)
	std.RegisterInterfaces(interfaceRegistry)

	bApp := baseapp.NewBaseApp(
		Name,
		logger,
		db,
		encodingConfig.TxConfig.TxDecoder(),
		baseAppOptions...,
	)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	bApp.SetTxEncoder(encodingConfig.TxConfig.TxEncoder())

	keys := storetypes.NewKVStoreKeys(
		authz.ModuleName, crisistypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		paramstypes.StoreKey, consensusparamtypes.StoreKey, ibcexported.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey, evidencetypes.StoreKey, circuittypes.StoreKey,
		ibctransfertypes.StoreKey, icahosttypes.StoreKey, capabilitytypes.StoreKey, group.StoreKey,
		icacontrollertypes.StoreKey,
		wasmxmoduletypes.StoreKey,
		websrvmoduletypes.StoreKey,
		networkmoduletypes.StoreKey,
		cosmosmodtypes.StoreKey,
		// this line is used by starport scaffolding # stargate/app/storeKey
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey, wasmxmoduletypes.TStoreKey)
	memKeys := storetypes.NewMemoryStoreKeys(capabilitytypes.MemStoreKey, wasmxmoduletypes.MemStoreKey)
	clessKeys := storetypes.NewConsensuslessStoreKeys(networkmoduletypes.CLessStoreKey, wasmxmoduletypes.MetaConsensusStoreKey, wasmxmoduletypes.SingleConsensusStoreKey)

	// register streaming services
	if err := bApp.RegisterStreamingServices(appOpts, keys); err != nil {
		panic(err)
	}

	app := &App{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		txConfig:          encodingConfig.TxConfig,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
		goRoutineGroup:    goRoutineGroup,
		goContextParent:   goContextParent,
		clessKeys:         clessKeys,
	}

	// TODO replace NewPermissionsForAddress with address by role
	permAddrs := make(map[string]authtypes.PermissionsForAddress)
	for name, perms := range maccPerms {
		permAddrs[name] = authtypes.NewPermissionsForAddress(name, perms)
		app.Logger().Info("module address", name, permAddrs[name].GetAddress().String())
	}

	app.ParamsKeeper = initParamsKeeper(
		appCodec,
		cdc,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[consensusparamtypes.StoreKey]),
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
		runtime.EventService{},
	)
	bApp.SetParamStore(app.ConsensusParamsKeeper.ParamsStore)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	scopedICAControllerKeeper := app.CapabilityKeeper.ScopeToModule(icacontrollertypes.SubModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	// this line is used by starport scaffolding # stargate/app/scopedKeeper

	// add keepers

	wasmconfig := wasmxmoduletypes.DefaultWasmConfig()
	app.WasmxKeeper = *wasmxmodulekeeper.NewKeeper(
		app.goRoutineGroup,
		app.goContextParent,
		appCodec,
		encodingConfig.TxConfig,
		keys[wasmxmoduletypes.StoreKey],
		memKeys[wasmxmoduletypes.MemStoreKey],
		tkeys[wasmxmoduletypes.TStoreKey],
		clessKeys[wasmxmoduletypes.MetaConsensusStoreKey],
		clessKeys[wasmxmoduletypes.SingleConsensusStoreKey],
		app.GetSubspace(wasmxmoduletypes.ModuleName),
		// TODO?
		// app.TransferKeeper,
		// distrkeeper.NewQuerier(app.DistrKeeper),
		// app.IBCKeeper.ChannelKeeper,
		wasmconfig,
		homePath,
		BaseDenom,
		permAddrs,
		app.interfaceRegistry,
		app.MsgServiceRouter(),
		app.GRPCQueryRouter(),
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
		app,
	)
	wasmxModule := wasmxmodule.NewAppModule(appCodec, app.WasmxKeeper)

	app.actionExecutor = networkmodulekeeper.NewActionExecutor(app, logger)
	app.NetworkKeeper = *networkmodulekeeper.NewKeeper(
		app.goRoutineGroup,
		app.goContextParent,
		appCodec,
		keys[networkmoduletypes.StoreKey],
		memKeys[networkmoduletypes.MemStoreKey],
		tkeys[networkmoduletypes.TStoreKey],
		clessKeys[networkmoduletypes.CLessStoreKey],
		app.GetSubspace(networkmoduletypes.ModuleName),
		&app.WasmxKeeper,
		app.actionExecutor,
		// TODO remove authority?
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
	)
	networkModule := networkmodule.NewAppModule(appCodec, app.NetworkKeeper, app)

	app.AccountKeeper = cosmosmodkeeper.NewKeeperAuth(
		appCodec,
		appCodec,
		keys[cosmosmodtypes.StoreKey], // TODO remove
		app.GetSubspace(cosmosmodtypes.ModuleName),
		&app.WasmxKeeper,
		app.NetworkKeeper,
		app.actionExecutor,
		// TODO what authority?
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
		app.interfaceRegistry,
		authcodec.NewBech32Codec(Bech32PrefixValAddr),
		authcodec.NewBech32Codec(Bech32PrefixConsAddr),
		authcodec.NewBech32Codec(Bech32PrefixAccAddr),
		permAddrs,

		// authtypes.ProtoBaseAccount,
		// runtime.NewKVStoreService(keys[authtypes.StoreKey]),
		// Bech32PrefixAccAddr,
	)
	app.AuthzKeeper = authzkeeper.NewKeeper(
		runtime.NewKVStoreService(keys[authzkeeper.StoreKey]),
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	app.StakingKeeper = cosmosmodkeeper.NewKeeperStaking(
		appCodec,
		appCodec,
		keys[cosmosmodtypes.StoreKey], // TODO remove
		app.GetSubspace(cosmosmodtypes.ModuleName),
		app.AccountKeeper,
		&app.WasmxKeeper,
		app.NetworkKeeper,
		app.actionExecutor,
		// TODO what authority?
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
		app.interfaceRegistry,
		authcodec.NewBech32Codec(Bech32PrefixValAddr),
		authcodec.NewBech32Codec(Bech32PrefixConsAddr),
	)
	app.BankKeeper = cosmosmodkeeper.NewKeeperBank(
		appCodec,
		appCodec,
		keys[cosmosmodtypes.StoreKey], // TODO remove
		app.GetSubspace(cosmosmodtypes.ModuleName),
		app.AccountKeeper,
		&app.WasmxKeeper,
		app.NetworkKeeper,
		app.actionExecutor,
		// TODO what authority?
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
		app.interfaceRegistry,
		authcodec.NewBech32Codec(Bech32PrefixValAddr),
		authcodec.NewBech32Codec(Bech32PrefixConsAddr),
	)
	app.GovKeeper = cosmosmodkeeper.NewKeeperGov(
		appCodec,
		appCodec,
		keys[cosmosmodtypes.StoreKey], // TODO remove
		app.GetSubspace(cosmosmodtypes.ModuleName),
		app.AccountKeeper,
		&app.WasmxKeeper,
		app.NetworkKeeper,
		app.actionExecutor,
		// TODO what authority?
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
		app.interfaceRegistry,
		authcodec.NewBech32Codec(Bech32PrefixValAddr),
		authcodec.NewBech32Codec(Bech32PrefixConsAddr),
	)
	app.DistrKeeper = cosmosmodkeeper.NewKeeperDistribution(
		appCodec,
		appCodec,
		keys[cosmosmodtypes.StoreKey], // TODO remove
		app.GetSubspace(cosmosmodtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		&app.WasmxKeeper,
		app.NetworkKeeper,
		app.actionExecutor,
		// TODO what authority?
		// TODO we have addressByRole now
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
		authtypes.FeeCollectorName,
		app.interfaceRegistry,
		authcodec.NewBech32Codec(Bech32PrefixValAddr),
		authcodec.NewBech32Codec(Bech32PrefixConsAddr),
	)
	app.SlashingKeeper = cosmosmodkeeper.NewKeeperSlashing(
		appCodec,
		appCodec,
		keys[cosmosmodtypes.StoreKey], // TODO remove
		app.GetSubspace(cosmosmodtypes.ModuleName),
		app.StakingKeeper,
		&app.WasmxKeeper,
		app.NetworkKeeper,
		app.actionExecutor,
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
	)
	cosmosmodModule := cosmosmod.NewAppModule(appCodec, appCodec, *app.BankKeeper, *app.StakingKeeper, *app.GovKeeper, *app.AccountKeeper, *app.SlashingKeeper, *app.DistrKeeper, app)

	// TODO remove
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[minttypes.StoreKey]),
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
	)

	app.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[crisistypes.StoreKey]),
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
		app.AccountKeeper.AddressCodec(),
	)

	// TODO remove
	app.CircuitKeeper = circuitkeeper.NewKeeper(appCodec, runtime.NewKVStoreService(keys[circuittypes.StoreKey]), authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(), app.AccountKeeper.AddressCodec())
	app.BaseApp.SetCircuitBreaker(&app.CircuitKeeper)

	groupConfig := group.DefaultConfig()
	/*
		Example of setting group params:
		groupConfig.MaxMetadataLen = 1000
	*/
	app.GroupKeeper = groupkeeper.NewKeeper(
		keys[group.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		groupConfig,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[feegrant.StoreKey]),
		app.AccountKeeper,
	)
	// set the governance module account as the authority for conducting upgrades
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		runtime.NewKVStoreService(keys[upgradetypes.StoreKey]),
		appCodec,
		homePath,
		app.BaseApp,
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
	)

	// ... other modules keepers

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibcexported.StoreKey],
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		scopedIBCKeeper,
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
	)

	// Create Transfer Keepers
	app.TransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
	)
	transferModule := transfer.NewAppModule(app.TransferKeeper)
	transferIBCModule := transfer.NewIBCModule(app.TransferKeeper)

	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey],
		app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		scopedICAHostKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
	)
	app.ICAControllerKeeper = icacontrollerkeeper.NewKeeper(
		appCodec, keys[icacontrollertypes.StoreKey],
		app.GetSubspace(icacontrollertypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // may be replaced with middleware such as ics29 fee
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.PortKeeper,
		scopedICAControllerKeeper,
		app.MsgServiceRouter(),
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
	)
	icaModule := ica.NewAppModule(&app.ICAControllerKeeper, &app.ICAHostKeeper)
	icaHostIBCModule := icahost.NewIBCModule(app.ICAHostKeeper)

	// Create evidence Keeper for to register the IBC light client misbehaviour evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[evidencetypes.StoreKey]),
		app.StakingKeeper,
		app.SlashingKeeper,
		app.AccountKeeper.AddressCodec(),
		runtime.ProvideCometInfoService(),
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	app.WebsrvKeeper = *websrvmodulekeeper.NewKeeper(
		appCodec,
		keys[websrvmoduletypes.StoreKey],
		memKeys[websrvmoduletypes.MemStoreKey],
		app.GetSubspace(websrvmoduletypes.ModuleName),
		&app.WasmxKeeper,
		app.Query,
		authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String(),
	)
	websrvModule := websrvmodule.NewAppModule(appCodec, app.WebsrvKeeper, app.AccountKeeper, app.BankKeeper)

	// this line is used by starport scaffolding # stargate/app/keeperDefinition

	/**** IBC Routing ****/

	// Sealing prevents other modules from creating scoped sub-keepers
	app.CapabilityKeeper.Seal()

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()
	ibcRouter.AddRoute(icahosttypes.SubModuleName, icaHostIBCModule).
		AddRoute(ibctransfertypes.ModuleName, transferIBCModule)
	// this line is used by starport scaffolding # ibc/app/router
	app.IBCKeeper.SetRouter(ibcRouter)

	/**** Module Hooks ****/

	// register hooks after all modules have been initialized

	// TODO hooks?
	// app.StakingKeeper.SetHooks(
	// 	stakingtypes.NewMultiStakingHooks(
	// 		// insert staking hooks receivers here
	// 		app.DistrKeeper.Hooks(),
	// 		app.SlashingKeeper.Hooks(),
	// 	),
	// )

	// TODO governance hooks? or we just read block events
	// app.GovKeeper.SetHooks(
	// 	govtypes.NewMultiGovHooks(
	// 	// insert governance hooks receivers here
	// 	),
	// )

	/**** Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app,
			encodingConfig.TxConfig,
		),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, minttypes.DefaultInflationCalculationFn, app.GetSubspace(minttypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		transferModule,
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		circuit.NewAppModule(appCodec, app.CircuitKeeper),

		// non sdk modules
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, true),
		ibc.NewAppModule(app.IBCKeeper),
		icaModule,

		// mythos modules
		wasmxModule,
		networkModule,
		cosmosmodModule,
		websrvModule,

		// sdk
		// crisis - always be last to make sure that it checks for all invariants and not only part of them
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
	)

	// BasicModuleManager defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration and genesis verification.
	// By default it is composed of all the module from the module manager.
	// Additionally, app module basics can be overwritten by passing them as argument.
	app.BasicModuleManager = module.NewBasicManagerFromManager(
		app.mm,
		map[string]module.AppModuleBasic{
			genutiltypes.ModuleName: genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		})
	// TODO - do we need this?
	// app.BasicModuleManager.RegisterLegacyAminoCodec(cdc)
	app.BasicModuleManager.RegisterInterfaces(interfaceRegistry)

	// upgrades should be run first
	app.mm.SetOrderPreBlockers(
		upgradetypes.ModuleName,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		crisistypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		wasmxmoduletypes.ModuleName,
		networktypes.ModuleName,
		cosmosmodtypes.ModuleName,
		websrvmoduletypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/beginBlockers
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		capabilitytypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		wasmxmoduletypes.ModuleName,
		networktypes.ModuleName,
		cosmosmodtypes.ModuleName,
		websrvmoduletypes.ModuleName,
		// this line is used by starport scaffolding # stargate/app/endBlockers
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	genesisModuleOrder := []string{
		// sdk
		capabilitytypes.ModuleName,

		// mythos
		wasmxmoduletypes.ModuleName,
		networkmoduletypes.ModuleName,
		cosmosmodtypes.ModuleName,

		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		consensusparamtypes.ModuleName,
		circuittypes.ModuleName,
		// ibc
		ibctransfertypes.ModuleName,
		ibcexported.ModuleName,
		icatypes.ModuleName,
		// mythos extra
		websrvmoduletypes.ModuleName,
	}

	app.mm.SetOrderInitGenesis(genesisModuleOrder...)
	app.mm.SetOrderExportGenesis(genesisModuleOrder...)

	// Uncomment if you want to set a custom migration order here.
	// app.mm.SetOrderMigrations(custom order)

	app.mm.RegisterInvariants(app.CrisisKeeper)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	err := app.mm.RegisterServices(app.configurator)
	if err != nil {
		panic(err)
	}

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	// Make sure it's called after `app.ModuleManager` and `app.configurator` are set.
	app.RegisterUpgradeHandlers()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	testdata_pulsar.RegisterQueryServer(app.GRPCQueryRouter(), testdata_pulsar.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	// NOTE: this is not required apps that don't use the simulator for fuzz testing
	// transactions
	overrideModules := map[string]module.AppModuleSimulation{}
	app.sm = module.NewSimulationManagerFromAppModules(app.mm.Modules, overrideModules)
	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)
	app.MountConsensuslessStores(clessKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetPreBlocker(app.PreBlocker)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler(encodingConfig.TxConfig)

	// In v0.46, the SDK introduces _postHandlers_. PostHandlers are like
	// antehandlers, but are run _after_ the `runMsgs` execution. They are also
	// defined as a chain, and have the same signature as antehandlers.
	//
	// In baseapp, postHandlers are run in the same store branch as `runMsgs`,
	// meaning that both `runMsgs` and `postHandler` state will be committed if
	// both are successful, and both will be reverted if any of the two fails.
	//
	// The SDK exposes a default postHandlers chain
	//
	// Please note that changing any of the anteHandler or postHandler chain is
	// likely to be a state-machine breaking change, which needs a coordinated
	// upgrade.
	app.setPostHandler()

	// At startup, after all modules have been registered, check that all prot
	// annotations are correct.
	// TODO reenable this! fix the libp2p proto issue
	// protoFiles, err := proto.MergedRegistry()
	// if err != nil {
	// 	panic(err)
	// }
	// err = msgservice.ValidateProtoAnnotations(protoFiles)
	// if err != nil {
	// 	// Once we switch to using protoreflect-based antehandlers, we might
	// 	// want to panic here instead of logging a warning.
	// 	fmt.Fprintln(os.Stderr, err.Error())
	// }

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			Exit(err.Error())
		}

		// TODO
		// // Initialize pinned codes in wasmvm as they are not persisted there
		// if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
		// 	panic(fmt.Sprintf("failed initialize pinned codes %s", err))
		// }
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	// this line is used by starport scaffolding # stargate/app/beforeInitReturn

	return app
}

func (app *App) setAnteHandler(txConfig client.TxConfig) {
	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: txConfig.SignModeHandler(),
			FeegrantKeeper:  app.FeeGrantKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			WasmxKeeper:     &app.WasmxKeeper,
			CircuitKeeper:   &app.CircuitKeeper,
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %w", err))
	}

	// Set the AnteHandler for the app
	app.SetAnteHandler(anteHandler)
}

func (app *App) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// PreBlocker application updates every pre block
func (app *App) PreBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (*sdk.ResponsePreBlock, error) {
	return app.mm.PreBlock(ctx)
}

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context, req *abci.RequestFinalizeBlock) (sdk.BeginBlock, error) {
	if app.LastBlockHeight() > 1 {
		reqbz, err := json.Marshal(req)
		if err != nil {
			return sdk.BeginBlock{}, fmt.Errorf("BeginBlocker cannot marshal RequestFinalizeBlock: %s", err.Error())
		}
		reqbas64 := base64.StdEncoding.EncodeToString(reqbz)
		msgbz := []byte(fmt.Sprintf(`{"RunHook":{"hook":"BeginBlock","data":"%s"}}`, reqbas64))
		_, err = app.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
			Sender:   wasmxmoduletypes.ROLE_CONSENSUS, // TODO role baseapp ?
			Contract: wasmxmoduletypes.ROLE_HOOKS,
			Msg:      msgbz,
		})
		if err != nil {
			return sdk.BeginBlock{}, fmt.Errorf("BeginBlock wasmx call failed: %s", err.Error())
		}
	}
	return app.mm.BeginBlock(ctx)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, metadata []byte) (sdk.EndBlock, error) {
	// only first block
	if app.LastBlockHeight() > 1 {
		metabase64 := base64.StdEncoding.EncodeToString(metadata)
		msgbz := []byte(fmt.Sprintf(`{"RunHook":{"hook":"EndBlock","data":"%s"}}`, metabase64))
		_, err := app.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
			Sender:   wasmxmoduletypes.ROLE_CONSENSUS, // TODO role baseapp ?
			Contract: wasmxmoduletypes.ROLE_HOOKS,
			Msg:      msgbz,
		})
		if err != nil {
			return sdk.EndBlock{}, fmt.Errorf("EndBlock wasmx call failed: %s", err.Error())
		}
	}
	return app.mm.EndBlock(ctx)
}

func (a *App) Configurator() module.Configurator {
	return a.configurator
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req *abci.RequestInitChain) (*abci.ResponseInitChain, error) {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedModuleAccountAddrs returns all the app's blocked module account
// addresses.
func (app *App) BlockedModuleAccountAddrs() map[string]bool {
	modAccAddrs := app.ModuleAccountAddrs()

	// allow the following addresses to receive funds
	delete(modAccAddrs, authtypes.NewModuleAddress(wasmxmoduletypes.ROLE_GOVERNANCE).String())

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// Used by network module
func (app *App) GetMKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// Used by network module
func (app *App) GetCLessKey(storeKey string) *storetypes.ConsensuslessStoreKey {
	return app.clessKeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	cmtservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	app.BasicModuleManager.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register app's OpenAPI routes.
	docs.RegisterOpenAPIService(Name, apiSvr.Router)
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	cmtservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// RegisterNodeService implements the Application.RegisterNodeService method.
func (app *App) RegisterNodeService(clientCtx client.Context, cfg config.Config) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter(), cfg)
}

// IBC Go TestingApp functions

// GetBaseApp implements the TestingApp interface.
func (app *App) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetStakingKeeper implements the TestingApp interface.
func (app *App) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetIBCKeeper implements the TestingApp interface.
func (app *App) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper implements the TestingApp interface.
func (app *App) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// TxConfig returns App's TxConfig
func (app *App) TxConfig() client.TxConfig {
	return app.txConfig
}

// GetTxConfig implements the TestingApp interface.
func (app *App) GetTxConfig() client.TxConfig {
	return app.TxConfig()
}

// AutoCliOpts returns the autocli options for the app.
func (app *App) AutoCliOpts() autocli.AppOptions {
	modules := make(map[string]appmodule.AppModule, 0)
	for _, m := range app.mm.Modules {
		if moduleWithName, ok := m.(module.HasName); ok {
			moduleName := moduleWithName.Name()
			if appModule, ok := moduleWithName.(appmodule.AppModule); ok {
				modules[moduleName] = appModule
			}
		}
	}

	return autocli.AppOptions{
		Modules:               modules,
		ModuleOptions:         runtimeservices.ExtractAutoCLIOptions(app.mm.Modules),
		AddressCodec:          authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
		ValidatorAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		ConsensusAddressCodec: authcodec.NewBech32Codec(sdk.GetConfig().GetBech32ConsensusAddrPrefix()),
	}
}

// DefaultGenesis returns a default genesis from the registered AppModuleBasic's.
func (a *App) DefaultGenesis() map[string]json.RawMessage {
	return a.BasicModuleManager.DefaultGenesis(a.appCodec)
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	// paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(wasmxmoduletypes.ModuleName)
	paramsKeeper.Subspace(cosmosmodtypes.ModuleName)
	paramsKeeper.Subspace(networkmoduletypes.ModuleName)
	paramsKeeper.Subspace(websrvmoduletypes.ModuleName)
	// this line is used by starport scaffolding # stargate/app/paramSubspace

	return paramsKeeper
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// For network grpc
func (app *App) GetNetworkKeeper() *networkmodulekeeper.Keeper {
	return &app.NetworkKeeper
}

func (app *App) GetActionExecutor() *networkmodulekeeper.ActionExecutor {
	return app.actionExecutor
}

func (app *App) GetGoContextParent() context.Context {
	return app.goContextParent
}

func (app *App) GetGoRoutineGroup() *errgroup.Group {
	return app.goRoutineGroup
}

func Exit(s string) {
	fmt.Printf(s + "\n")
	os.Exit(1)
}
