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

type KayrosErrorResponse struct {
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

type KayrosRecordResponse struct {
	Record KayrosRecord `json:"record"`
}

type KayrosRecordsResponse struct {
	Records []KayrosRecord `json:"records"`
	Count   int            `json:"count"`
	Error   string         `json:"error,omitempty"`
	Message string         `json:"message,omitempty"`
}

// KayrosRegistrationRequest is the request body for POST /api/grpc/single-hash
type KayrosRegistrationRequest struct {
	DataType wasmx.HexString `json:"data_type"`
	DataItem wasmx.HexString `json:"data_item"`
}

// KayrosRegistrationResponse is the response from the registration API
type KayrosRegistrationResponse struct {
	Success  bool   `json:"success"`
	Hash     string `json:"hash,omitempty"`
	TimeUUID string `json:"timeuuid,omitempty"`
	Encoding string `json:"encoding,omitempty"`
	Error    string `json:"error,omitempty"`
	Message  string `json:"message,omitempty"`
}

type LevelHashEntry struct {
	Position int    `json:"position"`
	HashHex  string `json:"hash_hex"`
	UuidHex  string `json:"uuid_hex"`
}

type ProofPathData struct {
	DataType    string   `json:"data_type"`
	HashItem    string   `json:"hash_item"`
	Proof       []string `json:"proof"`
	Root        string   `json:"root"`
	Position    int64    `json:"position"`
	Levels      int32    `json:"levels"`
	LevelCounts []int32  `json:"level_counts"`
	LevelStarts []int64  `json:"level_starts"`
}

type ProofPathResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
	ProofPathData
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
