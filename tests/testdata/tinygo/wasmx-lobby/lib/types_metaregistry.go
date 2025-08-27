package lib

import (
	consensus "github.com/loredanacirstea/wasmx-env-consensus/lib"
)

const CROSS_CHAIN_TIMEOUT_MS = 120000 // 2 min

// MetaregistryParams represents metaregistry module parameters
type MetaregistryParams struct {
	CurrentLevel int32 `json:"current_level"`
}

// MetaregistryChainConfigData represents chain configuration data
type MetaregistryChainConfigData struct {
	Config  consensus.ChainConfig `json:"config"`
	ChainId consensus.ChainId     `json:"chain_id"`
}

// Message types
type MsgInitialize struct {
	Params MetaregistryParams `json:"params"`
}

type MsgSetChainDataRequest struct {
	Data MetaregistryChainConfigData `json:"data"`
}

type MsgSetChainDataResponse struct{}

// Query types
type QueryGetChainDataRequest struct {
	ChainId string `json:"chain_id"`
}

type QueryGetChainDataResponse struct {
	Data MetaregistryChainConfigData `json:"data"`
}

type QueryGetSubChainRequest struct {
	ChainId string `json:"chainId"`
}

type QueryGetSubChainsByIdsRequest struct {
	Ids []string `json:"ids"`
}

type QuerySubChainConfigByIdsRequest struct {
	Ids []string `json:"ids"`
}
