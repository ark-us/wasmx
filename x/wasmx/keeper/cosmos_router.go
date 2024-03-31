package keeper

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	sdkerr "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	verifysig "mythos/v1/crypto/verifysig"
	cw8types "mythos/v1/x/wasmx/cw8/types"
	"mythos/v1/x/wasmx/types"
)

type WasmxCosmosHandler struct {
	Ctx             sdk.Context
	Keeper          *Keeper
	ContractAddress sdk.AccAddress
}

func (h *WasmxCosmosHandler) WithContext(newctx sdk.Context) {
	h.Ctx = newctx
}
func (h *WasmxCosmosHandler) SubmitCosmosQuery(reqQuery *abci.RequestQuery) ([]byte, error) {
	return h.Keeper.SubmitCosmosQuery(h.Ctx, reqQuery)
}
func (h *WasmxCosmosHandler) DecodeCosmosTx(bz []byte) ([]byte, error) {
	return h.Keeper.DecodeCosmosTx(bz)
}
func (h *WasmxCosmosHandler) VerifyCosmosTx(bz []byte) (bool, error) {
	return h.Keeper.VerifyCosmosTx(h.Ctx, bz)
}

func (h *WasmxCosmosHandler) ExecuteCosmosMsgAny(any *cdctypes.Any) ([]sdk.Event, []byte, error) {
	return h.Keeper.ExecuteCosmosMsgAny(h.Ctx, any, h.ContractAddress)
}
func (h *WasmxCosmosHandler) ExecuteCosmosMsgAnyBz(msgbz []byte) ([]sdk.Event, []byte, error) {
	any, err := h.Keeper.AnyFromBz(msgbz)
	if err != nil {
		return nil, nil, err
	}
	return h.Keeper.ExecuteCosmosMsgAny(h.Ctx, &any, h.ContractAddress)
}
func (h *WasmxCosmosHandler) ExecuteCosmosMsg(msg sdk.Msg) ([]sdk.Event, []byte, error) {
	return h.Keeper.ExecuteCosmosMsg(h.Ctx, msg, h.ContractAddress)
}
func (h *WasmxCosmosHandler) WasmVMQueryHandler(caller sdk.AccAddress, request cw8types.QueryRequest) ([]byte, error) {
	return h.Keeper.wasmVMQueryHandler.HandleQuery(h.Ctx, caller, request)
}
func (h *WasmxCosmosHandler) JSONCodec() codec.JSONCodec {
	return h.Keeper.cdc
}
func (h *WasmxCosmosHandler) GetAlias(addr sdk.AccAddress) (sdk.AccAddress, bool) {
	return h.Keeper.GetAlias(h.Ctx, addr)
}
func (h *WasmxCosmosHandler) GetAccount(addr sdk.AccAddress) sdk.AccountI {
	aliasAddr, found := h.Keeper.GetAlias(h.Ctx, addr)
	if found {
		addr = aliasAddr
	}
	acc := h.Keeper.GetAccount(h.Ctx, addr)
	return acc
}
func (h *WasmxCosmosHandler) GetCodeHash(contractAddress sdk.AccAddress) types.Checksum {
	_, codeInfo, _, err := h.Keeper.ContractInstance(h.Ctx, contractAddress)
	if err != nil {
		return types.EMPTY_BYTES32
	}
	return codeInfo.CodeHash
}
func (h *WasmxCosmosHandler) GetCode(contractAddress sdk.AccAddress) []byte {
	_, codeInfo, _, err := h.Keeper.ContractInstance(h.Ctx, contractAddress)
	if err != nil {
		return []byte{}
	}
	if len(codeInfo.InterpretedBytecodeRuntime) > 0 {
		return codeInfo.InterpretedBytecodeRuntime
	}
	code, err := h.Keeper.GetCode(codeInfo.CodeHash, codeInfo.Deps)
	if err != nil {
		return []byte{}
	}
	return code
}

func (h *WasmxCosmosHandler) GetCodeInfo(contractAddress sdk.AccAddress) types.CodeInfo {
	_, codeInfo, _, err := h.Keeper.ContractInstance(h.Ctx, contractAddress)
	if err != nil {
		return types.CodeInfo{CodeHash: types.EMPTY_BYTES32}
	}
	return codeInfo
}

func (h *WasmxCosmosHandler) GetContractInstance(contractAddress sdk.AccAddress) (types.ContractInfo, types.CodeInfo, []byte, error) {
	return h.Keeper.ContractInstance(h.Ctx, contractAddress)
}

// TODO
func (h *WasmxCosmosHandler) GetBlockHash(blockNumber uint64) types.Checksum {
	return types.EMPTY_BYTES32
}
func (h *WasmxCosmosHandler) ContractStore(ctx sdk.Context, storageType types.ContractStorageType, prefixStoreKey []byte) prefix.Store {
	return h.Keeper.ContractStore(ctx, storageType, prefixStoreKey)
}

// TODO provenance
func (h *WasmxCosmosHandler) Create(codeId uint64, creator sdk.AccAddress, initMsg []byte, label string, value *big.Int, funds sdk.Coins) (sdk.AccAddress, error) {
	if value != nil {
		funds = sdk.NewCoins(sdk.NewCoin(h.Keeper.denom, sdkmath.NewIntFromBigInt(value)))
	}
	address, _, err := h.Keeper.Instantiate(h.Ctx, codeId, creator, initMsg, funds, label)
	if err != nil {
		return nil, err
	}
	return address, nil
}

// TODO provenance
func (h *WasmxCosmosHandler) Create2(codeId uint64, creator sdk.AccAddress, initMsg []byte, salt types.Checksum, label string, value *big.Int, funds sdk.Coins) (sdk.AccAddress, error) {
	if value != nil {
		funds = sdk.NewCoins(sdk.NewCoin(h.Keeper.denom, sdkmath.NewIntFromBigInt(value)))
	}
	address, _, err := h.Keeper.Instantiate2(h.Ctx, codeId, creator, initMsg, funds, salt, false, label)
	return address, err
}
func (h *WasmxCosmosHandler) Deploy(bytecode []byte, sender sdk.AccAddress, provenance sdk.AccAddress, initMsg []byte, value *big.Int, deps []string, metadata types.CodeMetadata, label string, salt []byte) (codeId uint64, checksum []byte, contractAddress sdk.AccAddress, err error) {
	funds := sdk.NewCoins(sdk.NewCoin(h.Keeper.denom, sdkmath.NewIntFromBigInt(value)))
	return h.Keeper.CreateInterpreted(h.Ctx, sender, provenance, bytecode, deps, metadata, initMsg, funds, label, salt)
}

func (h *WasmxCosmosHandler) Execute(contractAddress sdk.AccAddress, sender sdk.AccAddress, execmsg []byte, value *big.Int, deps []string) (res []byte, err error) {
	funds := sdk.NewCoins(sdk.NewCoin(h.Keeper.denom, sdkmath.NewIntFromBigInt(value)))
	return h.Keeper.Execute(h.Ctx, contractAddress, sender, execmsg, funds, deps)
}

func (h *WasmxCosmosHandler) GetContractDependency(ctx sdk.Context, addr sdk.AccAddress) (types.ContractDependency, error) {
	return h.Keeper.GetContractDependency(ctx, addr)
}
func (h *WasmxCosmosHandler) CanCallSystemContract(ctx sdk.Context, addr sdk.AccAddress) bool {
	return h.Keeper.CanCallSystemContract(ctx, addr)
}
func (h *WasmxCosmosHandler) GetAddressOrRole(ctx sdk.Context, addressOrRole string) (sdk.AccAddress, error) {
	return h.Keeper.GetAddressOrRole(ctx, addressOrRole)
}
func (h *WasmxCosmosHandler) GetRoleByContractAddress(ctx sdk.Context, addr sdk.AccAddress) string {
	return h.Keeper.GetRoleByContractAddress(ctx, addr)
}

func (h *WasmxCosmosHandler) WithNewAddress(addr sdk.AccAddress) types.WasmxCosmosHandler {
	return h.Keeper.newCosmosHandler(h.Ctx, addr)
}

func (k *Keeper) newCosmosHandler(ctx sdk.Context, contractAddress sdk.AccAddress) types.WasmxCosmosHandler {
	return &WasmxCosmosHandler{
		Ctx:             ctx,
		Keeper:          k,
		ContractAddress: contractAddress,
	}
}

func (k *Keeper) SubmitCosmosQuery(ctx sdk.Context, reqQuery *abci.RequestQuery) ([]byte, error) {
	// TODO if we allow historical queries, at a certain block
	// use app.Query(queryReq)
	queryFn := k.grpcQueryRouter.Route(reqQuery.Path)
	res, err := safeQuery(ctx, reqQuery, queryFn)
	if err != nil {
		return nil, err
	}
	return res.Value, nil
}

func (h *Keeper) DecodeCosmosTx(bz []byte) ([]byte, error) {
	tx, err := h.txConfig.TxDecoder()(bz)
	if err != nil {
		return nil, err
	}
	txbz, err := h.txConfig.TxJSONEncoder()(tx)
	if err != nil {
		return nil, err
	}
	return txbz, nil
}

func (h *Keeper) VerifyCosmosTx(ctx sdk.Context, bz []byte) (bool, error) {
	tx, err := h.txConfig.TxDecoder()(bz)
	if err != nil {
		return false, err
	}
	ak := NewAccountKeeperVerifySig(h)
	valid, err := verifysig.VerifySignature(ctx, ak, tx, false, h.txConfig.SignModeHandler())

	if err != nil {
		return false, err
	}
	return valid, nil
}

func (k *Keeper) ExecuteCosmosMsgAny(ctx sdk.Context, any *cdctypes.Any, owner sdk.AccAddress) ([]sdk.Event, []byte, error) {
	// sdk.Msg
	var msg sdk.Msg
	err := k.cdc.UnpackAny(any, &msg)
	if err != nil {
		return nil, nil, err
	}
	return k.ExecuteCosmosMsg(ctx, msg, owner)
}

func (k *Keeper) ExecuteCosmosMsg(ctx sdk.Context, msg sdk.Msg, owner sdk.AccAddress) ([]sdk.Event, []byte, error) {
	anymsg, err := cdctypes.NewAnyWithValue(msg)
	if err != nil {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("wasmx cosmos message any")
	}
	signers, _, err := k.cdc.GetMsgAnySigners(anymsg)
	if err != nil || len(signers) == 0 {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("wasmx cosmos message signer missing")
	}
	authorized := bytes.Equal(signers[0], owner.Bytes())
	if !authorized {
		authorized = sdk.AccAddress(signers[0]).String() == k.authority
	}
	if !authorized {
		return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("wasmx cosmos message signer %s, expected %s", sdk.AccAddress(signers[0]).String(), owner.String())
	}
	// signers := msg.GetSigners()
	// if signers[0].String() != owner.String() {
	// 	return nil, nil, sdkerrors.ErrUnauthorized.Wrapf("wasmx cosmos message signer %s, expected %s", signers[0].String(), owner.String())
	// }

	return k.executeMsg(ctx, msg)
}

// Attempts to get the message handler from the router and if found will then execute the message.
// If the message execution is successful, the proto marshaled message response will be returned.
func (k *Keeper) executeMsg(ctx sdk.Context, msg sdk.Msg) ([]sdk.Event, []byte, error) {
	handler := k.msgRouter.Handler(msg)
	if handler == nil {
		return nil, nil, types.ErrInvalidRoute
	}
	// handler can panic with out of gas or other errors
	res, err := safeHandler(ctx, msg, handler)
	if err != nil {
		return nil, nil, err
	}

	return res.GetEvents(), res.Data, nil
}

// func (k *Keeper) newQueryHandler(ctx sdk.Context, contractAddress sdk.AccAddress) QueryHandler {
// 	return NewQueryHandler(ctx, k, contractAddress, k.gasRegister)
// }

func safeHandler(ctx sdk.Context, msg sdk.Msg, handler func(ctx sdk.Context, req sdk.Msg) (*sdk.Result, error)) (res *sdk.Result, err error) {
	defer func() {
		if r := recover(); r != nil {
			// TODO maybe there is a better way to get the cosmos sdk errors
			// and make them EVM errors
			switch x := r.(type) {
			case string:
				err = fmt.Errorf("failed to execute Cosmos message: %s", x)
			case error:
				err = sdkerr.Wrapf(x, "failed to execute Cosmos message")
			default:
				// Fallback err
				err = sdkerr.Wrapf(sdkerrors.ErrPanic, "unknown panic %v", r)
			}
			// invalidate rep
			res = nil
		}
	}()

	res, err = handler(ctx, msg)
	return res, err
}

func safeQuery(
	ctx sdk.Context,
	msg *abci.RequestQuery,
	handler func(ctx sdk.Context, req *abci.RequestQuery) (*abci.ResponseQuery, error),
) (res *abci.ResponseQuery, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = sdkerr.Wrapf(x, "failed to execute Cosmos message")
			default:
				// Fallback err
				err = sdkerr.Wrapf(sdkerrors.ErrPanic, "unknown panic %v", r)

			}
			// invalidate res
			res = nil
		}
	}()

	resp, err := handler(ctx, msg)
	return resp, err
}
