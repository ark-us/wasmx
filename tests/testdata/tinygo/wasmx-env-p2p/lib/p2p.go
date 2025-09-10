package lib

//go:wasmimport p2p start_node_with_identity
func StartNodeWithIdentity_(ptr int64) int64

//go:wasmimport p2p get_node_info
func GetNodeInfo_() int64

//go:wasmimport p2p connect_peer
func ConnectPeer_(ptr int64) int64

//go:wasmimport p2p send_message
func SendMessage_(ptr int64) int64

//go:wasmimport p2p send_message_to_peers
func SendMessageToPeers_(ptr int64) int64

//go:wasmimport p2p connect_chat_room
func ConnectChatRoom_(ptr int64) int64

//go:wasmimport p2p send_message_to_chat_room
func SendMessageToChatRoom_(ptr int64) int64

//go:wasmimport p2p close_node
func CloseNode_() int64

//go:wasmimport p2p disconnect_chat_room
func DisconnectChatRoom_(ptr int64) int64

//go:wasmimport p2p disconnect_peer
func DisconnectPeer_(ptr int64) int64

//go:wasmimport p2p start_state_sync_request
func StartStateSyncRequest_(ptr int64) int64

//go:wasmimport p2p start_state_sync_response
func StartStateSyncResponse_(ptr int64) int64
