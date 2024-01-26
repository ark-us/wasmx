package cosmosmod

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"mythos/v1/x/cosmosmod/keeper"
	"mythos/v1/x/cosmosmod/types"
	networktypes "mythos/v1/x/network/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// // this line is used by starport scaffolding # genesis/module/init
	// k.SetParams(ctx, genState.Params)
	// for _, contract := range genState.SystemContracts {
	// 	k.SetSystemContract(ctx, contract)
	// }

	var stakingGenesis stakingtypes.GenesisState = genState.Staking
	msgjson, err := json.Marshal(stakingGenesis)
	if err != nil {
		panic(err)
	}
	msgbz := []byte(fmt.Sprintf(`{"InitGenesis":%s}`, string(msgjson)))
	fmt.Println("----InitGenesis---", string(msgbz))
	_, err = k.NetworkKeeper.ExecuteContract(ctx, &networktypes.MsgExecuteContract{
		Sender:   wasmxtypes.ROLE_STAKING,
		Contract: wasmxtypes.ROLE_STAKING,
		Msg:      msgbz,
	})
	if err != nil {
		panic(err)
	}
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
