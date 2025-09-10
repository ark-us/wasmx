package lib

const MODULE_NAME = "multichain-registry"
const VERSION = "0.0.1"

// Calldata structure
type CallData struct {
	GetParams *MsgGetParams `json:"GetParams"`

	InitGenesis *MsgInitGenesis `json:"InitGenesis"`
}

type MsgInitGenesis struct{}

type MsgGetParams struct{}
