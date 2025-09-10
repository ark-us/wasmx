package lib

import (
    wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const MODULE_NAME = "blocks"

// BlockEntry mirrors the AS type with byte fields for base64 data
type BlockEntry struct {
    Index           int64               `json:"index"`
    ReaderContract  wasmx.Bech32String  `json:"readerContract"`
    WriterContract  wasmx.Bech32String  `json:"writerContract"`
    Data            []byte              `json:"data"`          // JSON-encoded RequestProcessProposalWithMetaInfo
    Header          []byte              `json:"header"`        // JSON-encoded Header
    ProposerAddress wasmx.Bech32String  `json:"proposer_address"`
    LastCommit      []byte              `json:"last_commit"`   // JSON-encoded BlockCommit
    Evidence        []byte              `json:"evidence"`      // JSON-encoded EvidenceData
    Result          string              `json:"result"`        // JSON-encoded ResponseFinalizeBlock (string for compatibility)
    ValidatorInfo   []byte              `json:"validator_info"` // JSON-encoded TendermintValidators
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

