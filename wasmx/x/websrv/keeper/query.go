package keeper

import (
	"context"
	"encoding/json"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdkerr "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	mcfg "mythos/v1/config"
	wasmxtypes "mythos/v1/x/wasmx/types"
	"mythos/v1/x/websrv/types"
)

var _ types.QueryServer = &Keeper{}

func (k *Keeper) ContractByRoute(c context.Context, req *types.QueryContractByRouteRequest) (*types.QueryContractByRouteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Path == "" {
		return nil, types.ErrEmptyRoute
	}
	contractAddress := k.GetRouteToContract(sdk.UnwrapSDKContext(c), req.Path)
	contractAddressStr, err := k.AddressCodec().BytesToString(contractAddress)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "alias: %s", mcfg.ERRORMSG_ACC_TOSTRING)
	}
	return &types.QueryContractByRouteResponse{
		ContractAddress: contractAddressStr,
	}, nil
}

func (k *Keeper) RouteByContract(c context.Context, req *types.QueryRouteByContractRequest) (*types.QueryRouteByContractResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	contractAddress, err := k.AddressCodec().StringToBytes(req.ContractAddress)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}
	path := k.GetContractToRoute(sdk.UnwrapSDKContext(c), contractAddress)
	return &types.QueryRouteByContractResponse{
		Path: path,
	}, nil
}

func (k *Keeper) GetAllOauthClients(c context.Context, req *types.QueryGetAllOauthClientsRequest) (*types.QueryGetAllOauthClientsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	clients := k.GetOauthClients(sdk.UnwrapSDKContext(c))
	return &types.QueryGetAllOauthClientsResponse{
		Clients: clients,
	}, nil
}

func (k *Keeper) GetOauthClient(c context.Context, req *types.QueryGetOauthClientRequest) (*types.QueryGetOauthClientResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	client, err := k.GetClientIdToInfo(sdk.UnwrapSDKContext(c), req.ClientId)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetOauthClientResponse{
		Client: client,
	}, nil
}

func (k *Keeper) GetOauthClientsByOwner(c context.Context, req *types.QueryGetOauthClientsByOwnerRequest) (*types.QueryGetOauthClientsByOwnerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	owner, err := k.AddressCodec().StringToBytes(req.Owner)
	if err != nil {
		return nil, sdkerrors.ErrInvalidAddress
	}

	clientIds, err := k.GetAddressToClients(sdk.UnwrapSDKContext(c), owner)
	if err != nil {
		return nil, err
	}
	return &types.QueryGetOauthClientsByOwnerResponse{
		ClientIds: clientIds,
	}, nil
}

func (k *Keeper) HttpGet(c context.Context, req *types.QueryHttpRequestGet) (*types.QueryHttpResponseGet, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	var request types.HttpRequest
	// err := k.cdc.Unmarshal(req.HttpRequest, &request)
	err := json.Unmarshal(req.HttpRequest, &request)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "could not unmarshal HttpRequest")
	}

	rsp, err := k.HttpGetInternal(sdk.UnwrapSDKContext(c), request)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "http get failed")
	}
	rspbz, err := json.Marshal(rsp)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "could not marshal HttpResponse")
	}
	return &types.QueryHttpResponseGet{Data: rspbz}, nil
}

func (k *Keeper) HttpGetInternal(ctx sdk.Context, req types.HttpRequest) (*types.HttpResponse, error) {
	headerMap := k.headersToMap(req)
	path := headerMap[types.Path_Info]
	contractAddress := k.GetMostSpecificRouteToContract(ctx, path)
	if contractAddress == nil {
		return nil, sdkerr.Wrapf(types.ErrRouteNotFound, "request path %s", path)
	}
	msg, err := types.RequestGetEncodeAbi(req)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot encode abi for HttpRequestGet")
	}
	msgExecute := wasmxtypes.WasmxExecutionMessage{Data: msg}
	msgExecuteBz, err := json.Marshal(msgExecute)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot marshal WasmxExecutionMessage")
	}
	contractAddressPrefixed := k.wasmx.AccBech32Codec().BytesToAccAddressPrefixed(contractAddress)
	senderPrefixed := k.wasmx.AccBech32Codec().BytesToAccAddressPrefixed(types.ModuleAddress)
	answ, err := k.wasmx.Query(ctx, contractAddressPrefixed, senderPrefixed, msgExecuteBz, nil, nil)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "querying route contract failed")
	}

	var contractResponse wasmxtypes.WasmxQueryResponse
	err = json.Unmarshal(answ, &contractResponse)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot unmarshal WasmxQueryResponse")
	}

	resp, err := types.ResponseGetDecodeAbi(contractResponse.Data)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot abi decode WasmxQueryResponse")
	}
	return resp, nil
}

func (k *Keeper) headersToMap(req types.HttpRequest) map[uint8]string {
	var headerMap = map[uint8]string{}
	for _, header := range req.Header {
		headerMap[header.HeaderType] = header.Value
	}
	return headerMap
}
