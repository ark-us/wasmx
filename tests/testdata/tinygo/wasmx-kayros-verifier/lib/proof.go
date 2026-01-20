package lib

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

type LevelProof struct {
	Level    int
	Position int
	Hashes   []string
}

// VerifyProof hashes data and verifies it exists in Kayros.
func VerifyProof(data []byte, dataType wasmx.HexString, hashAlgo string, cfg KayrosConfig) (bool, string) {
	hashAlgo = strings.ToLower(strings.TrimSpace(hashAlgo))
	if hashAlgo == "" || hashAlgo == "sha256" {
		return VerifyProofHash(dataType, wasmx.HexString(hex.EncodeToString(wasmx.Sha256(data))), cfg)
	}
	if hashAlgo == "keccak256" {
		return VerifyProofHash(dataType, wasmx.HexString(hex.EncodeToString(wasmx.Keccak256(data))), cfg)
	}
	return false, "unsupported hash algorithm"
}

// VerifyProofHash checks if a Kayros record exists for the dataType and dataHash.
func VerifyProofHash(dataType wasmx.HexString, dataHash wasmx.HexString, cfg KayrosConfig) (bool, string) {
	client := NewKayrosClient(cfg)
	record, err := client.GetRecord(dataType, dataHash)
	if err != nil || record == nil {
		return false, "no record found"
	}

	if record.DataTypeHex != string(dataType) || record.DataItemHex != string(dataHash) {
		return false, "record mismatch"
	}
	return true, ""
}

// VerifyProofWithInclusion hashes data and verifies the record + proof path inclusion.
func VerifyProofWithInclusion(data []byte, dataType wasmx.HexString, hashAlgo string, trustedRootHash string, trustedLevel int, trustedPosition int, cfg KayrosConfig) (bool, string) {
	hashAlgo = strings.ToLower(strings.TrimSpace(hashAlgo))
	if hashAlgo == "" || hashAlgo == "sha256" {
		return VerifyProofHashWithInclusion(dataType, wasmx.HexString(hex.EncodeToString(wasmx.Sha256(data))), hashAlgo, trustedRootHash, trustedLevel, trustedPosition, cfg)
	}
	if hashAlgo == "keccak256" {
		return VerifyProofHashWithInclusion(dataType, wasmx.HexString(hex.EncodeToString(wasmx.Keccak256(data))), hashAlgo, trustedRootHash, trustedLevel, trustedPosition, cfg)
	}
	return false, "unsupported hash algorithm"
}

// VerifyProofHashWithInclusion verifies record details and proof path against a trusted root hash.
func VerifyProofHashWithInclusion(dataType wasmx.HexString, dataHash wasmx.HexString, hashAlgo string, trustedRootHash string, trustedLevel int, trustedPosition int, cfg KayrosConfig) (bool, string) {
	client := NewKayrosClient(cfg)
	record, err := client.GetRecord(dataType, dataHash)
	if err != nil || record == nil {
		return false, "no record found"
	}
	if record.DataTypeHex != string(dataType) || record.DataItemHex != string(dataHash) {
		return false, "record mismatch"
	}

	var prev *KayrosRecord
	if strings.TrimSpace(record.PrevHashHex) != "" {
		prev, err = client.GetRecordByHash(record.PrevHashHex)
		if err != nil || prev == nil {
			return false, "previous record not found"
		}
	}

	if ok, errMsg := VerifyRecordHash(record, hashAlgo); !ok {
		return false, errMsg
	}
	if prev != nil {
		if ok, errMsg := VerifyRecordChainLink(record, prev); !ok {
			return false, errMsg
		}
	}
	if ok, errMsg := VerifyRecordUUID(record); !ok {
		return false, errMsg
	}

	proof, err := client.GetProofPath(record.HashItemHex)
	if err != nil || proof == nil {
		return false, "missing proof path"
	}
	if !strings.EqualFold(proof.HashItemHex, record.HashItemHex) {
		return false, "hash_item mismatch"
	}
	if !strings.EqualFold(proof.DataTypeHex, record.DataTypeHex) {
		return false, "data_type mismatch"
	}
	if !strings.EqualFold(proof.DataItemHex, record.DataItemHex) {
		return false, "data_item mismatch"
	}
	if trustedRootHash == "" {
		return false, "missing trusted root hash"
	}
	if !strings.EqualFold(proof.RootHashHex, trustedRootHash) {
		return false, "root hash mismatch"
	}

	if trustedLevel >= 0 && trustedPosition >= 0 {
		entry, err := client.GetLevelHash(trustedLevel, trustedPosition)
		if err != nil || entry == nil {
			return false, "trusted level hash not found"
		}
		if !strings.EqualFold(entry.HashHex, trustedRootHash) {
			return false, "trusted level hash mismatch"
		}
	}

	if ok, errMsg := VerifyProofPath(proof, hashAlgo); !ok {
		return false, errMsg
	}
	return true, ""
}

// VerifyRecordHash checks hash_item = hash(prev_hash || data_type || data_item).
func VerifyRecordHash(record *KayrosRecord, hashAlgo string) (bool, string) {
	if record == nil {
		return false, "record is nil"
	}
	if strings.TrimSpace(hashAlgo) == "" && strings.TrimSpace(record.HashType) != "" {
		hashAlgo = record.HashType
	}
	prevBytes, err := decodeHexOrEmpty(record.PrevHashHex)
	if err != nil {
		return false, "invalid prev_hash"
	}
	dataTypeBytes, err := decodeHex(record.DataTypeHex)
	if err != nil {
		return false, "invalid data_type"
	}
	dataItemBytes, err := decodeHex(record.DataItemHex)
	if err != nil {
		return false, "invalid data_item"
	}
	uuidBytes, err := decodeHex(record.UuidHex)
	if err != nil {
		return false, "invalid uuid"
	}
	if len(uuidBytes) != 16 {
		return false, "invalid uuid length"
	}

	payload := append(append(append(prevBytes, dataTypeBytes...), dataItemBytes...), uuidBytes...)
	computed, errMsg := hashBytes(payload, hashAlgo)
	if errMsg != "" {
		return false, errMsg
	}
	computedHex := hex.EncodeToString(computed)
	if !strings.EqualFold(computedHex, record.HashItemHex) {
		return false, fmt.Sprintf("hash mismatch computed=%s record=%s", computedHex, record.HashItemHex)
	}
	return true, ""
}

// VerifyRecordChainLink checks record.prev_hash equals prev.hash_item and data_type matches.
func VerifyRecordChainLink(record *KayrosRecord, prev *KayrosRecord) (bool, string) {
	if record == nil || prev == nil {
		return false, "missing record chain"
	}
	if !strings.EqualFold(record.DataTypeHex, prev.DataTypeHex) {
		return false, "data_type mismatch"
	}
	if !strings.EqualFold(record.PrevHashHex, prev.HashItemHex) {
		return false, "prev_hash mismatch"
	}
	return true, ""
}

// VerifyRecordTimestamp ensures timestamp is valid RFC3339Nano.
func VerifyRecordTimestamp(record *KayrosRecord) (bool, string) {
	if record == nil {
		return false, "record is nil"
	}
	if record.Timestamp == "" {
		return false, "missing timestamp"
	}
	if _, err := parseRecordTimestamp(record.Timestamp); err != nil {
		return false, "invalid timestamp: " + record.Timestamp
	}
	return true, ""
}

// VerifyRecordUUID ensures UUID timestamp matches record timestamp.
func VerifyRecordUUID(record *KayrosRecord) (bool, string) {
	if record == nil {
		return false, "record is nil"
	}
	recordTime, err := parseRecordTimestamp(record.Timestamp)
	if err != nil {
		return false, "invalid timestamp: " + record.Timestamp
	}
	ts, err := TimeuuidHexToTimestamp(record.UuidHex)
	if err != nil {
		return false, "invalid uuid"
	}
	uuidTime, err := parseRecordTimestamp(ts)
	if err != nil {
		return false, "invalid uuid timestamp"
	}
	recordTime = recordTime.UTC().Truncate(time.Millisecond)
	uuidTime = uuidTime.UTC().Truncate(time.Millisecond)
	if !recordTime.Equal(uuidTime) {
		return false, "uuid does not match timestamp"
	}
	return true, ""
}

func parseRecordTimestamp(value string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02 15:04:05.999 MST", value)
}

// VerifyLevelProof checks a single level rollup hash against Kayros.
func VerifyLevelProof(cfg KayrosConfig, proof LevelProof, hashAlgo string) (bool, string) {
	if proof.Level < 1 {
		return false, "invalid level"
	}
	if proof.Position < 0 {
		return false, "invalid position"
	}
	if len(proof.Hashes) == 0 {
		return false, "missing level hashes"
	}

	rollupBytes, errMsg := hashHexConcat(proof.Hashes, hashAlgo)
	if errMsg != "" {
		return false, errMsg
	}

	client := NewKayrosClient(cfg)
	expected, err := client.GetLevelHash(proof.Level, proof.Position)
	if err != nil || expected == nil {
		return false, "no level hash found"
	}
	if !strings.EqualFold(hex.EncodeToString(rollupBytes), expected.HashHex) {
		return false, "level hash mismatch"
	}
	return true, ""
}

// VerifyProofPath verifies the proof path against computed rollup hashes and root hash.
func VerifyProofPath(path *ProofPathData, hashAlgo string) (bool, string) {
	if path == nil {
		return false, "missing proof path"
	}
	if len(path.Proof) == 0 {
		return false, "empty proof path"
	}

	hashItem := strings.ToLower(strings.TrimSpace(path.HashItemHex))
	if hashItem == "" {
		return false, "missing hash_item"
	}
	rootHash := strings.ToLower(strings.TrimSpace(path.RootHashHex))
	if rootHash == "" {
		return false, "missing root hash"
	}

	for _, step := range path.Proof {
		if step.Position < 0 {
			return false, "invalid position"
		}
		if step.IndexInBlock < 0 || step.IndexInBlock >= 256 {
			return false, "invalid index_in_block"
		}
		stepHash := strings.ToLower(strings.TrimSpace(step.HashHex))
		if stepHash == "" {
			return false, "missing hash_hex"
		}
		expectedCount := step.SiblingCount
		if expectedCount == 0 {
			expectedCount = 256
		}
		if len(step.SiblingHashes) != expectedCount {
			return false, "sibling_count mismatch"
		}
		if len(step.SiblingHashes) == 0 {
			return false, "missing sibling hashes"
		}
		rollup, errMsg := hashHexConcat(step.SiblingHashes, hashAlgo)
		if errMsg != "" {
			return false, errMsg
		}
		if !strings.EqualFold(hex.EncodeToString(rollup), stepHash) {
			return false, "level rollup mismatch"
		}
	}

	last := strings.ToLower(strings.TrimSpace(path.Proof[len(path.Proof)-1].HashHex))
	if last == "" {
		return false, "missing final hash"
	}
	if !strings.EqualFold(last, rootHash) {
		return false, "root hash mismatch"
	}
	return true, ""
}

// VerifyKayrosRecordWithProof verifies the record and its proof path.
func VerifyKayrosRecordWithProof(record *KayrosRecord, prev *KayrosRecord, proof *ProofPathData, hashAlgo string) (bool, string) {
	if ok, errMsg := VerifyKayrosRecord(record, prev, hashAlgo, nil, KayrosConfig{}); !ok {
		return false, errMsg
	}
	if proof == nil {
		return false, "missing proof path"
	}
	if !strings.EqualFold(proof.HashItemHex, record.HashItemHex) {
		return false, "hash_item mismatch"
	}
	if !strings.EqualFold(proof.DataTypeHex, record.DataTypeHex) {
		return false, "data_type mismatch"
	}
	if !strings.EqualFold(proof.DataItemHex, record.DataItemHex) {
		return false, "data_item mismatch"
	}
	if ok, errMsg := VerifyProofPath(proof, hashAlgo); !ok {
		return false, errMsg
	}
	return true, ""
}

// VerifyKayrosRecord runs hash, chain, uuid, and level checks.
func VerifyKayrosRecord(record *KayrosRecord, prev *KayrosRecord, hashAlgo string, levelProofs []LevelProof, cfg KayrosConfig) (bool, string) {
	if ok, errMsg := VerifyRecordHash(record, hashAlgo); !ok {
		return false, errMsg
	}
	if ok, errMsg := VerifyRecordChainLink(record, prev); !ok {
		return false, errMsg
	}
	if ok, errMsg := VerifyRecordUUID(record); !ok {
		return false, errMsg
	}

	for _, proof := range levelProofs {
		if ok, errMsg := VerifyLevelProof(cfg, proof, hashAlgo); !ok {
			return false, errMsg
		}
	}
	return true, ""
}

func hashBytes(data []byte, hashAlgo string) ([]byte, string) {
	hashAlgo = strings.ToLower(strings.TrimSpace(hashAlgo))
	if hashAlgo == "" || hashAlgo == "sha256" {
		return wasmx.Sha256(data), ""
	}
	if hashAlgo == "keccak256" {
		return wasmx.Keccak256(data), ""
	}
	return nil, "unsupported hash algorithm"
}

func hashHexConcat(hashes []string, hashAlgo string) ([]byte, string) {
	payload := make([]byte, 0)
	for _, h := range hashes {
		bz, err := decodeHex(h)
		if err != nil {
			return nil, "invalid hash hex"
		}
		payload = append(payload, bz...)
	}
	return hashBytes(payload, hashAlgo)
}

func decodeHex(value string) ([]byte, error) {
	return hex.DecodeString(strings.TrimSpace(value))
}

func decodeHexOrEmpty(value string) ([]byte, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return []byte{}, nil
	}
	return hex.DecodeString(value)
}
