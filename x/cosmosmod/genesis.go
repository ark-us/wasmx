package cosmosmod

import (
	"fmt"

	errors "cosmossdk.io/errors"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	abci "github.com/cometbft/cometbft/abci/types"

	"mythos/v1/x/cosmosmod/keeper"
	"mythos/v1/x/cosmosmod/types"
	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) []abci.ValidatorUpdate {
	// initialize bank
	bankGenesis := genState.Bank
	bankmsgjson, err := k.JSONCodec().MarshalJSON(&bankGenesis)
	if err != nil {
		panic(err)
	}
	bankmsgbz := []byte(fmt.Sprintf(`{"InitGenesis":%s}`, string(bankmsgjson)))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_BANK,
		Contract: wasmxtypes.ROLE_BANK,
		Msg:      bankmsgbz,
	})
	if err != nil {
		panic(err)
	}

	// initialize staking
	stakingGenesis := genState.Staking
	msgjson, err := k.JSONCodec().MarshalJSON(&stakingGenesis)
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

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *stakingtypes.GenesisState {
	genState := &stakingtypes.GenesisState{}
	// k.IterateSystemContracts(ctx, func(contract types.SystemContract) bool {
	// 	genState.SystemContracts = append(genState.SystemContracts, contract)
	// 	return false
	// })
	return genState
}
