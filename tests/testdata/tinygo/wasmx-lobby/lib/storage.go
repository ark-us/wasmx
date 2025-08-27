package lib

import (
	"encoding/json"
	"strconv"

	sdkmath "cosmossdk.io/math"
	consensus "github.com/loredanacirstea/wasmx-env-consensus/lib"
	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const TEMP_NEW_CHAIN_REQUESTS = "newchain_requests"
const TEMP_NEW_CHAIN_RESPONSE = "newchain_response"
const SUBCHAIN_DATA = "subchain_data"
const SETUP_DATA = "setup_data"
const LAST_CHAIN_ID = "chainid_last"

func GetNewChainRequests() ([]MsgNewChainRequest, error) {
	value := wasmx.SLoad(TEMP_NEW_CHAIN_REQUESTS)
	if value == "" {
		return []MsgNewChainRequest{}, nil
	}
	var requests []MsgNewChainRequest
	if err := json.Unmarshal([]byte(value), &requests); err != nil {
		return nil, err
	}
	return requests, nil
}

func AddNewChainRequest(data MsgNewChainRequest) error {
	req, err := GetNewChainRequests()
	if err != nil {
		return err
	}

	// Check if already exists
	for _, r := range req {
		if r.Validator.ConsensusPublicKey == data.Validator.ConsensusPublicKey {
			return nil // already exists
		}
	}

	req = append(req, data)
	return SetNewChainRequests(req)
}

func SetNewChainRequests(data []MsgNewChainRequest) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	wasmx.SStore(TEMP_NEW_CHAIN_REQUESTS, string(jsonData))
	return nil
}

func GetNewChainResponse() (*MsgNewChainResponse, error) {
	value := wasmx.SLoad(TEMP_NEW_CHAIN_RESPONSE)
	if value == "" {
		return nil, nil
	}
	var response MsgNewChainResponse
	if err := json.Unmarshal([]byte(value), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func SetNewChainResponse(data MsgNewChainResponse) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	wasmx.SStore(TEMP_NEW_CHAIN_RESPONSE, string(jsonData))
	return nil
}

func GetChainSetupData() (*CurrentChainSetup, error) {
	value := wasmx.SLoad(SETUP_DATA)
	if value == "" {
		return nil, nil
	}
	var setup CurrentChainSetup
	if err := json.Unmarshal([]byte(value), &setup); err != nil {
		return nil, err
	}
	return &setup, nil
}

func SetChainSetupData(data CurrentChainSetup) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	wasmx.SStore(SETUP_DATA, string(jsonData))
	return nil
}

func GetSubChainData() (*MsgNewChainGenesisData, error) {
	value := wasmx.SLoad(SUBCHAIN_DATA)
	if value == "" {
		return nil, nil
	}
	var data MsgNewChainGenesisData
	if err := json.Unmarshal([]byte(value), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func SetSubChainData(data MsgNewChainGenesisData) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	wasmx.SStore(SUBCHAIN_DATA, string(jsonData))
	return nil
}

func GetChainIdLast() (consensus.ChainId, error) {
	value := wasmx.SLoad(LAST_CHAIN_ID)
	if value == "" {
		level, err := GetCurrentLevel()
		if err != nil {
			return consensus.ChainId{}, err
		}
		return consensus.ChainId{
			Full:      "",
			BaseName:  "",
			Level:     uint32(level),
			EvmID:     INIT_CHAIN_INDEX,
			ForkIndex: INIT_FORK_INDEX,
		}, nil
	}
	var chainId consensus.ChainId
	if err := json.Unmarshal([]byte(value), &chainId); err != nil {
		return consensus.ChainId{}, err
	}
	return chainId, nil
}

func SetChainIdLast(data consensus.ChainId) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	wasmx.SStore(LAST_CHAIN_ID, string(jsonData))
	return nil
}

func GetMinValidatorsCount() (int32, error) {
	value := wasmx.GetContextValue(KEY_MIN_VALIDATORS_COUNT)
	if value == "" {
		return 0, nil
	}
	val, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

func GetCurrentLevel() (int32, error) {
	value := wasmx.GetContextValue(KEY_CURRENT_LEVEL)
	if value == "" {
		return 0, nil
	}
	val, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(val), nil
}

func GetNextLevel() (int32, error) {
	level, err := GetCurrentLevel()
	if err != nil {
		return 0, err
	}
	return level + 1, nil
}

func GetParams() (Params, error) {
	currentLevel, err := GetCurrentLevel()
	if err != nil {
		return Params{}, err
	}

	minValidatorCount, err := GetMinValidatorsCount()
	if err != nil {
		return Params{}, err
	}

	erc20CodeIdStr := wasmx.GetContextValue(KEY_ERC20_CODE_ID)
	erc20CodeId, err := strconv.ParseUint(erc20CodeIdStr, 10, 64)
	if err != nil {
		return Params{}, err
	}

	derc20CodeIdStr := wasmx.GetContextValue(KEY_DERC20_CODE_ID)
	derc20CodeId, err := strconv.ParseUint(derc20CodeIdStr, 10, 64)
	if err != nil {
		return Params{}, err
	}

	initialBalanceStr := wasmx.GetContextValue(KEY_INITIAL_BALANCE)
	initialBalance, ok := sdkmath.NewIntFromString(initialBalanceStr)
	if !ok {
		return Params{}, err
	}

	enableEidCheckStr := wasmx.GetContextValue(KEY_ENABLE_EID_CHECK)
	enableEidCheck := enableEidCheckStr == "true"

	return Params{
		CurrentLevel:        currentLevel,
		MinValidatorsCount:  minValidatorCount,
		EnableEidCheck:      enableEidCheck,
		Erc20CodeId:         erc20CodeId,
		Derc20CodeId:        derc20CodeId,
		LevelInitialBalance: initialBalance,
	}, nil
}
