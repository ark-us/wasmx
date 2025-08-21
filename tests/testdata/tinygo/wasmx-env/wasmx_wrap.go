package wasmx

// #include <stdlib.h>
import "C"

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"time"

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

func GetEnv() []byte { return utils.PackedPtrToBytes(GetEnv_()) }

func GetCaller() Bech32String {
	bz := utils.PackedPtrToBytes(GetCaller_())
	return Bech32String(Bech32BytesToString(bz))
}

func GetAddress() Bech32String {
	bz := utils.PackedPtrToBytes(GetAddress_())
	return Bech32String(Bech32BytesToString(bz))
}

func GetCurrentBlock() BlockInfo {
	data := utils.PackedPtrToBytes(GetCurrentBlock_())
	v := BlockInfo{}
	err := json.Unmarshal(data, &v)
	if err != nil {
		panic(err)
	}
	return v
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

// log a message to the console using _log.
func Log(data []byte, topics [][32]byte) {
	encoded, _ := json.Marshal(WasmxLog{Data: data, Topics: topics})
	Log_(utils.BytesToPackedPtr(encoded))
}

func EmitCosmosEvents(events []Event) {
	bz, _ := json.Marshal(events)
	EmitCosmosEvents_(utils.BytesToPackedPtr(bz))
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

// storage helpers
func StorageDelete(key string) { StorageDelete_(utils.StringToPackedPtr(key)) }

func StorageDeleteRangeReq(r StorageDeleteRange) {
	bz, _ := json.Marshal(&r)
	StorageDeleteRange_(utils.BytesToPackedPtr(bz))
}

func StorageLoadRangeReq(r StorageRange) []string {
	bz, _ := json.Marshal(&r)
	out := utils.PackedPtrToBytes(StorageLoadRange_(utils.BytesToPackedPtr(bz)))
	var res []string
	_ = json.Unmarshal(out, &res)
	return res
}

func StorageLoadRangePairsReq(r StorageRange) StoragePairs {
	bz, _ := json.Marshal(&r)
	out := utils.PackedPtrToBytes(StorageLoadRangePairs_(utils.BytesToPackedPtr(bz)))
	var res StoragePairs
	_ = json.Unmarshal(out, &res)
	return res
}

func GetFinishData() []byte { return utils.PackedPtrToBytes(GetFinishData_()) }

func GetAccount(addr Bech32String) Account {
	bz := utils.PackedPtrToBytes(GetAccount_(utils.BytesToPackedPtr(Bech32StringToBytes(string(addr)))))
	var acc Account
	_ = json.Unmarshal(bz, &acc)
	return acc
}

func ExecuteCosmosMsg(msg string) CallResponse {
	out := utils.PackedPtrToBytes(ExecuteCosmosMsg_(utils.StringToPackedPtr(msg)))
	var resp CallResponse
	_ = json.Unmarshal(out, &resp)
	return resp
}

func CreateAccount(req CreateAccountRequest) Bech32String {
	// Msg is json string, must be base64-encoded like AS SDK
	req.Msg = base64.StdEncoding.EncodeToString([]byte(req.Msg))
	bz, _ := json.Marshal(&req)
	out := utils.PackedPtrToBytes(CreateAccount_(utils.BytesToPackedPtr(bz)))
	var resp CreateAccountResponse
	_ = json.Unmarshal(out, &resp)
	return resp.Address
}

func Create2Account(req Create2AccountRequest) Bech32String {
	req.Msg = base64.StdEncoding.EncodeToString([]byte(req.Msg))
	bz, _ := json.Marshal(&req)
	out := utils.PackedPtrToBytes(CreateAccount2_(utils.BytesToPackedPtr(bz)))
	var resp Create2AccountResponse
	_ = json.Unmarshal(out, &resp)
	return resp.Address
}

// crypto
func Sha256(dataBase64 string) string {
	out := utils.PackedPtrToBytes(Sha256_(utils.StringToPackedPtr(dataBase64)))
	return base64.StdEncoding.EncodeToString(out)
}

func MerkleHashSlices(slices []string) string {
	payload := struct {
		Slices []string `json:"slices"`
	}{Slices: slices}
	bz, _ := json.Marshal(&payload)
	out := utils.PackedPtrToBytes(MerkleHash_(utils.BytesToPackedPtr(bz)))
	return base64.StdEncoding.EncodeToString(out)
}

func Ed25519Sign(privKey []byte, msg []byte) []byte {
	sig := utils.PackedPtrToBytes(Ed25519Sign_(utils.BytesToPackedPtr(privKey), utils.BytesToPackedPtr(msg)))
	return sig
}

func Ed25519Verify(pubKey []byte, signature []byte, msg []byte) bool {
	r := Ed25519Verify_(utils.BytesToPackedPtr(pubKey), utils.BytesToPackedPtr(signature), utils.BytesToPackedPtr(msg))
	return r == 1
}

func Ed25519PubToHex(pubKey []byte) string {
	return string(utils.PackedPtrToBytes(Ed25519PubToHex_(utils.BytesToPackedPtr(pubKey))))
}

// addresses
func ValidateBech32Address(value string) bool {
	return ValidateBech32Address_(utils.StringToPackedPtr(value)) == 1
}

func AddrHumanize(bz []byte) string        { return Bech32BytesToString(bz) }
func AddrCanonicalize(value string) []byte { return Bech32StringToBytes(value) }
func AddrEquivalent(addr1, addr2 Bech32String) bool {
	return AddrEquivalent_(utils.BytesToPackedPtr(Bech32StringToBytes(string(addr1))), utils.BytesToPackedPtr(Bech32StringToBytes(string(addr2)))) == 1
}
func AddrHumanizeMC(bz []byte, prefix string) string {
	return string(utils.PackedPtrToBytes(AddrHumanizeMC_(utils.BytesToPackedPtr(bz), utils.StringToPackedPtr(prefix))))
}
func AddrCanonicalizeMC(value string) []byte {
	return utils.PackedPtrToBytes(AddrCanonicalizeMC_(utils.StringToPackedPtr(value)))
}

func GetAddressByRole(role string) Bech32String {
	return Bech32String(Bech32BytesToString(utils.PackedPtrToBytes(GetAddressByRole_(utils.StringToPackedPtr(role)))))
}
func GetRoleByAddress(addr Bech32String) string {
	return string(utils.PackedPtrToBytes(GetRoleByAddress_(utils.BytesToPackedPtr(Bech32StringToBytes(string(addr))))))
}

// cosmos tx helpers
func DecodeCosmosTxToJson(data []byte) []byte {
	return utils.PackedPtrToBytes(DecodeCosmosTxToJson_(utils.BytesToPackedPtr(data)))
}
func VerifyCosmosTx(encoded []byte) []byte {
	return utils.PackedPtrToBytes(VerifyCosmosTx_(utils.BytesToPackedPtr(encoded)))
}

func GetTimestamp() time.Time {
	data := GetCurrentBlock()
	// timestamp is in nanoseconds since unix epoch
	return time.Unix(0, int64(data.Timestamp))
	// return new Date(i64(data.timestamp / 1000000))
}
