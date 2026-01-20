package lib

import (
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const (
	MODULE_NAME = "kayros_verifier"
)

// KayrosConfig holds the configuration for Kayros API client
type KayrosConfig struct {
	ApiBaseUrl string `json:"api_base_url"`
	ApiUserKey string `json:"api_user_key"`
}

// KayrosRecord represents a record in the Kayros database
type KayrosRecord struct {
	DataType    string `json:"data_type"`
	DataTypeHex string `json:"data_type_hex"`
	DataItemHex string `json:"data_item_hex"`
	UuidHex     string `json:"uuid_hex"`
	HashItemHex string `json:"hash_item_hex"`
	PrevHashHex string `json:"prev_hash_hex,omitempty"`
	HashType    string `json:"hash_type"`
	Timestamp   string `json:"timestamp"`
}

// KayrosApiResponse is the common response structure from Kayros API
type KayrosApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// KayrosRecordResponse is the response for single record queries
type KayrosRecordResponse struct {
	KayrosApiResponse
	Data KayrosRecord `json:"data"`
}

// KayrosRecordsData is the data structure for multiple record queries
type KayrosRecordsData struct {
	Count   int            `json:"count"`
	Limit   int            `json:"limit"`
	Records []KayrosRecord `json:"records"`
}

// KayrosRecordsResponse is the response for multiple record queries
type KayrosRecordsResponse struct {
	KayrosApiResponse
	Data KayrosRecordsData `json:"data"`
}

// KayrosRegistrationRequest is the request body for POST /api/grpc/single-hash
type KayrosRegistrationRequest struct {
	DataType wasmx.HexString `json:"data_type"`
	DataItem wasmx.HexString `json:"data_item"`
}

// KayrosRegistrationResponse is the response from the registration API
type KayrosRegistrationResponse struct {
	Success      bool            `json:"success"`
	Message      string          `json:"message"`
	DataType     string          `json:"data_type"`
	DataItem     string          `json:"data_item"`
	ComputedHash wasmx.HexString `json:"computed_hash_hex"`
	TimeUUID     wasmx.HexString `json:"timeuuid_hex"`
	DataTypeHex  wasmx.HexString `json:"data_type_hex"`
	DataItemHex  wasmx.HexString `json:"data_item_hex"`
}

type KayrosRegistrationResponseWrap struct {
	KayrosApiResponse
	Data KayrosRegistrationResponse `json:"data"`
}

type LevelHashEntry struct {
	Position int    `json:"position"`
	HashHex  string `json:"hash_hex"`
	UuidHex  string `json:"uuid_hex"`
}

type LevelHashData struct {
	Level    int    `json:"level"`
	Position int    `json:"position"`
	HashHex  string `json:"hash_hex"`
}

type LevelHashResponse struct {
	KayrosApiResponse
	Data LevelHashData `json:"data"`
}

type LevelRangeData struct {
	Level   int              `json:"level"`
	Count   int              `json:"count"`
	Entries []LevelHashEntry `json:"entries"`
}

type LevelRangeResponse struct {
	KayrosApiResponse
	Data LevelRangeData `json:"data"`
}

type ProofPathEntry struct {
	Level         int      `json:"level"`
	Position      int      `json:"position"`
	HashHex       string   `json:"hash_hex"`
	IndexInBlock  int      `json:"index_in_block"`
	SiblingHashes []string `json:"sibling_hashes"`
	SiblingCount  int      `json:"sibling_count"`
}

type ProofPathData struct {
	HashItemHex string           `json:"hash_item_hex"`
	DataTypeHex string           `json:"data_type_hex"`
	DataItemHex string           `json:"data_item_hex"`
	RecordIndex int              `json:"record_index"`
	Proof       []ProofPathEntry `json:"proof"`
	RootHashHex string           `json:"root_hash_hex"`
}

type ProofPathResponse struct {
	KayrosApiResponse
	Data ProofPathData `json:"data"`
}

type VerifyResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type VerifyProofRequest struct {
	Data       []byte `json:"data"`
	DataType   []byte `json:"data_type"`
	HashAlgo   string `json:"hash_algo"`
	ApiBaseUrl string `json:"api_base_url"`
	ApiUserKey string `json:"api_user_key"`
}

type VerifyProofWithInclusionRequest struct {
	Data            []byte `json:"data"`
	DataType        []byte `json:"data_type"`
	HashAlgo        string `json:"hash_algo"`
	TrustedRootHash string `json:"trusted_root_hash"`
	TrustedLevel    int    `json:"trusted_level"`
	TrustedPosition int    `json:"trusted_position"`
	ApiBaseUrl      string `json:"api_base_url"`
	ApiUserKey      string `json:"api_user_key"`
}

type VerifyProofHashRequest struct {
	DataType   []byte `json:"data_type"`
	DataHash   []byte `json:"data_hash"`
	ApiBaseUrl string `json:"api_base_url"`
	ApiUserKey string `json:"api_user_key"`
}

type VerifyProofHashWithInclusionRequest struct {
	DataType        []byte `json:"data_type"`
	DataHash        []byte `json:"data_hash"`
	HashAlgo        string `json:"hash_algo"`
	TrustedRootHash string `json:"trusted_root_hash"`
	TrustedLevel    int    `json:"trusted_level"`
	TrustedPosition int    `json:"trusted_position"`
	ApiBaseUrl      string `json:"api_base_url"`
	ApiUserKey      string `json:"api_user_key"`
}

type VerifyRecordHashRequest struct {
	Record   KayrosRecord `json:"record"`
	HashAlgo string       `json:"hash_algo"`
}

type VerifyRecordChainLinkRequest struct {
	Record KayrosRecord `json:"record"`
	Prev   KayrosRecord `json:"prev"`
}

type VerifyRecordTimestampRequest struct {
	Record KayrosRecord `json:"record"`
}

type VerifyRecordUUIDRequest struct {
	Record KayrosRecord `json:"record"`
}

type VerifyLevelProofRequest struct {
	Proof      LevelProof `json:"proof"`
	HashAlgo   string     `json:"hash_algo"`
	ApiBaseUrl string     `json:"api_base_url"`
	ApiUserKey string     `json:"api_user_key"`
}

type VerifyProofPathRequest struct {
	Proof    ProofPathData `json:"proof"`
	HashAlgo string        `json:"hash_algo"`
}

type VerifyKayrosRecordRequest struct {
	Record      KayrosRecord `json:"record"`
	Prev        KayrosRecord `json:"prev"`
	HashAlgo    string       `json:"hash_algo"`
	LevelProofs []LevelProof `json:"level_proofs"`
	ApiBaseUrl  string       `json:"api_base_url"`
	ApiUserKey  string       `json:"api_user_key"`
}

type VerifyKayrosRecordWithProofRequest struct {
	Record   KayrosRecord  `json:"record"`
	Prev     KayrosRecord  `json:"prev"`
	Proof    ProofPathData `json:"proof"`
	HashAlgo string        `json:"hash_algo"`
}

type Calldata struct {
	VerifyProof                  *VerifyProofRequest                  `json:"verify_proof,omitempty"`
	VerifyProofWithInclusion     *VerifyProofWithInclusionRequest     `json:"verify_proof_with_inclusion,omitempty"`
	VerifyProofHash              *VerifyProofHashRequest              `json:"verify_proof_hash,omitempty"`
	VerifyProofHashWithInclusion *VerifyProofHashWithInclusionRequest `json:"verify_proof_hash_with_inclusion,omitempty"`
	VerifyRecordHash             *VerifyRecordHashRequest             `json:"verify_record_hash,omitempty"`
	VerifyRecordChainLink        *VerifyRecordChainLinkRequest        `json:"verify_record_chain_link,omitempty"`
	VerifyRecordTimestamp        *VerifyRecordTimestampRequest        `json:"verify_record_timestamp,omitempty"`
	VerifyRecordUUID             *VerifyRecordUUIDRequest             `json:"verify_record_uuid,omitempty"`
	VerifyLevelProof             *VerifyLevelProofRequest             `json:"verify_level_proof,omitempty"`
	VerifyProofPath              *VerifyProofPathRequest              `json:"verify_proof_path,omitempty"`
	VerifyKayrosRecord           *VerifyKayrosRecordRequest           `json:"verify_kayros_record,omitempty"`
	VerifyKayrosRecordWithProof  *VerifyKayrosRecordWithProofRequest  `json:"verify_kayros_record_with_proof,omitempty"`
}
