package lib

// CallData mirrors the storage contract API
type CallData struct {
	SetIndexedData              *CallDataSetIndexedData              `json:"setIndexedData,omitempty"`
	SetBlock                    *CallDataSetBlock                    `json:"setBlock,omitempty"`
	SetConsensusParams          *CallDataSetConsensusParams          `json:"setConsensusParams,omitempty"`
	SetIndexedTransactionByHash *CallDataSetIndexedTransactionByHash `json:"setIndexedTransactionByHash,omitempty"`
	BootstrapAfterStateSync     *CallDataBootstrap                   `json:"bootstrapAfterStateSync,omitempty"`

	GetIndexedData              *CallDataGetIndexedData              `json:"getIndexedData,omitempty"`
	GetLastBlockIndex           *CallDataGetLastBlockIndex           `json:"getLastBlockIndex,omitempty"`
	GetBlockByIndex             *CallDataGetBlockByIndex             `json:"getBlockByIndex,omitempty"`
	GetBlockByHash              *CallDataGetBlockByHash              `json:"getBlockByHash,omitempty"`
	GetIndexedTransactionByHash *CallDataGetIndexedTransactionByHash `json:"getIndexedTransactionByHash,omitempty"`
	GetConsensusParams          *CallDataGetConsensusParams          `json:"getConsensusParams,omitempty"`

	GetContextValue *CallDataGetContextValue `json:"getContextValue,omitempty"`
}

type CallDataGetContextValue struct {
	Key string `json:"key"`
}

type CallDataInstantiate struct {
	InitialBlockIndex int64 `json:"initialBlockIndex"`
}

type CallDataSetIndexedData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type CallDataGetIndexedData struct {
	Key string `json:"key"`
}

type CallDataGetLastBlockIndex struct{}

type CallDataGetBlockByIndex struct {
	Index int64 `json:"index"`
}

type CallDataGetBlockByHash struct {
	Hash []byte `json:"hash"`
}

type CallDataGetIndexedTransactionByHash struct {
	Hash []byte `json:"hash"`
}

type CallDataGetConsensusParams struct {
	Height int64 `json:"height"`
}

type IndexedTopic struct {
	Topic  string   `json:"topic"`
	Values []string `json:"values"`
}

type CallDataSetBlock struct {
	Value         string         `json:"value"`
	Hash          string         `json:"hash"`
	Txhashes      []string       `json:"txhashes"`
	IndexedTopics []IndexedTopic `json:"indexed_topics"`
}

type CallDataSetConsensusParams struct {
	Height int64  `json:"height"`
	Params []byte `json:"params"`
}

type CallDataSetIndexedTransactionByHash struct {
	Hash []byte             `json:"hash"`
	Data IndexedTransaction `json:"data"`
}

type LastBlockIndexResult struct {
	Index int64 `json:"index"`
}

type CallDataBootstrap struct {
	LastBlockHeight   int64  `json:"last_block_height"`
	LastHeightChanged int64  `json:"last_height_changed"`
	Params            []byte `json:"params"`
}
