package vmp2p

import (
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const ChatRoomBufSize = 128

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.
type ChatRoom struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages chan *ChatRoomMessage

	ctx   *Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	protocolID  string
	topicString string
	self        peer.ID
}

// JoinChatRoom tries to subscribe to the PubSub topic for the room name, returning
// a ChatRoom on success.
func JoinChatRoom(ctx *Context, ps *pubsub.PubSub, selfID peer.ID, protocolID string, topicString string) (*ChatRoom, error) {
	// join the pubsub topic
	topic, err := ps.Join(topicName(topicString))
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &ChatRoom{
		ctx:         ctx,
		ps:          ps,
		topic:       topic,
		sub:         sub,
		self:        selfID,
		protocolID:  protocolID,
		topicString: topicString,
		Messages:    make(chan *ChatRoomMessage, ChatRoomBufSize),
	}

	// start reading messages from the subscription in a loop
	go readLoop(cr)
	return cr, nil
}

// Publish sends a message to the pubsub topic.
func (cr *ChatRoom) Publish(message []byte) error {
	ctx := cr.ctx.Context.GoContextParent
	return cr.topic.Publish(ctx, message)
}

func (cr *ChatRoom) ListPeers() []peer.ID {
	return cr.ps.ListPeers(topicName(cr.topicString))
}

func (cr *ChatRoom) Unsubscribe() {
	cr.sub.Cancel()
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func readLoop(cr *ChatRoom) {
	logger := cr.ctx.Logger
	ctx := cr.ctx.Context.GoContextParent
	for {
		msg, err := cr.sub.Next(ctx)
		if err != nil {
			if err.Error() != ERROR_STREAM_RESET && err.Error() != ERROR_CTX_CANCELED {
				logger.Error("Error chat room message read ", "error", err.Error(), "topic", cr.topic)
			}
			// remove chat room; it will be reconnected when needed
			p2pctx, err := GetP2PContext(cr.ctx.Context)
			if err == nil {
				p2pctx.DeleteChatRoom(cr.protocolID, cr.topicString)
			}
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom.String() == cr.self.String() {
			continue
		}
		cm := &ChatRoomMessage{
			ContractMsg: msg.Data,
			RoomId:      cr.topicString,
			Sender:      NodeInfo{Id: msg.ReceivedFrom.String()},
			Timestamp:   time.Now(),
			ProtocolID:  cr.protocolID,
		}

		// send valid messages onto the Messages channel
		cr.Messages <- cm
	}
}

func topicName(topicString string) string {
	return "chat-room:" + topicString
}
