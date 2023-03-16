package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmxtypes "wasmx/x/wasmx/types"
	"wasmx/x/websrv/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ContractByRoute(c context.Context, req *types.QueryContractByRouteRequest) (*types.QueryContractByRouteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Path == "" {
		return nil, types.ErrEmptyRoute
	}
	contractAddress := k.GetRouteToContract(sdk.UnwrapSDKContext(c), req.Path)
	return &types.QueryContractByRouteResponse{
		ContractAddress: contractAddress.String(),
	}, nil
}

func (k Keeper) RouteByContract(c context.Context, req *types.QueryRouteByContractRequest) (*types.QueryRouteByContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	contractAddress, err := sdk.AccAddressFromBech32(req.ContractAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}
	path := k.GetContractToRoute(sdk.UnwrapSDKContext(c), contractAddress)
	return &types.QueryRouteByContractResponse{
		Path: path,
	}, nil
}

func (k Keeper) HttpGet(c context.Context, req *types.QueryHttpRequestGet) (*types.QueryHttpResponseGet, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var request types.HttpRequest
	k.cdc.Unmarshal(req.HttpRequest, &request)

	rsp, err := k.HttpGetInternal(sdk.UnwrapSDKContext(c), request)
	if err != nil {
		return nil, err
	}
	return &types.QueryHttpResponseGet{Data: rsp}, nil
}

func (k Keeper) HttpGetInternal(ctx sdk.Context, req types.HttpRequest) (*types.HttpResponse, error) {
	headerMap := k.headersToMap(req)
	path := headerMap[types.HeaderOption_PathInfo]
	contractAddress := k.GetMostSpecificRouteToContract(ctx, path)
	if contractAddress == nil {
		return nil, sdkerrors.Wrapf(types.ErrRouteNotFound, "request path %s", path)
	}
	msg, err := types.RequestGetEncodeAbi(req)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "cannot encode abi for HttpRequestGet")
	}
	msgExecute := wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgExecuteBz, err := json.Marshal(msgExecute)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "cannot marshal WasmxExecutionMessage")
	}
	answ, err := k.wasmx.Query(ctx, contractAddress, types.ModuleAddress, msgExecuteBz, nil, nil)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "querying route contract failed")
	}

	var contractResponse wasmxtypes.WasmxQueryResponse
	err = json.Unmarshal(answ, &contractResponse)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "cannot unmarshal WasmxQueryResponse")
	}

	return types.ResponseGetDecodeAbi(contractResponse.Data)
}

func (k Keeper) headersToMap(req types.HttpRequest) map[types.HeaderOption]string {
	var headerMap map[types.HeaderOption]string
	for _, header := range req.Header {
		bz, bint := header.HeaderType.EnumDescriptor()
		headerMap[header.HeaderType] = header.Value
	}
	return headerMap
}
