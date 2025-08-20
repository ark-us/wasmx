package sql

// #include <stdlib.h>
import "C"

import (
	"encoding/json"

	utils "github.com/loredanacirstea/wasmx-env-utils"
)

//go:wasm-module sql
//export wasmx_sql_i64_1
func wasmx_sql_i64_1() {}

// Host function imports
//
//go:wasmimport sql Connect
func Connect_(reqPtr int64) int64

//go:wasmimport sql Close
func Close_(reqPtr int64) int64

//go:wasmimport sql Ping
func Ping_(reqPtr int64) int64

//go:wasmimport sql Execute
func Execute_(reqPtr int64) int64

//go:wasmimport sql BatchAtomic
func BatchAtomic_(reqPtr int64) int64

//go:wasmimport sql Query
func Query_(reqPtr int64) int64

// SDK function wrappers
func Connect(req *SqlConnectionRequest) SqlConnectionResponse {
	bz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	ptr := Connect_(utils.BytesToPackedPtr(bz))
	resp := SqlConnectionResponse{}
	err = json.Unmarshal(utils.PackedPtrToBytes(ptr), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Close(req *SqlCloseRequest) SqlCloseResponse {
	bz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	ptr := Close_(utils.BytesToPackedPtr(bz))
	resp := SqlCloseResponse{}
	err = json.Unmarshal(utils.PackedPtrToBytes(ptr), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Ping(req *SqlPingRequest) SqlPingResponse {
	bz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	ptr := Ping_(utils.BytesToPackedPtr(bz))
	resp := SqlPingResponse{}
	err = json.Unmarshal(utils.PackedPtrToBytes(ptr), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Execute(req *SqlExecuteRequest) SqlExecuteResponse {
	bz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	ptr := Execute_(utils.BytesToPackedPtr(bz))
	resp := SqlExecuteResponse{}
	err = json.Unmarshal(utils.PackedPtrToBytes(ptr), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func BatchAtomic(req *SqlExecuteBatchRequest) SqlExecuteBatchResponse {
	bz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	ptr := BatchAtomic_(utils.BytesToPackedPtr(bz))
	resp := SqlExecuteBatchResponse{}
	err = json.Unmarshal(utils.PackedPtrToBytes(ptr), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}

func Query(req *SqlQueryRequest) SqlQueryResponse {
	bz, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	ptr := Query_(utils.BytesToPackedPtr(bz))
	resp := SqlQueryResponse{}
	err = json.Unmarshal(utils.PackedPtrToBytes(ptr), &resp)
	if err != nil {
		panic(err)
	}
	return resp
}
