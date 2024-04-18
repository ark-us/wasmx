package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmxtypes "mythos/v1/x/wasmx/types"
)

// WasmxKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type WasmxKeeper interface {
	Query(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)
	Execute(ctx sdk.Context, contractAddr sdk.AccAddress, senderAddr sdk.AccAddress, msg wasmxtypes.RawContractMessage, funds sdk.Coins, dependencies []string, inBackground bool) ([]byte, error)
	ExecuteEntryPoint(ctx sdk.Context, entryPoint string, contractAddress sdk.AccAddress, caller sdk.AccAddress, msg []byte, dependencies []string, inBackground bool) ([]byte, error)
	ContractInstance(ctx sdk.Context, contractAddress sdk.AccAddress) (wasmxtypes.ContractInfo, wasmxtypes.CodeInfo, []byte, error)
	GetAddressOrRole(ctx sdk.Context, addressOrRole string) (sdk.AccAddress, error)
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmxtypes.ContractInfo

	ExecuteCosmosMsg(ctx sdk.Context, msg sdk.Msg, owner sdk.AccAddress) ([]sdk.Event, []byte, error)

	SmartContractCall(c context.Context, req *wasmxtypes.QuerySmartContractCallRequest) (rsp *wasmxtypes.QuerySmartContractCallResponse, err error)
}

type WasmxWrapper interface {
	ExecuteContract(ctx sdk.Context, req *MsgExecuteContract) (*MsgExecuteContractResponse, error)
	// QueryContract
	QueryContract(ctx sdk.Context, req *MsgQueryContract) (*MsgQueryContractResponse, error)
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmxtypes.ContractInfo
}
