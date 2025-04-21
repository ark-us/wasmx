package vmsql

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
	goContextParent context.Context
}

func NewAppModuleBasic(goContextParent context.Context) AppModuleBasic {
	return AppModuleBasic{goContextParent: goContextParent}
}

// Name returns the name of the module as a string
func (AppModuleBasic) Name() string {
	return ModuleName
}

// RegisterLegacyAminoCodec registers the amino codec for the module, which is used to marshal and unmarshal structs to/from []byte in order to persist them in the module's KVStore
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterInterfaces registers a module's interface types and their concrete implementations as proto.Message
func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {}

// DefaultGenesis returns a default GenesisState for the module, marshalled to json.RawMessage. The default GenesisState need to be defined by the module developer and is primarily used for testing
func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return nil
}

// ValidateGenesis used to validate the GenesisState, given in its json.RawMessage form
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
}

// GetTxCmd returns the root Tx command for the module. The subcommands of this root command are used by end-users to generate new transactions containing messages defined in the module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns the root query command for the module. The subcommands of this root command are used by end-users to generate new queries to the subset of the state defined by the module
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

// AppModule implements the AppModule interface that defines the inter-dependent methods that modules need to implement
type AppModule struct {
	AppModuleBasic
}

func NewAppModule(goContextParent context.Context) AppModule {
	return AppModule{AppModuleBasic: AppModuleBasic{goContextParent: goContextParent}}
}

// RegisterServices registers a gRPC query service to respond to the module-specific gRPC queries
func (am AppModule) RegisterServices(cfg module.Configurator) {}

// RegisterInvariants registers the invariants of the module. If an invariant deviates from its predicted value, the InvariantRegistry triggers appropriate logic (most often the chain will be halted)
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// InitGenesis performs the module's genesis initialization. It returns no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the module's exported genesis state as raw JSON bytes.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return nil
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

func (am AppModule) BeginTransaction(ctx context.Context, txmode sdk.ExecMode, txBytes []byte) error {
	return nil
}

func (am AppModule) EndTransaction(ctx context.Context, txmode sdk.ExecMode, gInfo sdk.GasInfo, result *sdk.Result, anteEvents []abci.Event, txerr error) error {
	vctx, err := GetSqlContext(am.goContextParent)
	if err != nil {
		return err
	}
	shouldCommit := false
	if txmode == sdk.ExecModeFinalize {
		shouldCommit = true
	}
	for _, conn := range vctx.DbConnections {
		if conn.OpenSavepointTx == nil {
			continue
		}
		if shouldCommit {
			cmd := "RELEASE sp0"
			if txerr != nil {
				cmd = "ROLLBACK TO sp0"
			}
			_, err := conn.OpenSavepointTx.Exec(cmd)
			if err != nil {
				return fmt.Errorf("db tx command failed: %s, %s", cmd, err.Error())
			}
			err = conn.OpenSavepointTx.Commit()
			if err != nil {
				return fmt.Errorf("cannot commit sql open tx: %s", err.Error())
			}
		} else {
			err = conn.OpenSavepointTx.Rollback()
			if err != nil {
				return fmt.Errorf("cannot rollback sql open tx: %s", err.Error())
			}
		}
		conn.OpenSavepointTx = nil
		conn.SavePointMap = make(map[string]bool, 0)
	}
	return nil
}

func (am AppModule) BeginSubCall(ctx context.Context, level uint32, index uint32, isquery bool) error {
	vctx, err := GetSqlContext(am.goContextParent)
	if err != nil {
		return err
	}
	for _, conn := range vctx.DbConnections {
		if conn.OpenSavepointTx == nil {
			continue
		}
		savepoint := buildSavepoint(level, index)
		cmd := fmt.Sprintf("SAVEPOINT %s", savepoint)
		conn.SavePointMap[savepoint] = true
		_, err := conn.OpenSavepointTx.Exec(cmd)
		if err != nil {
			return fmt.Errorf("cannot add savepoint: %s, %s", savepoint, err.Error())
		}
	}
	return nil
}

func (am AppModule) EndSubCall(_ context.Context, level uint32, index uint32, isquery bool, txerr error) error {
	// check err & rollback to previous savepoint
	vctx, err := GetSqlContext(am.goContextParent)
	if err != nil {
		return err
	}
	for _, conn := range vctx.DbConnections {
		if conn.OpenSavepointTx == nil {
			continue
		}
		savepoint := buildSavepoint(level, index)
		if !conn.hasSavePoint(savepoint) {
			continue
		}
		cmd := fmt.Sprintf("RELEASE sp%d_%d", level, index)
		if isquery || txerr != nil {
			cmd = fmt.Sprintf("ROLLBACK TO sp%d_%d", level, index)
		}
		_, err := conn.OpenSavepointTx.Exec(cmd)
		if err != nil {
			return fmt.Errorf("db tx command failed: %s, %s", cmd, err.Error())
		}
	}
	return nil
}

func buildSavepoint(level, index uint32) string {
	return fmt.Sprintf("sp%d_%d", level, index)
}
