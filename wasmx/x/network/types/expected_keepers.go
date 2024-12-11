package types

import (
	context "context"

	address "cosmossdk.io/core/address"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	mcodec "wasmx/v1/codec"
	wasmxtypes "wasmx/v1/x/wasmx/types"
)

// WasmxKeeper defines a subset of methods implemented by the cosmos-sdk account keeper
type WasmxKeeper interface {
	Query(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg wasmxtypes.RawContractMessage, funds sdk.Coins, deps []string) ([]byte, error)
	Execute(ctx sdk.Context, contractAddr mcodec.AccAddressPrefixed, senderAddr mcodec.AccAddressPrefixed, msg wasmxtypes.RawContractMessage, funds sdk.Coins, dependencies []string, inBackground bool) ([]byte, error)
	ExecuteContract(goCtx context.Context, msg *wasmxtypes.MsgExecuteContract) (*wasmxtypes.MsgExecuteContractResponse, error)
	ExecuteEntryPoint(ctx sdk.Context, entryPoint string, contractAddress mcodec.AccAddressPrefixed, caller mcodec.AccAddressPrefixed, msg []byte, dependencies []string, inBackground bool) ([]byte, error)
	ContractInstance(ctx sdk.Context, contractAddress sdk.AccAddress) (wasmxtypes.ContractInfo, wasmxtypes.CodeInfo, []byte, error)
	GetAddressOrRole(ctx sdk.Context, addressOrRole string) (mcodec.AccAddressPrefixed, error)
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmxtypes.ContractInfo

	ExecuteCosmosMsg(ctx sdk.Context, msg sdk.Msg, owner mcodec.AccAddressPrefixed) ([]sdk.Event, []byte, error)

	SmartContractCall(c context.Context, req *wasmxtypes.QuerySmartContractCallRequest) (rsp *wasmxtypes.QuerySmartContractCallResponse, err error)

	AddressCodec() address.Codec
	ValidatorAddressCodec() address.Codec
	ConsensusAddressCodec() address.Codec
	AccBech32Codec() mcodec.AccBech32Codec
}

type WasmxWrapper interface {
	ExecuteContract(ctx sdk.Context, req *MsgExecuteContract) (*MsgExecuteContractResponse, error)
	// QueryContract
	QueryContract(ctx sdk.Context, req *MsgQueryContract) (*MsgQueryContractResponse, error)
	GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmxtypes.ContractInfo

	Codec() codec.Codec
}
