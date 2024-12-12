package cosmosmod

import (
	"fmt"

	errors "cosmossdk.io/errors"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/loredanacirstea/wasmx/x/cosmosmod/keeper"
	"github.com/loredanacirstea/wasmx/x/cosmosmod/types"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, bank keeper.KeeperBank, gov keeper.KeeperGov, staking keeper.KeeperStaking, auth keeper.KeeperAuth, slashing keeper.KeeperSlashing, distribution keeper.KeeperDistribution, genState types.GenesisState) []abci.ValidatorUpdate {
	InitGenesisBank(ctx, bank, genState.Bank)
	InitGenesisGov(ctx, gov, genState.Gov)
	InitGenesisAuth(ctx, auth, genState.Auth)

	// we instantiate slashing before staking because we currently run
	// some actions from ApplyAndReturnValidatorSetUpdates
	// in staking genesis init, which calls slashing hooks
	InitGenesisSlashing(ctx, slashing, genState.Slashing)
	InitGenesisDistribution(ctx, distribution, genState.Distribution)

	updates := InitGenesisStaking(ctx, staking, genState.Staking)

	return updates
}

func InitGenesisBank(ctx sdk.Context, k keeper.KeeperBank, genState types.BankGenesisState) {
	k.Logger(ctx).Info("initializing bank genesis")
	msgjson, err := k.JSONCodec().MarshalJSON(&genState)
	if err != nil {
		panic(err)
	}
	msgbz := []byte(fmt.Sprintf(`{"InitGenesis":%s}`, string(msgjson)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      msgbz,
	})
	if err != nil {
		panic(err)
	}
	k.Logger(ctx).Info("initialized bank genesis")
}

func InitGenesisGov(ctx sdk.Context, k keeper.KeeperGov, genState types.GovGenesisState) {
	k.Logger(ctx).Info("initializing gov genesis")
	msgjson, err := k.JSONCodec().MarshalJSON(&genState)
	if err != nil {
		panic(err)
	}
	msgbz := []byte(fmt.Sprintf(`{"InitGenesis":%s}`, string(msgjson)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_GOVERNANCE,
		Contract: wasmxtypes.ROLE_GOVERNANCE,
		Msg:      msgbz,
	})
	if err != nil {
		panic(err)
	}
	k.Logger(ctx).Info("initialized gov genesis")
}

func InitGenesisStaking(ctx sdk.Context, k keeper.KeeperStaking, genState types.StakingGenesisState) []abci.ValidatorUpdate {
	k.Logger(ctx).Info("initializing staking genesis")
	msgjson, err := k.JSONCodec().MarshalJSON(&genState)
	if err != nil {
		panic(err)
	}

	msgbz := []byte(fmt.Sprintf(`{"InitGenesis":%s}`, string(msgjson)))
	res, err := k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_STAKING,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		panic(err)
	}
	var response types.InitGenesisResponse
	err = k.JSONCodec().UnmarshalJSON(res.Data, &response)
	if err != nil {
		panic(err)
	}
	k.Logger(ctx).Info("initialized staking genesis")
	updates := make([]abci.ValidatorUpdate, len(response.Updates))
	for i, upd := range response.Updates {
		var pkI cryptotypes.PubKey
		err = k.InterfaceRegistry.UnpackAny(upd.PubKey, &pkI)
		if err != nil {
			panic(err)
		}
		pk, ok := upd.PubKey.GetCachedValue().(cryptotypes.PubKey)
		if !ok {
			panic(errors.Wrapf(sdkerrors.ErrInvalidType, "expecting cryptotypes.PubKey, got %T", pk))
		}
		tmPk, err := cryptocodec.ToCmtProtoPublicKey(pk)
		if err != nil {
			panic(err)
		}
		updates[i] = abci.ValidatorUpdate{
			PubKey: tmPk,
			Power:  upd.Power,
		}
	}
	return updates
}

// TODO
// permAddrs    map[string]types.PermissionsForAddress
// bech32Prefix string
func InitGenesisAuth(ctx sdk.Context, k keeper.KeeperAuth, genState types.AuthGenesisState) {
	k.Logger(ctx).Info("initializing auth genesis")
	genState.BaseAccountTypeurl = sdk.MsgTypeURL(&types.BaseAccount{})
	genState.ModuleAccountTypeurl = sdk.MsgTypeURL(&types.ModuleAccount{})
	msgjson, err := k.JSONCodec().MarshalJSON(&genState)
	if err != nil {
		panic(err)
	}
	msgbz := []byte(fmt.Sprintf(`{"InitGenesis":%s}`, string(msgjson)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_AUTH,
		Contract: wasmxtypes.ROLE_AUTH,
		Msg:      msgbz,
	})
	if err != nil {
		panic(err)
	}
	// this should just create the feecollector account
	k.GetModuleAccount(ctx, authtypes.FeeCollectorName)

	k.Logger(ctx).Info("initialized auth genesis")
}

func InitGenesisSlashing(ctx sdk.Context, k keeper.KeeperSlashing, genState slashingtypes.GenesisState) {
	k.Logger(ctx).Info("initializing slashing genesis")
	msgjson, err := k.JSONCodec().MarshalJSON(&genState)
	if err != nil {
		panic(err)
	}
	msgbz := []byte(fmt.Sprintf(`{"InitGenesis":%s}`, string(msgjson)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_SLASHING,
		Contract: wasmxtypes.ROLE_SLASHING,
		Msg:      msgbz,
	})
	if err != nil {
		panic(err)
	}
	k.Logger(ctx).Info("initialized slashing genesis")
}

func InitGenesisDistribution(ctx sdk.Context, k keeper.KeeperDistribution, genState types.DistributionGenesisState) {
	k.Logger(ctx).Info("initializing distribution genesis")
	msgjson, err := k.JSONCodec().MarshalJSON(&genState)
	if err != nil {
		panic(err)
	}
	msgbz := []byte(fmt.Sprintf(`{"InitGenesis":%s}`, string(msgjson)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_DISTRIBUTION,
		Contract: wasmxtypes.ROLE_DISTRIBUTION,
		Msg:      msgbz,
	})
	if err != nil {
		panic(err)
	}
	k.Logger(ctx).Info("initialized distribution genesis")
}

// TODO
// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, bank keeper.KeeperBank, gov keeper.KeeperGov, staking keeper.KeeperStaking, auth keeper.KeeperAuth, slashing keeper.KeeperSlashing, distribution keeper.KeeperDistribution) *stakingtypes.GenesisState {
	genState := &stakingtypes.GenesisState{}
	// TODO
	// k.IterateSystemContracts(ctx, func(contract types.SystemContract) bool {
	// 	genState.SystemContracts = append(genState.SystemContracts, contract)
	// 	return false
	// })
	return genState
}
