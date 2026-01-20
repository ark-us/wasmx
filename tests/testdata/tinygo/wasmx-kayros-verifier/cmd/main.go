package main

import (
	"encoding/hex"
	"encoding/json"

	wasmx "github.com/loredanacirstea/wasmx-env/lib"
	verifier "github.com/loredanacirstea/wasmx-kayros-verifier/lib"
)

//go:wasm-module wasmx
//export memory_ptrlen_i64_1
func Memory_ptrlen_i64_1() {}

//go:wasm-module wasmx
//export wasmx_env_i64_2
func Wasmx_env_i64_2() {}

//go:wasm-module httpclient
//export wasmx_httpclient_i64_1
func Wasmx_httpclient_i64_1() {}

func respond(ok bool, errMsg string) {
	resp := verifier.VerifyResponse{Ok: ok, Error: errMsg}
	bz, err := json.Marshal(resp)
	if err != nil {
		wasmx.Revert([]byte("failed to marshal response: " + err.Error()))
	}
	wasmx.Finish(bz)
}

func configFrom(baseUrl string, userKey string) verifier.KayrosConfig {
	return verifier.KayrosConfig{
		ApiBaseUrl: baseUrl,
		ApiUserKey: userKey,
	}
}

//go:wasm-module wasmx-kayros-verifier
//export instantiate
func Instantiate() {
	databz := wasmx.GetCallData()
	if len(databz) == 0 {
		wasmx.Finish([]byte{})
		return
	}
	var payload map[string]any
	if err := json.Unmarshal(databz, &payload); err != nil {
		wasmx.Revert([]byte("invalid call data: " + err.Error() + ": " + string(databz)))
	}
	wasmx.Finish([]byte{})
}

func main() {
	databz := wasmx.GetCallData()
	calld := &verifier.Calldata{}
	if err := json.Unmarshal(databz, calld); err != nil {
		wasmx.Revert([]byte("invalid call data: " + err.Error() + ": " + string(databz)))
	}

	switch {
	case calld.VerifyProof != nil:
		req := *calld.VerifyProof
		if len(req.Data) == 0 {
			respond(false, "missing data")
			return
		}
		dataTypeHex := wasmx.HexString(hex.EncodeToString(req.DataType))
		ok, errMsg := verifier.VerifyProof(req.Data, dataTypeHex, req.HashAlgo, configFrom(req.ApiBaseUrl, req.ApiUserKey))
		respond(ok, errMsg)
		return
	case calld.VerifyProofWithInclusion != nil:
		req := *calld.VerifyProofWithInclusion
		if len(req.Data) == 0 {
			respond(false, "missing data")
			return
		}
		dataTypeHex := wasmx.HexString(hex.EncodeToString(req.DataType))
		ok, errMsg := verifier.VerifyProofWithInclusion(
			req.Data,
			dataTypeHex,
			req.HashAlgo,
			req.TrustedRootHash,
			req.TrustedLevel,
			req.TrustedPosition,
			configFrom(req.ApiBaseUrl, req.ApiUserKey),
		)
		respond(ok, errMsg)
		return
	case calld.VerifyProofHash != nil:
		req := *calld.VerifyProofHash
		dataTypeHex := wasmx.HexString(hex.EncodeToString(req.DataType))
		dataHashHex := wasmx.HexString(hex.EncodeToString(req.DataHash))
		ok, errMsg := verifier.VerifyProofHash(dataTypeHex, dataHashHex, configFrom(req.ApiBaseUrl, req.ApiUserKey))
		respond(ok, errMsg)
		return
	case calld.VerifyProofHashWithInclusion != nil:
		req := *calld.VerifyProofHashWithInclusion
		dataTypeHex := wasmx.HexString(hex.EncodeToString(req.DataType))
		dataHashHex := wasmx.HexString(hex.EncodeToString(req.DataHash))
		ok, errMsg := verifier.VerifyProofHashWithInclusion(
			dataTypeHex,
			dataHashHex,
			req.HashAlgo,
			req.TrustedRootHash,
			req.TrustedLevel,
			req.TrustedPosition,
			configFrom(req.ApiBaseUrl, req.ApiUserKey),
		)
		respond(ok, errMsg)
		return
	case calld.VerifyRecordHash != nil:
		req := *calld.VerifyRecordHash
		ok, errMsg := verifier.VerifyRecordHash(&req.Record, req.HashAlgo)
		respond(ok, errMsg)
		return
	case calld.VerifyRecordChainLink != nil:
		req := *calld.VerifyRecordChainLink
		ok, errMsg := verifier.VerifyRecordChainLink(&req.Record, &req.Prev)
		respond(ok, errMsg)
		return
	case calld.VerifyRecordTimestamp != nil:
		req := *calld.VerifyRecordTimestamp
		ok, errMsg := verifier.VerifyRecordTimestamp(&req.Record)
		respond(ok, errMsg)
		return
	case calld.VerifyRecordUUID != nil:
		req := *calld.VerifyRecordUUID
		ok, errMsg := verifier.VerifyRecordUUID(&req.Record)
		respond(ok, errMsg)
		return
	case calld.VerifyLevelProof != nil:
		req := *calld.VerifyLevelProof
		ok, errMsg := verifier.VerifyLevelProof(configFrom(req.ApiBaseUrl, req.ApiUserKey), req.Proof, req.HashAlgo)
		respond(ok, errMsg)
		return
	case calld.VerifyProofPath != nil:
		req := *calld.VerifyProofPath
		ok, errMsg := verifier.VerifyProofPath(&req.Proof, req.HashAlgo)
		respond(ok, errMsg)
		return
	case calld.VerifyKayrosRecord != nil:
		req := *calld.VerifyKayrosRecord
		ok, errMsg := verifier.VerifyKayrosRecord(&req.Record, &req.Prev, req.HashAlgo, req.LevelProofs, configFrom(req.ApiBaseUrl, req.ApiUserKey))
		respond(ok, errMsg)
		return
	case calld.VerifyKayrosRecordWithProof != nil:
		req := *calld.VerifyKayrosRecordWithProof
		ok, errMsg := verifier.VerifyKayrosRecordWithProof(&req.Record, &req.Prev, &req.Proof, req.HashAlgo)
		respond(ok, errMsg)
		return
	}

	wasmx.Revert(append([]byte("invalid function call data: "), databz...))
}
