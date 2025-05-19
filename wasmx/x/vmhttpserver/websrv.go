package vmhttpserver

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	sdkerr "cosmossdk.io/errors"
	"cosmossdk.io/log"

	cfg "github.com/loredanacirstea/wasmx/config"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
	wasmxtypes "github.com/loredanacirstea/wasmx/x/wasmx/types"
)

type WebsrvServer struct {
	coreHandler    wasmxtypes.WasmxCosmosHandler
	parentCtx      context.Context
	logger         log.Logger
	cfg            *WebsrvConfig
	actionExecutor cfg.ActionExecutor
	senderAddress  string
}

func NewWebsrvServer(
	coreHandler wasmxtypes.WasmxCosmosHandler,
	parentCtx context.Context,
	logger log.Logger,
	config *WebsrvConfig,
	actionExecutor cfg.ActionExecutor,
	senderAddress string,
) *WebsrvServer {
	return &WebsrvServer{
		coreHandler:    coreHandler,
		parentCtx:      parentCtx,
		logger:         logger.With(log.ModuleKey, "websrv"),
		cfg:            config,
		actionExecutor: actionExecutor,
		senderAddress:  senderAddress,
	}
}

func (k *WebsrvServer) Route(w http.ResponseWriter, r *http.Request) {
	response, err := k.HandleContractRoute(r)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}
	if response.Error != "" {
		io.WriteString(w, response.Error)
		return
	}
	for key, values := range response.Data.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.Write([]byte(response.Data.Data))
}

func (k *WebsrvServer) HandleContractRoute(r *http.Request) (*HttpResponseWrap, error) {
	// this can be a quick way to get the handling contract
	// if we need additional parsing of the route, we do it in the
	// contract starting the webserver
	contractAddress, ok := k.cfg.RouteToContractAddress[r.URL.Path]
	if !ok {
		// set default as the contract starting the webserver
		contractAddress = k.senderAddress
	}

	body := []byte{}
	var err error
	if r.ContentLength <= k.cfg.RequestBodyMaxSize {
		body, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, sdkerr.Wrapf(err, "incoming http request: cannot read body")
		}
	}

	req := HttpRequestIncoming{
		Method:        r.Method,
		Url:           r.URL.String(),
		RemoteAddr:    r.RemoteAddr,
		RequestURI:    r.RequestURI,
		Header:        r.Header,
		ContentLength: r.ContentLength,
		Data:          body,
	}

	httpReqBz, err := json.Marshal(req)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot marshal HttpRequestGet")
	}

	cb := func(goctx context.Context) (any, error) {
		// TODO consider not using reentry, just use normal contract execution...
		msg := &networktypes.MsgReentry{
			Sender:     k.senderAddress,
			Contract:   contractAddress,
			EntryPoint: ENTRY_POINT_HTTP_SERVER,
			Msg:        httpReqBz,
		}
		_, res, err := k.coreHandler.ExecuteCosmosMsg(msg)
		if err != nil {
			return nil, sdkerr.Wrapf(err, "http server incoming request execution failed")
		}
		return res, nil
	}
	bapp := k.actionExecutor.GetBaseApp()
	newctx, cancel := context.WithCancel(k.parentCtx)
	defer cancel()
	resp, err := k.actionExecutor.Execute(newctx, bapp.LastBlockHeight(), cb)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "Websrv HttpGet failed")
	}

	var requestResp networktypes.MsgReentryResponse
	err = requestResp.Unmarshal(resp.([]byte))
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot unmarshal MsgReentryResponse")
	}

	respHttp := &HttpResponseWrap{}
	err = json.Unmarshal(requestResp.Data, respHttp)
	if err != nil {
		return nil, sdkerr.Wrapf(err, "cannot unmarshal HttpResponseWrap")
	}

	return respHttp, nil
}
