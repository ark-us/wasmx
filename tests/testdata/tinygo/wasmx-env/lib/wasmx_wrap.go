package wasmx

// #include <stdlib.h>
import "C"

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func RevertWithModule(module string, msg string) {
	LoggerError(module, "revert", []string{"error", msg})
	msg = fmt.Sprintf(`%s: %s`, module, msg)
	Revert_(utils.BytesToPackedPtr([]byte(msg)))
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
	if err := json.Unmarshal(data, &v); err != nil {
		RevertWithModule("wasmx-env:GetCurrentBlock", err.Error())
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

func CallInternal(addrBech32 Bech32String, value *sdkmath.Int, calldata []byte, gasLimit *big.Int, isQuery bool, moduleName string) (bool, []byte) {
	LoggerDebugExtended(moduleName+":wasmx_env", "call", []string{"to", string(addrBech32), "calldata", string(calldata), "is_query", fmt.Sprintf("%t", isQuery)})

	req := &SimpleCallRequestRaw{
		To:       addrBech32,
		Value:    value,
		GasLimit: gasLimit,
		Calldata: calldata,
		IsQuery:  isQuery,
	}

	reqbz, err := json.Marshal(req)
	if err != nil {
		RevertWithModule(moduleName+":CallInternal", err.Error())
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

func Call(addrBech32 Bech32String, value *sdkmath.Int, calldata []byte, gasLimit *big.Int, moduleName string) (bool, []byte) {
	return CallInternal(addrBech32, value, calldata, gasLimit, false, moduleName)
}

func CallSimple(addrBech32 Bech32String, calldata []byte, isQuery bool, moduleName string) (bool, []byte) {
	return CallInternal(addrBech32, nil, calldata, bigInt(DEFAULT_GAS_TX), isQuery, moduleName)
}

func CallStatic(addrBech32 Bech32String, calldata []byte, gasLimit *big.Int, moduleName string) (bool, []byte) {
	return CallInternal(addrBech32, nil, calldata, gasLimit, false, moduleName)
}

// log a message to the console using _log.
func Log(data []byte, topics [][32]byte) {
	encoded, err := json.Marshal(WasmxLog{Data: data, Topics: topics})
	if err != nil {
		RevertWithModule("wasmx-env:Log", err.Error())
	}
	Log_(utils.BytesToPackedPtr(encoded))
}

func EmitCosmosEvents(events []Event) {
	bz, err := json.Marshal(events)
	if err != nil {
		RevertWithModule("wasmx-env:EmitCosmosEvents", err.Error())
	}
	EmitCosmosEvents_(utils.BytesToPackedPtr(bz))
}

func LoggerInfo(module string, msg string, parts []string) {
	msg = fmt.Sprintf(`%s: %s`, module, msg)
	LoggerInfo_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerError(module string, msg string, parts []string) {
	msg = fmt.Sprintf(`%s: %s`, module, msg)
	LoggerError_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerDebug(module string, msg string, parts []string) {
	msg = fmt.Sprintf(`%s: %s`, module, msg)
	LoggerDebug_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerDebugExtended(module string, msg string, parts []string) {
	msg = fmt.Sprintf(`%s: %s`, module, msg)
	LoggerDebugExtended_(LoggerDataToPackedPtr(msg, parts))
}

func LoggerDataToPackedPtr(msg string, parts []string) int64 {
	databz, err := json.Marshal(&LoggerLog{Msg: msg, Parts: parts})
	if err != nil {
		RevertWithModule("wasmx-env:LoggerDataToPackedPtr", "cannot marshal LoggerLog: "+err.Error())
	}
	return utils.BytesToPackedPtr(databz)
}

// storage helpers
func StorageDelete(key string) { StorageDelete_(utils.StringToPackedPtr(key)) }

func StorageDeleteRangeReq(r StorageDeleteRange) {
	bz, err := json.Marshal(&r)
	if err != nil {
		RevertWithModule("wasmx-env:StorageDeleteRangeReq", err.Error())
	}
	StorageDeleteRange_(utils.BytesToPackedPtr(bz))
}

// Storage helper functions to match AS versions
func SStore(key, value string) {
	StorageStore([]byte(key), []byte(value))
}

func SLoad(key string) string {
	return string(StorageLoad([]byte(key)))
}

func SDelete(key string) {
	StorageDelete(key)
}

func SDeleteRange(keyStart, keyEnd string) {
	req := StorageDeleteRange{
		StartKey: base64.StdEncoding.EncodeToString([]byte(keyStart)),
		EndKey:   base64.StdEncoding.EncodeToString([]byte(keyEnd)),
	}
	StorageDeleteRangeReq(req)
}

func StorageLoadRangeReq(r StorageRange) []string {
	bz, err := json.Marshal(&r)
	if err != nil {
		RevertWithModule("wasmx-env:StorageLoadRangeReq", err.Error())
	}
	out := utils.PackedPtrToBytes(StorageLoadRange_(utils.BytesToPackedPtr(bz)))
	var res []string
	if err := json.Unmarshal(out, &res); err != nil {
		RevertWithModule("wasmx-env:StorageLoadRangeReq", err.Error())
	}
	return res
}

func StorageLoadRangePairsReq(r StorageRange) StoragePairs {
	bz, err := json.Marshal(&r)
	if err != nil {
		RevertWithModule("wasmx-env:StorageLoadRangePairsReq", err.Error())
	}
	out := utils.PackedPtrToBytes(StorageLoadRangePairs_(utils.BytesToPackedPtr(bz)))
	var res StoragePairs
	if err := json.Unmarshal(out, &res); err != nil {
		RevertWithModule("wasmx-env:StorageLoadRangePairsReq", err.Error())
	}
	return res
}

func GetFinishData() []byte { return utils.PackedPtrToBytes(GetFinishData_()) }

func GetAccount(addr Bech32String) Account {
	bz := utils.PackedPtrToBytes(GetAccount_(utils.BytesToPackedPtr(Bech32StringToBytes(string(addr)))))
	var acc Account
	if err := json.Unmarshal(bz, &acc); err != nil {
		RevertWithModule("wasmx-env:GetAccount", err.Error())
	}
	return acc
}

func ExecuteCosmosMsg(msg string, moduleName ...string) CallResponse {
	module := ""
	if len(moduleName) > 0 {
		module = moduleName[0]
	}
	LoggerDebugExtended(module+":wasmx_env", "executeCosmosMsg", []string{"msg", msg})
	out := utils.PackedPtrToBytes(ExecuteCosmosMsg_(utils.StringToPackedPtr(msg)))
	LoggerDebugExtended(module+":wasmx_env", "executeCosmosMsg", []string{"response", string(out)})
	var resp CallResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		RevertWithModule(module+":ExecuteCosmosMsg", err.Error())
	}
	return resp
}

// CallEvm mirrors AS callEvm: do not base64-encode calldata, send as-is
func CallEvm(req CallRequest, moduleName ...string) CallResponse {
	mod := ""
	if len(moduleName) > 0 {
		mod = moduleName[0]
	}
	LoggerDebugExtended(mod+":wasmx_env", "call", []string{"to", req.To, "calldata", req.Calldata})
	bz, err := json.Marshal(&req)
	if err != nil {
		RevertWithModule(mod+":CallEvm", err.Error())
	}
	out := utils.PackedPtrToBytes(Call_(utils.BytesToPackedPtr(bz)))
	var resp CallResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		RevertWithModule(mod+":CallEvm", err.Error())
	}
	return resp
}

func CreateAccount(req CreateAccountRequest, moduleName ...string) Bech32String {
	// Msg is json string, must be base64-encoded like AS SDK
	req.Msg = base64.StdEncoding.EncodeToString([]byte(req.Msg))
	bz, err := json.Marshal(&req)
	if err != nil {
		RevertWithModule("wasmx-env:CreateAccount", err.Error())
	}
	module := ""
	if len(moduleName) > 0 {
		module = moduleName[0]
	}
	LoggerDebugExtended(module+":wasmx_env", "createAccount", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(CreateAccount_(utils.BytesToPackedPtr(bz)))
	var resp CreateAccountResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		RevertWithModule("wasmx-env:CreateAccount", err.Error())
	}
	return resp.Address
}

func Create2Account(req Create2AccountRequest, moduleName ...string) Bech32String {
	req.Msg = base64.StdEncoding.EncodeToString([]byte(req.Msg))
	bz, err := json.Marshal(&req)
	if err != nil {
		RevertWithModule("wasmx-env:Create2Account", err.Error())
	}
	module := ""
	if len(moduleName) > 0 {
		module = moduleName[0]
	}
	LoggerDebugExtended(module+":wasmx_env", "create2Account", []string{"request", string(bz)})
	out := utils.PackedPtrToBytes(CreateAccount2_(utils.BytesToPackedPtr(bz)))
	var resp Create2AccountResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		RevertWithModule("wasmx-env:Create2Account", err.Error())
	}
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
	bz, err := json.Marshal(&payload)
	if err != nil {
		RevertWithModule("wasmx-env:MerkleHashSlices", err.Error())
	}
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

// VerifyCosmosTxFromBase64 verifies a cosmos transaction from base64-encoded data
func VerifyCosmosTxFromBase64(encodedTx string) VerifyCosmosTxResponse {
	data, err := base64.StdEncoding.DecodeString(encodedTx)
	if err != nil {
		RevertWithModule("wasmx-env:VerifyCosmosTxFromBase64", err.Error())
	}
	result := VerifyCosmosTx(data)
	var resp VerifyCosmosTxResponse
	if err := json.Unmarshal(result, &resp); err != nil {
		RevertWithModule("wasmx-env:VerifyCosmosTxFromBase64", err.Error())
	}
	return resp
}

// DecodeCosmosTxFromBytes decodes a cosmos transaction from bytes to JSON
func DecodeCosmosTxFromBytes(data []byte) SignedTransaction {
	result := DecodeCosmosTxToJson(data)
	var tx SignedTransaction
	if err := json.Unmarshal(result, &tx); err != nil {
		RevertWithModule("wasmx-env:DecodeCosmosTxFromBytes", err.Error())
	}
	return tx
}

func GetTimestamp() time.Time {
	data := GetCurrentBlock()
	// timestamp is in nanoseconds since unix epoch
	return time.Unix(0, int64(data.Timestamp))
	// return new Date(i64(data.timestamp / 1000000))
}
