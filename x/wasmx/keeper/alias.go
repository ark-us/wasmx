package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
	cchtypes "mythos/v1/x/wasmx/types/contract_handler"
	"mythos/v1/x/wasmx/types/contract_handler/alias"
)

func (k Keeper) GetAlias(ctx sdk.Context, addr sdk.AccAddress) (sdk.AccAddress, bool) {
	addrEth := types.EvmAddressFromAcc(addr)
	querymsg := alias.GetCosmosAddressRequest{EthAddress: addrEth, CoinType: cchtypes.COIN_TYPE_ETH}
	resp, err := k.ContractHandler().Query(ctx, cchtypes.ContractHandlerMessage{
		Role:   types.ROLE_ALIAS,
		Method: "getCosmosAddress",
		Sender: sdk.AccAddress([]byte("wasmx")),
		Msg:    querymsg,
	})
	if err != nil {
		return nil, false
	}
	decodedResp := resp.([]interface{})[0].(alias.GetCosmosAddressResponse)
	aliasAddr := types.AccAddressFromEvm(decodedResp.CosmAddress)
	return aliasAddr, decodedResp.Found
}
