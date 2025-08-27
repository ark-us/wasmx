package lib

import (
	"encoding/json"
	"strconv"

	sdkmath "cosmossdk.io/math"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const (
	SPLIT                     = "."
	PARAMS_KEY                = "params"
	CHAIN_IDS                 = "chainids"
	CHAIN_VALIDATORS          = "chain_validators."
	CHAIN_VALIDATOR_ADDRESSES = "chain_validatoraddresses."
	DATA_KEY                  = "chain_data."
	VALIDATOR_CHAINS          = "validator_chains."
	LEVEL_LAST                = "level_last"
	LEVEL_CHAIN_IDS           = "level_chainids."
	CURRENT_LEVEL             = "currentlevel"
	LAST_CHAIN_ID             = "chainid_last"

	INITIAL_LEVEL int32 = 1
)

// Key builders
func GetDataKey(chainId string) string               { return DATA_KEY + chainId }
func GetValidatorsKey(chainId string) string         { return CHAIN_VALIDATORS + chainId }
func GetValidatorAddressesKey(chainId string) string { return CHAIN_VALIDATOR_ADDRESSES + chainId }
func GetValidatorChainIdsKey(validatorAddress string) string {
	return VALIDATOR_CHAINS + validatorAddress
}
func GetLevelChainIdsKey(levelIndex int32) string {
	return LEVEL_CHAIN_IDS + strconv.FormatInt(int64(levelIndex), 10)
}

// Validator chains per address
func GetValidatorChains(validatorAddress string) []string {
	value := wasmx.SLoad(GetValidatorChainIdsKey(validatorAddress))
	if value == "" {
		return []string{}
	}
	var out []string
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func AddValidatorChain(validatorAddress, chainId string) {
	chainIds := GetValidatorChains(validatorAddress)
	for _, id := range chainIds {
		if id == chainId {
			return
		}
	}
	chainIds = append(chainIds, chainId)
	SetValidatorChains(validatorAddress, chainIds)
}

func SetValidatorChains(validatorAddress string, chainIds []string) {
	bz, _ := json.Marshal(&chainIds)
	wasmx.SStore(GetValidatorChainIdsKey(validatorAddress), string(bz))
}

// Chain data
func GetChainData(chainId string) *SubChainData {
	value := wasmx.SLoad(GetDataKey(chainId))
	if value == "" {
		return nil
	}
	var out SubChainData
	if err := json.Unmarshal([]byte(value), &out); err != nil {
		Revert("cannot decode chain data: " + err.Error())
		return nil
	}
	return &out
}

func SetChainData(data SubChainData) {
	chainId := data.Data.InitChainRequest.ChainID
	bz, _ := json.Marshal(&data)
	wasmx.SStore(GetDataKey(chainId), string(bz))
}

// Validators and addresses
func AddChainValidator(chainId string, validatorAddress wasmx.Bech32String, genTx []byte) {
	AddChainValidatorAddress(chainId, validatorAddress)
	value := GetChainValidators(chainId)
	value = append(value, genTx)
	SetChainValidators(chainId, value)
}

func GetChainValidators(chainId string) [][]byte {
	value := wasmx.SLoad(GetValidatorsKey(chainId))
	if value == "" {
		return [][]byte{}
	}
	var out [][]byte
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func SetChainValidators(chainId string, genTxs [][]byte) {
	bz, _ := json.Marshal(&genTxs)
	wasmx.SStore(GetValidatorsKey(chainId), string(bz))
}

func AddChainValidatorAddress(chainId string, addr wasmx.Bech32String) {
	value := GetChainValidatorAddresses(chainId)
	for _, a := range value {
		if a == addr {
			Revert("validator address already included: " + string(addr))
		}
	}
	value = append(value, addr)
	SetChainValidatorAddresses(chainId, value)
}

func GetChainValidatorAddresses(chainId string) []wasmx.Bech32String {
	value := wasmx.SLoad(GetValidatorAddressesKey(chainId))
	if value == "" {
		return []wasmx.Bech32String{}
	}
	var out []wasmx.Bech32String
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func SetChainValidatorAddresses(chainId string, addrs []wasmx.Bech32String) {
	bz, _ := json.Marshal(&addrs)
	wasmx.SStore(GetValidatorAddressesKey(chainId), string(bz))
}

// Levels
func GetLevelChainIds(levelIndex int32) []string {
	value := wasmx.SLoad(GetLevelChainIdsKey(levelIndex))
	if value == "" {
		return []string{}
	}
	var out []string
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func AddLevelChainId(levelIndex int32, chainId string) {
	value := GetLevelChainIds(levelIndex)
	for _, id := range value {
		if id == chainId {
			Revert("chain_id " + chainId + " already included in level " + strconv.FormatInt(int64(levelIndex), 10))
		}
	}
	value = append(value, chainId)
	SetLevelChainIds(levelIndex, value)
	lastLevel := GetLevelLast()
	if lastLevel < levelIndex {
		SetLevelLast(levelIndex)
	}
}

func SetLevelChainIds(levelIndex int32, chainIds []string) {
	bz, _ := json.Marshal(&chainIds)
	wasmx.SStore(GetLevelChainIdsKey(levelIndex), string(bz))
}

// Chain ids list
func AddChainId(id string) {
	ids := GetChainIds()
	for _, v := range ids {
		if v == id {
			return
		}
	}
	ids = append(ids, id)
	bz, _ := json.Marshal(&ids)
	wasmx.SStore(CHAIN_IDS, string(bz))
}

func GetChainIds() []string {
	value := wasmx.SLoad(CHAIN_IDS)
	if value == "" {
		return []string{}
	}
	var out []string
	_ = json.Unmarshal([]byte(value), &out)
	return out
}

func SetChainIds(data []string) {
	bz, _ := json.Marshal(&data)
	wasmx.SStore(CHAIN_IDS, string(bz))
}

// Last/Current level
func GetLevelLast() int32 {
	valuestr := wasmx.SLoad(LEVEL_LAST)
	if valuestr == "" {
		return INITIAL_LEVEL
	}
	v, err := strconv.ParseInt(valuestr, 10, 32)
	if err != nil {
		Revert("invalid level_last: " + err.Error())
		return INITIAL_LEVEL
	}
	return int32(v)
}

func SetLevelLast(id int32) { wasmx.SStore(LEVEL_LAST, strconv.FormatInt(int64(id), 10)) }

func GetCurrentLevel() int32 {
	valuestr := wasmx.SLoad(CURRENT_LEVEL)
	if valuestr == "" {
		return 0
	}
	v, err := strconv.ParseInt(valuestr, 10, 32)
	if err != nil {
		Revert("invalid currentlevel: " + err.Error())
		return 0
	}
	return int32(v)
}

func SetCurrentLevel(level int32) { wasmx.SStore(CURRENT_LEVEL, strconv.FormatInt(int64(level), 10)) }

func GetChainIdLast() uint64 {
	valuestr := wasmx.SLoad(LAST_CHAIN_ID)
	if valuestr == "" {
		return 0
	}
	v, err := strconv.ParseUint(valuestr, 10, 64)
	if err != nil {
		Revert("invalid chainid_last: " + err.Error())
		return 0
	}
	return v
}

func SetChainIdLast(id uint64) { wasmx.SStore(LAST_CHAIN_ID, strconv.FormatUint(id, 10)) }

// Params
func GetParams() Params {
	value := wasmx.SLoad(PARAMS_KEY)
	if value == "" {
		init := sdkmath.ZeroInt()
		if v, ok := sdkmath.NewIntFromString(DEFAULT_INITIAL_BALANCE); ok {
			init = v
		} else {
			Revert("invalid DEFAULT_INITIAL_BALANCE")
		}
		return Params{
			MinValidatorsCount:  DEFAULT_MIN_VALIDATORS_COUNT,
			EnableEidCheck:      DEFAULT_EID_CHECK,
			Erc20CodeId:         DEFAULT_ERC20_CODE_ID,
			Derc20CodeId:        DEFAULT_DERC20_CODE_ID,
			LevelInitialBalance: init,
		}
	}
	var p Params
	if err := json.Unmarshal([]byte(value), &p); err != nil {
		Revert("cannot decode params: " + err.Error())
	}
	return p
}

func SetParams(data Params) {
	bz, _ := json.Marshal(&data)
	wasmx.SStore(PARAMS_KEY, string(bz))
}
