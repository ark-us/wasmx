package lib

//go:wasmimport p2p StartNodeWithIdentity
func StartNodeWithIdentity_(ptr int64) int64

//go:wasmimport p2p GetNodeInfo
func GetNodeInfo_() int64

//go:wasmimport p2p ConnectPeer
func ConnectPeer_(ptr int64) int64

//go:wasmimport p2p SendMessage
func SendMessage_(ptr int64) int64

//go:wasmimport p2p SendMessageToPeers
func SendMessageToPeers_(ptr int64) int64

//go:wasmimport p2p ConnectChatRoom
func ConnectChatRoom_(ptr int64) int64

//go:wasmimport p2p SendMessageToChatRoom
func SendMessageToChatRoom_(ptr int64) int64

//go:wasmimport p2p CloseNode
func CloseNode_() int64

//go:wasmimport p2p DisconnectChatRoom
func DisconnectChatRoom_(ptr int64) int64

//go:wasmimport p2p DisconnectPeer
func DisconnectPeer_(ptr int64) int64

//go:wasmimport p2p StartStateSyncRequest
func StartStateSyncRequest_(ptr int64) int64

//go:wasmimport p2p StartStateSyncResponse
func StartStateSyncResponse_(ptr int64) int64
