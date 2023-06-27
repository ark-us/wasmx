package server

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/tendermint/tendermint/libs/log"

	"mythos/v1/x/wasmx/server/config"
	"mythos/v1/x/wasmx/types"
	wasmxtypes "mythos/v1/x/wasmx/types"
)

type JsonRpcServer struct {
	ctx         context.Context
	clientCtx   client.Context
	queryClient types.QueryClient // gRPC query client
	logger      log.Logger
	chainID     *big.Int
	longChainID string
	cfg         *config.JsonRpcConfig
}

// NewJsonRpcServer creates a new JSON-RPC instance for cosmos and ethereum namespaces
func NewJsonRpcServer(
	ctx *server.Context,
	logger log.Logger,
	clientCtx client.Context,
	config *config.JsonRpcConfig,
) *JsonRpcServer {
	chainID, err := wasmxtypes.ParseChainID(clientCtx.ChainID)
	if err != nil {
		panic(err)
	}

	// appConf, err := config.GetConfig(ctx.Viper)
	// if err != nil {
	// 	panic(err)
	// }

	// algos, _ := clientCtx.Keyring.SupportedAlgorithms()
	// if !algos.Contains(hd.EthSecp256k1) {
	// 	kr, err := keyring.New(
	// 		sdk.KeyringServiceName(),
	// 		viper.GetString(flags.FlagKeyringBackend),
	// 		clientCtx.KeyringDir,
	// 		clientCtx.Input,
	// 		clientCtx.Codec,
	// 		hd.EthSecp256k1Option(),
	// 	)
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	clientCtx = clientCtx.WithKeyring(kr)
	// }

	return &JsonRpcServer{
		ctx:         context.Background(),
		clientCtx:   clientCtx,
		queryClient: types.NewQueryClient(clientCtx),
		logger:      logger.With("module", "json-rpc"),
		chainID:     chainID,
		longChainID: clientCtx.ChainID,
		cfg:         config,
	}
}

func (k JsonRpcServer) Route(w http.ResponseWriter, r *http.Request) {
	var response []byte
	fmt.Println("--Route--", r)
	// var err error
	// switch r.Method {
	// case "GET":
	// 	response, err = k.RouteGET(w, r)
	// default:
	// 	err = fmt.Errorf("JSON RPC method %s not implemented", r.Method)
	// }
	// if err != nil {
	// 	io.WriteString(w, err.Error())
	// 	return
	// }
	w.Write(response)
}

// // ServeHTTP serves JSON-RPC requests over HTTP.
// func (s *JsonRpcServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	// Permit dumb empty requests for remote health-checks (AWS)
// 	if r.Method == http.MethodGet && r.ContentLength == 0 && r.URL.RawQuery == "" {
// 		w.WriteHeader(http.StatusOK)
// 		return
// 	}
// 	if code, err := validateRequest(r); err != nil {
// 		http.Error(w, err.Error(), code)
// 		return
// 	}

// 	// Create request-scoped context.
// 	connInfo := PeerInfo{Transport: "http", RemoteAddr: r.RemoteAddr}
// 	connInfo.HTTP.Version = r.Proto
// 	connInfo.HTTP.Host = r.Host
// 	connInfo.HTTP.Origin = r.Header.Get("Origin")
// 	connInfo.HTTP.UserAgent = r.Header.Get("User-Agent")
// 	ctx := r.Context()
// 	ctx = context.WithValue(ctx, peerInfoContextKey{}, connInfo)

// 	// All checks passed, create a codec that reads directly from the request body
// 	// until EOF, writes the response to w, and orders the server to process a
// 	// single request.
// 	w.Header().Set("content-type", contentType)
// 	codec := newHTTPServerConn(r, w)
// 	defer codec.close()
// 	s.serveSingleRequest(ctx, codec)
// }

// func (k JsonRpcServer) RouteGET(w http.ResponseWriter, r *http.Request) ([]byte, error) {
// 	// TODO query params
// 	params := []types.RequestQueryParam{}
// 	header := []types.HeaderItem{}

// 	header = append(header, types.HeaderItem{
// 		HeaderType: types.Query_String,
// 		Value:      r.URL.RawQuery,
// 	})
// 	header = append(header, types.HeaderItem{
// 		HeaderType: types.Path_Info,
// 		Value:      strings.ToLower(r.URL.Path),
// 	})

// 	for key, values := range r.Header {
// 		headerType, ok := types.HeaderStringToType[key]
// 		if ok {
// 			header = append(header, types.HeaderItem{
// 				HeaderType: headerType,
// 				Value:      strings.Join(values, ","),
// 			})
// 		}
// 	}

// 	paramPairs := strings.Split(r.URL.RawQuery, "&")
// 	for _, pair := range paramPairs {
// 		if len(pair) == 0 {
// 			continue
// 		}
// 		paramArr := strings.Split(pair, "=")
// 		if len(paramArr) >= 2 {
// 			params = append(params, types.RequestQueryParam{
// 				Key:   paramArr[0],
// 				Value: paramArr[1],
// 			})
// 		}
// 	}

// 	req := types.HttpRequest{
// 		Header:      header,
// 		QueryParams: params,
// 	}
// 	response, err := k.HandleContractRoute(req)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(err, "RouteGET failed")
// 	}
// 	for _, header := range response.Header {
// 		w.Header().Set(types.HeaderTypeToString[header.HeaderType], header.Value)
// 	}
// 	return []byte(response.Content), nil
// }

// func (k JsonRpcServer) HandleContractRoute(req types.HttpRequest) (*types.HttpResponse, error) {
// 	httpReqBz, err := json.Marshal(req)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(err, "cannot marshal HttpRequestGet")
// 	}
// 	websrvQuery := &types.QueryHttpRequestGet{
// 		HttpRequest: httpReqBz,
// 	}
// 	respGet, err := k.queryClient.HttpGet(k.ctx, websrvQuery)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(err, "Websrv HttpGet failed")
// 	}

// 	var requestResp types.HttpResponse
// 	err = json.Unmarshal(respGet.Data, &requestResp)
// 	if err != nil {
// 		return nil, sdkerrors.Wrapf(err, "cannot unmarshal HttpResponse")
// 	}

// 	return &requestResp, nil
// }
