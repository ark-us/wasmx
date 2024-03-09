package vm

import (
	"encoding/json"

	"github.com/libp2p/go-libp2p/core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const ChatRoomBufSize = 128

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.
type ChatRoom struct {
	ctx   *Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string
}

// ChatMessage gets converted to/from JSON and sent in the body of pubsub messages.
type ChatMessage struct {
	Message    []byte
	SenderID   string
	SenderNick string
}

// JoinChatRoom tries to subscribe to the PubSub topic for the room name, returning
// a ChatRoom on success.
func JoinChatRoom(ctx *Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, roomName string) (*ChatRoom, error) {
	// join the pubsub topic
	topic, err := ps.Join(topicName(roomName))
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &ChatRoom{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		nick:     nickname,
		roomName: roomName,
	}

	// start reading messages from the subscription in a loop
	go readLoop(cr)
	return cr, nil
}

// Publish sends a message to the pubsub topic.
func (cr *ChatRoom) Publish(message []byte) error {
	m := ChatMessage{
		Message:    message,
		SenderID:   cr.self.String(),
		SenderNick: cr.nick,
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	ctx := cr.ctx.Context.GoContextParent
	return cr.topic.Publish(ctx, msgBytes)
}

func (cr *ChatRoom) ListPeers() []peer.ID {
	return cr.ps.ListPeers(topicName(cr.roomName))
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func readLoop(cr *ChatRoom) {
	logger := cr.ctx.Context.Ctx.Logger()
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
				delete(p2pctx.ChatRooms, cr.roomName)
			}
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == cr.self {
			continue
		}
		cm := new(ChatMessage)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			continue
		}
		cr.ctx.handleMessage(cm.Message)
	}
}

func topicName(roomName string) string {
	return "chat-room:" + roomName
}
