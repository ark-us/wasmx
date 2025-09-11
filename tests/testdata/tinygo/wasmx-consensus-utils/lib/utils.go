package lib

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	consensus "github.com/loredanacirstea/wasmx-env-consensus/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	staking "github.com/loredanacirstea/wasmx-staking/lib"
)

// MaxKeyLength and MaxValueLength mirror the AS constants
const (
	MaxKeyLength   = 131071     // 128K - 1
	MaxValueLength = 2147483647 // 2G - 1
)

// POWER_REDUCTION must match staking module semantics
const POWER_REDUCTION int64 = 1_000_000

// GetValidatorsHash: validators will be sorted by power (DESC) and then address (ASC)
func GetValidatorsHash(validators []staking.Validator) ([]byte, error) {
	info, err := GetActiveValidatorInfo(validators)
	if err != nil {
		return nil, err
	}
	return consensus.ValidatorsHash(info)
}

// IsValidatorInactive when jailed or not bonded
func IsValidatorInactive(v staking.Validator) bool {
	return v.Jailed || v.Status != staking.BondedS
}

// GetActiveValidatorInfo maps staking validators to Tendermint validators
func GetActiveValidatorInfo(validators []staking.Validator) ([]consensus.TendermintValidator, error) {
	out := make([]consensus.TendermintValidator, 0, len(validators))
	for _, v := range validators {
		fmt.Println("--GetActiveValidatorInfo.OperatorAddress--", v.OperatorAddress)
		if IsValidatorInactive(v) {
			continue
		}
		if v.ConsensusPubkey == nil {
			wasmx.RevertWithModule("consensus-utils", "validator missing consensus key "+string(v.OperatorAddress))
			return nil, nil
		}
		fmt.Println("--GetActiveValidatorInfo.ConsensusPubkey--", v.ConsensusPubkey.TypeUrl, v.ConsensusPubkey.Value)
		// hex address from pubkey bytes
		key := v.ConsensusPubkey.GetKey().Key
		addrhex := wasmx.Ed25519PubToHex(key)
		fmt.Println("--GetActiveValidatorInfo.addrhex--", addrhex, hex.EncodeToString(addrhex))
		pow := getPower(v.Tokens)
		out = append(out, consensus.TendermintValidator{
			OperatorAddress:  v.OperatorAddress,
			HexAddress:       wasmx.HexString(hex.EncodeToString(addrhex)),
			PubKey:           v.ConsensusPubkey,
			VotingPower:      pow,
			ProposerPriority: 0,
		})
	}
	return out, nil
}

func getPower(tokens sdkmath.Int) int64 {
	// divide by POWER_REDUCTION
	if tokens.IsNil() {
		return 0
	}
	q := tokens.QuoRaw(POWER_REDUCTION)
	return q.Int64()
}

// GetSortedBlockCommits reorders signatures by active validator order
func GetSortedBlockCommits(last consensus.BlockCommit, activeSorted []consensus.TendermintValidator) (consensus.BlockCommit, error) {
	sigmap := make(map[string]consensus.CommitSig, len(last.Signatures))
	for _, s := range last.Signatures {
		sigmap[string(s.ValidatorAddress)] = s
	}
	sigs := make([]consensus.CommitSig, len(activeSorted))
	for i, v := range activeSorted {
		s, ok := sigmap[string(v.HexAddress)]
		if !ok {
			return consensus.BlockCommit{}, wasmxError("sorted validator address not found: " + string(v.HexAddress) + " - " + string(v.OperatorAddress))
		}
		sigs[i] = s
	}
	return consensus.BlockCommit{Height: last.Height, Round: last.Round, BlockID: last.BlockID, Signatures: sigs}, nil
}

// CleanAbsentCommits zeros timestamps for Absent flags
func CleanAbsentCommits(last consensus.BlockCommit) consensus.BlockCommit {
	for i := range last.Signatures {
		if last.Signatures[i].BlockIDFlag == consensus.BlockIDFlagAbsent {
			last.Signatures[i].Timestamp = nil
		}
	}
	return last
}

// SortTendermintValidators by power desc, then hex address asc
func SortTendermintValidators(vals []consensus.TendermintValidator) []consensus.TendermintValidator {
	out := make([]consensus.TendermintValidator, len(vals))
	copy(out, vals)
	sort.Slice(out, func(i, j int) bool {
		if out[i].VotingPower != out[j].VotingPower {
			return out[i].VotingPower > out[j].VotingPower
		}
		return sortHexAddr(string(out[i].HexAddress), string(out[j].HexAddress)) < 0
	})
	return out
}

// FilterAndSortCommitSignatures keeps signatures from active validator set (order preserved)
func FilterAndSortCommitSignatures(sigs []consensus.CommitSig, infos []consensus.TendermintValidator) []consensus.CommitSig {
	active := make(map[string]struct{}, len(infos))
	for _, v := range infos {
		active[string(v.HexAddress)] = struct{}{}
	}
	out := make([]consensus.CommitSig, 0, len(sigs))
	for _, s := range sigs {
		if _, ok := active[string(s.ValidatorAddress)]; ok {
			out = append(out, s)
		}
	}
	return out
}

func sortHexAddr(hex1, hex2 string) int {
	// compare by bytes length and lexicographically
	b1 := []byte(hex1)
	b2 := []byte(hex2)
	if len(b1) < len(b2) {
		return -1
	}
	if len(b1) > len(b2) {
		return 1
	}
	if hex1 < hex2 {
		return -1
	}
	if hex1 > hex2 {
		return 1
	}
	return 0
}

// GetTxsHash computes merkle hash from tx slices (encoded as base64 strings)
func GetTxsHash(txs [][]byte) []byte {
	slices := make([]string, len(txs))
	for i := range txs {
		slices[i] = base64.StdEncoding.EncodeToString(txs[i])
	}
	return wasmx.MerkleHash(slices)
}

// GetConsensusParamsHash hashes consensus params
func GetConsensusParamsHash(params consensus.ConsensusParams) ([]byte, error) {
	return consensus.ConsensusParamsHash(params)
}

// GetEvidenceHash placeholder (empty merkle)
func GetEvidenceHash(_ consensus.Evidence) []byte { return wasmx.MerkleHash([]string{}) }

// GetCommitHash merkle hash of signatures
func GetCommitHash(last consensus.BlockCommit) []byte {
	values := make([]string, len(last.Signatures))
	for i := range last.Signatures {
		values[i] = base64.StdEncoding.EncodeToString(last.Signatures[i].Signature)
	}
	return wasmx.MerkleHash(values)
}

// GetResultsHash merkle hash of JSON-encoded ExecTxResult
func GetResultsHash(results []consensus.ExecTxResult) []byte {
	values := make([]string, len(results))
	for i := range results {
		bz, _ := json.Marshal(&results[i])
		values[i] = base64.StdEncoding.EncodeToString(bz)
	}
	return wasmx.MerkleHash(values)
}

// GetHeaderHash delegates to consensus wrapper
func GetHeaderHash(header consensus.Header) ([]byte, error) { return consensus.HeaderHash(header) }

// Event topics and indexing helpers
func GetEventTopic(eventName, eventAttribute, eventValue string) string {
	return eventName + "." + eventAttribute + "='" + eventValue + "'"
}

type IndexedTopic struct {
	Topic  string   `json:"topic"`
	Values []string `json:"values"`
}

func ExtractIndexedTopics(finalizeResp consensus.ResponseFinalizeBlock, txhashes [][]byte) []IndexedTopic {
	topicMap := map[string][]string{}
	push := func(topic string, txhash []byte) {
		if len(topic) > MaxKeyLength {
			return
		}
		arr := topicMap[topic]
		arr = append(arr, base64.StdEncoding.EncodeToString(txhash))
		topicMap[topic] = arr
	}
	for i := range finalizeResp.TxResults {
		res := finalizeResp.TxResults[i]
		for _, ev := range res.Events {
			for _, attr := range ev.Attributes {
				if attr.Index {
					topic := GetEventTopic(ev.Type, attr.Key, attr.Value)
					push(topic, txhashes[i])
				}
			}
			push(ev.Type, txhashes[i])
		}
	}
	out := make([]IndexedTopic, 0, len(topicMap))
	for k, v := range topicMap {
		out = append(out, IndexedTopic{Topic: k, Values: v})
	}
	return out
}

// Simple error helper
func wasmxError(msg string) error {
	wasmx.LoggerError("consensus-utils", msg, nil)
	return &utilsError{msg}
}

type utilsError struct{ s string }

func (e *utilsError) Error() string { return e.s }
