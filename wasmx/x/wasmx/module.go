package wasmx

import (
	"context"
	"encoding/json"
	"fmt"

	// this line is used by starport scaffolding # 1

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	cdcaddress "cosmossdk.io/core/address"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/rand"
	"github.com/cosmos/cosmos-sdk/types/address"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	mcodec "wasmx/v1/codec"
	mcfg "wasmx/v1/config"
	"wasmx/v1/multichain"
	"wasmx/v1/x/wasmx/client/cli"
	"wasmx/v1/x/wasmx/keeper"
	"wasmx/v1/x/wasmx/types"
	memc "wasmx/v1/x/wasmx/vm/memory/common"
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
	wasmVmMeta memc.IWasmVmMeta
	cdc        codec.BinaryCodec
	ccdc       codec.Codec
	addrCodec  cdcaddress.Codec
	appCreator multichain.NewAppCreator
}

func NewAppModuleBasic(wasmVmMeta memc.IWasmVmMeta, cdc codec.BinaryCodec, ccdc codec.Codec, addrCodec cdcaddress.Codec, appCreator multichain.NewAppCreator) AppModuleBasic {
	return AppModuleBasic{wasmVmMeta: wasmVmMeta, cdc: cdc, ccdc: ccdc, addrCodec: addrCodec, appCreator: appCreator}
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
	feeCollector, _ := a.addrCodec.BytesToString(authtypes.NewModuleAddress(mcfg.FEE_COLLECTOR))
	mintAddress, _ := a.addrCodec.BytesToString(authtypes.NewModuleAddress("mint"))
	bootstrapAccount, _ := a.addrCodec.BytesToString(sdk.AccAddress(rand.Bytes(address.Len)))

	return cdc.MustMarshalJSON(types.DefaultGenesisState(a.addrCodec.(mcodec.AccBech32Codec), bootstrapAccount, feeCollector, mintAddress, 3, false, "{}"))
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
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetTxCmd returns the root Tx command for the module. The subcommands of this root command are used by end-users to generate new transactions containing messages defined in the module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd(a.wasmVmMeta, a.ccdc.InterfaceRegistry().SigningContext().AddressCodec(), a.appCreator)
}

// GetQueryCmd returns the root query command for the module. The subcommands of this root command are used by end-users to generate new queries to the subset of the state defined by the module
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd(a.wasmVmMeta, types.StoreKey, a.ccdc.InterfaceRegistry().SigningContext().AddressCodec())
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

func NewAppModule(
	wasmVmMeta memc.IWasmVmMeta,
	cdc codec.Codec,
	ccdc codec.Codec,
	keeper keeper.Keeper,
	appCreator multichain.NewAppCreator,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(wasmVmMeta, cdc, ccdc, keeper.AddressCodec(), appCreator),
		keeper:         keeper,
	}
}

// RegisterServices registers a gRPC query service to respond to the module-specific gRPC queries
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(&am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), &am.keeper)
}

// RegisterInvariants registers the invariants of the module. If an invariant deviates from its predicted value, the InvariantRegistry triggers appropriate logic (most often the chain will be halted)
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the module's genesis initialization. It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	var genState types.GenesisState
	// Initialize global index to index in genesis state
	cdc.MustUnmarshalJSON(gs, &genState)

	bootstrapAccountAddr, err := am.keeper.AccBech32Codec().StringToAccAddressPrefixed(genState.BootstrapAccountAddress)
	if err != nil {
		panic(fmt.Sprintf("bootstrap account: %+v", err))
	}

	if err := am.keeper.BootstrapSystemContracts(ctx, bootstrapAccountAddr, genState.SystemContracts, genState.CompiledFolderPath); err != nil {
		panic(fmt.Sprintf("bootstrap system contracts: %+v", err))
	}

	InitGenesis(ctx, am.keeper, genState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

// ConsensusVersion is a sequence number for state-breaking change of the module. It should be incremented on each consensus-breaking change introduced by the module. To avoid wrong/empty versions, the initial version should be set to 1
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock contains the logic that is automatically triggered at the beginning of each block
func (am AppModule) BeginBlock(_ context.Context) error {
	return nil
}

// EndBlock contains the logic that is automatically triggered at the end of each block
func (am AppModule) EndBlock(_ context.Context) error {
	return nil
}

func (m AppModule) IsAppModule() {}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}
