package keeper

import (
	"context"
	"runtime/debug"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"mythos/v1/x/wasmx/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ContractInfo(c context.Context, req *types.QueryContractInfoRequest) (*types.QueryContractInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	contractAddr, err := sdk.AccAddressFromBech32(req.Address)
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
func (k Keeper) ContractsByCode(c context.Context, req *types.QueryContractsByCodeRequest) (*types.QueryContractsByCodeResponse, error) {
	// if req == nil {
	// 	return nil, status.Error(codes.InvalidArgument, "empty request")
	// }
	// if req.CodeId == 0 {
	// 	return nil, sdkerrors.Wrap(types.ErrInvalid, "code id")
	// }
	// ctx := sdk.UnwrapSDKContext(c)
	// r := make([]string, 0)

	// prefixStore := prefix.NewStore(ctx.KVStore(q.storeKey), types.GetContractByCodeIDSecondaryIndexPrefix(req.CodeId))
	// pageRes, err := query.FilteredPaginate(prefixStore, req.Pagination, func(key []byte, value []byte, accumulate bool) (bool, error) {
	// 	if accumulate {
	// 		var contractAddr sdk.AccAddress = key[types.AbsoluteTxPositionLen:]
	// 		r = append(r, contractAddr.String())
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

func (k Keeper) AllContractState(c context.Context, req *types.QueryAllContractStateRequest) (*types.QueryAllContractStateResponse, error) {
	// if req == nil {
	// 	return nil, status.Error(codes.InvalidArgument, "empty request")
	// }
	// contractAddr, err := sdk.AccAddressFromBech32(req.Address)
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

func (k Keeper) RawContractState(c context.Context, req *types.QueryRawContractStateRequest) (*types.QueryRawContractStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	contractAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	if !k.HasContractInfo(ctx, contractAddr) {
		return nil, types.ErrNotFound
	}
	rsp := k.QueryRaw(ctx, contractAddr, req.QueryData)
	return &types.QueryRawContractStateResponse{Data: rsp}, nil
}

// func (k Keeper) SmartContractState(c context.Context, req *types.QuerySmartContractStateRequest) (rsp *types.QuerySmartContractStateResponse, err error) {
// if req == nil {
// 	return nil, status.Error(codes.InvalidArgument, "empty request")
// }
// if err := req.QueryData.ValidateBasic(); err != nil {
// 	return nil, status.Error(codes.InvalidArgument, "invalid query data")
// }
// contractAddr, err := sdk.AccAddressFromBech32(req.Address)
// if err != nil {
// 	return nil, err
// }
// ctx := sdk.UnwrapSDKContext(c).WithGasMeter(sdk.NewGasMeter(q.queryGasLimit))
// // recover from out-of-gas panic
// defer func() {
// 	if r := recover(); r != nil {
// 		switch rType := r.(type) {
// 		case sdk.ErrorOutOfGas:
// 			err = sdkerrors.Wrapf(sdkerrors.ErrOutOfGas,
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

func (k Keeper) SmartContractCall(c context.Context, req *types.QuerySmartContractCallRequest) (rsp *types.QuerySmartContractCallResponse, err error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.QueryData.ValidateBasic(); err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid query data")
	}
	contractAddr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}
	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, err
	}
	// TODO validate deps
	ctx := sdk.UnwrapSDKContext(c).WithGasMeter(sdk.NewGasMeter(k.queryGasLimit))
	// recover from out-of-gas panic
	defer func() {
		if r := recover(); r != nil {
			switch rType := r.(type) {
			case sdk.ErrorOutOfGas:
				err = sdkerrors.Wrapf(sdkerrors.ErrOutOfGas,
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

	bz, err := k.Query(ctx, contractAddr, sender, req.QueryData, req.Funds, req.Dependencies)
	switch {
	case err != nil:
		return nil, err
	case bz == nil:
		return nil, types.ErrNotFound
	}
	return &types.QuerySmartContractCallResponse{Data: bz}, nil
}

func (k Keeper) Code(c context.Context, req *types.QueryCodeRequest) (*types.QueryCodeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.CodeId == 0 {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "code id")
	}
	rsp, err := queryCode(sdk.UnwrapSDKContext(c), req.CodeId, k)
	switch {
	case err != nil:
		return nil, err
	case rsp == nil:
		return nil, types.ErrNotFound
	}
	return &types.QueryCodeResponse{
		CodeInfo: rsp.CodeInfo,
		Data:     rsp.Data,
	}, nil
}

func (k Keeper) CodeInfo(c context.Context, req *types.QueryCodeInfoRequest) (*types.QueryCodeInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.CodeId == 0 {
		return nil, sdkerrors.Wrap(types.ErrInvalid, "code id")
	}
	res := k.GetCodeInfo(sdk.UnwrapSDKContext(c), req.CodeId)
	if res == nil {
		return nil, types.ErrNotFound
	}
	return &types.QueryCodeInfoResponse{
		CodeInfo: res,
	}, nil
}

func (k Keeper) Codes(c context.Context, req *types.QueryCodesRequest) (*types.QueryCodesResponse, error) {
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

func queryContractInfo(ctx sdk.Context, addr sdk.AccAddress, keeper Keeper) (*types.QueryContractInfoResponse, error) {
	info := keeper.GetContractInfo(ctx, addr)
	if info == nil {
		return nil, types.ErrNotFound
	}
	return &types.QueryContractInfoResponse{
		Address:      addr.String(),
		ContractInfo: *info,
	}, nil
}

func queryCode(ctx sdk.Context, codeID uint64, keeper Keeper) (*types.QueryCodeResponse, error) {
	if codeID == 0 {
		return nil, nil
	}
	res := keeper.GetCodeInfo(ctx, codeID)
	if res == nil {
		// nil, nil leads to 404 in rest handler
		return nil, nil
	}

	code, err := keeper.GetByteCode(ctx, codeID)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "loading wasm code")
	}

	return &types.QueryCodeResponse{CodeInfo: res, Data: code}, nil
}

func (k Keeper) ContractsByCreator(c context.Context, req *types.QueryContractsByCreatorRequest) (*types.QueryContractsByCreatorResponse, error) {
	// if req == nil {
	// 	return nil, status.Error(codes.InvalidArgument, "empty request")
	// }
	// ctx := sdk.UnwrapSDKContext(c)
	// contracts := make([]string, 0)

	// creatorAddress, err := sdk.AccAddressFromBech32(req.CreatorAddress)
	// if err != nil {
	// 	return nil, err
	// }
	// prefixStore := prefix.NewStore(ctx.KVStore(q.storeKey), types.GetContractsByCreatorPrefix(creatorAddress))
	// pageRes, err := query.FilteredPaginate(prefixStore, req.Pagination, func(key []byte, _ []byte, accumulate bool) (bool, error) {
	// 	if accumulate {
	// 		accAddres := sdk.AccAddress(key[types.AbsoluteTxPositionLen:])
	// 		contracts = append(contracts, accAddres.String())
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
