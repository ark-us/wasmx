package lib

import (
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
)

const storageKeyGasPrice = "gas_price"

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

// LoadGasPrice loads the gas price from storage
func LoadGasPrice() *wasmx.Coin {
	key := []byte(storageKeyGasPrice)
	data := wasmx.StorageLoad(key)
	if len(data) == 0 {
		return &wasmx.Coin{}
	}

	var coin wasmx.Coin
	if err := json.Unmarshal(data, &coin); err != nil {
		return &wasmx.Coin{}
	}
	return &coin
}

// SaveGasPrice saves the gas price to storage
func SaveGasPrice(coin wasmx.Coin) error {
	key := []byte(storageKeyGasPrice)
	data, err := json.Marshal(coin)
	if err != nil {
		return err
	}
	wasmx.StorageStore(key, data)
	return nil
}

// InitGenesis initializes the contract with genesis state
func InitGenesis(msg *InitGenesisRequest) []byte {
	LoggerInfo("InitGenesis called", nil)

	// Gas price is required
	if msg.GasPrice.Amount.IsNil() {
		return marshalJSON(InitGenesisResponse{
			Success: false,
			Error:   "gas price is required",
		})
	}
	if msg.GasPrice.Amount.IsZero() {
		return marshalJSON(InitGenesisResponse{
			Success: false,
			Error:   "gas price must be greater than zero",
		})
	}
	if err := SaveGasPrice(msg.GasPrice); err != nil {
		return marshalJSON(InitGenesisResponse{
			Success: false,
			Error:   "Failed to save gas price",
		})
	}

	LoggerInfo("Genesis initialized", nil)
	return marshalJSON(InitGenesisResponse{Success: true})
}

func marshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		LoggerError("Failed to marshal JSON", []string{"error", err.Error()})
		return []byte("{}")
	}
	return data
}
