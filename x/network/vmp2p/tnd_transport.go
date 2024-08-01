package vmp2p

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	cmtconn "github.com/cometbft/cometbft/p2p/conn"

	"github.com/cosmos/gogoproto/proto"

	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
)

var CUSTOM_HANDLER_STATESYNC = "statesync"

type Transport struct{}

type peerConfig struct{}

// Listening address.
func (sw *Transport) NetAddress() p2p.NetAddress {
	fmt.Println("--Transport.NetAddress-")
	return p2p.NetAddress{}
}

// Accept returns a newly connected Peer.
func (sw *Transport) Accept(cfgi interface{}) (p2p.Peer, error) {
	fmt.Println("--Transport.Accept-")
	// cfg := cfgi.(peerConfig)
	return nil, nil
}

// Dial connects to the Peer for the address.
func (sw *Transport) Dial(netAddress p2p.NetAddress, cfgi interface{}) (p2p.Peer, error) {
	fmt.Println("--Transport.Dial-")
	return nil, nil
}

// Cleanup any resources associated with Peer.
func (sw *Transport) Cleanup(p2p.Peer) {
	fmt.Println("--Transport.Cleanup-")
}

type Peer struct {
	MultiAddress string
	Peer         *peerstore.AddrInfo
	Stream       network.Stream
	P2pCtx       *P2PContext
	ProtocolId   string
}

func (p *Peer) FlushStop() {
	fmt.Println("--Peer.FlushStop-")
}

// peer's cryptographic ID
func (p *Peer) ID() p2p.ID {
	return p2p.ID(p.MultiAddress)
}

// remote IP of the connection
func (p *Peer) RemoteIP() net.IP {
	fmt.Println("--Peer.RemoteIP-")
	return []byte{}
}

// remote address of the connection
func (p *Peer) RemoteAddr() net.Addr {
	fmt.Println("--Peer.RemoteAddr-")
	return nil
}

func (p *Peer) IsOutbound() bool {
	fmt.Println("--Peer.IsOutbound-")
	return false
}

// do we redial this peer when we disconnect
func (p *Peer) IsPersistent() bool {
	fmt.Println("--Peer.IsPersistent-")
	return false
}

// close original connection
func (p *Peer) CloseConn() error {
	fmt.Println("--Peer.CloseConn-")
	return nil
}

// peer's info
func (p *Peer) NodeInfo() p2p.NodeInfo {
	fmt.Println("--Peer.NodeInfo-")
	return nil
}

func (p *Peer) Status() cmtconn.ConnectionStatus {
	fmt.Println("--Peer.Status-")
	return cmtconn.ConnectionStatus{}
}

// actual address of the socket
func (p *Peer) SocketAddr() *p2p.NetAddress {
	fmt.Println("--Peer.SocketAddr-")
	return nil
}

type WrapMsg struct {
	Msg       []byte `json:"msg"`
	ChannelID byte   `json:"channel_id"`
}

func (p *Peer) Send(e p2p.Envelope) bool {
	msg := e.Message

	if w, ok := msg.(p2p.Wrapper); ok {
		msg = w.Wrap()
	}
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return false
	}

	msgWrapBytes, err := json.Marshal(&WrapMsg{Msg: msgBytes, ChannelID: e.ChannelID})
	if err != nil {
		return false
	}

	msgReq := ContractMessage{
		Msg:             msgWrapBytes,
		ContractAddress: CUSTOM_HANDLER_STATESYNC,
		SenderAddress:   CUSTOM_HANDLER_STATESYNC,
	}
	err = sendMessageToPeersInternal(p.Stream, msgReq)
	if err != nil {
		if err.Error() == ERROR_STREAM_RESET {
			// we just remove the stream from the mapping
			// if later needed, it will try to reconnect
			p.P2pCtx.DeletePeer(p.ProtocolId, p.MultiAddress)
		}
		return false
	}
	return true
}

func (p *Peer) TrySend(p2p.Envelope) bool {
	fmt.Println("--Peer.TrySend-")
	return false
}

func (p *Peer) Set(string, interface{}) {
	fmt.Println("--Peer.Set-")
}

func (p *Peer) Get(a string) interface{} {
	fmt.Println("--Peer.Get-", a)
	return nil
}

func (p *Peer) SetRemovalFailed() {
	fmt.Println("--Peer.SetRemovalFailed-")
}

func (p *Peer) GetRemovalFailed() bool {
	fmt.Println("--Peer.GetRemovalFailed-")
	return false
}

func (p *Peer) IsRunning() bool {
	fmt.Println("--Peer.IsRunning-")
	return true
}

func (p *Peer) OnReset() error {
	fmt.Println("--Peer.OnReset-")
	return nil
}
func (p *Peer) OnStart() error {
	fmt.Println("--Peer.OnStart-")
	return nil
}
func (p *Peer) OnStop() {
	fmt.Println("--Peer.OnStop-")
}

func (p *Peer) Quit() <-chan struct{} {
	fmt.Println("--Peer.Quit-")
	return nil
}
func (p *Peer) Reset() error {
	fmt.Println("--Peer.Reset-")
	return nil
}
func (p *Peer) SetLogger(l log.Logger) {
	fmt.Println("--Peer.SetLogger-")
}

func (p *Peer) Start() error {
	fmt.Println("--Peer.Start-")
	return nil
}
func (p *Peer) Stop() error {
	fmt.Println("--Peer.Stop-")
	return nil
}
func (p *Peer) String() string {
	fmt.Println("--Peer.String-")
	return p.MultiAddress
}
func (p *Peer) Wait() {
	fmt.Println("--Peer.Wait-")
}

func NewPeer(
	multiaddress string,
	stream network.Stream,
	peerInfo *peerstore.AddrInfo,
	protocolId string,
	p2pctx *P2PContext,
) *Peer {
	p := &Peer{
		MultiAddress: multiaddress,
		Stream:       stream,
		Peer:         peerInfo,
		ProtocolId:   protocolId,
		P2pCtx:       p2pctx,
	}
	return p
}
