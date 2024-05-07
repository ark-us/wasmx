package vmp2p

import (
	"bufio"
	"encoding/json"
	"fmt"
	"time"

	log "cosmossdk.io/log"

	network "github.com/libp2p/go-libp2p/core/network"

	mcodec "mythos/v1/codec"
	networktypes "mythos/v1/x/network/types"
	vmtypes "mythos/v1/x/wasmx/vm"
)

var STREAM_MAIN = "mainstream"

// main stream
func (c *Context) handleStream(stream network.Stream) {
	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readDataStd(c.Context, c.Context.Ctx.Logger(), rw, STREAM_MAIN, c.handleContractMessage)
}

// peer stream
func (c *Context) listenPeerStream(stream network.Stream, peeraddrstr string) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readDataStd(c.Context, c.Context.Ctx.Logger(), rw, peeraddrstr, c.handleContractMessage)
	c.Context.Ctx.Logger().Debug("Connected to:", peeraddrstr)
}

func (c *Context) handleContractMessage(msgbz []byte, frompeer string) {
	var msg ContractMessage
	err := json.Unmarshal(msgbz, &msg)
	if err != nil {
		c.Context.Ctx.Logger().Debug(fmt.Sprintf("ContractMessage unmarshal failed: %s; err: %s", string(msgbz), err.Error()))
	}
	netmsg := P2PMessage{
		Message:   msg.Msg,
		Timestamp: time.Now(),
		RoomId:    "",
		Sender:    NodeInfo{Ip: frompeer},
	}
	c.handleMessage(netmsg, msg.ContractAddress, msg.SenderAddress)
}

func (c *Context) handleChatRoomMessage(crmsg *ChatRoomMessage) {
	var msg ContractMessage
	err := json.Unmarshal(crmsg.ContractMsg, &msg)
	if err != nil {
		c.Context.Ctx.Logger().Debug(fmt.Sprintf("chat room message unmarshal failed: %s; err: %s", string(crmsg.ContractMsg), err.Error()))
	}
	netmsg := P2PMessage{
		Message:   msg.Msg,
		Timestamp: crmsg.Timestamp,
		RoomId:    crmsg.RoomId,
		Sender:    crmsg.Sender,
	}
	c.handleMessage(netmsg, msg.ContractAddress, msg.SenderAddress)
}

func (c *Context) handleMessage(netmsg P2PMessage, contractAddress mcodec.AccAddressPrefixed, senderAddress mcodec.AccAddressPrefixed) {
	netmsgbz, err := json.Marshal(netmsg)
	if err != nil {
		c.Context.Ctx.Logger().Error("cannot marshall P2PMessage", "error", err.Error())
		return
	}
	contractAddressStr := contractAddress.String()
	senderAddressStr := senderAddress.String()

	c.Context.Ctx.Logger().Debug("p2p received message", "message", string(netmsgbz), "sender", senderAddressStr, "contract", contractAddressStr)

	msgtosend := &networktypes.MsgP2PReceiveMessageRequest{
		Sender:   senderAddressStr,
		Contract: contractAddressStr,
		Data:     netmsgbz,
	}
	_, _, err = c.Context.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		c.Context.Ctx.Logger().Error(err.Error())
	}
}

func readDataStd(ctx *vmtypes.Context, logger log.Logger, rw *bufio.ReadWriter, frompeer string, handleMessage func(msg []byte, frompeer string)) {
	logger.Debug("reading stream data from peer", "peer", frompeer)
	goCtx := ctx.GoContextParent
out:
	for {
		msgbz, err := rw.ReadBytes('\n')
		if err != nil {
			if err.Error() != ERROR_STREAM_RESET {
				logger.Error("Error reading from buffer", "error", err.Error(), "peer", frompeer)
			}
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
			handleMessage(msgbz, frompeer)
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
