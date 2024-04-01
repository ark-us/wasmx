package cosmosmod

import (
	"context"
	"encoding/json"
	"fmt"

	// this line is used by starport scaffolding # 1

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	networktypes "mythos/v1/x/network/types"

	"mythos/v1/x/cosmosmod/client/cli"
	"mythos/v1/x/cosmosmod/keeper"
	"mythos/v1/x/cosmosmod/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

// AppModuleBasic implements the AppModuleBasic interface that defines the independent methods a Cosmos SDK module needs to implement.
type AppModuleBasic struct {
	cdc  codec.BinaryCodec
	ccdc codec.Codec
}

func NewAppModuleBasic(cdc codec.BinaryCodec, ccdc codec.Codec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc, ccdc: ccdc}
}

// Name returns the name of the module as a string
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the amino codec for the module, which is used to marshal and unmarshal structs to/from []byte in order to persist them in the module's KVStore
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers a module's interface types and their concrete implementations as proto.Message
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

// DefaultGenesis returns a default GenesisState for the module, marshalled to json.RawMessage. The default GenesisState need to be defined by the module developer and is primarily used for testing
func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState("myt", 18, "mythos"))
}

// ValidateGenesis used to validate the GenesisState, given in its json.RawMessage form
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	banktypes.RegisterQueryHandlerClient(context.Background(), mux, banktypes.NewQueryClient(clientCtx))
	stakingtypes.RegisterQueryHandlerClient(context.Background(), mux, stakingtypes.NewQueryClient(clientCtx))
	govtypes1.RegisterQueryHandlerClient(context.Background(), mux, govtypes1.NewQueryClient(clientCtx))
	slashingtypes.RegisterQueryHandlerClient(context.Background(), mux, slashingtypes.NewQueryClient(clientCtx))
	distributiontypes.RegisterQueryHandlerClient(context.Background(), mux, distributiontypes.NewQueryClient(clientCtx))
}

// GetTxCmd returns the root Tx command for the module. The subcommands of this root command are used by end-users to generate new transactions containing messages defined in the module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd(a.ccdc.InterfaceRegistry().SigningContext().ValidatorAddressCodec(), a.ccdc.InterfaceRegistry().SigningContext().AddressCodec())
}

// GetQueryCmd returns the root query command for the module. The subcommands of this root command are used by end-users to generate new queries to the subset of the state defined by the module
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(a.ccdc.InterfaceRegistry().SigningContext().AddressCodec())
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement
type AppModule struct {
	AppModuleBasic
	bank         keeper.KeeperBank
	staking      keeper.KeeperStaking
	gov          keeper.KeeperGov
	auth         keeper.KeeperAuth
	slashing     keeper.KeeperSlashing
	distribution keeper.KeeperDistribution
	app          networktypes.BaseApp
}

func NewAppModule(
	cdc codec.Codec,
	ccdc codec.Codec,
	bank keeper.KeeperBank,
	staking keeper.KeeperStaking,
	gov keeper.KeeperGov,
	auth keeper.KeeperAuth,
	slashing keeper.KeeperSlashing,
	distribution keeper.KeeperDistribution,
	app networktypes.BaseApp,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc, ccdc),
		bank:           bank,
		staking:        staking,
		gov:            gov,
		auth:           auth,
		slashing:       slashing,
		distribution:   distribution,
		app:            app,
	}
}

// RegisterServices registers a gRPC query service to respond to the module-specific gRPC queries
func (am AppModule) RegisterServices(cfg module.Configurator) {
	banktypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgBankServerImpl(&am.bank))
	stakingtypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgStakingServerImpl(&am.staking))
	govtypes1.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgGovServerImpl(&am.gov))
	authtypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgAuthServerImpl(&am.auth))
	slashingtypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgSlashingServerImpl(&am.slashing))
	distributiontypes.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgDistributionServerImpl(&am.distribution))

	banktypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerierBank(&am.bank))
	stakingtypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerierStaking(&am.staking))
	govtypes1.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerierGov(&am.gov))
	authtypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerierAuth(&am.auth))
	slashingtypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerierSlashing(&am.slashing))
	distributiontypes.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerierDistribution(&am.distribution))
}

// RegisterInvariants registers the invariants of the module. If an invariant deviates from its predicted value, the InvariantRegistry triggers appropriate logic (most often the chain will be halted)
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the module's genesis initialization. It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	return InitGenesis(ctx, am.bank, am.gov, am.staking, am.auth, am.slashing, am.distribution, genState)
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.bank, am.gov, am.staking, am.auth, am.slashing, am.distribution)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion is a sequence number for state-breaking change of the module. It should be incremented on each consensus-breaking change introduced by the module. To avoid wrong/empty versions, the initial version should be set to 1
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock contains the logic that is automatically triggered at the beginning of each block
func (am AppModule) BeginBlock(_ sdk.Context) {}

// EndBlock contains the logic that is automatically triggered at the end of each block
func (am AppModule) EndBlock(_ sdk.Context) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (m AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}
