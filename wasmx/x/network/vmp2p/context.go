package vmp2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	log "cosmossdk.io/log"

	network "github.com/libp2p/go-libp2p/core/network"

	mcodec "github.com/loredanacirstea/wasmx/codec"
	networktypes "github.com/loredanacirstea/wasmx/x/network/types"
)

// main stream
func (c *Context) handleStream(stream network.Stream) {
	// Create a buffer stream for non-blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	peeraddr := stream.Conn().RemoteMultiaddr().String() + "/p2p/" + stream.Conn().RemotePeer().String()
	go readDataStd(c.Context.GoContextParent, c.Logger, rw, string(stream.Protocol()), peeraddr, c.handleContractMessage)
}

// peer stream
func (c *Context) listenPeerStream(stream network.Stream, peeraddrstr string) {
	rw := bufio.NewReadWriter(bufio.NewReader(stream), bufio.NewWriter(stream))
	go readDataStd(c.Context.GoContextParent, c.Logger, rw, string(stream.Protocol()), peeraddrstr, c.handleContractMessage)
	c.Logger.Debug("Connected to:", peeraddrstr)
}

func (c *Context) handleContractMessage(msgbz []byte, frompeer string) {
	var msg ContractMessage
	err := json.Unmarshal(msgbz, &msg)
	if err != nil {
		c.Logger.Debug(fmt.Sprintf("ContractMessage unmarshal failed: %s; err: %s", string(msgbz), err.Error()))
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
		c.Logger.Debug(fmt.Sprintf("chat room message unmarshal failed: %s; err: %s", string(crmsg.ContractMsg), err.Error()))
	}
	netmsg := P2PMessage{
		Message:   msg.Msg,
		Timestamp: crmsg.Timestamp,
		RoomId:    crmsg.RoomId,
		Sender:    crmsg.Sender,
	}
	c.handleMessage(netmsg, msg.ContractAddress, msg.SenderAddress)
}

func (c *Context) handleMessage(netmsg P2PMessage, contractAddress string, senderAddress string) {
	netmsgbz, err := json.Marshal(netmsg)
	if err != nil {
		c.Logger.Error("cannot marshal P2PMessage", "error", err.Error())
		return
	}

	c.Logger.Debug("p2p received message", "sender", senderAddress, "contract", contractAddress, "topic", netmsg.RoomId, "frompeer", netmsg.Sender.Ip)

	// handle custom messages
	p2pctx, err := GetP2PContext(c.Context.GoContextParent)
	if err == nil {
		handler := p2pctx.GetCustomHandler(contractAddress)
		if handler != nil {
			handler(netmsg, contractAddress, senderAddress)
			return
		}
	}

	// TODO roles should be bech32 compatible
	contractAddressPrefixed, err := c.Context.CosmosHandler.GetAddressOrRole(c.Context.Ctx, contractAddress)
	if err == nil {
		contractAddress = contractAddressPrefixed.String()
	} else {
		// if the message is from another chain, contract address may be without prefix
		// and sender address will have the prefix of the other chain
		contractBz, _, err := mcodec.GetFromBech32Unsafe(contractAddress)
		if err != nil {
			contractAddressPrefixed := c.Context.CosmosHandler.AccBech32Codec().BytesToAccAddressPrefixed(contractBz)
			contractAddress = contractAddressPrefixed.String()
		}
	}

	senderAddressPrefixed, err := c.Context.CosmosHandler.GetAddressOrRole(c.Context.Ctx, senderAddress)
	if err == nil {
		senderAddress = senderAddressPrefixed.String()
	} else {
		_, senderPrefix, err := mcodec.GetFromBech32Unsafe(senderAddress)
		ourprefix := c.Context.CosmosHandler.AccBech32Codec().Prefix()
		if senderPrefix != ourprefix {
			c.Logger.Info("p2p received message from address with different prefix", "prefix", senderPrefix, "sender", senderAddress, "ourprefix", ourprefix, "err", err, "msg", string(netmsgbz))
			senderAddress, _ = networktypes.CrossChainAddress(senderAddress, ourprefix)
		}
	}

	msgtosend := &networktypes.MsgP2PReceiveMessageRequest{
		Sender:   senderAddress,
		Contract: contractAddress,
		Data:     netmsgbz,
	}
	_, _, err = c.Context.CosmosHandler.ExecuteCosmosMsg(msgtosend)
	if err != nil {
		c.Logger.Error(err.Error())
	}
}

func readDataStd(goContextParent context.Context, logger log.Logger, rw *bufio.ReadWriter, protocolID string, frompeer string, handleMessage func(msg []byte, frompeer string)) {
	logger.Debug("reading stream data from peer", "peer", frompeer)
out:
	for {
		msgbz, err := rw.ReadBytes('\n')
		if err != nil {
			if err.Error() != ERROR_STREAM_RESET {
				logger.Error("Error reading from buffer", "error", err.Error(), "peer", frompeer)
			}
			p2pctx, err := GetP2PContext(goContextParent)
			if err == nil {
				p2pctx.DeletePeer(protocolID, frompeer)
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
		case <-goContextParent.Done():
			logger.Debug("stopping peer libp2p connection", "peer", frompeer)
			break out
		default:
			// nothing
		}
	}
}
