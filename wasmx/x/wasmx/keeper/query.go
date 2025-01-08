package keeper

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkerr "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	"github.com/loredanacirstea/wasmx/x/wasmx/types"
	cchtypes "github.com/loredanacirstea/wasmx/x/wasmx/types/contract_handler"
)

var _ types.QueryServer = &Keeper{}

func (k *Keeper) ContractInfo(c context.Context, req *types.QueryContractInfoRequest) (*types.QueryContractInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	contractAddr, err := k.accBech32Codec.StringToAccAddressPrefixed(req.Address)
	if err != nil {
		return nil, err
	}
	rsp, err := queryContractInfo(sdk.UnwrapSDKContext(c), contractAddr, k)
	switch {
	case err != nil:
		return nil, err
	case rsp == nil:
		return nil, types.ErrNotFound
	}
	return rsp, nil
}

// ContractsByCode lists all smart contracts for a code id
func (k *Keeper) ContractsByCode(c context.Context, req *types.QueryContractsByCodeRequest) (*types.QueryContractsByCodeResponse, error) {
	// if req == nil {
	// 	return nil, status.Error(codes.InvalidArgument, "empty request")
	// }
	// if req.CodeId == 0 {
	// 	return nil, sdkerr.Wrap(types.ErrInvalid, "code id")
	// }
	// ctx := sdk.UnwrapSDKContext(c)
	// r := make([]string, 0)

	// prefixStore := prefix.NewStore(ctx.KVStore(q.storeKey), types.GetContractByCodeIDSecondaryIndexPrefix(req.CodeId))
	// pageRes, err := query.FilteredPaginate(prefixStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
	// 	if accumulate {
	// 		var contractAddr sdk.AccAddress = key[types.AbsoluteTxPositionLen:]
	// 		r = append(r, contractAddr.String()) // TODO codec stringify
	// 	}
	// 	return true, nil
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// return &types.QueryContractsByCodeResponse{
	// 	Contracts:  r,
	// 	Pagination: pageRes,
	// }, nil
	return &types.QueryContractsByCodeResponse{}, nil
}

func (k *Keeper) AllContractState(c context.Context, req *types.QueryAllContractStateRequest) (*types.QueryAllContractStateResponse, error) {
	// if req == nil {
	// 	return nil, status.Error(codes.InvalidArgument, "empty request")
	// }
	// contractAddr, err := k.AddressCodec().StringToBytes(req.Address)
	// if err != nil {
	// 	return nil, err
	// }
	// ctx := sdk.UnwrapSDKContext(c)
	// if !q.keeper.HasContractInfo(ctx, contractAddr) {
	// 	return nil, types.ErrNotFound
	// }

	// r := make([]types.Model, 0)
	// prefixStore := prefix.NewStore(ctx.KVStore(q.storeKey), types.GetContractStorePrefix(contractAddr))
	// pageRes, err := query.FilteredPaginate(prefixStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
	// 	if accumulate {
	// 		r = append(r, types.Model{
	// 			Key:   key,
	// 			Value: value,
	// 		})
	// 	}
	// 	return true, nil
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// return &types.QueryAllContractStateResponse{
	// 	Models:     r,
	// 	Pagination: pageRes,
	// }, nil
	return nil, nil
}

func (k *Keeper) RawContractState(c context.Context, req *types.QueryRawContractStateRequest) (*types.QueryRawContractStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	contractAddr, err := k.accBech32Codec.StringToAccAddressPrefixed(req.Address)
	if err != nil {
		return nil, err
	}

	if !k.HasContractInfo(ctx, contractAddr) {
		return nil, types.ErrNotFound
	}
	rsp := k.QueryRaw(ctx, contractAddr, req.QueryData)
	return &types.QueryRawContractStateResponse{Data: rsp}, nil
}

// func (k *Keeper) SmartContractState(c context.Context, req *types.QuerySmartContractStateRequest) (rsp *types.QuerySmartContractStateResponse, err error) {
// if req == nil {
// 	return nil, status.Error(codes.InvalidArgument, "empty request")
// }
// if err := req.QueryData.ValidateBasic(); err != nil {
// 	return nil, status.Error(codes.InvalidArgument, "invalid query data")
// }
// contractAddr, err := k.AddressCodec().StringToBytes(req.Address)
// if err != nil {
// 	return nil, err
// }
// ctx := sdk.UnwrapSDKContext(c).WithGasMeter(sdk.NewGasMeter(q.queryGasLimit))
// // recover from out-of-gas panic
// defer func() {
// 	if r := recover(); r != nil {
// 		switch rType := r.(type) {
// 		case storetypes.ErrorOutOfGas:
// 			err = sdkerr.Wrapf(sdkerrors.ErrOutOfGas,
// 				"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
// 				rType.Descriptor, ctx.GasMeter().Limit(), ctx.GasMeter().GasConsumed(),
// 			)
// 		default:
// 			err = sdkerrors.ErrPanic
// 		}
// 		rsp = nil
// 		moduleLogger(ctx).
// 			Debug("smart query contract",
// 				"error", "recovering panic",
// 				"contract-address", req.Address,
// 				"stacktrace", string(debug.Stack()))
// 	}
// }()

// bz, err := q.keeper.QuerySmart(ctx, contractAddr, req.QueryData)
// switch {
// case err != nil:
// 	return nil, err
// case bz == nil:
// 	return nil, types.ErrNotFound
// }
// return &types.QuerySmartContractStateResponse{Data: bz}, nil
// }

func (k *Keeper) SmartContractCall(c context.Context, req *types.QuerySmartContractCallRequest) (rsp *types.QuerySmartContractCallResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	// TODO validate deps
	ctx := sdk.UnwrapSDKContext(c).WithGasMeter(storetypes.NewGasMeter(k.queryGasLimit))
	if err := req.QueryData.ValidateBasic(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid query data")
	}
	contractAddr, err := k.GetAddressOrRole(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	sender, err := k.accBech32Codec.StringToAccAddressPrefixed(req.Sender)
	if err != nil {
		return nil, err
	}
	// recover from out-of-gas panic
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case storetypes.ErrorOutOfGas:
				err = sdkerr.Wrapf(sdkerrors.ErrOutOfGas,
					"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
					rType.Descriptor, ctx.GasMeter().Limit(), ctx.GasMeter().GasConsumed(),
				)
			default:
				err = sdkerrors.ErrPanic
			}
			rsp = nil
			k.Logger(ctx).
				Debug("smart query contract",
					"error", "recovering panic",
					"contract-address", req.Address,
					"stacktrace", string(debug.Stack()))
		}
	}()

	if types.IsSystemAddress(contractAddr.Bytes()) && !k.CanCallSystemContract(ctx, sender) {
		return nil, sdkerr.Wrap(types.ErrUnauthorizedAddress, "cannot call system address")
	}

	bz, err := k.Query(ctx, contractAddr, sender, req.QueryData, req.Funds, req.Dependencies)
	switch {
	case err != nil:
		return nil, err
	case bz == nil:
		return nil, types.ErrNotFound
	}
	return &types.QuerySmartContractCallResponse{Data: bz}, nil
}

func (k *Keeper) CallEth(c context.Context, req *types.QueryCallEthRequest) (rsp *types.QueryCallEthResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.QueryData.ValidateBasic(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid query data")
	}
	senderBech32 := req.Sender
	if senderBech32 == "" {
		senderBech32 = "mythos1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqvvnu6d"
	}
	sender, err := k.accBech32Codec.StringToAccAddressPrefixed(senderBech32)
	if err != nil {
		return nil, err
	}

	// TODO validate deps
	ctx := sdk.UnwrapSDKContext(c).WithGasMeter(storetypes.NewGasMeter(k.queryGasLimit))
	ctx = ctx.WithValue(cchtypes.CONTEXT_COIN_TYPE_KEY, cchtypes.COIN_TYPE_ETH)
	// recover from out-of-gas panic
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case storetypes.ErrorOutOfGas:
				err = sdkerr.Wrapf(sdkerrors.ErrOutOfGas,
					"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
					rType.Descriptor, ctx.GasMeter().Limit(), ctx.GasMeter().GasConsumed(),
				)
			default:
				err = sdkerrors.ErrPanic
			}
			rsp = nil
			k.Logger(ctx).
				Debug("smart query contract",
					"error", "recovering panic",
					"contract-address", req.Address,
					"stacktrace", string(debug.Stack()))
		}
	}()

	aliasAddr, found := k.GetAlias(ctx, sender)
	if found {
		sender = aliasAddr
	}

	if req.Address == "" {
		deps := []string{types.INTERPRETER_EVM_SHANGHAI}
		msg := types.WasmxExecutionMessage{Data: []byte{}}
		msgbz, err := json.Marshal(msg)
		if err != nil {
			sdkerr.Wrap(err, "ExecuteEth could not marshal data")
		}
		tempCtx, _ := ctx.CacheContext()

		_, _, _, err = k.Deploy(tempCtx, sender, req.QueryData, deps, types.CodeMetadata{}, msgbz, req.Funds, "")
		if err != nil {
			return nil, err
		}

		return &types.QueryCallEthResponse{Data: []byte{}, GasUsed: ctx.GasMeter().GasConsumed()}, nil
	}

	contractAddr, err := k.accBech32Codec.StringToAccAddressPrefixed(req.Address)
	if err != nil {
		return nil, err
	}
	if types.IsSystemAddress(contractAddr.Bytes()) && !k.CanCallSystemContract(ctx, sender) {
		return nil, sdkerr.Wrap(types.ErrUnauthorizedAddress, "cannot call system address")
	}
	bz, err := k.Query(ctx, contractAddr, sender, req.QueryData, req.Funds, req.Dependencies)
	switch {
	case err != nil:
		return nil, err
	case bz == nil:
		return nil, types.ErrNotFound
	}
	return &types.QueryCallEthResponse{Data: bz, GasUsed: ctx.GasMeter().GasConsumed()}, nil
}

func (k *Keeper) DebugContractCall(c context.Context, req *types.QueryDebugContractCallRequest) (rsp *types.QueryDebugContractCallResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	// TODO validate deps
	ctx := sdk.UnwrapSDKContext(c).WithGasMeter(storetypes.NewGasMeter(k.queryGasLimit))
	if err := req.QueryData.ValidateBasic(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid query data")
	}
	contractAddr, err := k.GetAddressOrRole(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	sender, err := k.accBech32Codec.StringToAccAddressPrefixed(req.Sender)
	if err != nil {
		return nil, err
	}

	// recover from out-of-gas panic
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case storetypes.ErrorOutOfGas:
				err = sdkerr.Wrapf(sdkerrors.ErrOutOfGas,
					"out of gas in location: %v; gasWanted: %d, gasUsed: %d",
					rType.Descriptor, ctx.GasMeter().Limit(), ctx.GasMeter().GasConsumed(),
				)
			default:
				err = sdkerrors.ErrPanic
			}
			rsp = nil
			k.Logger(ctx).
				Debug("smart query contract",
					"error", "recovering panic",
					"contract-address", req.Address,
					"stacktrace", string(debug.Stack()))
		}
	}()

	if !k.CanCallSystemContract(ctx, sender) {
		return nil, sdkerr.Wrap(types.ErrUnauthorizedAddress, "debug is non-deterministic and cannot be called as part of a transaction")
	}
	bz, memsnapshot, errMsg := k.QueryDebug(ctx, contractAddr, sender, req.QueryData, req.Funds, req.Dependencies)
	return &types.QueryDebugContractCallResponse{Data: bz, MemorySnapshot: memsnapshot, ErrorMessage: errMsg}, nil
}

func (k *Keeper) Code(c context.Context, req *types.QueryCodeRequest) (*types.QueryCodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.CodeId == 0 {
		return nil, sdkerr.Wrap(types.ErrInvalid, "code id")
	}
	rsp, err := queryCode(sdk.UnwrapSDKContext(c), req.CodeId, k)
	switch {
	case err != nil:
		return nil, err
	case rsp == nil:
		return nil, types.ErrNotFound
	}
	return &types.QueryCodeResponse{
		CodeInfoPB: rsp.CodeInfoPB,
		Data:       rsp.Data,
	}, nil
}

func (k *Keeper) CodeInfo(c context.Context, req *types.QueryCodeInfoRequest) (*types.QueryCodeInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.CodeId == 0 {
		return nil, sdkerr.Wrap(types.ErrInvalid, "code id")
	}
	ctx := sdk.UnwrapSDKContext(c)
	res, err := k.GetCodeInfo(ctx, req.CodeId)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, types.ErrNotFound
	}
	return &types.QueryCodeInfoResponse{
		CodeInfoPB: res.ToProto(),
	}, nil
}

func (k *Keeper) Codes(c context.Context, req *types.QueryCodesRequest) (*types.QueryCodesResponse, error) {
	// if req == nil {
	// 	return nil, status.Error(codes.InvalidArgument, "empty request")
	// }
	// ctx := sdk.UnwrapSDKContext(c)
	// r := make([]types.CodeInfo, 0)
	// prefixStore := prefix.NewStore(ctx.KVStore(q.storeKey), types.CodeKeyPrefix)
	// pageRes, err := query.FilteredPaginate(prefixStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
	// 	if accumulate {
	// 		var c types.CodeInfo
	// 		if err := q.cdc.Unmarshal(value, &c); err != nil {
	// 			return false, err
	// 		}
	// 		r = append(r, types.CodeInfoResponse{
	// 			CodeID:                binary.BigEndian.Uint64(key),
	// 			Creator:               c.Creator,
	// 			DataHash:              c.CodeHash,
	// 			InstantiatePermission: c.InstantiateConfig,
	// 		})
	// 	}
	// 	return true, nil
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// return &types.QueryCodesResponse{CodeInfos: r, Pagination: pageRes}, nil
	return nil, nil
}

func queryContractInfo(ctx sdk.Context, addr mcodec.AccAddressPrefixed, keeper *Keeper) (*types.QueryContractInfoResponse, error) {
	info, err := keeper.GetContractInfo(ctx, addr)
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, types.ErrNotFound
	}
	return &types.QueryContractInfoResponse{
		Address:        addr.String(),
		ContractInfoPB: *info.ToProto(),
	}, nil
}

func queryCode(ctx sdk.Context, codeID uint64, keeper *Keeper) (*types.QueryCodeResponse, error) {
	if codeID == 0 {
		return nil, nil
	}
	res, err := keeper.GetCodeInfo(ctx, codeID)
	if err != nil {
		return nil, err
	}
	if res == nil {
		// nil, nil leads to 404 in rest handler
		return nil, nil
	}

	code, err := keeper.GetByteCode(ctx, codeID)
	if err != nil {
		return nil, sdkerr.Wrap(err, "loading wasm code")
	}

	return &types.QueryCodeResponse{CodeInfoPB: res.ToProto(), Data: code}, nil
}

func (k *Keeper) ContractsByCreator(c context.Context, req *types.QueryContractsByCreatorRequest) (*types.QueryContractsByCreatorResponse, error) {
	// if req == nil {
	// 	return nil, status.Error(codes.InvalidArgument, "empty request")
	// }
	// ctx := sdk.UnwrapSDKContext(c)
	// contracts := make([]string, 0)

	// creatorAddress, err := k.AddressCodec().StringToBytes(req.CreatorAddress)
	// if err != nil {
	// 	return nil, err
	// }
	// prefixStore := prefix.NewStore(ctx.KVStore(q.storeKey), types.GetContractsByCreatorPrefix(creatorAddress))
	// pageRes, err := query.FilteredPaginate(prefixStore, req.Pagination, func(key []byte, _ []byte, accumulate bool) (bool, error) {
	// 	if accumulate {
	// 		accAddres := sdk.AccAddress(key[types.AbsoluteTxPositionLen:])
	// 		contracts = append(contracts, accAddres.String()) // TODO codec stringify
	// 	}
	// 	return true, nil
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// return &types.QueryContractsByCreatorResponse{
	// 	ContractAddresses: contracts,
	// 	Pagination:        pageRes,
	// }, nil
	return nil, nil
}
