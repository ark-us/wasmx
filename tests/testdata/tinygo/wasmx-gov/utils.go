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

func LoggerInfo(msg string, parts []string)          { wasmx.LoggerInfo("gov: "+msg, parts) }
func LoggerError(msg string, parts []string)         { wasmx.LoggerError("gov: "+msg, parts) }
func LoggerDebug(msg string, parts []string)         { wasmx.LoggerDebug("gov: "+msg, parts) }
func LoggerDebugExtended(msg string, parts []string) { wasmx.LoggerDebugExtended("gov: "+msg, parts) }

func Revert(message string) { wasmx.Revert([]byte(message)) }

// decimalCount returns the count of non-trailing decimal digits in string like "0.5000"
func decimalCount(val string) int {
	dot := -1
	for i := 0; i < len(val); i++ {
		if val[i] == '.' {
			dot = i
			break
		}
	}
	if dot < 0 {
		return 0
	}
	dec := val[dot+1:]
	// strip trailing zeros
	j := len(dec)
	for j > 0 && dec[j-1] == '0' {
		j--
	}
	if j == 0 {
		return 0
	}
	return j
}

// strToScaledInt converts a decimal string to an integer scaled by 10^dec
func strToScaledInt(val string, dec int) Big {
	// e.g., val="0.5", dec=1 -> 5
	if dec == 0 {
		// integer-like
		return NewBigFromString(val)
	}
	// Remove dot, keep dec digits
	out := make([]byte, 0, len(val))
	seen := false
	after := 0
	for i := 0; i < len(val); i++ {
		c := val[i]
		if c == '.' {
			seen = true
			continue
		}
		out = append(out, c)
		if seen {
			after++
		}
	}
	// pad with zeros if fewer decimals than dec
	for after < dec {
		out = append(out, '0')
		after++
	}
	// trim to dec digits after dot originally
	if after > dec {
		out = out[:len(out)-(after-dec)]
	}
	// remove leading zeros
	k := 0
	for k < len(out)-1 && out[k] == '0' {
		k++
	}
	s := string(out[k:])
	if s == "" {
		s = "0"
	}
	return NewBigFromString(s)
}

// Helpers for bank/contract calls
const defaultGasLimit = int64(50_000_000)

func callBank(calldata string, isQuery bool) ResponseBz {
	var ok bool
	var data []byte
	if isQuery {
		ok, data = wasmx.CallStatic(wasmx.Bech32String("bank"), []byte(calldata), big.NewInt(defaultGasLimit))
	} else {
		ok, data = wasmx.Call(wasmx.Bech32String("bank"), nil, []byte(calldata), big.NewInt(defaultGasLimit))
	}
	return ResponseBz{Success: ok, Data: data}
}

func callContract(addr Bech32String, calldata string, isQuery bool) ResponseBz {
	var ok bool
	var data []byte
	if isQuery {
		ok, data = wasmx.CallStatic(wasmx.Bech32String(addr), []byte(calldata), big.NewInt(defaultGasLimit))
	} else {
		ok, data = wasmx.Call(wasmx.Bech32String(addr), nil, []byte(calldata), big.NewInt(defaultGasLimit))
	}
	return ResponseBz{Success: ok, Data: data}
}

// Bank/Stake helpers
func bankSendCoinFromAccountToModule(from Bech32String, to Bech32String, coins []Coin) {
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
	resp := callBank(string(bz), false)
	if !resp.Success {
		Revert("could not transfer coins by bank")
	}
}

func getTokenAddress(denom string) Bech32String {
	// {"GetAddressByDenom":{"denom":"..."}}
	payload := struct {
		Req struct {
			Denom string `json:"denom"`
		} `json:"GetAddressByDenom"`
	}{}
	payload.Req.Denom = denom
	bz, _ := json.Marshal(&payload)
	resp := callBank(string(bz), true)
	if !resp.Success {
		Revert("could not get staking token address")
	}
	var out struct {
		Address string `json:"address"`
	}
	_ = json.Unmarshal(resp.Data, &out)
	if out.Address == "" {
		Revert("could not find staking token address: " + denom)
	}
	return Bech32String(out.Address)
}

func callGetStake(tokenAddress Bech32String, delegator Bech32String) Big {
	// {"balanceOf":{"owner":"..."}}
	payload := struct {
		Q struct {
			Owner string `json:"owner"`
		} `json:"balanceOf"`
	}{}
	payload.Q.Owner = string(delegator)
	bz, _ := json.Marshal(&payload)
	resp := callContract(tokenAddress, string(bz), true)
	if !resp.Success {
		Revert("delegation not found")
	}
	var out struct {
		Balance Coin `json:"balance"`
	}
	_ = json.Unmarshal(resp.Data, &out)
	return out.Balance.Amount
}

func callGetTotalStake() Big {
	params := getParams()
	tokenAddress := getTokenAddress(params.MinDeposit[0].Denom)
	payload := struct {
		Q struct{} `json:"totalSupply"`
	}{}
	bz, _ := json.Marshal(&payload)
	resp := callContract(tokenAddress, string(bz), true)
	if !resp.Success {
		Revert("delegation not found")
	}
	var out struct {
		Supply Coin `json:"supply"`
	}
	_ = json.Unmarshal(resp.Data, &out)
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
func getStake(voter Bech32String) Big {
	params := getParams()
	addr := getTokenAddress(params.MinDeposit[0].Denom)
	return callGetStake(addr, voter)
}
