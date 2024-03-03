package types

import (
	context "context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmxtypes "mythos/v1/x/wasmx/types"
)

// AccountKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type AccountKeeper interface {
	// Return a new account with the next account number and the specified address. Does not save the new account to the store.
	NewAccountWithAddress(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Retrieve an account from the store.
	// GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	// Set an account in the store.
	SetAccount(ctx context.Context, acc sdk.AccountI)
	AddressCodec() address.Codec
	HasAccount(ctx context.Context, acc sdk.AccAddress) bool
}

// WasmxKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type WasmxKeeper interface {
	Query(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)
	Execute(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, dependencies []string) ([]byte, error)
	ExecuteEntryPoint(ctx sdk.Context, entryPoint string, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, dependencies []string) ([]byte, error)
	ContractInstance(ctx sdk.Context, contractAddress sdk.AccAddress) (wasmxtypes.ContractInfo, wasmxtypes.CodeInfo, []byte, error)
	GetAddressOrRole(ctx sdk.Context, addressOrRole string) (sdk.AccAddress, error)
	IterateContractState(ctx sdk.Context, contractAddress sdk.AccAddress, cb func(key, value []byte) bool)
}
