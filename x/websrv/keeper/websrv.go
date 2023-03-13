package keeper

import (
	"fmt"
	"io"
	"net/http"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

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
	var response []byte
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
	w.Write(response)
}

func (k Keeper) RouteGET(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	params := []types.RequestParam{}
	req := types.HttpRequestGet{
		Url: &types.RequestUrl{
			Path:   r.URL.Path,
			Params: params,
		},
	}
	content, contentType, err := k.HandleContractRoute(req)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "RouteGET failed")
	}
	w.Header().Set("Content-Type", contentType)
	return content, nil
}

func (k Keeper) HandleContractRoute(req types.HttpRequestGet) ([]byte, string, error) {
	httpReqBz, err := req.Marshal()
	if err != nil {
		return nil, "", sdkerrors.Wrapf(err, "cannot marshal HttpRequestGet")
	}
	websrvQuery := types.QueryHttpGetRequest{
		HttpRequest: httpReqBz,
	}
	websrvQueryBz, err := websrvQuery.Marshal()
	if err != nil {
		return nil, "", sdkerrors.Wrapf(err, "cannot marshal QueryHttpGetRequest")
	}
	reqQuery := abci.RequestQuery{
		Path: "/wasmx.websrv.Query/HttpGet",
		Data: websrvQueryBz,
	}
	abcires := k.query(reqQuery)
	if abcires.IsErr() {
		return nil, "", sdkerrors.Wrapf(types.ErrRouteInternalError, "log: %s", abcires.GetLog())
	}

	var respGet types.QueryHttpGetResponse
	err = respGet.Unmarshal(abcires.Value)
	if err != nil {
		return nil, "", sdkerrors.Wrapf(err, "cannot unmarshal QueryHttpGetResponse")
	}

	return respGet.Data.Content, respGet.Data.ContentType, nil
}
