package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmxtypes "mythos/v1/x/wasmx/types"
)

// WasmxKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type WasmxKeeper interface {
	Query(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)
	Execute(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, dependencies []string) ([]byte, error)
	ExecuteEntryPoint(ctx sdk.Context, entryPoint string, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, dependencies []string) ([]byte, error)
	ContractInstance(ctx sdk.Context, contractAddress sdk.AccAddress) (wasmxtypes.ContractInfo, wasmxtypes.CodeInfo, []byte, error)
	GetAddressOrRole(ctx sdk.Context, addressOrRole string) (sdk.AccAddress, error)
}

type WasmxWrapper interface {
	ExecuteContract(ctx sdk.Context, req *MsgExecuteContract) (*MsgExecuteContractResponse, error)
	// QueryContract
	QueryContract(ctx sdk.Context, req *MsgQueryContract) (*MsgQueryContractResponse, error)
}
