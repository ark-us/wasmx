package gov

import (
	"encoding/json"
	"math/big"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

func LoggerInfo(msg string, parts []string) {
	wasmx.LoggerInfo(MODULE_NAME, msg, parts)
}

func LoggerError(msg string, parts []string) {
	wasmx.LoggerError(MODULE_NAME, msg, parts)
}

func LoggerDebug(msg string, parts []string) {
	wasmx.LoggerDebug(MODULE_NAME, msg, parts)
}

func LoggerDebugExtended(msg string, parts []string) {
	wasmx.LoggerDebugExtended(MODULE_NAME, msg, parts)
}

func Revert(message string) {
	wasmx.RevertWithModule(MODULE_NAME, message)
}

// Big integer utilities
func NewBigZero() *big.Int {
	return new(big.Int)
}

func NewBigFromString(s string) *big.Int {
	z := new(big.Int)
	z.SetString(s, 10)
	return z
}

func NewBigFromUint64(i uint64) *big.Int {
	return new(big.Int).SetUint64(i)
}

func NewBigPow10(exp int) *big.Int {
	result := new(big.Int)
	base := big.NewInt(10)
	exponent := big.NewInt(int64(exp))
	return result.Exp(base, exponent, nil)
}

// Utility constants
const MaxMetadataLen = 10000

type groupVoterPowerResponse struct {
	IsMember bool   `json:"is_member"`
	UserID   string `json:"user_id,omitempty"`
	Power    string `json:"power"`
	Denom    string `json:"denom,omitempty"`
	Token    string `json:"token,omitempty"`
}

func getGroupVoterPower(groupContract wasmx.Bech32String, groupID string, voter wasmx.Bech32String) (groupVoterPowerResponse, bool) {
	payload := struct {
		Query struct {
			GroupID string `json:"group_id"`
			Voter   string `json:"voter"`
		} `json:"query_get_voter_power"`
	}{}
	payload.Query.GroupID = groupID
	payload.Query.Voter = string(voter)
	bz, _ := json.Marshal(&payload)
	ok, resp := wasmx.CallSimple(groupContract, bz, true, MODULE_NAME)
	if !ok {
		return groupVoterPowerResponse{}, false
	}
	var out groupVoterPowerResponse
	if err := json.Unmarshal(resp, &out); err != nil {
		return groupVoterPowerResponse{}, false
	}
	return out, true
}
