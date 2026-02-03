package lib

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/sha3"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

type LevelProof struct {
	DataType string
	Level    int
	Position int
	Hashes   []string
}

// VerifyProof hashes data and verifies it exists in Kayros.
func VerifyProof(data []byte, dataType string, hashAlgo string, cfg KayrosConfig) (bool, string) {
	hashBytes, errMsg := hashBytes(data, hashAlgo)
	if errMsg != "" {
		return false, errMsg
	}
	return VerifyProofHash(dataType, hashBytes, cfg)
}

// VerifyProofHash checks if a Kayros record exists for the dataType and dataHash.
func VerifyProofHash(dataType string, dataHash []byte, cfg KayrosConfig) (bool, string) {
	client := NewKayrosClient(cfg)
	record, err := client.GetRecord(dataType, dataHash)
	if err != nil {
		return false, "get record failed: " + err.Error()
	}
	if record == nil {
		return false, "no record found"
	}

	if !dataTypeMatches(record, dataType) || !bytesEqual(record.DataItem, dataHash) {
		return false, fmt.Sprintf(
			"record mismatch data_type=%s record_data_type=%s data_hash=%s record_data_item=%s",
			dataType,
			record.DataType,
			hex.EncodeToString(dataHash),
			hex.EncodeToString(record.DataItem),
		)
	}
	return true, ""
}

// VerifyProofWithInclusion hashes data and verifies the record + proof path inclusion.
func VerifyProofWithInclusion(data []byte, dataType string, hashAlgo string, trustedRootHash string, trustedLevel int, trustedPosition int, cfg KayrosConfig) (bool, string) {
	hashBytes, errMsg := hashBytes(data, hashAlgo)
	if errMsg != "" {
		return false, errMsg
	}
	return VerifyProofHashWithInclusion(dataType, hashBytes, trustedRootHash, trustedLevel, trustedPosition, false, cfg)
}

// VerifyProofWithInclusionDetailed hashes data and verifies the record + proof path inclusion with proof metadata.
func VerifyProofWithInclusionDetailed(data []byte, dataType string, hashAlgo string, trustedRootHash string, trustedLevel int, trustedPosition int, cfg KayrosConfig) VerifyProofWithInclusionResponse {
	hashBytes, errMsg := hashBytes(data, hashAlgo)
	if errMsg != "" {
		return VerifyProofWithInclusionResponse{Ok: false, Error: errMsg, Pending: true, MaxLevel: -1, MaxLevelPosition: -1, MaxLevelHash: ""}
	}
	ok, errMsg, pending, maxLevel, maxLevelPosition, maxLevelHash := verifyProofHashWithInclusionDetailed(
		dataType,
		hashBytes,
		trustedRootHash,
		trustedLevel,
		trustedPosition,
		false,
		cfg,
	)
	return VerifyProofWithInclusionResponse{
		Ok:               ok,
		Error:            errMsg,
		Pending:          pending,
		MaxLevel:         maxLevel,
		MaxLevelPosition: maxLevelPosition,
		MaxLevelHash:     maxLevelHash,
	}
}

// VerifyProofHashWithInclusion verifies record details and proof path against a trusted root hash.
func VerifyProofHashWithInclusion(dataType string, dataHash []byte, trustedRootHash string, trustedLevel int, trustedPosition int, verifyDbExistence bool, cfg KayrosConfig) (bool, string) {
	ok, errMsg, _, _, _, _ := verifyProofHashWithInclusionDetailed(dataType, dataHash, trustedRootHash, trustedLevel, trustedPosition, verifyDbExistence, cfg)
	return ok, errMsg
}

// VerifyProofHashWithInclusionDetailed verifies record + proof inclusion and returns proof metadata.
func VerifyProofHashWithInclusionDetailed(dataType string, dataHash []byte, trustedRootHash string, trustedLevel int, trustedPosition int, verifyDbExistence bool, cfg KayrosConfig) VerifyProofWithInclusionResponse {
	ok, errMsg, pending, maxLevel, maxLevelPosition, maxLevelHash := verifyProofHashWithInclusionDetailed(
		dataType,
		dataHash,
		trustedRootHash,
		trustedLevel,
		trustedPosition,
		verifyDbExistence,
		cfg,
	)
	return VerifyProofWithInclusionResponse{
		Ok:               ok,
		Error:            errMsg,
		Pending:          pending,
		MaxLevel:         maxLevel,
		MaxLevelPosition: maxLevelPosition,
		MaxLevelHash:     maxLevelHash,
	}
}

func verifyProofHashWithInclusionDetailed(dataType string, dataHash []byte, trustedRootHash string, trustedLevel int, trustedPosition int, verifyDbExistence bool, cfg KayrosConfig) (bool, string, bool, int, int64, string) {
	client := NewKayrosClient(cfg)
	record, err := client.GetRecord(dataType, dataHash)
	if err != nil {
		return false, "get record failed: " + err.Error(), true, -1, -1, ""
	}
	if record == nil {
		return false, "no record found", true, -1, -1, ""
	}
	if !dataTypeMatches(record, dataType) || !bytesEqual(record.DataItem, dataHash) {
		return false, fmt.Sprintf(
			"record mismatch data_type=%s record_data_type=%s data_hash=%s record_data_item=%s",
			dataType,
			record.DataType,
			hex.EncodeToString(dataHash),
			hex.EncodeToString(record.DataItem),
		), true, -1, -1, ""
	}

	var prev *KayrosRecord
	if len(record.PrevHash) > 0 {
		prev, err = client.GetRecordByHash(record.DataType, hex.EncodeToString(record.PrevHash))
		if err != nil || prev == nil {
			return false, "previous record not found", true, -1, -1, ""
		}
	}

	hashAlgo := strings.ToLower(strings.TrimSpace(record.HashType))
	if hashAlgo == "" {
		return false, "missing record hash type", true, -1, -1, ""
	}
	if ok, errMsg := VerifyRecordBasics(record, prev); !ok {
		return false, errMsg, true, -1, -1, ""
	}

	proof, err := client.GetProofPath(record.DataType, hex.EncodeToString(record.HashItem))
	if err != nil || proof == nil {
		return false, "missing proof path", true, -1, -1, ""
	}
	if !bytesEqual(record.HashItem, proof.HashItem) {
		return false, fmt.Sprintf(
			"hash_item mismatch record=%s proof=%s",
			hex.EncodeToString(record.HashItem),
			hex.EncodeToString(proof.HashItem),
		), true, -1, -1, ""
	}
	if proof.DataType != "" && record.DataType != "" && !strings.EqualFold(proof.DataType, record.DataType) {
		return false, fmt.Sprintf("data_type mismatch record=%s proof=%s", record.DataType, proof.DataType), true, -1, -1, ""
	}
	pending, maxLevel, maxLevelPosition, maxLevelHash := proofInclusionMeta(proof)
	if !pending {
		if trustedRootHash == "" {
			return false, "missing trusted root hash", pending, maxLevel, maxLevelPosition, maxLevelHash
		}
		if !strings.EqualFold(proof.Root, trustedRootHash) {
			return false, fmt.Sprintf("root hash mismatch proof=%s trusted=%s", proof.Root, trustedRootHash), pending, maxLevel, maxLevelPosition, maxLevelHash
		}

		if trustedLevel >= 0 && trustedPosition >= 0 {
			entry, err := client.GetLevelHash(record.DataType, trustedLevel, trustedPosition)
			if err != nil || entry == nil {
				return false, "trusted level hash not found", pending, maxLevel, maxLevelPosition, maxLevelHash
			}
			if !strings.EqualFold(entry.HashHex, trustedRootHash) {
				return false, fmt.Sprintf("trusted level hash mismatch expected=%s got=%s", trustedRootHash, entry.HashHex), pending, maxLevel, maxLevelPosition, maxLevelHash
			}
		}
	}

	if verifyDbExistence {
		if ok, errMsg := VerifyProofHashesExistInDB(cfg, proof); !ok {
			return false, errMsg, pending, maxLevel, maxLevelPosition, maxLevelHash
		}
	}

	if !pending {
		if ok, errMsg := VerifyProofPath(proof, hashAlgo); !ok {
			return false, errMsg, pending, maxLevel, maxLevelPosition, maxLevelHash
		}
	}
	if ok, errMsg := VerifyProofTargetPosition(proof, hex.EncodeToString(record.HashItem)); !ok {
		return false, errMsg, pending, maxLevel, maxLevelPosition, maxLevelHash
	}
	return true, "", pending, maxLevel, maxLevelPosition, maxLevelHash
}

// VerifyProofHashesExistInDB checks that all proof hashes exist in the database at their levels.
func VerifyProofHashesExistInDB(cfg KayrosConfig, path *ProofPathData) (bool, string) {
	if path == nil {
		return false, "missing proof path"
	}
	if len(path.Proof) == 0 {
		return false, "empty proof path"
	}
	levelCounts, errMsg := normalizeLevelCounts(path.LevelCounts, int(path.Levels), len(path.Proof))
	if errMsg != "" {
		return false, errMsg
	}
	if len(path.LevelStarts) > 0 && len(path.LevelStarts) != len(levelCounts) {
		return false, "level starts length mismatch"
	}
	client := NewKayrosClient(cfg)
	offset := 0
	for levelIdx, count := range levelCounts {
		if count <= 0 {
			return false, "invalid level count"
		}
		if offset+count > len(path.Proof) {
			return false, "proof length mismatch"
		}
		start := int64(0)
		if len(path.LevelStarts) > levelIdx {
			start = path.LevelStarts[levelIdx]
		}
		hashes := path.Proof[offset : offset+count]
		if ok, errMsg := client.VerifyHashBatch(path.DataType, levelIdx, start, hashes); !ok {
			return false, fmt.Sprintf("db existence check failed level=%d: %s", levelIdx, errMsg)
		}
		offset += count
	}
	return true, ""
}

// VerifyRecordHash checks hash_item = hash(prev_hash || data_type || data_item).
func VerifyRecordHash(record *KayrosRecord) (bool, string) {
	if record == nil {
		return false, "record is nil"
	}
	hashAlgo := strings.ToLower(strings.TrimSpace(record.HashType))
	if hashAlgo == "" {
		return false, "missing record hash type"
	}
	computed, errMsg := computeRecordHash(record, hashAlgo)
	if errMsg != "" {
		return false, errMsg
	}
	computedHex := hex.EncodeToString(computed)
	if !strings.EqualFold(computedHex, hex.EncodeToString(record.HashItem)) {
		return false, fmt.Sprintf("hash mismatch computed=%s record=%s data_type=%s", computedHex, hex.EncodeToString(record.HashItem), record.DataType)
	}
	return true, ""
}

// VerifyRecordChainLink checks record.prev_hash equals prev.hash_item and data_type matches.
func VerifyRecordChainLink(record *KayrosRecord, prev *KayrosRecord) (bool, string) {
	if record == nil || prev == nil {
		return false, "missing record chain"
	}
	if !strings.EqualFold(record.DataType, prev.DataType) {
		return false, fmt.Sprintf("data_type mismatch record=%s prev=%s", record.DataType, prev.DataType)
	}
	if !bytesEqual(record.PrevHash, prev.HashItem) {
		return false, fmt.Sprintf(
			"prev_hash mismatch record_prev=%s prev_hash_item=%s",
			hex.EncodeToString(record.PrevHash),
			hex.EncodeToString(prev.HashItem),
		)
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
	_ = hashAlgo
	if strings.TrimSpace(proof.DataType) == "" {
		return false, "missing data_type"
	}
	if proof.Level < 1 {
		return false, "invalid level"
	}
	if proof.Position < 0 {
		return false, "invalid position"
	}
	if len(proof.Hashes) == 0 {
		return false, "missing level hashes"
	}

	rollupBytes, errMsg := hashHexConcat(proof.Hashes, "sha256")
	if errMsg != "" {
		return false, errMsg
	}

	client := NewKayrosClient(cfg)
	expected, err := client.GetLevelHash(proof.DataType, proof.Level, proof.Position)
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
	_ = hashAlgo
	if path == nil {
		return false, "missing proof path"
	}
	if len(path.Proof) == 0 {
		return false, "empty proof path"
	}

	rootHash := strings.ToLower(strings.TrimSpace(path.Root))
	if rootHash == "" {
		return false, "missing root hash"
	}

	levelCounts, errMsg := normalizeLevelCounts(path.LevelCounts, int(path.Levels), len(path.Proof))
	if errMsg != "" {
		return false, errMsg
	}
	if len(path.LevelStarts) > 0 && len(path.LevelStarts) != len(levelCounts) {
		return false, "level starts length mismatch"
	}

	offset := 0
	var lastRollup string
	var prevRollup string
	currentPos := path.Position
	for levelIdx, count := range levelCounts {
		if count <= 0 {
			return false, "invalid level count"
		}
		if offset+count > len(path.Proof) {
			return false, "proof length mismatch"
		}
		levelHashes := path.Proof[offset : offset+count]
		if prevRollup != "" {
			index, errMsg := levelIndexForPosition(levelIdx, currentPos, count, path.LevelStarts)
			if errMsg != "" {
				return false, errMsg
			}
			if !strings.EqualFold(levelHashes[index], prevRollup) {
				foundAt := -1
				for i, h := range levelHashes {
					if strings.EqualFold(h, prevRollup) {
						foundAt = i
						break
					}
				}
				if foundAt >= 0 {
					return false, fmt.Sprintf(
						"level hash mismatch level=%d index=%d expected=%s got=%s found_at=%d",
						levelIdx,
						index,
						prevRollup,
						levelHashes[index],
						foundAt,
					)
				}
				return false, fmt.Sprintf(
					"level hash mismatch level=%d index=%d expected=%s got=%s",
					levelIdx,
					index,
					prevRollup,
					levelHashes[index],
				)
			}
		}

		isLast := levelIdx == len(levelCounts)-1
		if isLast && count == 1 {
			lastRollup = strings.ToLower(strings.TrimSpace(levelHashes[0]))
		} else {
			rollup, errMsg := hashHexConcat(levelHashes, "sha256")
			if errMsg != "" {
				return false, errMsg
			}
			prevRollup = hex.EncodeToString(rollup)
			if isLast {
				lastRollup = prevRollup
			}
		}
		offset += count
		currentPos = currentPos / 256
	}

	if lastRollup == "" {
		return false, "missing final hash"
	}
	if !strings.EqualFold(lastRollup, rootHash) {
		return false, fmt.Sprintf("root hash mismatch computed=%s root=%s", lastRollup, rootHash)
	}
	return true, ""
}

// VerifyRecordBasics validates record hash, chain link (when prev provided), and UUID.
func VerifyRecordBasics(record *KayrosRecord, prev *KayrosRecord) (bool, string) {
	if ok, errMsg := VerifyRecordHash(record); !ok {
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
	return true, ""
}

func computeRecordHash(record *KayrosRecord, hashAlgo string) ([]byte, string) {
	if record == nil {
		return nil, "record is nil"
	}
	prevBytes := record.PrevHash
	dataTypeBytes := []byte(record.DataType)
	dataItemBytes := record.DataItem
	uuidBytes, err := decodeHex(record.UuidHex)
	if err != nil {
		return nil, "invalid uuid"
	}
	if len(uuidBytes) != 16 {
		return nil, "invalid uuid length"
	}

	payload := append(append(append(prevBytes, dataTypeBytes...), dataItemBytes...), uuidBytes...)
	return hashBytes(payload, hashAlgo)
}

func stringTo32Bytes(s string) []byte {
	b := make([]byte, 32)
	data := []byte(s)
	if len(data) > 32 {
		copy(b, data[:32])
	} else {
		copy(b, data)
	}
	return b
}

// VerifyProofTargetPosition checks that the target hash appears at its level-0 position.
func VerifyProofTargetPosition(path *ProofPathData, targetHashHex string) (bool, string) {
	if path == nil {
		return false, "missing proof path"
	}
	if len(path.Proof) == 0 {
		return false, "empty proof path"
	}
	targetHashHex = strings.ToLower(strings.TrimSpace(targetHashHex))
	if targetHashHex == "" {
		return false, "missing target hash"
	}

	levelCounts, errMsg := normalizeLevelCounts(path.LevelCounts, int(path.Levels), len(path.Proof))
	if errMsg != "" {
		return false, errMsg
	}
	if len(levelCounts) == 0 || levelCounts[0] <= 0 {
		return false, "invalid level count"
	}
	if path.Position < 0 {
		return false, "invalid position"
	}
	count0 := levelCounts[0]
	if len(path.LevelStarts) > 0 && len(path.LevelStarts) != len(levelCounts) {
		return false, "level starts length mismatch"
	}
	index, errMsg := levelIndexForPosition(0, path.Position, count0, path.LevelStarts)
	if errMsg != "" {
		return false, errMsg
	}
	if !strings.EqualFold(path.Proof[index], targetHashHex) {
		return false, fmt.Sprintf(
			"target hash not found at expected position index=%d expected=%s got=%s",
			index,
			targetHashHex,
			path.Proof[index],
		)
	}
	return true, ""
}

func levelIndexForPosition(levelIdx int, currentPos int64, count int, levelStarts []int64) (int, string) {
	if count <= 0 {
		return 0, "invalid level count"
	}
	var start int64
	if len(levelStarts) > levelIdx {
		start = levelStarts[levelIdx]
	} else {
		start = (currentPos / int64(count)) * int64(count)
	}
	idx := currentPos - start
	if idx < 0 || idx >= int64(count) {
		return 0, "proof index out of range"
	}
	return int(idx), ""
}

// VerifyKayrosRecordWithProof verifies the record and its proof path.
func VerifyKayrosRecordWithProof(record *KayrosRecord, prev *KayrosRecord, proof *ProofPathData, hashAlgo string) (bool, string) {
	if ok, errMsg := VerifyKayrosRecord(record, prev, hashAlgo, nil, KayrosConfig{}); !ok {
		return false, errMsg
	}
	if proof == nil {
		return false, "missing proof path"
	}
	if !bytesEqual(record.HashItem, proof.HashItem) {
		return false, fmt.Sprintf(
			"hash_item mismatch record=%s proof=%s",
			hex.EncodeToString(record.HashItem),
			hex.EncodeToString(proof.HashItem),
		)
	}
	if proof.DataType != "" && record.DataType != "" && !strings.EqualFold(proof.DataType, record.DataType) {
		return false, fmt.Sprintf("data_type mismatch record=%s proof=%s", record.DataType, proof.DataType)
	}
	if ok, errMsg := VerifyProofPath(proof, hashAlgo); !ok {
		return false, errMsg
	}
	return true, ""
}

// VerifyKayrosRecord runs hash, chain, uuid, and level checks.
func VerifyKayrosRecord(record *KayrosRecord, prev *KayrosRecord, hashAlgo string, levelProofs []LevelProof, cfg KayrosConfig) (bool, string) {
	if ok, errMsg := VerifyRecordBasics(record, prev); !ok {
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
	switch hashAlgo {
	case "", "sha256":
		return wasmx.Sha256(data), ""
	case "sha3_256", "sha3-256":
		sum := sha3.Sum256(data)
		return sum[:], ""
	case "keccak256":
		return wasmx.Keccak256(data), ""
	default:
		return nil, "unsupported hash algorithm: " + hashAlgo
	}
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

func normalizeLevelCounts(counts []int32, levels int, proofLen int) ([]int, string) {
	if proofLen <= 0 {
		return nil, "empty proof path"
	}
	if len(counts) > 0 {
		out := make([]int, len(counts))
		total := 0
		for i, c := range counts {
			if c <= 0 {
				return nil, "invalid level count"
			}
			out[i] = int(c)
			total += out[i]
		}
		if total != proofLen {
			return nil, "proof length mismatch"
		}
		return out, ""
	}

	if levels <= 0 {
		return []int{proofLen}, ""
	}
	if levels == 1 {
		return []int{proofLen}, ""
	}
	remaining := proofLen - 256*(levels-1)
	if remaining <= 0 {
		return nil, "proof length mismatch"
	}
	out := make([]int, levels)
	for i := 0; i < levels-1; i++ {
		out[i] = 256
	}
	out[levels-1] = remaining
	return out, ""
}

func proofInclusionMeta(path *ProofPathData) (bool, int, int64, string) {
	if path == nil || len(path.LevelCounts) == 0 {
		return true, -1, -1, ""
	}
	levelCounts, errMsg := normalizeLevelCounts(path.LevelCounts, int(path.Levels), len(path.Proof))
	if errMsg != "" || len(levelCounts) == 0 {
		return true, -1, -1, ""
	}
	positionPath := make([]int64, 0, len(levelCounts))
	currentPos := path.Position
	positionPath = append(positionPath, currentPos)
	for i := 0; i < len(levelCounts)-1; i++ {
		currentPos = currentPos / 256
		positionPath = append(positionPath, currentPos)
	}
	maxLevel := len(levelCounts) - 1
	maxLevelPosition := positionPath[maxLevel]
	maxLevelHash := ""
	if strings.TrimSpace(path.Root) != "" {
		maxLevelHash = strings.ToLower(strings.TrimSpace(path.Root))
	} else {
		levelHashes := proofLevelHashes(path.Proof, levelCounts, maxLevel)
		levelStart := int64(0)
		if len(path.LevelStarts) > maxLevel {
			levelStart = path.LevelStarts[maxLevel]
		}
		index := int(maxLevelPosition - levelStart)
		if index >= 0 && index < len(levelHashes) {
			maxLevelHash = strings.ToLower(strings.TrimSpace(levelHashes[index]))
		}
	}

	pending := false
	if len(levelCounts) < 2 {
		pending = true
	} else {
		level1Start := int64(0)
		if len(path.LevelStarts) > 1 {
			level1Start = path.LevelStarts[1]
		}
		level1Index := int(positionPath[1] - level1Start)
		if level1Index < 0 || level1Index >= levelCounts[1] {
			pending = true
		}
	}

	return pending, maxLevel, maxLevelPosition, maxLevelHash
}

func proofLevelHashes(all []string, levelCounts []int, level int) []string {
	if level < 0 || level >= len(levelCounts) {
		return nil
	}
	offset := 0
	for i := 0; i < level; i++ {
		offset += levelCounts[i]
	}
	count := levelCounts[level]
	if offset+count > len(all) || count <= 0 {
		return nil
	}
	return all[offset : offset+count]
}

func dataTypeMatches(record *KayrosRecord, dataType string) bool {
	if record == nil {
		return false
	}
	if record.DataType != "" && strings.EqualFold(record.DataType, dataType) {
		return true
	}
	return record.DataTypeHex != "" && strings.EqualFold(record.DataTypeHex, hex.EncodeToString([]byte(dataType)))
}

func bytesEqual(left []byte, right []byte) bool {
	return bytes.Equal(left, right)
}
