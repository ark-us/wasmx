package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "mythos/v1/codec"
	"mythos/v1/x/wasmx/types"
	cchtypes "mythos/v1/x/wasmx/types/contract_handler"
	"mythos/v1/x/wasmx/types/contract_handler/alias"
)

func (k *Keeper) GetAlias(ctx sdk.Context, addr mcodec.AccAddressPrefixed) (mcodec.AccAddressPrefixed, bool) {
	addrEth := types.EvmAddressFromAcc(addr.Bytes())
	querymsg := alias.GetCosmosAddressRequest{EthAddress: addrEth, CoinType: cchtypes.COIN_TYPE_ETH}
	resp, err := k.ContractHandler().Query(ctx, cchtypes.ContractHandlerMessage{
		Role:   types.ROLE_ALIAS,
		Method: "getCosmosAddress",
		Sender: sdk.AccAddress([]byte("wasmx")),
		Msg:    querymsg,
	})
	if err != nil {
		return mcodec.AccAddressPrefixed{}, false
	}
	decodedResp := resp.([]interface{})[0].(alias.GetCosmosAddressResponse)
	aliasAddr := types.AccAddressFromEvm(decodedResp.CosmAddress)
	return k.accBech32Codec.BytesToAccAddressPrefixed(aliasAddr), decodedResp.Found
}
