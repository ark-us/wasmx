package vmp2p

import (
	"fmt"
	"net"
	"time"

	"github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/p2p"
)

type Switch struct{}

func (sw *Switch) AddPersistentPeers(addrs []string) error {
	fmt.Println("--Switch.AddPersistentPeers-", addrs)
	return nil
}

func (sw *Switch) AddPrivatePeerIDs(ids []string) error {
	fmt.Println("--Switch.AddPrivatePeerIDs-", ids)
	return nil
}

func (sw *Switch) AddReactor(name string, reactor p2p.Reactor) p2p.Reactor {
	fmt.Println("--Switch.AddReactor-", name)
	return nil
}

func (sw *Switch) AddUnconditionalPeerIDs(ids []string) error {
	fmt.Println("--Switch.AddUnconditionalPeerIDs-", ids)
	return nil
}

func (sw *Switch) Broadcast(e p2p.Envelope) chan bool {
	fmt.Println("--Switch.Broadcast-", e)
	return nil
}

func (sw *Switch) DialPeerWithAddress(addr *p2p.NetAddress) error {
	fmt.Println("--Switch.DialPeerWithAddress-", addr)
	return nil
}

func (sw *Switch) DialPeersAsync(peers []string) error {
	fmt.Println("--Switch.DialPeersAsync-", peers)
	return nil
}

func (sw *Switch) IsDialingOrExistingAddress(addr *p2p.NetAddress) bool {
	fmt.Println("--Switch.IsDialingOrExistingAddress-", addr)
	return true
}

func (sw *Switch) IsPeerPersistent(na *p2p.NetAddress) bool {
	fmt.Println("--Switch.IsPeerPersistent-", na)
	return true
}

func (sw *Switch) IsPeerUnconditional(id p2p.ID) bool {
	fmt.Println("--Switch.IsPeerUnconditional-", id)
	return true
}

// func (bs *service.BaseService) IsRunning() bool {
// 	return true
// }

func (sw *Switch) MarkPeerAsGood(peer p2p.Peer) {
	fmt.Println("--Switch.MarkPeerAsGood-", peer)
}

func (sw *Switch) MaxNumOutboundPeers() int {
	fmt.Println("--Switch.MaxNumOutboundPeers-")
	return 1000
}

func (sw *Switch) NetAddress() *p2p.NetAddress {
	fmt.Println("--Switch.NetAddress-")
	return nil
}

func (sw *Switch) NodeInfo() p2p.NodeInfo {
	fmt.Println("--Switch.NodeInfo-")
	return nil
}

func (sw *Switch) NumPeers() (outbound int, inbound int, dialing int) {
	fmt.Println("--Switch.NumPeers-", outbound, inbound, dialing)
	return 0, 0, 0
}

// func (bs *service.BaseService) OnReset() error {
// 	return nil
// }

func (sw *Switch) OnStart() error {
	fmt.Println("--Switch.OnStart-")
	return nil
}

func (sw *Switch) OnStop() {
	fmt.Println("--Switch.OnStop-")
}

func (sw *Switch) Peers() p2p.IPeerSet {
	fmt.Println("--Switch.Peers-")
	return nil
}

// func (bs *service.BaseService) Quit() <-chan struct{} {
// 	return nil
// }

func (sw *Switch) Reactor(name string) p2p.Reactor {
	fmt.Println("--Switch.Reactor-", name)
	return nil
}

func (sw *Switch) Reactors() map[string]p2p.Reactor {
	fmt.Println("--Switch.Reactors-")
	return nil
}

func (sw *Switch) RemoveReactor(name string, reactor p2p.Reactor) {
	fmt.Println("--Switch.RemoveReactor-", name, reactor)
}

// func (bs *service.BaseService) Reset() error {
// 	return nil
// }

func (sw *Switch) SetAddrBook(addrBook p2p.AddrBook) {
	fmt.Println("--Switch.SetAddrBook-", addrBook)
}

// func (bs *service.BaseService) SetLogger(l log.Logger) {
// 	return nil
// }

func (sw *Switch) SetNodeInfo(nodeInfo p2p.NodeInfo) {
	fmt.Println("--Switch.SetNodeInfo-", nodeInfo)
}

func (sw *Switch) SetNodeKey(nodeKey *p2p.NodeKey) {
	fmt.Println("--Switch.SetNodeKey-", nodeKey)
}

// func (bs *service.BaseService) Start() error {
// 	return nil
// }

// func (bs *service.BaseService) Stop() error {
// 	return nil
// }

func (sw *Switch) StopPeerForError(peer p2p.Peer, reason interface{}) {
	fmt.Println("--Switch.StopPeerForError-", peer, reason)
}

func (sw *Switch) StopPeerGracefully(peer p2p.Peer) {
	fmt.Println("--Switch.StopPeerGracefully-", peer)
}

// func (bs *service.BaseService) String() string {
// 	return ""
// }

// func (bs *service.BaseService) Wait() {
// 	return nil
// }

func (sw *Switch) acceptRoutine() {
	fmt.Println("--Switch.acceptRoutine-")
}

func (sw *Switch) addOutboundPeerWithConfig(addr *p2p.NetAddress, cfg *config.P2PConfig) error {
	fmt.Println("--Switch.addOutboundPeerWithConfig-", addr, cfg)
	return nil
}

func (sw *Switch) addPeer(p p2p.Peer) error {
	return nil
}

func (sw *Switch) addPeerWithConnection(conn net.Conn) error {
	return nil
}

func (sw *Switch) dialPeersAsync(netAddrs []*p2p.NetAddress) {
}

func (sw *Switch) filterPeer(p p2p.Peer) error {
	return nil
}

func (sw *Switch) randomSleep(interval time.Duration) {
}

func (sw *Switch) reconnectToPeer(addr *p2p.NetAddress) {
}

func (sw *Switch) stopAndRemovePeer(peer p2p.Peer, reason interface{}) {
}
