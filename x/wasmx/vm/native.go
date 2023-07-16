package vm

import (
	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mythos/v1/x/wasmx/types"
)

type NativeMapI map[string]func(context *Context, input []byte) ([]byte, error)

var NativeMap = NativeMapI{
	"0x0000000000000000000000000000000000000001": Secp256k1RecoverNative,
	"0x0000000000000000000000000000000000000022": SecretSharing,
	"0x0000000000000000000000000000000000000025": ProxyInterfaces,
}

func (m NativeMapI) IsPrecompile(contractAddress sdk.AccAddress) bool {
	hexaddr := types.EvmAddressFromAcc(contractAddress).Hex()
	_, found := m[hexaddr]
	return found
}

func (m NativeMapI) Execute(context *Context, contractAddress sdk.AccAddress, input []byte) ([]byte, error) {
	hexaddr := types.EvmAddressFromAcc(contractAddress).Hex()
	precompile, found := m[hexaddr]
	if !found {
		return nil, sdkerr.Wrapf(sdkerr.Error{}, "native precompile not found")
	}
	return precompile(context, input)
}
