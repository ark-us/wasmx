package wasmx

// #include <stdlib.h>
import "C"

import (
	"encoding/json"
	"math/big"

	sdkmath "cosmossdk.io/math"
	utils "github.com/loredanacirstea/wasmx-env-utils"
)

func StorageStore(key, value []byte) {
	Log([]byte("storagestore"), [][32]byte{})
	StorageStore_(utils.BytesToPackedPtr(key), utils.BytesToPackedPtr(value))
}

func StorageLoad(key []byte) []byte {
	packed := StorageLoad_(utils.BytesToPackedPtr(key))
	return utils.PackedPtrToBytes(packed)
}

func GetCallData() []byte {
	packed := GetCallData_()
	return utils.PackedPtrToBytes(packed)
}

func SetFinishData(data []byte) {
	SetFinishData_(utils.BytesToPackedPtr(data))
}

func SetReturnData(data []byte) {
	SetReturnData_(utils.BytesToPackedPtr(data))
}

func Finish(data []byte) {
	Finish_(utils.BytesToPackedPtr(data))
}

func Revert(data []byte) {
	Revert_(utils.BytesToPackedPtr(data))
}

func GetChainId() string {
	return string(utils.PackedPtrToBytes(GetChainId_()))
}

func Bech32StringToBytes(addrBech32 string) []byte {
	ptr := Bech32StringToBytes_(utils.StringToPackedPtr(addrBech32))
	return utils.PackedPtrToBytes(ptr)
}

func Bech32BytesToString(addr []byte) string {
	packed := Bech32BytesToString_(utils.BytesToPackedPtr(utils.PaddLeftTo32(addr)))
	data := utils.PackedPtrToBytes(packed)
	return string(data)
}

func CallInternal(addrBech32 Bech32String, value *sdkmath.Int, calldata []byte, gasLimit *big.Int, isQuery bool) (bool, []byte) {
	req := &SimpleCallRequestRaw{
		To:       addrBech32,
		Value:    value,
		GasLimit: gasLimit,
		Calldata: calldata,
		IsQuery:  isQuery,
	}

	reqbz, err := json.Marshal(req)
	if err != nil {
		Revert([]byte(err.Error()))
	}
	ptr := utils.BytesToPackedPtr(reqbz)
	packed := Call_(ptr)
	res := utils.PackedPtrToBytes(packed)

	var calld CallResult
	err = json.Unmarshal(res, &calld)
	if err != nil {
		Revert([]byte("Cannot decode json: " + err.Error()))
	}
	return calld.Success == 0, calld.Data
}

func Call(addrBech32 Bech32String, value *sdkmath.Int, calldata []byte, gasLimit *big.Int) (bool, []byte) {
	return CallInternal(addrBech32, value, calldata, gasLimit, false)
}

func CallStatic(addrBech32 Bech32String, calldata []byte, gasLimit *big.Int) (bool, []byte) {
	return CallInternal(addrBech32, nil, calldata, gasLimit, false)
}

type WasmxLog struct {
	Data   []byte
	Topics [][32]byte
}

// log a message to the console using _log.
func Log(data []byte, topics [][32]byte) {
	encoded, _ := json.Marshal(WasmxLog{Data: data, Topics: topics})
	Log_(utils.BytesToPackedPtr(encoded))
}

func LoggerInfo(msg string, parts []string) {
	LoggerInfo_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerError(msg string, parts []string) {
	LoggerError_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerDebug(msg string, parts []string) {
	LoggerDebug_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerDebugExtended(msg string, parts []string) {
	LoggerDebugExtended_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerDataToPackedPtr(msg string, parts []string) int64 {
	databz, err := json.Marshal(&LoggerLog{Msg: msg, Parts: parts})
	if err != nil {
		Revert([]byte("cannot marshal LoggerLog" + err.Error()))
	}
	return utils.BytesToPackedPtr(databz)
}
