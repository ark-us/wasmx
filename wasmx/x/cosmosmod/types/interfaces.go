package types

import (
	context "context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

// AccountKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type AccountKeeper interface {
	// Return a new account with the next account number and the specified address. Does not save the new account to the store.
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Set an account in the store.
	SetAccount(ctx context.Context, acc sdk.AccountI)
	AddressCodec() address.Codec
	HasAccount(ctx context.Context, acc sdk.AccAddress) bool
}

// WasmxKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type WasmxKeeper interface {
	Query(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg wasmxtypes.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)
	Execute(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg wasmxtypes.RawContractMessage, funds sdk.Coins, dependencies []string, inBackground bool) ([]byte, error)
	ExecuteEntryPoint(ctx sdk.Context, entryPoint string, contractAddress mcodec.AccAddressPrefixed, caller mcodec.AccAddressPrefixed, msg []byte, dependencies []string, inBackground bool) ([]byte, error)
	ContractInstance(ctx sdk.Context, contractAddress mcodec.AccAddressPrefixed) (wasmxtypes.ContractInfo, wasmxtypes.CodeInfo, []byte, error)
	GetAddressOrRole(ctx sdk.Context, addressOrRole string) (mcodec.AccAddressPrefixed, error)
	IterateContractState(ctx sdk.Context, contractAddress sdk.AccAddress, cb func(key, value []byte) bool)
}
