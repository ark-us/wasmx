package keeper

import (
	"wasmx/x/wasmx/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) Create(ctx sdk.Context, senderAddr sdk.AccAddress, wasmByteCode []byte) (uint64, []byte, error) {
	return 1, nil, nil
}

func (k Keeper) Instantiate(ctx sdk.Context, codeId uint64, senderAddr sdk.AccAddress, msg types.RawContractMessage, label string, funds sdk.Coins) (sdk.AccAddress, []byte, error) {
	return nil, nil, nil
}

func (k Keeper) Instantiate2(ctx sdk.Context, codeId uint64, senderAddr sdk.AccAddress, msg types.RawContractMessage, label string, funds sdk.Coins, salt []byte, fixMsg bool) (sdk.AccAddress, []byte, error) {
	return nil, nil, nil
}

func (k Keeper) Execute(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return nil, nil
}

func (k Keeper) ExecuteWithOrigin(ctx sdk.Context, originAddr sdk.AccAddress, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return nil, nil
}

func (k Keeper) ExecuteDelegate(ctx sdk.Context, originAddr sdk.AccAddress, codeContractAddr sdk.AccAddress, storageContractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg types.RawContractMessage, funds sdk.Coins) ([]byte, error) {
	return nil, nil
}
