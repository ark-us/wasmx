package keeper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	wasmxtypes "wasmx/x/wasmx/types"
	"wasmx/x/websrv/types"
)

func (k Keeper) Init() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", k.Route)

	err := http.ListenAndServe(":80", mux)
	if err != nil {
		return sdkerrors.Wrapf(err, "websrv could not start")
	}
	return nil
}

func (k Keeper) Route(w http.ResponseWriter, r *http.Request) {
	var response string
	var err error
	switch r.Method {
	case "GET":
		response, err = k.RouteGET(w, r)
	default:
		err = fmt.Errorf("websrv method %s not implemented", r.Method)
	}
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	io.WriteString(w, response)
}

func (k Keeper) RouteGET(w http.ResponseWriter, r *http.Request) (string, error) {
	params := []types.RequestParam{}
	req := types.HttpRequestGet{
		Url: &types.RequestUrl{
			Path:   r.URL.Path,
			Params: params,
		},
	}
	return k.HandleContractRoute(req)
}

func (k Keeper) HandleContractRoute(req types.HttpRequestGet) (string, error) {
	httpReqBz, err := req.Marshal()
	if err != nil {
		return "", sdkerrors.Wrapf(err, "cannot marshal HttpRequestGet")
	}
	websrvQuery := types.QueryHttpGetRequest{
		HttpRequest: httpReqBz,
	}
	websrvQueryBz, err := websrvQuery.Marshal()
	if err != nil {
		return "", sdkerrors.Wrapf(err, "cannot marshal QueryHttpGetRequest")
	}
	reqQuery := abci.RequestQuery{
		Path: "/wasmx.websrv.Query/HttpGet",
		Data: websrvQueryBz,
	}
	abcires := k.query(reqQuery)
	if abcires.IsErr() {
		return "", sdkerrors.Wrapf(types.ErrRouteInternalError, "log: %s", abcires.GetLog())
	}

	var respGet types.QueryHttpGetResponse
	err = respGet.Unmarshal(abcires.Value)
	if err != nil {
		return "", sdkerrors.Wrapf(err, "cannot unmarshal QueryHttpGetResponse")
	}

	var data wasmxtypes.WasmxQueryResponse
	err = json.Unmarshal(respGet.Data.Content, &data)
	if err != nil {
		return "", sdkerrors.Wrapf(err, "cannot unmarshal WasmxQueryResponse")
	}

	answ, err := types.ResponseGetDecodeAbi(data.Data)
	if err != nil {
		return "", sdkerrors.Wrapf(err, "cannot decode abi for WasmxQueryResponse")
	}

	return answ, nil
}

func (k Keeper) HttpGetInternal(ctx sdk.Context, req types.HttpRequestGet) (*types.HttpRequestGetResponse, error) {
	contractAddress := k.GetMostSpecificRouteToContract(ctx, req.Url.Path)
	if contractAddress == nil {
		return nil, sdkerrors.Wrapf(types.ErrRouteNotFound, "request path %s", req.Url.Path)
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
	return &types.HttpRequestGetResponse{Content: answ, ContentType: "text"}, nil
}
