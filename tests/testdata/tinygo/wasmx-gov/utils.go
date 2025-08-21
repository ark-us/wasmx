package main

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strconv"

	wasmx "github.com/loredanacirstea/wasmx-env"
)

const MaxMetadataLen = 255

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func itoa(i int) string      { return strconv.FormatInt(int64(i), 10) }
func u64toa(i uint64) string { return strconv.FormatUint(i, 10) }

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
	wasmx.Revert([]byte(message))
}

// parseDecimalToBig converts a decimal string to a scaled Big integer using Go's superior decimal parsing
// This replaces the manual string parsing that was necessary in AssemblyScript
func parseDecimalToBig(val string, scale int) Big {
	// Use Go's big.Rat for accurate decimal parsing
	rat := new(big.Rat)
	if _, ok := rat.SetString(val); !ok {
		// fallback to zero for invalid strings
		return NewBigZero()
	}
	
	// Scale by 10^scale to convert to integer
	scaler := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil))
	scaled := new(big.Rat).Mul(rat, scaler)
	
	// Convert to integer (truncating any remaining fractional part)
	result := new(big.Int).Div(scaled.Num(), scaled.Denom())
	return Big{Int: result}
}

// Helpers for bank/contract calls
const defaultGasLimit = int64(50_000_000)

func callBank(calldata string, isQuery bool) (bool, []byte) {
	return wasmx.CallInternal(wasmx.Bech32String(wasmx.ROLE_BANK), nil, []byte(calldata), big.NewInt(defaultGasLimit), isQuery, MODULE_NAME)
}

// Bank/Stake helpers
func bankSendCoinFromAccountToModule(from wasmx.Bech32String, to wasmx.Bech32String, coins []Coin) {
	// {"SendCoinsFromAccountToModule": { ... banktypes.MsgSend ... }}
	// We only need the envelope; the host will route the message
	// Construct minimal MsgSend: {"from_address":"...","to_address":"...","amount":[{"denom":"...","amount":"..."}]}
	payload := struct {
		Send struct {
			From   string `json:"from_address"`
			To     string `json:"to_address"`
			Amount []Coin `json:"amount"`
		} `json:"SendCoinsFromAccountToModule"`
	}{}
	payload.Send.From = string(from)
	payload.Send.To = string(to)
	payload.Send.Amount = coins
	bz, _ := json.Marshal(&payload)
	ok, _ := callBank(string(bz), false)
	if !ok {
		Revert("could not transfer coins by bank")
	}
}

func getTokenAddress(denom string) wasmx.Bech32String {
	// {"GetAddressByDenom":{"denom":"..."}}
	payload := struct {
		Req struct {
			Denom string `json:"denom"`
		} `json:"GetAddressByDenom"`
	}{}
	payload.Req.Denom = denom
	bz, _ := json.Marshal(&payload)
	ok, resp := callBank(string(bz), true)
	if !ok {
		Revert("could not get staking token address")
	}
	var out struct {
		Address string `json:"address"`
	}
	_ = json.Unmarshal(resp, &out)
	if out.Address == "" {
		Revert("could not find staking token address: " + denom)
	}
	return wasmx.Bech32String(out.Address)
}

func callGetStake(tokenAddress wasmx.Bech32String, delegator wasmx.Bech32String) Big {
	// {"balanceOf":{"owner":"..."}}
	payload := struct {
		Q struct {
			Owner string `json:"owner"`
		} `json:"balanceOf"`
	}{}
	payload.Q.Owner = string(delegator)
	bz, _ := json.Marshal(&payload)
	ok, resp := wasmx.CallSimple(tokenAddress, bz, true, MODULE_NAME)
	if !ok {
		Revert("delegation not found")
	}
	var out struct {
		Balance Coin `json:"balance"`
	}
	_ = json.Unmarshal(resp, &out)
	return out.Balance.Amount
}

func callGetTotalStake() Big {
	params := getParams()
	tokenAddress := getTokenAddress(params.MinDeposit[0].Denom)
	payload := struct {
		Q struct{} `json:"totalSupply"`
	}{}
	bz, _ := json.Marshal(&payload)
	ok, resp := wasmx.CallSimple(tokenAddress, bz, true, MODULE_NAME)
	if !ok {
		Revert("delegation not found")
	}
	var out struct {
		Supply Coin `json:"supply"`
	}
	_ = json.Unmarshal(resp, &out)
	return out.Supply.Amount
}

// executeProposal runs cosmos messages included in proposal.Messages (base64-encoded json)
func executeProposal(p Proposal) Response {
	for _, m := range p.Messages {
		// message is base64-encoded JSON
		msgbz, err := base64.StdEncoding.DecodeString(m)
		if err != nil {
			return Response{Success: false, Data: "invalid message encoding"}
		}
		data := wasmx.ExecuteCosmosMsg(string(msgbz))
		if data.Success > 0 {
			return Response{Success: false, Data: data.Data}
		}
	}
	return Response{Success: true, Data: ""}
}

// Wrapper to match AS getStake signature
func getStake(voter wasmx.Bech32String) Big {
	params := getParams()
	addr := getTokenAddress(params.MinDeposit[0].Denom)
	return callGetStake(addr, voter)
}
