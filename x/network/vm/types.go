package vm

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"

	sdk "github.com/cosmos/cosmos-sdk/types"

	vmtypes "mythos/v1/x/wasmx/vm"
)

const HOST_WASMX_ENV_P2P_VER1 = "wasmx_p2p_1"

var HOST_WASMX_ENV_EXPORT = "wasmx_p2p_"

var HOST_WASMX_ENV_P2P = "p2p"

type ContextKey string

const P2PContextKey ContextKey = "p2p-context"

type P2PMessage struct {
	Msg             []byte         `json:"msg"`
	ContractAddress sdk.AccAddress `json:"contract_address"`
}

type Peer struct {
	Id   string `json:"id"`
	Port string `json:"port"`
	Host string `json:"host"`
}

type P2PContext struct {
	Node    *host.Host
	Streams map[string]network.Stream
}

type StartNodeWithIdentityRequest struct {
	Port       string `json:"port"`
	ProtocolId string `json:"protocolId"`
	PrivateKey []byte `json:"pk"`
}

type StartNodeWithIdentityResponse struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
}

type ConnectPeerRequest struct {
	ProtocolId string `json:"protocolId"`
	Peer       string `json:"peer"`
}

type ConnectPeerResponse struct{}

type SendMessageToPeersRequest struct {
	Msg        []byte   `json:"msg"`
	ProtocolId string   `json:"protocolId"`
	Peers      []string `json:"peers"`
}

type SendMessageToPeersResponse struct{}

type SendMessageRequest struct {
	Msg []byte `json:"msg"`
}

type SendMessageResponse struct{}

type MsgStart struct {
	PrivateKey []byte `json:"pk"`
	ProtocolId string `json:"protocolId"`
	Node       Peer   `json:"node"`
	Peers      []Peer `json:"peers"`
}

type MsgStart2 struct {
	ProtocolIdd string `json:"protocolIdd"`
	ProtocolId  string `json:"protocolId"`
	PK          []byte `json:"pk"`
}

type CalldataStart struct {
	Start MsgStart `json:"start"`
}

func WithP2PEmptyContext(ctx context.Context) context.Context {
	p2pctx := &P2PContext{Streams: make(map[string]network.Stream)}
	return context.WithValue(ctx, P2PContextKey, p2pctx)
}

func WithP2PContext(ctx context.Context, p2pctx *P2PContext) context.Context {
	return context.WithValue(ctx, P2PContextKey, p2pctx)
}

func GetP2PContext(ctx *vmtypes.Context) (*P2PContext, error) {
	p2pctx_ := ctx.GoContextParent.Value(P2PContextKey)
	p2pctx := (p2pctx_).(*P2PContext)
	if p2pctx == nil {
		return nil, fmt.Errorf("p2p context not set")
	}
	return p2pctx, nil
}
