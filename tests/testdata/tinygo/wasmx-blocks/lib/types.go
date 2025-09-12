package lib

import wasmx "github.com/loredanacirstea/wasmx-env/lib"

const MODULE_NAME = "blocks"

type BlockEntry struct {
	Index           int64              `json:"index"`
	ReaderContract  []byte             `json:"readerContract"`
	WriterContract  []byte             `json:"writerContract"`
	Data            []byte             `json:"data"` // RequestProcessProposalWithMetaInfo
	ProposerAddress wasmx.Bech32String `json:"proposer_address"`
	Header          []byte             `json:"header"`         // Block Header
	LastCommit      []byte             `json:"last_commit"`    // BlockCommit
	Evidence        []byte             `json:"evidence"`       // EvidenceData
	Result          []byte             `json:"result"`         // ResponseFinalizeBlock
	ValidatorInfo   []byte             `json:"validator_info"` // cometbfttypes.Validator
}

// IndexedTransaction represents a tx location in a block
type IndexedTransaction struct {
	Height int64  `json:"height"`
	Index  uint32 `json:"index"`
}

// ConsensusParamsInfo stores params at a specific height
type ConsensusParamsInfo struct {
	Height            int64  `json:"height"`
	LastHeightChanged int64  `json:"last_height_changed"`
	Params            []byte `json:"params"` // JSON-encoded consensus params
}
