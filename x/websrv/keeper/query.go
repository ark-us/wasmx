package keeper

import (
	"context"
	"encoding/json"

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
	// err := k.cdc.Unmarshal(req.HttpRequest, &request)
	err := json.Unmarshal(req.HttpRequest, &request)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "could not unmarshal HttpRequest")
	}

	rsp, err := k.HttpGetInternal(sdk.UnwrapSDKContext(c), request)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "http get failed")
	}
	rspbz, err := json.Marshal(rsp)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "could not marshal HttpResponse")
	}
	return &types.QueryHttpResponseGet{Data: rspbz}, nil
}

func (k Keeper) HttpGetInternal(ctx sdk.Context, req types.HttpRequest) (*types.HttpResponse, error) {
	headerMap := k.headersToMap(req)
	path := headerMap[types.Path_Info]
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

	resp, err := types.ResponseGetDecodeAbi(contractResponse.Data)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "cannot abi decode WasmxQueryResponse")
	}
	return resp, nil
}

func (k Keeper) headersToMap(req types.HttpRequest) map[uint8]string {
	var headerMap = map[uint8]string{}
	for _, header := range req.Header {
		headerMap[header.HeaderType] = header.Value
	}
	return headerMap
}
