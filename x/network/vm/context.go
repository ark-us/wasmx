package vm

import (
	"bufio"
	"encoding/json"
	"fmt"

	log "cosmossdk.io/log"

	network "github.com/libp2p/go-libp2p/core/network"

	networktypes "mythos/v1/x/network/types"
)

// main stream
func (c *Context) handleStream(stream network.Stream) {
	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readDataStd(c.Context.Ctx.Logger(), rw, "mainstream", c.handleMessage)
}

// peer stream
func (c *Context) listenPeerStream(stream network.Stream, peeraddrstr string) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readDataStd(c.Context.Ctx.Logger(), rw, "peerstream: "+peeraddrstr, c.handleMessage)
	c.Context.Ctx.Logger().Debug("Connected to:", peeraddrstr)
}

func (c *Context) handleMessage(msgbz []byte) {
	var msg P2PMessage
	err := json.Unmarshal(msgbz, &msg)
	if err != nil {
		c.Context.Ctx.Logger().Debug(fmt.Sprintf("p2p message unmarshal failed: %s; err: %s", string(msgbz), err.Error()))
	}

	msgtosend := &networktypes.MsgP2PReceiveMessageRequest{
		Sender:   msg.SenderAddress.String(),
		Contract: msg.ContractAddress.String(),
		Data:     msg.Msg,
	}
	_, _, err = c.Context.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		c.Context.Ctx.Logger().Error(err.Error())
	}
}

func readDataStd(logger log.Logger, rw *bufio.ReadWriter, frompeer string, handleMessage func(msg []byte)) {
	logger.Debug("reading stream data from peer", "peer", frompeer)
	for {
		msgbz, err := rw.ReadBytes('\n')
		if err != nil {
			logger.Error("Error reading from buffer", "peer", frompeer)
			return
		}

		if len(msgbz) == 0 {
			return
		}
		if string(msgbz) != "\n" {
			handleMessage(msgbz)
		}
	}
}
