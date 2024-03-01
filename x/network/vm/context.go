package vm

import (
	"bufio"
	"encoding/json"
	"fmt"

	log "cosmossdk.io/log"

	network "github.com/libp2p/go-libp2p/core/network"

	networktypes "mythos/v1/x/network/types"
	vmtypes "mythos/v1/x/wasmx/vm"
)

var STREAM_MAIN = "mainstream"

// main stream
func (c *Context) handleStream(stream network.Stream) {
	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readDataStd(c.Context, c.Context.Ctx.Logger(), rw, STREAM_MAIN, c.handleMessage)
}

// peer stream
func (c *Context) listenPeerStream(stream network.Stream, peeraddrstr string) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readDataStd(c.Context, c.Context.Ctx.Logger(), rw, peeraddrstr, c.handleMessage)
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

func readDataStd(ctx *vmtypes.Context, logger log.Logger, rw *bufio.ReadWriter, frompeer string, handleMessage func(msg []byte)) {
	logger.Debug("reading stream data from peer", "peer", frompeer)
	goCtx := ctx.GoContextParent
out:
	for {
		msgbz, err := rw.ReadBytes('\n')
		if err != nil {
			logger.Error("Error reading from buffer", "error", err.Error(), "peer", frompeer)
			// remove stream if this is a direct peer stream
			if frompeer != STREAM_MAIN {
				p2pctx, err := GetP2PContext(ctx)
				if err == nil {
					delete(p2pctx.Streams, frompeer)
				}
			}
			return
		}

		if len(msgbz) == 0 {
			return
		}
		if string(msgbz) != "\n" {
			handleMessage(msgbz)
		}

		// if parent context is closing, stop receiving messages
		select {
		case <-goCtx.Done():
			logger.Debug("stopping peer libp2p connection", "peer", frompeer)
			break out
		default:
			// nothing
		}
	}
}
