package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/libs/log"

	wasmxtypes "wasmx/v1/x/wasmx/types"
	"wasmx/v1/x/websrv/server/config"
	"wasmx/v1/x/websrv/types"
)

type WebsrvServer struct {
	ctx         context.Context
	clientCtx   client.Context
	queryClient types.QueryClient // gRPC query client
	logger      log.Logger
	chainID     *big.Int
	longChainID string
	cfg         *config.WebsrvConfig
}

// NewBackend creates a new Backend instance for cosmos and ethereum namespaces
func NewWebsrvServer(
	ctx *server.Context,
	logger log.Logger,
	clientCtx client.Context,
	config *config.WebsrvConfig,
) *WebsrvServer {
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

	return &WebsrvServer{
		ctx:         context.Background(),
		clientCtx:   clientCtx,
		queryClient: types.NewQueryClient(clientCtx),
		logger:      logger.With("module", "backend"),
		chainID:     chainID,
		longChainID: clientCtx.ChainID,
		cfg:         config,
	}
}

func (k WebsrvServer) Route(w http.ResponseWriter, r *http.Request) {
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

func (k WebsrvServer) RouteGET(w http.ResponseWriter, r *http.Request) ([]byte, error) {
	// TODO query params
	params := []types.RequestQueryParam{}
	header := []types.HeaderItem{}

	header = append(header, types.HeaderItem{
		HeaderType: types.Query_String,
		Value:      r.URL.RawQuery,
	})
	header = append(header, types.HeaderItem{
		HeaderType: types.Path_Info,
		Value:      strings.ToLower(r.URL.Path),
	})

	for key, values := range r.Header {
		headerType, ok := types.HeaderStringToType[key]
		if ok {
			header = append(header, types.HeaderItem{
				HeaderType: headerType,
				Value:      strings.Join(values, ","),
			})
		}
	}

	paramPairs := strings.Split(r.URL.RawQuery, "&")
	for _, pair := range paramPairs {
		if len(pair) == 0 {
			continue
		}
		paramArr := strings.Split(pair, "=")
		if len(paramArr) >= 2 {
			params = append(params, types.RequestQueryParam{
				Key:   paramArr[0],
				Value: paramArr[1],
			})
		}
	}

	req := types.HttpRequest{
		Header:      header,
		QueryParams: params,
	}
	response, err := k.HandleContractRoute(req)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "RouteGET failed")
	}
	for _, header := range response.Header {
		w.Header().Set(types.HeaderTypeToString[header.HeaderType], header.Value)
	}
	return []byte(response.Content), nil
}

func (k WebsrvServer) HandleContractRoute(req types.HttpRequest) (*types.HttpResponse, error) {
	httpReqBz, err := json.Marshal(req)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "cannot marshal HttpRequestGet")
	}
	websrvQuery := &types.QueryHttpRequestGet{
		HttpRequest: httpReqBz,
	}
	respGet, err := k.queryClient.HttpGet(k.ctx, websrvQuery)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "Websrv HttpGet failed")
	}

	var requestResp types.HttpResponse
	err = json.Unmarshal(respGet.Data, &requestResp)
	if err != nil {
		return nil, sdkerrors.Wrapf(err, "cannot unmarshal HttpResponse")
	}

	return &requestResp, nil
}
