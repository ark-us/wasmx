package vmp2p

import (
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/p2p/conn"
	sm "github.com/cometbft/cometbft/state"
)

type blockSyncReactor interface {
	SwitchToBlockSync(sm.State) error
}

type MockBlockSyncReactor struct{}

func (*MockBlockSyncReactor) SwitchToBlockSync(sm.State) error {
	return nil
}

// SetSwitch allows setting a switch.
func (*MockBlockSyncReactor) SetSwitch(*p2p.Switch) {

}

// GetChannels returns the list of MConnection.ChannelDescriptor. Make sure
// that each ID is unique across all the reactors added to the switch.
func (*MockBlockSyncReactor) GetChannels() []*conn.ChannelDescriptor {
	return nil
}

// InitPeer is called by the switch before the peer is started. Use it to
// initialize data for the peer (e.g. peer state).
//
// NOTE: The switch won't call AddPeer nor RemovePeer if it fails to start
// the peer. Do not store any data associated with the peer in the reactor
// itself unless you don't want to have a state, which is never cleaned up.
func (*MockBlockSyncReactor) InitPeer(peer p2p.Peer) p2p.Peer {
	// TODO init peer, this is temp
	return p2p.CreateRandomPeer(true)
}

// AddPeer is called by the switch after the peer is added and successfully
// started. Use it to start goroutines communicating with the peer.
func (*MockBlockSyncReactor) AddPeer(peer p2p.Peer) {

}

// RemovePeer is called by the switch when the peer is stopped (due to error
// or other reason).
func (*MockBlockSyncReactor) RemovePeer(peer p2p.Peer, reason interface{}) {

}

// Receive is called by the switch when an envelope is received from any connected
// peer on any of the channels registered by the reactor
func (*MockBlockSyncReactor) Receive(p2p.Envelope) {

}

func (*MockBlockSyncReactor) IsRunning() bool {
	return true
}
func (*MockBlockSyncReactor) OnReset() error {
	return nil

}
func (*MockBlockSyncReactor) OnStart() error {
	return nil
}
func (*MockBlockSyncReactor) OnStop() {

}
func (*MockBlockSyncReactor) Quit() <-chan struct{} {
	return nil
}
func (*MockBlockSyncReactor) Reset() error {
	return nil
}
func (*MockBlockSyncReactor) SetLogger(log.Logger) {

}
func (*MockBlockSyncReactor) Start() error {
	return nil
}
func (*MockBlockSyncReactor) Stop() error {
	return nil
}
func (*MockBlockSyncReactor) String() string {
	return ""
}
